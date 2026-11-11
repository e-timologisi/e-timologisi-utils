package invoiceusers

// DiasAuth belongs to the existing invoice-users document.
// OrgCode is the five digit DIAS organisation code (9nnnn) issued to the tenant through its bank.
// There is no default: an empty OrgCode means RF generation is disabled for the tenant.
type DiasAuth struct {
	OrgCode      string `json:"org_code"`
	AutoGenerate bool   `json:"auto_generate"`
}
