package spec

// VaultKubernetesLogin is a structure that specifies the details needed to
// complete a Vault login operation using the Kubernetes authentication method.
type VaultKubernetesLogin struct {
	// MountPoint contains the path where the Kubernetes authentication method to
	// use is mounted.
	MountPoint string `json:"mount-point"`

	// Role contains the name of the backend role in the Kubernetes authentication
	// method.
	Role string `json:"role"`

	// JWTPath contains the local file-system path to use to load the Kubernetes
	// Service Account key file.
	JWTPath string `json:"jwt-path"`
}
