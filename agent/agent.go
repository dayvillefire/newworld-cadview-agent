package agent

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	ErrNotAuthorized = errors.New("not authorized")
)

// Agent is the NewWorld cadview access client. It needs to have Init()
// successfully called on it before it can perform any actions.
type Agent struct {
	// Debug turns debug logging on. Be very sure you want to do this, as it
	// is very verbose.
	Debug bool
	// BaseUrl specifies the URL of the cadview instance with a trailing
	// slash, like "https://cadview.somepsap.org/".
	BaseUrl string
	// Username is the login user credential for the cadview instance.
	Username string
	// Password is the login password credential for the cadview instance.
	Password string
	// FDID is the ORI/FDID associated with the login credentials.
	FDID string
	reqMap  map[string]string
	urlMap  map[string]string
	bodyMap map[string][]byte
	attr    map[string]string
	auth    OidcObj

	initialized bool
	cancelled   bool
	wg          *sync.WaitGroup
	l           sync.Mutex
}

// Init logs in and initializes the agent via HTTP OIDC flow (no browser).
func (a *Agent) Init() error {
	if a.initialized {
		return fmt.Errorf("already initialized")
	}

	// Apply environment variable overrides.
	a.LoadConfigFromEnv()

	// Initialize all maps to avoid NPE
	a.reqMap = map[string]string{}
	a.urlMap = map[string]string{}
	a.bodyMap = map[string][]byte{}
	a.attr = map[string]string{}
	if a.wg == nil {
		a.wg = &sync.WaitGroup{}
	}

	if err := a.authenticate(); err != nil {
		log.Printf("ERR: authenticate: %s", err.Error())
		return err
	}

	if a.Debug {
		log.Printf("DEBUG: auth : %#v", a.auth)
	}

	a.initialized = true
	return nil
}

func (a *Agent) Run() {
	// Keepalive ping every 15 seconds.
	go func() {
		for {
			if a.Debug {
				log.Printf("Run(): Ping()")
			}
			err := a.Ping()
			if err != nil {
				log.Printf("Run(): %s", err.Error())
			}
			for i := 0; i < 15; i++ {
				time.Sleep(time.Second)
				if a.cancelled {
					return
				}
			}
		}
	}()

	// Proactive token refresh every 8 minutes.
	// ensureValidToken handles its own locking.
	go func() {
		for {
			for i := 0; i < 120; i++ {
				time.Sleep(time.Second)
				if a.cancelled {
					return
				}
			}
			if a.Debug {
				log.Printf("Run(): proactive token refresh")
			}
			if err := a.ensureValidToken(); err != nil {
				log.Printf("Run(): token refresh failed: %s", err.Error())
			}
		}
	}()
}

func (a *Agent) ActiveCalls() ([]CallObj, error) {
	if !a.initialized {
		err := a.Init()
		if err != nil {
			return []CallObj{}, err
		}
	}
	return a.GetActiveCalls()
}

func (a *Agent) ClearedCalls(from, to time.Time, ori string) ([]CallObj, error) {
	if !a.initialized {
		if a.Debug {
			log.Printf("ClearedCalls: !initialized")
		}
		err := a.Init()
		if err != nil {
			return []CallObj{}, err
		}
	}
	return a.GetClearedCalls(from, to, ori)
}

func (a *Agent) RetrieveCADCall(call CallObj) (CADCall, error) {
	var err error
	out := CADCall{}
	if call.CallID == 0 {
		return out, fmt.Errorf("no call presented")
	}

	out.Call, err = a.GetCallDetails(call)
	if err != nil {
		return out, err
	}

	callId := fmt.Sprintf("%d", call.CallID)

	out.Incidents, err = a.GetCallIncidents(callId)
	if err == nil {
		if a.Debug {
			log.Printf(" --> Incidents : %#v", out.Incidents)
		}
	} else {
		log.Printf("ERR: GetCallIncidents: %s", err.Error())
	}

	{
		units, err := a.GetCallUnits(callId)
		if err == nil {
			if a.Debug {
				log.Printf(" --> Units : %#v", units)
			}
		} else {
			log.Printf("ERR: GetCallUnits: %s", err.Error())
		}
		out.Units = append(out.Units, units...)
	}

	{
		unitlogs, err := a.GetCallUnitLogs(callId)
		if err == nil {
			if a.Debug {
				log.Printf(" --> Unit Logs : %#v", unitlogs)
			}
		} else {
			log.Printf("ERR: GetCallUnitLogs: %s", err.Error())
		}
		out.UnitLogs = append(out.UnitLogs, unitlogs...)
	}

	{
		narratives, err := a.GetCallNarratives(callId)
		if err == nil {
			if a.Debug {
				log.Printf(" --> Narratives : %#v", narratives)
			}
		} else {
			log.Printf("ERR: GetCallNarratives: %s", err.Error())
		}
		out.Narratives = append(out.Narratives, narratives...)
	}

	{
		logs, err := a.GetCallLogs(callId)
		if err == nil {
			if a.Debug {
				log.Printf(" --> Logs : %#v", logs)
			}
		} else {
			log.Printf("ERR: GetCallLogs: %s", err.Error())
		}
		out.Logs = append(out.Logs, logs...)
	}

	return out, err
}

// authorizedGet uses the current authentication mechanism to
func (a *Agent) authorizedGet(url string) ([]byte, error) {
	if err := a.ensureValidToken(); err != nil {
		return []byte{}, fmt.Errorf("ensureValidToken: %w", err)
	}

	if a.auth.TokenType == "" {
		return []byte{}, fmt.Errorf("not authenticated")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Authorization", a.auth.TokenType+" "+a.auth.AccessToken)
	if req.Body != nil {
		defer req.Body.Close()
	}

	if a.Debug {
		log.Printf("DEBUG: authorizedGet: Headers : %#v", req.Header)
	}

	res, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	body, err := io.ReadAll(res.Body)
	if res.Body != nil {
		defer res.Body.Close()
	}
	defer res.Body.Close()

	// Check for not being authorized
	if err == nil {
		if len(body) < 1 || body[0] == '<' {
			err = ErrNotAuthorized
		}
	}

	return body, err
}

func (a *Agent) SetAuth(auth OidcObj) {
	if a.Debug {
		log.Printf("SetAuth: %#v", auth)
	}
	a.auth = auth
}

func (a *Agent) GetAuth() OidcObj {
	return a.auth
}

func (a *Agent) MakeCopy() *Agent {
	return &Agent{
		Debug:    a.Debug,
		BaseUrl:  a.BaseUrl,
		Username: a.Username,
		Password: a.Password,
		FDID:     a.FDID,
		wg:       a.wg,
	}
}

func (a *Agent) WaitGroup() *sync.WaitGroup {
	return a.wg
}

func (a *Agent) Cancel() {
	a.cancelled = true
}

func (a *Agent) TransferAuthFrom(a2 *Agent) {
	if a.Debug {
		log.Printf("TransferAuthFrom: %s (old) -> %s (new)", a.auth.AccessToken, a2.auth.AccessToken)
	}
	a.auth = a2.auth
}
