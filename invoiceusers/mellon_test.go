package invoiceusers

import (
	"encoding/json"
	"testing"
)

func TestUserMellonFieldsRoundTrip(t *testing.T) {
	var user User
	raw := []byte(`{"email":"alice@example.invalid","mellon_auth":{"api_key":"encrypted-key","linked":true,"environment":"uat"},"mellon_pos_terminals":[{"value":"362450","terminal_id":"054889019","label":"Desk","selected":true}],"epay_auth":{"apiKey":"existing-epay"}}`)
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
	if restored.MellonAuth.APIKey != "encrypted-key" || !restored.MellonAuth.Linked || len(restored.MellonPosTerminals) != 1 || restored.MellonPosTerminals[0].TerminalID != "054889019" || restored.EPayAuth.ApiKey != "existing-epay" {
		t.Fatal("user fields were lost")
	}
	var legacy User
	if err := json.Unmarshal([]byte(`{"email":"old@example.invalid"}`), &legacy); err != nil {
		t.Fatal(err)
	}
	if legacy.MellonAuth.Linked || legacy.MellonAuth.APIKey != "" {
		t.Fatal("old profile should remain unlinked")
	}
}
