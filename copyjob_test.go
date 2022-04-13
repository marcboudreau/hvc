package hvc

import (
	"errors"
	"testing"

	vault "github.com/hashicorp/vault/api"
	"github.com/marcboudreau/hvc/spec"
	"github.com/stretchr/testify/assert"
)

func TestNewCopyJob(t *testing.T) {
	for _, testcase := range []struct {
		spec          *spec.CopyJob
		errorAssert   func(assert.TestingT, error, ...interface{}) bool
		copyJobAssert func(assert.TestingT, interface{}, ...interface{}) bool
	}{
		// Happy path!
		{
			spec: &spec.CopyJob{
				Target: &spec.Vault{
					Address: "http://localhost:8200",
					Login: &spec.VaultLogin{
						Token: "root",
					},
				},
				Sources: map[string]*spec.Vault{
					"s1": {
						Address: "http://localhost:8300",
						Login: &spec.VaultLogin{
							Token: "root",
						},
					},
				},
				Copies: []*spec.Copy{
					{
						MountPoint: "kv",
						Path:       "p1",
						Values: map[string]*spec.CopyValue{
							"t1": {
								Source:     "s1",
								MountPoint: "kv",
								Path:       "p1",
								Key:        "k1",
							},
						},
					},
				},
			},
			errorAssert:   assert.NoError,
			copyJobAssert: assert.NotNil,
		},
		// Error bad target Vault spec
		{
			spec: &spec.CopyJob{
				Target: &spec.Vault{
					Address: "http://localhost:8200",
				},
				Sources: map[string]*spec.Vault{},
				Copies:  []*spec.Copy{},
			},
			errorAssert:   assert.Error,
			copyJobAssert: assert.Nil,
		},
		// Error bad source Vault spec
		{
			spec: &spec.CopyJob{
				Target: &spec.Vault{
					Address: "http://localhost:8200",
					Login: &spec.VaultLogin{
						Token: "root",
					},
				},
				Sources: map[string]*spec.Vault{
					"s1": {
						Address: "http://localhost:8200",
					},
				},
				Copies: []*spec.Copy{},
			},
			errorAssert:   assert.Error,
			copyJobAssert: assert.Nil,
		},
		// Error copy value referencing non-existant source vault
		{
			spec: &spec.CopyJob{
				Target: &spec.Vault{
					Address: "http://localhost:8200",
					Login: &spec.VaultLogin{
						Token: "root",
					},
				},
				Sources: map[string]*spec.Vault{},
				Copies: []*spec.Copy{
					{
						MountPoint: "kv",
						Path:       "p1",
						Values: map[string]*spec.CopyValue{
							"t1": {
								Source:     "s1", // non-existant
								MountPoint: "kv",
								Path:       "p1",
								Key:        "k1",
							},
						},
					},
				},
			},
			errorAssert:   assert.Error,
			copyJobAssert: assert.Nil,
		},
	} {
		copyJob, err := NewCopyJob(testcase.spec)
		testcase.errorAssert(t, err)
		testcase.copyJobAssert(t, copyJob)
	}
}

func TestCopyJobExecute(t *testing.T) {
	for _, testcase := range []struct {
		copyJob     *CopyJob
		errorAssert func(assert.TestingT, error, ...interface{}) bool
	}{
		// Happy path!
		{
			copyJob: &CopyJob{
				Target: &FakeVault{
					name: "_target",
					readResponses: []FakeVaultResponse{
						{
							secret: &vault.Secret{
								Data: map[string]interface{}{
									"updated_time": "2000-01-01T00:00:00.000000000Z",
								},
							},
							err: nil,
						},
					},
					writeResponses: []FakeVaultResponse{
						{
							secret: &vault.Secret{},
							err:    nil,
						},
					},
				},
				Copies: []*Copy{
					{
						MountPoint: "kv",
						Path:       "p1",
						Values: map[string]*CopyValue{
							"t1": {
								Source: &FakeVault{
									name: "s1",
									readResponses: []FakeVaultResponse{
										// metadata read
										{
											secret: &vault.Secret{
												Data: map[string]interface{}{
													"updated_time": "2022-01-01T00:00:00.000000000Z",
												},
											},
											err: nil,
										},
										// data read
										{
											secret: &vault.Secret{
												Data: map[string]interface{}{
													"k1": "value",
												},
											},
											err: nil,
										},
									},
								},
								MountPoint: "kv",
								Path:       "p1",
								Key:        "k1",
							},
						},
					},
				},
			},
			errorAssert: assert.NoError,
		},
		// Happy path no need to update
		{
			copyJob: &CopyJob{
				Target: &FakeVault{
					name: "_target",
					readResponses: []FakeVaultResponse{
						{
							secret: &vault.Secret{
								Data: map[string]interface{}{
									"updated_time": "2022-01-01T00:00:00.000000000Z",
								},
							},
							err: nil,
						},
					},
					writeResponses: []FakeVaultResponse{
						{
							secret: &vault.Secret{},
							err:    nil,
						},
					},
				},
				Copies: []*Copy{
					{
						MountPoint: "kv",
						Path:       "p1",
						Values: map[string]*CopyValue{
							"t1": {
								Source: &FakeVault{
									name: "s1",
									readResponses: []FakeVaultResponse{
										// metadata read
										{
											secret: &vault.Secret{
												Data: map[string]interface{}{
													"updated_time": "2012-01-01T00:00:00.000000000Z",
												},
											},
											err: nil,
										},
										// data read
										{
											secret: &vault.Secret{
												Data: map[string]interface{}{
													"k1": "value",
												},
											},
											err: nil,
										},
									},
								},
								MountPoint: "kv",
								Path:       "p1",
								Key:        "k1",
							},
						},
					},
				},
			},
			errorAssert: assert.NoError,
		},
		// Error reading target updated_time
		{
			copyJob: &CopyJob{
				Target: &FakeVault{
					name: "_target",
					readResponses: []FakeVaultResponse{
						{
							secret: nil,
							err:    errors.New("error"),
						},
					},
				},
				Copies: []*Copy{
					{
						MountPoint: "kv",
						Path:       "p1",
						Values: map[string]*CopyValue{
							"t1": {
								Source: &FakeVault{
									name: "s1",
								},
								MountPoint: "kv",
								Path:       "p1",
								Key:        "k1",
							},
						},
					},
				},
			},
			errorAssert: assert.Error,
		},
		// Error reading source updated_time
		{
			copyJob: &CopyJob{
				Target: &FakeVault{
					name: "_target",
					readResponses: []FakeVaultResponse{
						{
							secret: &vault.Secret{
								Data: map[string]interface{}{
									"updated_time": "2000-01-01T00:00:00.000000000Z",
								},
							},
							err: nil,
						},
					},
				},
				Copies: []*Copy{
					{
						MountPoint: "kv",
						Path:       "p1",
						Values: map[string]*CopyValue{
							"t1": {
								Source: &FakeVault{
									name: "s1",
									readResponses: []FakeVaultResponse{
										// metadata read
										{
											secret: nil,
											err:    errors.New("error"),
										},
									},
								},
								MountPoint: "kv",
								Path:       "p1",
								Key:        "k1",
							},
						},
					},
				},
			},
			errorAssert: assert.Error,
		},
		// Error reading source value
		{
			copyJob: &CopyJob{
				Target: &FakeVault{
					name: "_target",
					readResponses: []FakeVaultResponse{
						{
							secret: &vault.Secret{
								Data: map[string]interface{}{
									"updated_time": "2000-01-01T00:00:00.000000000Z",
								},
							},
							err: nil,
						},
					},
				},
				Copies: []*Copy{
					{
						MountPoint: "kv",
						Path:       "p1",
						Values: map[string]*CopyValue{
							"t1": {
								Source: &FakeVault{
									name: "s1",
									readResponses: []FakeVaultResponse{
										// metadata read
										{
											secret: &vault.Secret{
												Data: map[string]interface{}{
													"updated_time": "2022-01-01T00:00:00.000000000Z",
												},
											},
											err: nil,
										},
										// data read
										{
											secret: nil,
											err:    errors.New("error"),
										},
									},
								},
								MountPoint: "kv",
								Path:       "p1",
								Key:        "k1",
							},
						},
					},
				},
			},
			errorAssert: assert.Error,
		},
		// Error writing target secret
		{
			copyJob: &CopyJob{
				Target: &FakeVault{
					name: "_target",
					readResponses: []FakeVaultResponse{
						{
							secret: &vault.Secret{
								Data: map[string]interface{}{
									"updated_time": "2000-01-01T00:00:00.000000000Z",
								},
							},
							err: nil,
						},
					},
					writeResponses: []FakeVaultResponse{
						{
							secret: nil,
							err:    errors.New("error"),
						},
					},
				},
				Copies: []*Copy{
					{
						MountPoint: "kv",
						Path:       "p1",
						Values: map[string]*CopyValue{
							"t1": {
								Source: &FakeVault{
									name: "s1",
									readResponses: []FakeVaultResponse{
										// metadata read
										{
											secret: &vault.Secret{
												Data: map[string]interface{}{
													"updated_time": "2022-01-01T00:00:00.000000000Z",
												},
											},
											err: nil,
										},
										// data read
										{
											secret: &vault.Secret{
												Data: map[string]interface{}{
													"k1": "value",
												},
											},
											err: nil,
										},
									},
								},
								MountPoint: "kv",
								Path:       "p1",
								Key:        "k1",
							},
						},
					},
				},
			},
			errorAssert: assert.Error,
		},
	} {
		testcase.errorAssert(t, testcase.copyJob.Execute())
	}
}
