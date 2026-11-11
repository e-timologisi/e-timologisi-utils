package invoiceusers

import (
	"encoding/json"
	"testing"
)

func TestUserDiasFieldsRoundTrip(t *testing.T) {
	var user User
	raw := []byte(`{"email":"alice@example.invalid","dias_auth":{"org_code":"90001","auto_generate":true},"mellon_auth":{"linked":true}}`)
	if err := json.Unmarshal(raw, &user); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	var restored User
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.DiasAuth.OrgCode != "90001" || !restored.DiasAuth.AutoGenerate || !restored.MellonAuth.Linked {
		t.Fatal("user fields were lost")
	}
}

func TestLegacyUserHasNoDiasOrgCode(t *testing.T) {
	var legacy User
	if err := json.Unmarshal([]byte(`{"email":"old@example.invalid"}`), &legacy); err != nil {
		t.Fatal(err)
	}
	if legacy.DiasAuth.OrgCode != "" || legacy.DiasAuth.AutoGenerate {
		t.Fatal("old profile must have RF disabled")
	}
}
