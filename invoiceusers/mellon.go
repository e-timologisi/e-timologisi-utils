package invoiceusers

// MellonAuth belongs to the existing invoice-users document.
// APIKey contains the encrypted merchant key; API responses redact it.
type MellonAuth struct {
	APIKey      string `json:"api_key,omitempty"`
	Linked      bool   `json:"linked"`
	Environment string `json:"environment"`
}

type MellonPosTerminal struct {
	Value      string `json:"value"`       // Mellon backend Id for transaction endpoints
	TerminalID string `json:"terminal_id"` // Physical TID for provider signatures
	Label      string `json:"label"`
	Merchant   string `json:"merchant"`
	Acquirer   string `json:"acquirer"`
	Selected   bool   `json:"selected"`
}
