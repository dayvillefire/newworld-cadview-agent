package agent

import "testing"

func Test_Agent_Refresh(t *testing.T) {
	a := Agent{
		Username: DEFAULT_USERNAME,
		Password: DEFAULT_PASSWORD,
		LoginUrl: DEFAULT_URL,
		FDID:     DEFAULT_FDID,
	}
	err := a.Init()
	if err != nil {
		t.Fatalf("ERR: %s", err.Error())
	}

}

/*
func Test_Agent(t *testing.T) {
	a := Agent{
		Username: DEFAULT_USERNAME,
		Password: DEFAULT_PASSWORD,
		LoginUrl: DEFAULT_URL,
		FDID:     DEFAULT_FDID,
	}
	err := a.Init()
	if err != nil {
		t.Fatalf("ERR: %s", err.Error())
	}
	oris, err := a.GetORIs()
	if err != nil {
		t.Fatalf("ERR: %s", err.Error())
	}

	calls, err := a.GetClearedCalls(
		parseDate("10/13/2022 00:00:00"),
		parseDate("10/14/2022 23:59:59"),
		FDIDToORI(oris, a.FDID),
	)
	if err != nil {
		t.Fatalf("ERR: %s", err.Error())
	}
	t.Logf("Calls : %#v", calls)

	for _, c := range calls {
		o, err := a.RetrieveCADCall(c)
		if err != nil {
			t.Logf("ERR: %s", err.Error())
			continue
		}
		t.Logf("Call[%s] : %#v", c.IncidentNumber, o)
	}
}
*/
