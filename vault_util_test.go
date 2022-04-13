package hvc

import (
	"errors"

	vault "github.com/hashicorp/vault/api"
)

type FakeVault struct {
	Vault

	name           string
	readResponses  []FakeVaultResponse
	writeResponses []FakeVaultResponse
}

type FakeVaultResponse struct {
	secret *vault.Secret
	err    error
}

func (p *FakeVault) InitializeClient() error {
	return nil
}

func (p *FakeVault) Read(path string) (*vault.Secret, error) {
	response := p.readResponses[0]
	p.readResponses = p.readResponses[1:]

	return response.secret, response.err
}

func (p *FakeVault) Write(path string, data map[string]interface{}) (*vault.Secret, error) {
	response := p.writeResponses[0]
	p.writeResponses = p.writeResponses[1:]

	return response.secret, response.err
}

func (p *FakeVault) Name() string {
	return p.name
}

type UninitializableVault struct {
	name string
}

func (p *UninitializableVault) InitializeClient() error {
	return errors.New("error")
}

func (p *UninitializableVault) Read(path string) (*vault.Secret, error) {
	return nil, nil
}

func (p *UninitializableVault) Write(path string, data map[string]interface{}) (*vault.Secret, error) {
	return nil, nil
}

func (p *UninitializableVault) Name() string {
	return p.name
}
