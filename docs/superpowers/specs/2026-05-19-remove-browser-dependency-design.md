# Remove Browser Dependency from CadView Agent

**Date:** 2026-05-19
**Status:** Design approved
**Goal:** Replace chromedp/Chrome browser login with pure HTTP OIDC authorization code + PKCE flow, eliminating the browser dependency entirely.

## Motivation

The `Agent.Init()` method currently launches a Chrome browser (embedded or remote via CDP) to automate login through the web form and extract OIDC tokens from `localStorage`. This adds ~50MB+ of Chrome dependencies, slows startup, and introduces fragility from browser automation. All subsequent API calls use pure HTTP with bearer tokens — the browser is only needed for the initial login.

## Architecture

### What changes
- `agent.go` `Init()`: replaces chromedp browser automation with a multi-step HTTP OIDC authorization code + PKCE flow
- New method `authenticate()`: handles the HTTP login/authorization/token-exchange sequence
- New method `refreshToken()`: implements the currently-stubbed token refresh using `grant_type=refresh_token`
- `OidcObj`: adds `RefreshToken string` field
- `Agent` struct: drops `CDP` field

### What stays the same
- `authorizedGet()`: still sends `Authorization: Bearer <access_token>`
- All API methods (`GetActiveCalls`, `GetCallDetails`, `GetCallIncidents`, `GetCallUnits`, `GetCallUnitLogs`, `GetCallNarratives`, `GetCallLogs`, `GetORIs`, `Ping`, `IsAuthorized`): unchanged
- `Run()` loop: pings to keep session alive, unchanged
- `OidcObj` struct (with addition of `RefreshToken`)

### Dependencies removed
- `github.com/chromedp/chromedp`
- `github.com/chromedp/cdproto/*`

### Dependencies added
- `crypto/sha256` + `crypto/rand` (stdlib) for PKCE
- `net/http/cookiejar` (stdlib) for cookie management
- `github.com/joho/godotenv` (optional, for `.env` loading)

## New Login Flow: `authenticate()`

### Step 1: GET login page, extract CSRF token
```
GET {BaseUrl}newworld.cadview/account/login
→ Parse HTML for <input name="__RequestVerificationToken">
→ Store cookies via cookie jar
```

### Step 2: POST credentials
```
POST {BaseUrl}newworld.cadview/account/login
  Content-Type: application/x-www-form-urlencoded
  Body: Username={user}&Password={pass}&__RequestVerificationToken={token}
→ Do not auto-follow redirect; capture Location header
→ Expected: 302 to /connect/authorize?...
```

### Step 3: Generate PKCE parameters
```
code_verifier = base64url(random 43-128 bytes)
code_challenge = base64url(sha256(code_verifier))
state = random 32-char hex
```

### Step 4: GET authorize endpoint
```
GET {BaseUrl}newworld.cadview/connect/authorize?
    client_id=NewWorld.CadView2&
    redirect_uri={BaseUrl}NewWorld.CadView/silent-refresh.html&
    response_type=code&
    scope=openid cadviewapi.consumer&
    code_challenge={code_challenge}&
    code_challenge_method=S256&
    state={state}&
    nonce={random}
→ Session cookie from step 2 identifies the authenticated session
→ Expected: 302 to silent-refresh.html?code={code}&state={state}
→ Extract code from query string; verify state matches
```

### Step 5: Exchange code for tokens
```
POST {BaseUrl}newworld.cadview/connect/token
  Content-Type: application/x-www-form-urlencoded
  Body: grant_type=authorization_code&
        code={code}&
        redirect_uri={BaseUrl}NewWorld.CadView/silent-refresh.html&
        code_verifier={code_verifier}&
        client_id=NewWorld.CadView2
→ Response: { access_token, id_token, refresh_token, expires_in, token_type, scope }
→ Store into a.auth
```

## Token Refresh: `refreshToken()`

```
POST {BaseUrl}newworld.cadview/connect/token
  grant_type=refresh_token&
  refresh_token={stored_refresh_token}&
  client_id=NewWorld.CadView2
→ Response: new access_token, id_token, expires_in, optionally new refresh_token
→ Update a.auth with new values
```

Proactive refresh: if `expires_at - now < 300s` (5 min), refresh before the next API call.

## Error Handling

| Failure point | Behavior |
|---|---|
| Login page unreachable | Return error immediately |
| CSRF token not found in page | Return error with detail |
| Login POST rejected (bad credentials) | Return `ErrNotAuthorized` |
| Authorize redirect lacks `code` param | Return error (IdP may not support code flow) |
| Token exchange fails | Return error with response details |
| Refresh token rejected by IdP | Fall back to full `authenticate()` re-login |

## Configuration

Credentials loaded from environment variables, with hardcoded defaults as fallback:

| Variable | Default |
|---|---|
| `CADVIEW_BASE_URL` | `https://cadview.qvec.org/` |
| `CADVIEW_USERNAME` | `WA` |
| `CADVIEW_PASSWORD` | `wadispatch` |
| `CADVIEW_FDID` | `04042` |

A `.env` file (gitignored) at the project root provides credentials for testing. `os.Getenv` handles lookup; `godotenv` auto-loads the `.env` file.

## Testing Strategy

1. **Unit tests**: mock `http.RoundTripper` to verify PKCE generation, redirect following, CSRF parsing, token extraction. No real credentials needed.
2. **Integration test**: loads credentials from `.env`, runs the full `authenticate()` against a real CadView instance, verifies an API call succeeds.
3. **Existing tests** (`Test_Agent_Refresh`, `Test_Agent_API`): updated to work with the new flow.

## Files Modified

- `agent/agent.go`: replace `Init()` browser code with HTTP flow; add `authenticate()` and `refreshToken()`; remove `CDP` field
- `agent/api.go`: implement `refreshToken()` replacing the stub
- `agent/obj.go`: add `RefreshToken` field to `OidcObj`
- `agent/const.go`: keep defaults, add env var key constants
- `agent/util.go`: add PKCE helpers (`generateCodeVerifier`, `generateCodeChallenge`, `generateState`, `generateNonce`)
- `agent/agent_test.go`: update tests
- `go.mod` / `go.sum`: remove chromedp deps, add godotenv (optional)

## Sequence Diagram

```
Agent.Init()
  → authenticate()
    → GET login page → extract CSRF token
    → POST login credentials → capture auth cookie
    → GET /connect/authorize?response_type=code&... → capture code from redirect
    → POST /connect/token?grant_type=authorization_code → store tokens
  → refreshToken() (if needed)
  → return nil

Later API calls:
  → authorizedGet() → uses bearer token as before
  → if token near expiry → refreshToken() → POST /connect/token?grant_type=refresh_token
  → if refresh fails → authenticate() again
```
