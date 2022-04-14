package hvc

import (
	"testing"

	"github.com/marcboudreau/hvc/spec"
	"github.com/stretchr/testify/assert"
)

func TestNewVault(t *testing.T) {
	for _, testcase := range []struct {
		spec        *spec.Vault
		errorAssert func(assert.TestingT, error, ...interface{}) bool
		vaultAssert func(assert.TestingT, interface{}, ...interface{}) bool
	}{
		// Happy path with Token!
		{
			spec: &spec.Vault{
				Address: "http://localhost:8200",
				Login: &spec.VaultLogin{
					Token: "root",
				},
			},
			errorAssert: assert.NoError,
			vaultAssert: assert.NotNil,
		},
		// Missing Login section
		{
			spec: &spec.Vault{
				Address: "http://localhost:8200",
			},
			errorAssert: assert.Error,
			vaultAssert: assert.Nil,
		},
		// No Token
		{
			spec: &spec.Vault{
				Address: "http://localhost:8200",
				Login:   &spec.VaultLogin{},
			},
			errorAssert: assert.Error,
			vaultAssert: assert.Nil,
		},
	} {
		vault, err := NewVault(testcase.spec, "test")
		testcase.errorAssert(t, err)
		testcase.vaultAssert(t, vault)
	}
}

func TestVaultName(t *testing.T) {
	vault, err := NewVault(&spec.Vault{
		Address: "http://localhost:8200",
		Login: &spec.VaultLogin{
			Token: "root",
		},
	}, "test")
	assert.NoError(t, err)
	assert.NotNil(t, vault)
	assert.Equal(t, "test", vault.Name())
}

// func TestIntegrationNewVault(t *testing.T) {
// 	handleIntegrationTest(t)

// }

func TestNewVaultHandlesDefault(t *testing.T) {
	for _, testcase := range []struct {
		spec            *spec.Vault
		expectedAddress string
	}{
		// Address omitted
		{
			spec: &spec.Vault{
				Login: &spec.VaultLogin{
					Token: "root",
				},
			},
			expectedAddress: "https://127.0.0.1:8200",
		},
	} {
		vault, _ := NewVault(testcase.spec, "test")
		assert.Equal(t, testcase.expectedAddress, vault.(*realVault).client.Address())
	}
}
