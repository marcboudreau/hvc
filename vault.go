package hvc

import (
	"context"
	"errors"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	k8sauth "github.com/hashicorp/vault/api/auth/kubernetes"
	"github.com/marcboudreau/hvc/spec"
)

// Vault is an interface that defines the methods needed to interact with a
// Vault server.
type Vault interface {
	Name() string
	Read(string) (*vault.Secret, error)
	Write(string, map[string]interface{}) (*vault.Secret, error)
}

// realVault is an object that creates an API Client connection to a real
// Vault server.
type realVault struct {
	Vault

	name   string
	client *vault.Client
}

// NewVault creates a Vault connection using the provided spec.Vault object.
// This function creates the API Client object and then resolves the contained
// VaultLogin object to obtain a valid Vault token and sets it in the client.
func NewVault(spec *spec.Vault, name string) (Vault, error) {
	vaultClient, err := vault.NewClient(nil)
	if err != nil {
		return nil, err
	}

	vaultClient.SetAddress(spec.Address)

	if spec.Login != nil {
		if spec.Login.Token != "" {
			vaultClient.SetToken(spec.Login.Token)
		} else if spec.Login.Kubernetes != nil {
			auth, err := k8sauth.NewKubernetesAuth(
				spec.Login.Kubernetes.Role,
				k8sauth.WithServiceAccountTokenPath(spec.Login.Kubernetes.JWTPath),
				k8sauth.WithMountPath(spec.Login.Kubernetes.MountPoint),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize Kubernetes authentication method: %w", err)
			}

			_, err = vaultClient.Auth().Login(context.TODO(), auth)
			if err != nil {
				return nil, fmt.Errorf("failed to authentication with Vault server: %w", err)
			}
		}
	}

	if vaultClient.Token() == "" {
		return nil, errors.New("no Vault token obtained")
	}

	return &realVault{
		client: vaultClient,
		name:   name,
	}, nil
}

func (p *realVault) Name() string {
	return p.name
}

// Read uses the receiver's client field to dispatch a corresponding Read
// call.
func (p *realVault) Read(path string) (*vault.Secret, error) {
	return p.client.Logical().Read(path)
}

// Write uses the receiver's client field to dispatch a corresponding Write
// call.
func (p *realVault) Write(path string, data map[string]interface{}) (*vault.Secret, error) {
	return p.client.Logical().Write(path, data)
}
