package hvc

import (
	"errors"
	"fmt"
	"time"

	"github.com/marcboudreau/hvc/spec"
)

// Copy is a structure that defines how a secret in the target Vault server
// should be copied.
type Copy struct {
	// MountPoint is the path where the target secret's KV secrets engine is
	// mounted.
	MountPoint string

	// Path is the path of the target secret within the KV secrets engine.
	Path string

	// SourceSecret is the CopySource, which defines what source secret values
	// are used to update the target secret.
	SourceSecret CopySource
}

// NewCopy creates a Copy structure using the provided spec.Copy structure and
// map of source names to Vault interfaces.
func NewCopy(spec *spec.Copy, sources map[string]Vault) (*Copy, error) {
	targetMountPoint := spec.MountPoint
	if targetMountPoint == "" {
		targetMountPoint = "kv"
	}

	// Make sure that Path is provided.
	if spec.Path == "" {
		return nil, errors.New("copy element must provide a target secret path")
	}

	copy := &Copy{
		MountPoint: targetMountPoint,
		Path:       spec.Path,
	}

	if spec.Secret != nil {
		// Make sure that if Secret is not nil, Values is empty.
		if len(spec.Values) != 0 {
			return nil, errors.New("copy element cannot contain both secret and values")
		}

		// Make sure the single Secret is referencing an existing Vault
		vault, found := sources[spec.Secret.Source]
		if !found {
			return nil, fmt.Errorf("secret is referencing a non-existing source Vault %s", spec.Secret.Source)
		}

		sourceMountPoint := spec.Secret.MountPoint
		if sourceMountPoint == "" {
			sourceMountPoint = "kv"
		}

		sourcePath := spec.Secret.Path
		if sourcePath == "" {
			sourcePath = spec.Path
		}

		copy.SourceSecret = &CopySourceSecret{
			secret: &CopyValue{
				Source:     vault,
				MountPoint: sourceMountPoint,
				Path:       sourcePath,
			},
		}
	} else {
		copyValues := make(map[string]*CopyValue)
		for k, v := range spec.Values {
			sourceVault := sources[v.Source]
			if sourceVault == nil {
				return nil, fmt.Errorf("secret value for target secret key %s is referencing a non-existing source Vault %s", k, v.Source)
			}

			mountPoint := v.MountPoint
			if v.MountPoint == "" {
				mountPoint = "kv"
			}

			path := v.Path
			if v.Path == "" {
				path = spec.Path
			}

			key := v.Key
			if v.Key == "" {
				key = k
			}

			copyValues[k] = &CopyValue{
				Source:     sourceVault,
				MountPoint: mountPoint,
				Path:       path,
				Key:        key,
			}
		}

		copy.SourceSecret = &CopySourceValues{
			values: copyValues,
		}
	}

	return copy, nil
}

// TargetUpdateTime retrieves the updated_time value from the target secret's
// metadata.
func (p *Copy) TargetUpdateTime(target Vault) (time.Time, error) {
	// Get the metadata of the secret in the target Vault server
	secret, err := target.Read(fmt.Sprintf("%s/metadata/%s", p.MountPoint, p.Path))
	if err != nil {
		return time.Unix(0, 0), fmt.Errorf("failed to retrieve target secret %q metadata: %w", p.Name(), err)
	}

	// If the secret in the target Vault server doesn't exist, secret will be nil
	// and so will err.
	if secret == nil {
		return time.Unix(0, 0), nil
	}

	// Parse the retrieved time
	updatedTime := secret.Data["updated_time"].(string)
	targetTime, err := time.Parse(time.RFC3339Nano, updatedTime)
	if err != nil {
		return time.Unix(0, 0), fmt.Errorf("failed to parse the retrieved value for the updated_time %s: %w", updatedTime, err)
	}

	return targetTime, nil
}

// DetermineNeedToCopy retrieves the metadata for every source secret referenced
// by a Copy structure and compares the updated time from each with the provided
// target time. If any source updated time is more recent than the target time,
// the function will return true, otherwise it will return false. If an error is
// encountered, false and the error will be returned.
func (p *Copy) DetermineNeedToCopy(targetTime time.Time) (bool, error) {
	sourceTime, err := p.SourceSecret.DetermineUpdatedTime()
	if err != nil {
		return false, err
	}

	return targetTime.Before(sourceTime), nil
}

// UpdateTargetSecret updates the target secret referenced in the receiver using
// the provided target Vault interface.
func (p *Copy) UpdateTargetSecret(target Vault) error {
	targetData, err := p.SourceSecret.RetrieveSourceValues()
	if err != nil {
		return err
	}

	_, err = target.Write(fmt.Sprintf("%s/data/%s", p.MountPoint, p.Path), map[string]interface{}{"data": targetData})
	if err != nil {
		return fmt.Errorf("failed to update target secret %q: %w", p.Name(), err)
	}

	return nil
}

// Name returns a canonical name for the receiver.
func (p *Copy) Name() string {
	return fmt.Sprintf("%s/%s", p.MountPoint, p.Path)
}

// Execute executes the copy operation of the receiver using the provided target
// Vault interface. The function uses the provided index and channel to report
// any errors encountered.
func (p *Copy) Execute(target Vault, index int, ch chan error) {
	// Get the metadata of the secret in the target Vault server
	targetTime, err := p.TargetUpdateTime(target)
	if err != nil {
		ch <- fmt.Errorf("failed to execute copy %d: %w", index, err)
		return
	}

	needsUpdate, err := p.DetermineNeedToCopy(targetTime)
	if err != nil {
		ch <- fmt.Errorf("failed to execute copy %d: %w", index, err)
		return
	}

	if needsUpdate {
		err = p.UpdateTargetSecret(target)
		if err != nil {
			ch <- fmt.Errorf("failed to execute copy %d: %w", index, err)
			return
		}
	}

	ch <- nil
}
