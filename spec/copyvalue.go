package spec

// CopyValue defines the source for a value within the target secret.
type CopyValue struct {
	// Source is the name of the defined Vault structure in the Sources field of
	// the CopyJob structure.
	Source string `json:"source"`

	// MountPoint is the path where the KV secrets engine is mounted in the source
	// Vault server.
	MountPoint string `json:"mount-point"`

	// Path is the path of the secret being copied within the KV secrets engine in
	// the source Vault server.
	Path string `json:"path"`

	// Key specifies which value within the secret being copied to copy to the
	// target Vault server.
	Key string `json:"key"`
}
