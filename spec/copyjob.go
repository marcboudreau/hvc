package spec

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// CopyJob contains an entire specification of all secret values that need
// to be copied from one or more source Vault servers to a target Vault server.
type CopyJob struct {
	// Target is a VaultSpec structure that contains the details to establish an
	// API Client connection and authenticate with the target Vault server.
	Target *Vault `json:"target"`

	// Sources is a map of source names to VaultSpec structures.  These VaultSpec
	// structures contain the details to establish API Client connections to the
	// source Vault servers needed by this job.
	Sources map[string]*Vault `json:"sources"`

	// Copies is an array of CopySpec structures that define how each secret
	// should be copied.
	Copies []*Copy `json:"copies"`
}

func LoadSpec(in io.Reader) (*CopyJob, error) {
	specBytes, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec: %w", err)
	}

	spec := string(specBytes)

	spec = os.ExpandEnv(spec)

	var copyJob CopyJob
	if err := json.NewDecoder(strings.NewReader(spec)).Decode(&copyJob); err != nil {
		return nil, fmt.Errorf("failed to decode JSON spec: %w", err)
	}

	return &copyJob, nil
}
