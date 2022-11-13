package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/domstorage"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

var (
	debug    = flag.Bool("debug", false, "Debugging enabled")
	loginUrl = flag.String("url", "", "Login URL (no trailing slash)")
	username = flag.String("username", "", "Username")
	password = flag.String("password", "", "Password")
	fdid     = flag.String("fdid", "", "FDID")

	reqMap  = map[string]network.RequestID{}
	urlMap  = map[string]string{}
	bodyMap = map[string][]byte{}
	attr    = map[string]string{}
	auth    = OidcObj{}

	wg sync.WaitGroup
	l  sync.Mutex
)

func main() {
	flag.Parse()

	// Defaults
	if *loginUrl == "" {
		*loginUrl = DEFAULT_URL
	}
	if *username == "" {
		*username = DEFAULT_USERNAME
	}
	if *password == "" {
		*password = DEFAULT_PASSWORD
	}

	var _ctx context.Context
	var _cancel context.CancelFunc
	if *debug {
		_ctx, _cancel = chromedp.NewContext(context.Background(), chromedp.WithDebugf(log.Printf))
	} else {
		_ctx, _cancel = chromedp.NewContext(context.Background())
	}
	defer _cancel()

	ctx, cancel := context.WithTimeout(_ctx, 60*time.Second)
	defer cancel()

	// ensure that the browser process is started
	if err := chromedp.Run(ctx); err != nil {
		panic(err)
	}

	// Listen to all network events and save content for whatever comes in
	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent:
			if unwantedTraffic(ev.Request.URL) {
				break
			}
			if *debug {
				log.Printf("EventRequestWillBeSent: %v: %v", ev.RequestID, ev.Request.URL)
			}
			l.Lock()
			reqMap[ev.Request.URL] = ev.RequestID
			l.Unlock()
		case *network.EventResponseReceived:
			if unwantedTraffic(ev.Response.URL) {
				break
			}
			if *debug {
				log.Printf("EventResponseReceived: %v: %v", ev.RequestID, ev.Response.URL)
			}
			l.Lock()
			urlMap[ev.RequestID.String()] = ev.Response.URL
			l.Unlock()
		case *network.EventLoadingFinished:
			if *debug {
				log.Printf("EventLoadingFinished: %v", ev.RequestID)
			}
			wg.Add(1)
			go func() {
				c := chromedp.FromContext(ctx)
				body, err := network.GetResponseBody(ev.RequestID).Do(cdp.WithExecutor(ctx, c.Target))
				if err != nil {
					defer wg.Done()
					return
				}

				l.Lock()
				url := urlMap[ev.RequestID.String()]
				bodyMap[url] = body
				l.Unlock()

				if *debug {
					log.Printf("%s: %s", url, string(body))
				}

				defer wg.Done()
			}()
		}
	})

	// Use a Chrome web browser to log in to the interface and obtain the
	// appropriate authentication token from local storage.

	if err := chromedp.Run(ctx,
		chromedp.Navigate(*loginUrl),
		chromedp.Tasks{
			// Login sequence
			chromedp.WaitVisible("//input[@id='Username']"),
			chromedp.SendKeys("//input[@id='Username']", *username),
			chromedp.SendKeys("//input[@id='passwordField']", *password),
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
					SecurityOrigin: "https://" + strings.Split(*loginUrl, "/")[2],
					IsLocalStorage: true,
				}).Do(ctx)

				if err != nil {
					return err
				}

				//log.Printf("localStorage entries: %#v", entries)
				for _, entry := range entries {
					if strings.HasPrefix(entry[0], "oidc.user:") {
						err = json.Unmarshal([]byte(entry[1]), &auth)
						//log.Printf("JSON user obj : %s", entry[1])
					}
				}

				return err
			}),
		},
	); err != nil {
		log.Fatal("ERR: Failed to login:", err)
	}

	if *debug {
		log.Printf("DEBUG: Wait for all data to be received.")
	}
	wg.Wait()

	if *debug {
		log.Printf("attr : %#v", attr)
		log.Printf("urlMap : %#v", urlMap)
		log.Printf("/api/CadView/GetAllUserSettings : %s", string(bodyMap[*loginUrl+"/api/CadView/GetAllUserSettings"]))
	}

	if *debug {
		log.Printf("auth : %#v", auth)
	}

	// active calls -----------------------------------------------------------------------------------------------------

	calls, err := GetActiveCalls()
	if err != nil {
		panic(err)
	}
	for _, call := range calls {
		log.Printf("CALL : %#v", call)

		callId := fmt.Sprintf("%d", call.CallID)

		incidents, err := GetCallIncidents(callId)
		if err == nil {
			log.Printf(" --> Incidents : %#v", incidents)
		}

		units, err := GetCallUnits(callId)
		if err == nil {
			log.Printf(" --> Units : %#v", units)
		}

		unitLogs, err := GetCallUnitLogs(callId)
		if err == nil {
			log.Printf(" --> Unit Logs : %#v", unitLogs)
		}

		narratives, err := GetCallNarratives(callId)
		if err == nil {
			log.Printf(" --> Narratives : %#v", narratives)
		}

		logs, err := GetCallLogs(callId)
		if err == nil {
			log.Printf(" --> Logs : %#v", logs)
		}

	}

	// search calls -----------------------------------------------------------------------------------------------------

	/*
		orimap, err := GetORIs()
		if err != nil {
			panic(err)
		}
		ori := fdidToORI("04042")
	*/
}
