package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"
)

/*
func (a *Agent) RefreshToken() error {
	// https://cadview.qvec.org/newworld.cadview/connect/authorize
	// ?client_id=NewWorld.CadView2
	// &redirect_uri=https%3A%2F%2Fcadview.qvec.org%2FNewWorld.CadView%2Fsilent-refresh.html
	// &response_type=id_token%20token
	// &scope=openid%20cadviewapi.consumer
	// &state=58ed0b44b15e4aa8a6774c52249e6a2b
	// &nonce=8977509cfe84419980e2d2c6b30fae4b
	//&prompt=none

	// TODO: FIXME: IMPLEMENT: XXX
	// If this isn't implemented, everything goes stale in 15m

	return nil
}
*/

func (a *Agent) IsAuthorized() error {
	// https://cadview.qvec.org/NewWorld.CadView/api/CadView/IsAuthorized

	var out bool
	url := a.BaseUrl + "NewWorld.CadView/api/CadView/IsAuthorized"
	body, err := a.authorizedGet(url)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &out)
	return err
}

func (a *Agent) Ping() error {
	// https://cadview.qvec.org/NewWorld.CadView/api/CadView/Ping

	var out bool
	url := a.BaseUrl + "NewWorld.CadView/api/CadView/Ping"
	body, err := a.authorizedGet(url)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &out)
	return err
}

func (a *Agent) GetORIs() ([]ORIObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/CadView/GetOrisForClearedCallSearch

	var out []ORIObj
	url := a.BaseUrl + "NewWorld.CadView/api/CadView/GetOrisForClearedCallSearch"
	body, err := a.authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

func (a *Agent) GetActiveCalls() ([]CallObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetActiveCalls

	var out []CallObj
	url := a.BaseUrl + "NewWorld.CadView/api/Call/GetActiveCalls"
	body, err := a.authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

func (a *Agent) GetClearedCalls(fromDate time.Time, toDate time.Time, ori string) ([]CallObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/SearchClearedCalls?
	// fromDate=5/7/2021,%2012:00:00%20AM
	// &toDate=11/13/2022,%2011:59:59%20PM
	// &ori=28&includeCanceledCalls=true

	v := url.Values{}
	v.Add("fromDate", fromDate.Format(dateSearchFormat))
	v.Add("toDate", toDate.Format(dateSearchFormat))
	v.Add("ori", ori)
	v.Add("includeCanceledCalls", "true")

	var out []CallObj
	url := a.BaseUrl + "NewWorld.CadView/api/Call/SearchClearedCalls?" + v.Encode()
	body, err := a.authorizedGet(url)
	if err != nil {
		return out, err
	}
	if a.Debug {
		log.Printf("DEBUG: %s", string(body))
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

// GetCallDetails appropriately populates a CallObj. Results from
// GetClearedCalls(), etc, do not produce complete CallObj records.
func (a *Agent) GetCallDetails(cobj CallObj) (CallObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCall?id=591039

	v := url.Values{}
	v.Add("id", fmt.Sprintf("%d", cobj.CallID))

	var out CallObj
	url := a.BaseUrl + "NewWorld.CadView/api/Call/GetCall?" + v.Encode()
	body, err := a.authorizedGet(url)
	if err != nil {
		return out, err
	}
	if a.Debug {
		log.Printf("DEBUG: %s", string(body))
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

func (a *Agent) GetCallIncidents(callID string) ([]IncidentObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCallIncidents?id=573613

	var out []IncidentObj
	url := a.BaseUrl + "NewWorld.CadView/api/Call/GetCallIncidents?id=" + callID
	body, err := a.authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	if err == nil {
		numCallID, _ := strconv.Atoi(callID)
		for k := range out {
			out[k].CallID = int64(numCallID)
		}
	}
	return out, err
}

func (a *Agent) GetCallLogs(callID string) ([]CallLogObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCallLogs?id=573613

	var out []CallLogObj
	url := a.BaseUrl + "NewWorld.CadView/api/Call/GetCallLog?id=" + callID
	body, err := a.authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	if err == nil {
		numCallID, _ := strconv.Atoi(callID)
		for k := range out {
			out[k].CallID = int64(numCallID)
		}
	}
	return out, err
}

func (a *Agent) GetCallNarratives(callID string) ([]NarrativeObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCallNarratives?id=573613

	var out []NarrativeObj
	url := a.BaseUrl + "NewWorld.CadView/api/Call/GetCallNarratives?id=" + callID
	body, err := a.authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	if err == nil {
		numCallID, _ := strconv.Atoi(callID)
		for k := range out {
			out[k].CallID = int64(numCallID)
		}
	}
	return out, err
}

func (a *Agent) GetCallUnits(callID string) ([]UnitObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCallUnits?id=573613

	var out []UnitObj
	url := a.BaseUrl + "NewWorld.CadView/api/Call/GetCallUnits?id=" + callID
	body, err := a.authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	if err == nil {
		numCallID, _ := strconv.Atoi(callID)
		for k := range out {
			out[k].CallID = int64(numCallID)
		}
	}
	return out, err
}

func (a *Agent) GetCallUnitLogs(callID string) ([]UnitLogObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCallUnitLogs?id=573613

	var out []UnitLogObj
	url := a.BaseUrl + "NewWorld.CadView/api/Call/GetCallUnitLogs?id=" + callID
	body, err := a.authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	if err == nil {
		numCallID, _ := strconv.Atoi(callID)
		for k := range out {
			out[k].CallID = int64(numCallID)
		}
	}
	return out, err
}
