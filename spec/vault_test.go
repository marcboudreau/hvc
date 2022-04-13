package spec

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVaultParseJSON(t *testing.T) {

	for _, testcase := range []struct {
		json            string
		errorAssert     func(assert.TestingT, error, ...interface{}) bool
		expectedAddress string
		expectedLogin   *VaultLogin
	}{
		// Invalid JSON testcase
		{
			json:        `{%f0s93`,
			errorAssert: assert.Error,
		},
		// Valid JSON testcase
		{
			json:            `{"address":"a"}`,
			errorAssert:     assert.NoError,
			expectedAddress: "a",
		},
		// Valid JSON testcase with Login using token
		{
			json:            `{"address":"http://vault:8200","login":{"token":"tkn"}}`,
			errorAssert:     assert.NoError,
			expectedAddress: "http://vault:8200",
			expectedLogin: &VaultLogin{
				Token: "tkn",
			},
		},
		// Valid JSON testcase with Login using Kubernetes method
		{
			json:            `{"address":"http://vault:8200","login":{"kubernetes":{"role":"my-role","jwt-path":"/home/jwt"}}}`,
			errorAssert:     assert.NoError,
			expectedAddress: "http://vault:8200",
			expectedLogin: &VaultLogin{
				Kubernetes: &VaultKubernetesLogin{
					Role:    "my-role",
					JWTPath: "/home/jwt",
				},
			},
		},
	} {
		var v Vault

		testcase.errorAssert(t, json.NewDecoder(strings.NewReader(testcase.json)).Decode(&v))
		assert.Equal(t, testcase.expectedAddress, v.Address)
		assert.Equal(t, testcase.expectedLogin, v.Login)
		if testcase.expectedLogin != nil {
			assert.Equal(t, testcase.expectedLogin.Token, v.Login.Token)
			if testcase.expectedLogin.Kubernetes != nil {
				assert.NotNil(t, v.Login.Kubernetes)
				assert.Equal(t, testcase.expectedLogin.Kubernetes.Role, v.Login.Kubernetes.Role)
				assert.Equal(t, testcase.expectedLogin.Kubernetes.JWTPath, v.Login.Kubernetes.JWTPath)
			} else {
				assert.Nil(t, v.Login.Kubernetes)
			}
		} else {
			assert.Nil(t, v.Login)
		}
	}
}
