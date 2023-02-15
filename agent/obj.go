package agent

import "gorm.io/gorm"

type CADCall struct {
	gorm.Model `json:"-"`
	ID         int64          `json:"id" gorm:"primaryKey"`
	Call       CallObj        `json:"call" db:"-" gorm:"foreignKey:CallID"`
	Logs       []CallLogObj   `json:"logs" db:"-" gorm:"foreignKey:CallID"`
	Incidents  []IncidentObj  `json:"incidents" db:"-" gorm:"foreignKey:CallID"`
	Narratives []NarrativeObj `json:"narratives" db:"-" gorm:"foreignKey:CallID"`
	Units      []UnitObj      `json:"units" db:"-" gorm:"foreignKey:CallID"`
	UnitLogs   []UnitLogObj   `json:"unit_logs" db:"-" gorm:"foreignKey:CallID"`
}

type CallObj struct {
	gorm.Model         `json:"-"`
	ArrivedDateTime    string   `json:"arrivedDateTime"`
	CallID             int64    `json:"callId" gorm:"index;size:64"`
	CallNumber         int      `json:"callNumber"`
	CallPriority       string   `json:"callPriority"`
	CallSource         string   `json:"callSource"` // "911"
	CallStatus         string   `json:"callStatus"`
	CallType           string   `json:"callType"`   //"Sick Person"
	CallTypeID         int      `json:"callTypeId"` // 110
	CommonName         string   `json:"commonName"`
	ClosedFlag         bool     `json:"closedFlag"`          // false
	CreatedDateTime    string   `json:"createDateTime"`      // "11/13/2022 10:25:54"
	DispatchedDateTime string   `json:"dispatchedDateTime"`  // "11/13/2022 10:27:43"
	FireCallType       string   `json:"fireCallType"`        // "Sick Person"
	FireCallTypeID     string   `json:"fireCallTypeId"`      // "110"
	IncidentNumber     string   `json:"incidentNumber"`      // "2022-00000345"
	LatitudeY          float64  `json:"latitudeY"`           // 41.9026307589760000
	LongitudeX         float64  `json:"longitudeX"`          // -71.9467412712122000
	Location           string   `json:"location"`            // "120 FREEDLEY RD, Pomfret"
	NatureOfCall       string   `json:"natureOfCall"`        // "/GENERAL WEAKNESS/ UNIVERSAL PRECAUTIONS/ "
	PrimaryUnit        string   `json:"primaryUnit"`         // "STA70"
	Quadrant           string   `json:"quadrant"`            // "POMFRET B"
	AllowedORI         []string `json:"allowedOri" gorm:"-"` // ["04040-561","04090"]
	// "foregroundB":68,"foregroundG":68,"foregroundR":68
	// "district":null,"emsCallType":null,"emsCallTypeId":null
	// policeCallType":null,"policeCallTypeId":null,
	// "station":null,"agencyTypes":"Fire","isPendingPolice":false,"isPendingFire":false,"isPendingEms":false
}

type CallLogObj struct {
	gorm.Model        `json:"-"`
	CallID            int64  `json:"call_id" gorm:"index;size:64"`
	ID                string `json:"id"`                // "19889617"
	LogDateTime       string `json:"logDateTime"`       // "11/13/2022 11:46:26"
	ActionDescription string `json:"actionDescription"` // "Agency Context Added"
	Description       string `json:"description"`       // "Fire Call Type Added. Call Type: <NEW CALL>, Status: In Progress, Priority: 1"
	FirstName         string `json:"firstName"`         // "Justin"
	LastName          string `json:"lastName"`          // "jdeloge"
	Machine           string `json:"machine"`           // "EK-DISPATCH-002"
}

type IncidentObj struct {
	gorm.Model     `json:"-"`
	CallID         int64  `json:"call_id" gorm:"index;size:64"`
	ID             string `json:"id"`             //  "-466119"
	IncidentNumber string `json:"incidentNumber"` // "2022-00000282"
	ORI            string `json:"ori"`            // "FM"
	Department     string `json:"department"`     // "Fire Marshals"
	Abbreviation   string `json:"abbreviation"`   // "FM"
	AgencyType     string `json:"agencyType"`     // "Fire"
}

type NarrativeObj struct {
	gorm.Model    `json:"-"`
	CallID        int64  `json:"call_id"  gorm:"index;size:64"`
	ID            string `json:"id"`            // "1502821"
	Narrative     string `json:"narrative"`     // "fire extinguished."
	EnteredDate   string `json:"enteredDate"`   // "11/13/2022 12:12:20"
	FirstName     string `json:"firstName"`     // "Justin"
	LastName      string `json:"lastName"`      // "jdeloge"
	Machine       string `json:"machine"`       // "EK-DISPATCH-002"
	NarrativeType string `json:"narrativeType"` // "User Entry"
}

type OidcObj struct {
	gorm.Model   `json:"-"`
	IDToken      string `json:"id_token"`
	SessionState string `json:"session_state"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresAt    int64  `json:"expires_at"`
	Profile      struct {
		AuthTime int64  `json:"auth_time"`
		IDP      string `json:"idp"`
		SHash    string `json:"s_hash"`
		Sid      string `json:"sid"`
		Sub      string `json:"sub"`
	} `json:"profile"`
}

type ORIObj struct {
	gorm.Model `json:"-"`
	ORI        string `json:"oriId"`      // "26"
	FDID       string `json:"value"`      // "04040"
	AgencyName string `json:"agencyName"` // "Urban Renawal Technican Team"
}

type UnitObj struct {
	gorm.Model             `json:"-"`
	CallID                 int64  `json:"call_id" gorm:"index;size:64"`
	ID                     string `json:"id"`                     // "3132121"
	ORI                    string `json:"ori"`                    // "FM"
	UnitNumber             string `json:"unitNumber"`             // "FM161"
	DispatchDateTime       string `json:"dispatchDateTime"`       // "11/13/2022 12:19:46"
	EnrouteDateTime        string `json:"enrouteDateTime"`        // "11/13/2022 12:19:46"
	StagedDateTime         string `json:"stagedDateTime"`         // ""
	AtPatientDateTime      string `json:"atPatientDateTime"`      // ""
	ArriveDateTime         string `json:"arriveDateTime"`         // "11/13/2022 12:37:08"
	TransportDateTime      string `json:"transportDateTime"`      // ""
	AtHospitalDateTime     string `json:"atHospitalDateTime"`     // ""
	DepartHospitalDateTime string `json:"departHospitalDateTime"` // ""
	ClearDateTime          string `json:"clearDateTime"`          // ""
}

type UnitLogObj struct {
	gorm.Model  `json:"-"`
	CallID      int64  `json:"call_id" gorm:"index;size:64"`
	ID          string `json:"id"`          // "15131983"
	LogDateTime string `json:"logDateTime"` // "11/13/2022 12:19:46"
	Action      string `json:"action"`      // "Unit Status Change"
	Description string `json:"description"` // "RESPONDING"
	UnitNumber  string `json:"unitNumber"`  // "FM161"
	Status      string `json:"status"`      // "RESPONDING"
	FirstName   string `json:"firstName"`   // "Deanna"
	LastName    string `json:"lastName"`    // "ddf"
	Machine     string `json:"machine"`     // "EK-DISPATCH-001"
}
