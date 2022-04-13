package hvc

import "fmt"

// CopyValue is a structure that is used to copy a specific value from a source
// secret into a target secret.
type CopyValue struct {
	// Source is the connection to the source Vault server.
	Source Vault

	// MountPoint is the path where the KV secrets engine is mounted in the source
	// Vault server.
	MountPoint string

	// Path is the path of the source secret within the KV secrets engine from
	// which the value is retrieved.
	Path string

	// Key is the key of the value in the source secret that should be copied to
	// the target secret.
	Key string
}

// Name returns a canonical name for the receiver.
func (p *CopyValue) Name() string {
	return fmt.Sprintf("%s: %s/%s", p.Source.Name(), p.MountPoint, p.Path)
}
