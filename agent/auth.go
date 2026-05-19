package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

// authenticate performs the full OIDC authorization code + PKCE flow over
// pure HTTP — no browser required. On success it populates a.auth.
func (a *Agent) authenticate() error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("cookiejar: %w", err)
	}

	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Step 1: GET login page, extract CSRF token
	loginURL := a.BaseUrl + "newworld.cadview/account/login"
	resp, err := client.Get(loginURL)
	if err != nil {
		return fmt.Errorf("GET login page: %w", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	token := extractCSRF(string(body))
	if token == "" {
		if a.Debug {
			log.Printf("DEBUG: login page body: %s", string(body))
		}
		return fmt.Errorf("could not find CSRF token on login page")
	}
	if a.Debug {
		log.Printf("DEBUG: CSRF token = %s", token)
	}

	// Step 2: POST credentials
	form := url.Values{}
	form.Set("Username", a.Username)
	form.Set("Password", a.Password)
	form.Set("__RequestVerificationToken", token)

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build login POST: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("POST login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login POST returned %d", resp.StatusCode)
	}

	// Step 3: Generate PKCE parameters
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return fmt.Errorf("generate verifier: %w", err)
	}
	codeChallenge := generateCodeChallenge(codeVerifier)
	state, err := generateState()
	if err != nil {
		return fmt.Errorf("generate state: %w", err)
	}
	nonce, err := generateNonce()
	if err != nil {
		return fmt.Errorf("generate nonce: %w", err)
	}

	// Step 4: GET authorize endpoint
	redirURI := a.BaseUrl + "NewWorld.CadView/silent-refresh.html"
	authURL := fmt.Sprintf(
		"%snewworld.cadview/connect/authorize?"+
			"client_id=NewWorld.CadView2"+
			"&redirect_uri=%s"+
			"&response_type=code"+
			"&scope=openid%%20cadviewapi.consumer"+
			"&code_challenge=%s"+
			"&code_challenge_method=S256"+
			"&state=%s"+
			"&nonce=%s",
		a.BaseUrl, url.QueryEscape(redirURI),
		codeChallenge, state, nonce,
	)

	resp, err = client.Get(authURL)
	if err != nil {
		return fmt.Errorf("GET authorize: %w", err)
	}
	defer resp.Body.Close()

	loc, err := resp.Location()
	if err != nil {
		return fmt.Errorf("authorize redirect: %w", err)
	}

	code := loc.Query().Get("code")
	returnedState := loc.Query().Get("state")

	if code == "" {
		return fmt.Errorf("no authorization code in redirect: %s", loc.String())
	}
	if returnedState != state {
		return fmt.Errorf("state mismatch: expected %s, got %s", state, returnedState)
	}

	if a.Debug {
		log.Printf("DEBUG: authorization code received")
	}

	// Step 5: Exchange code for tokens
	tokenForm := url.Values{}
	tokenForm.Set("grant_type", "authorization_code")
	tokenForm.Set("code", code)
	tokenForm.Set("redirect_uri", redirURI)
	tokenForm.Set("code_verifier", codeVerifier)
	tokenForm.Set("client_id", "NewWorld.CadView2")

	tokenURL := a.BaseUrl + "newworld.cadview/connect/token"
	req, err = http.NewRequest("POST", tokenURL, strings.NewReader(tokenForm.Encode()))
	if err != nil {
		return fmt.Errorf("build token POST: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenClient := &http.Client{Jar: jar}
	resp, err = tokenClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST token: %w", err)
	}
	defer resp.Body.Close()

	tokenBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(tokenBody))
	}

	var auth OidcObj
	if err := json.Unmarshal(tokenBody, &auth); err != nil {
		return fmt.Errorf("unmarshal token response: %w", err)
	}

	if auth.AccessToken == "" {
		return fmt.Errorf("token response missing access_token: %s", string(tokenBody))
	}

	a.auth = auth

	if a.Debug {
		log.Printf("DEBUG: auth successful, expires_at=%d, has_refresh_token=%v",
			a.auth.ExpiresAt, a.auth.RefreshToken != "")
	}

	return nil
}

// extractCSRF searches HTML for a __RequestVerificationToken input value.
func extractCSRF(html string) string {
	re := regexp.MustCompile(
		`<input[^>]*name=["']__RequestVerificationToken["'][^>]*value=["']([^"']+)["']`,
	)
	m := re.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}
