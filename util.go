package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	dateSearchFormat = "1/2/2006,03:04:05 PM"
	dateFormat       = "1/2/2006 15:04:05"
)

// fdidToORI converts an FDID to a CAD system internal ORI used for searching.
// It requires an ORIObj array.
func fdidToORI(orimap []ORIObj, fdid string) string {
	for _, ori := range orimap {
		if ori.FDID == fdid {
			return ori.ORI
		}
	}
	return ""
}

func parseDate(dt string) time.Time {
	t, err := time.Parse(dateFormat, dt)
	if err != nil {
		log.Printf("parseDate: %s could not be parsed, using now()", dt)
		return time.Now()
	}
	return t
}

// authorizedGet uses the current authentication mechanism to
func authorizedGet(url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Authorization", auth.TokenType+" "+auth.AccessToken)
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

// unwantedTraffic determines if a URL should be stored in memory or not
func unwantedTraffic(url string) bool {
	return !strings.HasPrefix(url, "http") ||
		strings.HasSuffix(url, ".css") ||
		strings.HasSuffix(url, ".js") ||
		strings.HasSuffix(url, ".svg") ||
		strings.HasSuffix(url, ".woff2")
}
