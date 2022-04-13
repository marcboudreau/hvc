package hvc

import (
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

	// Values is a map of target secret key names to CopyValue objects that define
	// which source secret value to copy.
	Values map[string]*CopyValue
}

func NewCopy(spec *spec.Copy, sources map[string]Vault) (*Copy, error) {
	copyValues := make(map[string]*CopyValue)
	for k, v := range spec.Values {
		sourceVault := sources[v.Source]
		if sourceVault == nil {
			return nil, fmt.Errorf("secret value %q is referencing a non-existing source Vault", k)
		}

		copyValues[k] = &CopyValue{
			Source:     sourceVault,
			MountPoint: v.MountPoint,
			Path:       v.Path,
			Key:        v.Key,
		}
	}

	return &Copy{
		MountPoint: spec.MountPoint,
		Path:       spec.Path,
		Values:     copyValues,
	}, nil
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

func (p *Copy) DetermineNeedToCopy(targetTime time.Time) (bool, error) {
	// Update time cache
	sourceUpdateTimes := make(map[string]time.Time)

	needsUpdate := false

	// Get the metadta of every secret source value
	for _, value := range p.Values {
		// Check if this value's secret has already been examined.
		if _, found := sourceUpdateTimes[value.Name()]; !found {
			secret, err := value.Source.Read(fmt.Sprintf("%s/metadata/%s", value.MountPoint, value.Path))
			if err != nil {
				return false, fmt.Errorf("failed to retrieve source secret %q metadata: %w", value.Name(), err)
			}

			sourceTime, _ := time.Parse(time.RFC3339Nano, secret.Data["updated_time"].(string))
			if sourceTime.After(targetTime) {
				needsUpdate = true
			}

			sourceUpdateTimes[value.Name()] = sourceTime
		}
	}

	return needsUpdate, nil
}

func (p *Copy) UpdateTargetSecret(target Vault) error {
	targetData := make(map[string]interface{})

	for targetKey, value := range p.Values {
		secret, err := value.Source.Read(fmt.Sprintf("%s/data/%s", value.MountPoint, value.Path))
		if err != nil {
			return fmt.Errorf("failed to retrieve source secret %q data: %w", value.Name(), err)
		}

		secretData, found := secret.Data["data"].(map[string]interface{})
		if !found {
			return fmt.Errorf("source secret %q does not contain a data key", value.Name())
		}

		sourceValue, found := secretData[value.Key]
		if !found {
			return fmt.Errorf("key %s does not exist in source secret %q", value.Key, value.Name())
		}

		targetData[targetKey] = sourceValue
	}

	_, err := target.Write(fmt.Sprintf("%s/data/%s", p.MountPoint, p.Path), map[string]interface{}{"data": targetData})
	if err != nil {
		return fmt.Errorf("failed to update target secret %q: %w", p.Name(), err)
	}

	return nil
}

func (p *Copy) Name() string {
	return fmt.Sprintf("%s/%s", p.MountPoint, p.Path)
}

func (p *Copy) Execute(target Vault) error {
	// Get the metadata of the secret in the target Vault server
	targetTime, err := p.TargetUpdateTime(target)
	if err != nil {
		return err
	}

	needsUpdate, err := p.DetermineNeedToCopy(targetTime)
	if err != nil {
		return err
	}

	if needsUpdate {
		return p.UpdateTargetSecret(target)
	}

	return nil
}
