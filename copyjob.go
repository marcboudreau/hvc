package hvc

import (
	"fmt"
	"sync"

	"github.com/marcboudreau/hvc/spec"
)

// CopyJob is a structure that is built from a CopyJobSpec. It contains fully
// initialized API Client connections to Vault servers and resolved references
// to them in the CopyValue objects.
type CopyJob struct {
	// Target specifies the connection to the target Vault server.
	Target Vault

	// Copies is an array of Copy objects that define what needs to be copied
	// to the target Vault server.
	Copies []*Copy
}

// NewCopyJob creates a CopyJob structure using the data in the provided
// CopyJobSpec object.
func NewCopyJob(spec *spec.CopyJob) (*CopyJob, error) {
	copyJob := &CopyJob{}

	targetVault, err := NewVault(spec.Target, "_target")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize target Vault: %w", err)
	}

	copyJob.Target = targetVault

	sourceVaults := make(map[string]Vault)
	for sourceVaultKey, sourceVaultSpec := range spec.Sources {
		sourceVault, err := NewVault(sourceVaultSpec, sourceVaultKey)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize source Vault %q: %w", sourceVaultKey, err)
		}

		sourceVaults[sourceVaultKey] = sourceVault
	}

	copyJob.Copies = make([]*Copy, len(spec.Copies))
	for i, copySpec := range spec.Copies {
		copy, err := NewCopy(copySpec, sourceVaults)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve copy %d: %w", i+1, err)
		}

		copyJob.Copies[i] = copy
	}

	return copyJob, nil
}

// Execute copies the secret values referenced in the provided Copy
// structure using the receiver's configured target and source Vault
// connections.
func (p *CopyJob) Execute() []error {
	ch := make(chan error)

	waitGroup := sync.WaitGroup{}

	for i, copy := range p.Copies {
		waitGroup.Add(1)
		go func(copy *Copy, i int, ch chan error) {
			copy.Execute(p.Target, i, ch)
			waitGroup.Done()
		}(copy, i, ch)
	}

	errorSlice := []error{}
	for range p.Copies {
		err := <-ch
		if err != nil {
			errorSlice = append(errorSlice, err)
		}
	}

	return errorSlice
}
