package spec

// Vault is a structure that specifies the connection information for a Vault
// server and includes a VaultLogin structure to provide a login strategy.
type Vault struct {
	// Address contains the scheme, host, and port address of the Vault server.
	Address string `json:"address"`

	// Login is a VaultLogin object that provides the details on how to obtain
	// a valid Vault token.
	Login *VaultLogin `json:"login"`
}
