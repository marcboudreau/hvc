package spec

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadSpec(t *testing.T) {
	for _, testcase := range []struct {
		input         io.Reader
		errorAssert   func(assert.TestingT, error, ...interface{}) bool
		copyJobAssert func(assert.TestingT, interface{}, ...interface{}) bool
	}{
		// Happy path!
		{
			input:         strings.NewReader(`{"target":{"address":"http://localhost:8200"},"sources":{"s1":{"address":"http://localhost:8300"}},"copies":[{}]}`),
			errorAssert:   assert.NoError,
			copyJobAssert: assert.NotNil,
		},
		// Error
		{
			input:         strings.NewReader(`{6fl*@`),
			errorAssert:   assert.Error,
			copyJobAssert: assert.Nil,
		},
	} {
		copyJob, err := LoadSpec(testcase.input)
		testcase.errorAssert(t, err)
		testcase.copyJobAssert(t, copyJob)
	}
}

func TestEnvironmentVariableExpansionLoadSpec(t *testing.T) {
	spec := `{"target":{"address":"${TARGET_VAULT_ADDR}"},"sources":{"s1":{"address":"${SOURCE_VAULT_ADDR}"}},"copies":[{}]}`

	os.Setenv("TARGET_VAULT_ADDR", "http://target:8200")
	os.Setenv("SOURCE_VAULT_ADDR", "http://source:8200")

	copyJob, err := LoadSpec(strings.NewReader(spec))
	assert.NoError(t, err)
	assert.NotNil(t, copyJob)
	assert.Equal(t, "http://target:8200", copyJob.Target.Address)
	assert.Equal(t, "http://source:8200", copyJob.Sources["s1"].Address)
}
