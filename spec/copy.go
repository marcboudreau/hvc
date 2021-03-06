package spec

// Copy contains the specification for a single secret in the target Vault
// server including all of the source values used to update this secret.
type Copy struct {
	// MountPoint is the path where the KV secrets engine is mounted in the target
	// Vault server.
	MountPoint string `json:"mount-point"`

	// Path is the path of the copied secret within the KV secrets engine in the
	// target Vault server.
	Path string `json:"path"`

	// Values is a map of secret keys to CopyValue structures, which define the
	// source of the secret value. Only one of Values and Secret can be used for
	// any Copy instance.
	Values map[string]*CopyValue `json:"values"`

	// Secret is a structure which defines an entire source secret to copy (all of
	// its keys). Only one of Values and Secret can be used for any Copy instance.
	Secret *CopyValue `json:"secret"`
}
