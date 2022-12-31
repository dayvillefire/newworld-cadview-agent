package agent

import (
	"encoding/json"
	"log"
	"net/url"
	"time"
)

func (a *Agent) Ping() error {
	// https://cadview.qvec.org/NewWorld.CadView/api/CadView/Ping

	var out bool
	url := a.LoginUrl + "/api/CadView/Ping"
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
	url := a.LoginUrl + "/api/CadView/GetOrisForClearedCallSearch"
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
	url := a.LoginUrl + "/api/Call/GetActiveCalls"
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
	url := a.LoginUrl + "/api/Call/SearchClearedCalls?" + v.Encode()
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
	url := a.LoginUrl + "/api/Call/GetCallIncidents?id=" + callID
	body, err := a.authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

func (a *Agent) GetCallLogs(callID string) ([]CallLogObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCallLogs?id=573613

	var out []CallLogObj
	url := a.LoginUrl + "/api/Call/GetCallLog?id=" + callID
	body, err := a.authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

func (a *Agent) GetCallNarratives(callID string) ([]NarrativeObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCallNarratives?id=573613

	var out []NarrativeObj
	url := a.LoginUrl + "/api/Call/GetCallNarratives?id=" + callID
	body, err := a.authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

func (a *Agent) GetCallUnits(callID string) ([]UnitObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCallUnits?id=573613

	var out []UnitObj
	url := a.LoginUrl + "/api/Call/GetCallUnits?id=" + callID
	body, err := a.authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

func (a *Agent) GetCallUnitLogs(callID string) ([]UnitLogObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCallUnitLogs?id=573613

	var out []UnitLogObj
	url := a.LoginUrl + "/api/Call/GetCallUnitLogs?id=" + callID
	body, err := a.authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}
