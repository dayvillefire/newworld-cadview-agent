package agent

import (
	"log"
	"strings"
	"time"
)

const (
	dateSearchFormat = "1/2/2006,03:04:05 PM"
	dateFormat       = "1/2/2006 15:04:05"
)

// FDIDToORI converts an FDID to a CAD system internal ORI used for searching.
// It requires an ORIObj array.
func FDIDToORI(orimap []ORIObj, fdid string) string {
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

// unwantedTraffic determines if a URL should be stored in memory or not
func unwantedTraffic(url string) bool {
	return !strings.HasPrefix(url, "http") ||
		strings.HasSuffix(url, ".css") ||
		strings.HasSuffix(url, ".js") ||
		strings.HasSuffix(url, ".svg") ||
		strings.HasSuffix(url, ".woff2")
}
