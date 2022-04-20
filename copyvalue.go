package hvc

import (
	"errors"
	"fmt"
	"time"
)

// CopySource is an interface that defines the methods needed to retrieve source
// values from secrets to update a target secret.
type CopySource interface {
	DetermineUpdatedTime() (time.Time, error)
	RetrieveSourceValues() (map[string]interface{}, error)
}

// CopySourceValues implements the CopySource interface and uses a map of
// strings to CopyValue objects to offer fine grained control over which source
// secret keys are used to update the target secret.
type CopySourceValues struct {
	CopySource

	values map[string]*CopyValue
}

// CopySourceSecret implements the CopySource interface and uses a CopyValue
// object to refer to a single source secret which is entirely copied to the
// target secret.
type CopySourceSecret struct {
	CopySource

	secret *CopyValue
}

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

// DetermineUpdatedTime retrieves the updated_time value from the single source
// secret's metadata.
func (p *CopySourceSecret) DetermineUpdatedTime() (time.Time, error) {
	secret, err := p.secret.Source.Read(fmt.Sprintf("%s/metadata/%s", p.secret.MountPoint, p.secret.Path))
	if err != nil {
		return time.Unix(0, 0), fmt.Errorf("failed to retrieve source secret %q metadata: %w", p.secret.Name(), err)
	}

	if secret == nil {
		return time.Unix(0, 0), errors.New("source secret %q does not exist")
	}

	updatedTime := secret.Data["updated_time"].(string)
	sourceTime, err := time.Parse(time.RFC3339Nano, updatedTime)
	if err != nil {
		return time.Unix(0, 0), fmt.Errorf("failed to parse updated_time value %s: %w", updatedTime, err)
	}

	return sourceTime, nil
}

// DetermineUpdatedTime retrieves the updated_time value from each of the source
// secrets and returns the greatest of those values.
func (p *CopySourceValues) DetermineUpdatedTime() (time.Time, error) {
	// Update time cache
	sourceUpdateTimes := make(map[string]time.Time)

	// Remember the maximum updated_time found
	maxUpdateTime := time.Unix(0, 0)

	// Get the metadta of every secret source value
	for _, value := range p.values {
		// Check if this value's secret has already been examined.
		if _, found := sourceUpdateTimes[value.Name()]; !found {
			secret, err := value.Source.Read(fmt.Sprintf("%s/metadata/%s", value.MountPoint, value.Path))
			if err != nil {
				return time.Unix(0, 0), fmt.Errorf("failed to retrieve source secret %q metadata: %w", value.Name(), err)
			}

			if secret == nil {
				return time.Unix(0, 0), fmt.Errorf("source secret %q does not exist", value.Name())
			}

			updatedTime := secret.Data["updated_time"].(string)
			sourceTime, err := time.Parse(time.RFC3339Nano, updatedTime)
			if err != nil {
				return time.Unix(0, 0), fmt.Errorf("failed to parse updated_time value %s for secret %q: %w", updatedTime, value.Name(), err)
			}

			sourceUpdateTimes[value.Name()] = sourceTime

			if sourceTime.After(maxUpdateTime) {
				maxUpdateTime = sourceTime
			}
		}
	}

	return maxUpdateTime, nil
}

func (p *CopySourceSecret) RetrieveSourceValues() (map[string]interface{}, error) {
	secret, err := p.secret.Source.Read(fmt.Sprintf("%s/data/%s", p.secret.MountPoint, p.secret.Path))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve source secret %q values: %w", p.secret.Name(), err)
	}

	if secret == nil {
		return nil, fmt.Errorf("source secret %q does not exist", p.secret.Name())
	}

	return secret.Data["data"].(map[string]interface{}), nil
}

func (p *CopySourceValues) RetrieveSourceValues() (map[string]interface{}, error) {
	secretValues := make(map[string]interface{})

	for k, v := range p.values {
		secret, err := v.Source.Read(fmt.Sprintf("%s/data/%s", v.MountPoint, v.Path))
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve source secret %q values: %w", v.Name(), err)
		}

		if secret == nil {
			return nil, fmt.Errorf("source secret %q does not exist", v.Name())
		}

		if secret.Data == nil || secret.Data["data"] == nil {
			return nil, fmt.Errorf("source secret %q values are missing", v.Name())
		}

		data := secret.Data["data"].(map[string]interface{})
		value, found := data[v.Key]
		if !found {
			return nil, fmt.Errorf("missing key %s in source secret %q", v.Key, v.Name())
		}

		secretValues[k] = value
	}

	return secretValues, nil
}
