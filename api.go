package main

import (
	"encoding/json"
	"net/url"
	"time"
)

func GetORIs() ([]ORIObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/CadView/GetOrisForClearedCallSearch

	var out []ORIObj
	url := *loginUrl + "/api/CadView/GetOrisForClearedCallSearch"
	body, err := authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

func GetActiveCalls() ([]CallObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetActiveCalls

	var out []CallObj
	url := *loginUrl + "/api/Call/GetActiveCalls"
	body, err := authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

func GetClearedCalls(fromDate time.Time, toDate time.Time, ori ORIObj) ([]CallObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/SearchClearedCalls?
	// fromDate=5/7/2021,%2012:00:00%20AM
	// &toDate=11/13/2022,%2011:59:59%20PM
	// &ori=28&includeCanceledCalls=true

	v := url.Values{}
	v.Add("fromDate", fromDate.Format(dateSearchFormat))
	v.Add("toDate", toDate.Format(dateSearchFormat))
	v.Add("ori", ori.ORI)
	v.Add("includeCanceledCalls", "true")

	var out []CallObj
	url := *loginUrl + "/api/CadView/GetOrisForClearedCallSearch?" + v.Encode()
	body, err := authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

func GetCallIncidents(callID string) ([]IncidentObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCallIncidents?id=573613

	var out []IncidentObj
	url := *loginUrl + "/api/Call/GetCallIncidents?id=" + callID
	body, err := authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

func GetCallLogs(callID string) ([]CallLogObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCallLogs?id=573613

	var out []CallLogObj
	url := *loginUrl + "/api/Call/GetCallLogs?id=" + callID
	body, err := authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

func GetCallNarratives(callID string) ([]NarrativeObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCallNarratives?id=573613

	var out []NarrativeObj
	url := *loginUrl + "/api/Call/GetCallNarratives?id=" + callID
	body, err := authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

func GetCallUnits(callID string) ([]UnitObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCallUnits?id=573613

	var out []UnitObj
	url := *loginUrl + "/api/Call/GetCallUnits?id=" + callID
	body, err := authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}

func GetCallUnitLogs(callID string) ([]UnitLogObj, error) {
	// https://cadview.qvec.org/NewWorld.CadView/api/Call/GetCallUnitLogs?id=573613

	var out []UnitLogObj
	url := *loginUrl + "/api/Call/GetCallUnitLogs?id=" + callID
	body, err := authorizedGet(url)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(body, &out)
	return out, err
}
