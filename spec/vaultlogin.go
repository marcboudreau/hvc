package spec

// VaultLogin is a structure that specifies the method to obtain a Vault token.
// The structure contains multiple strategies, but only one should be used.
type VaultLogin struct {
	// Token contains a valid Vault token provided to this application.
	Token string `json:"token"`
	// Kubernetes is a VaultKubernetesLogin object that specifies the details to
	// complete a Vault login operation using the Kubernetes authentication
	// method.
	Kubernetes *VaultKubernetesLogin `json:"kubernetes"`
}
