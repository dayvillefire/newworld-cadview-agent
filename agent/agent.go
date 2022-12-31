package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/domstorage"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Agent struct {
	Debug    bool
	LoginUrl string
	Username string
	Password string
	FDID     string

	reqMap  map[string]network.RequestID
	urlMap  map[string]string
	bodyMap map[string][]byte
	attr    map[string]string
	auth    OidcObj

	initialized bool
	cancelled   bool
	wg          sync.WaitGroup
	l           sync.Mutex
}

// Init logs in and initializes the agent
func (a *Agent) Init() error {
	if a.initialized {
		return fmt.Errorf("already initialized")
	}

	// Initialize all maps to avoid NPE
	a.reqMap = map[string]network.RequestID{}
	a.urlMap = map[string]string{}
	a.bodyMap = map[string][]byte{}
	a.attr = map[string]string{}

	var _ctx context.Context
	var _cancel context.CancelFunc
	if a.Debug {
		_ctx, _cancel = chromedp.NewContext(context.Background(), chromedp.WithDebugf(log.Printf))
	} else {
		_ctx, _cancel = chromedp.NewContext(context.Background())
	}
	defer _cancel()

	ctx, cancel := context.WithTimeout(_ctx, 60*time.Second)
	defer cancel()

	// ensure that the browser process is started
	if err := chromedp.Run(ctx); err != nil {
		log.Printf("ERR: %s", err.Error())
		return err
	}

	// Listen to all network events and save content for whatever comes in
	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent:
			if unwantedTraffic(ev.Request.URL) {
				break
			}
			if a.Debug {
				log.Printf("EventRequestWillBeSent: %v: %v", ev.RequestID, ev.Request.URL)
			}
			a.l.Lock()
			a.reqMap[ev.Request.URL] = ev.RequestID
			a.l.Unlock()
		case *network.EventResponseReceived:
			if unwantedTraffic(ev.Response.URL) {
				break
			}
			if a.Debug {
				log.Printf("EventResponseReceived: %v: %v", ev.RequestID, ev.Response.URL)
			}
			a.l.Lock()
			a.urlMap[ev.RequestID.String()] = ev.Response.URL
			a.l.Unlock()
		case *network.EventLoadingFinished:
			if a.Debug {
				log.Printf("EventLoadingFinished: %v", ev.RequestID)
			}
			a.wg.Add(1)
			go func() {
				c := chromedp.FromContext(ctx)
				body, err := network.GetResponseBody(ev.RequestID).Do(cdp.WithExecutor(ctx, c.Target))
				if err != nil {
					defer a.wg.Done()
					return
				}

				a.l.Lock()
				url := a.urlMap[ev.RequestID.String()]
				a.bodyMap[url] = body
				a.l.Unlock()

				if a.Debug {
					log.Printf("%s: %s", url, string(body))
				}

				defer a.wg.Done()
			}()
		}
	})

	// Use a Chrome web browser to log in to the interface and obtain the
	// appropriate authentication token from local storage.

	if err := chromedp.Run(ctx,
		chromedp.Navigate(a.LoginUrl),
		chromedp.Tasks{
			// Login sequence
			chromedp.WaitVisible("//input[@id='Username']"),
			chromedp.SendKeys("//input[@id='Username']", a.Username),
			chromedp.SendKeys("//input[@id='passwordField']", a.Password),
			chromedp.Submit("//button[@id='loginbtn']"),

			// Don't continue until the dashboard is visible
			chromedp.WaitVisible(`//*[contains(., 'Dashboard')]`),

			chromedp.ActionFunc(func(ctx context.Context) error {
				// if the default profile is not loaded,
				// it just gets the entries added by the navigation action in the previous step.
				// it's possible that the js code to add cache entries is executed after this action,
				// and this action gets nothing.
				// in this case, it's better to listen to the DOMStorage events.
				entries, err := domstorage.GetDOMStorageItems(&domstorage.StorageID{
					SecurityOrigin: "https://" + strings.Split(a.LoginUrl, "/")[2],
					IsLocalStorage: true,
				}).Do(ctx)

				if err != nil {
					return err
				}

				//log.Printf("localStorage entries: %#v", entries)
				for _, entry := range entries {
					if strings.HasPrefix(entry[0], "oidc.user:") {
						err = json.Unmarshal([]byte(entry[1]), &(a.auth))
						//log.Printf("JSON user obj : %s", entry[1])
					}
				}

				return err
			}),
		},
	); err != nil {
		log.Printf("ERR: Failed to login: %s", err.Error())
		return err
	}

	if a.Debug {
		log.Printf("DEBUG: Wait for all data to be received.")
	}
	a.wg.Wait()

	if a.Debug {
		log.Printf("attr : %#v", a.attr)
		log.Printf("urlMap : %#v", a.urlMap)
		log.Printf("/api/CadView/GetAllUserSettings : %s", string(a.bodyMap[a.LoginUrl+"/api/CadView/GetAllUserSettings"]))
	}

	if a.Debug {
		log.Printf("auth : %#v", a.auth)
	}

	a.initialized = true

	return nil
}

func (a *Agent) Run() {
	go func() {
		for {
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
		err := a.Init()
		if err != nil {
			return []CallObj{}, err
		}
	}
	return a.GetClearedCalls(from, to, ori)
}

func (a *Agent) RetrieveCADCall(call CallObj) (CADCall, error) {
	out := CADCall{}
	if call.CallID == 0 {
		return out, fmt.Errorf("no call presented")
	}
	out.Call = call

	callId := fmt.Sprintf("%d", call.CallID)
	var err error

	out.Incidents, err = a.GetCallIncidents(callId)
	if err == nil {
		if a.Debug {
			log.Printf(" --> Incidents : %#v", out.Incidents)
		}
	}

	out.Units, err = a.GetCallUnits(callId)
	if err == nil {
		if a.Debug {
			log.Printf(" --> Units : %#v", out.Units)
		}
	}

	out.UnitLogs, err = a.GetCallUnitLogs(callId)
	if err == nil {
		if a.Debug {
			log.Printf(" --> Unit Logs : %#v", out.UnitLogs)
		}
	}

	out.Narratives, err = a.GetCallNarratives(callId)
	if err == nil {
		if a.Debug {
			log.Printf(" --> Narratives : %#v", out.Narratives)
		}
	}

	out.Logs, err = a.GetCallLogs(callId)
	if err == nil {
		if a.Debug {
			log.Printf(" --> Logs : %#v", out.Logs)
		}
	}

	return out, err
}

// authorizedGet uses the current authentication mechanism to
func (a *Agent) authorizedGet(url string) ([]byte, error) {
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

	//log.Printf("headers : %#v", req.Header)

	res, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if res.Body != nil {
		defer res.Body.Close()
	}
	defer res.Body.Close()
	return body, err
}

func (a *Agent) SetAuth(auth OidcObj) {
	a.auth = auth
}

func (a *Agent) GetAuth() OidcObj {
	return a.auth
}
