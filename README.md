# hvc - Hashicorp Vault Copier


<img align="left" alt="Vault Copying" src="./copy.png" width="67" height="80" style="margin:5px"/>

The **hvc** application copies secrets from one or more source Vault servers to
a single target Vault server.

Each copied secret can be sourced from a single secret (exact copy) or it can be
sourced using multiple keys from different secrets.  The source secrets can also
span different source Vault servers.

## Usage

This section explains how to use the **hvc** application.  The application uses
a **Copy Job Specification** file to model which secrets are copied to a target
Vault.

### Running

The **hvc** application can be run in a Docker container as follows:

```
$ docker run -t -v $PWD/spec.json:/hvc/spec.json ghcr.io/marcboudreau/hvc /hvc/spec.json
```

In this example, the *spec.json* file situated in the current working directory
is mounted into the container and used as the Copy Job Specification.

The **hvc** application can be run as a Kubernetes Job as demonstrated in the
[kubernetes_example.md](./kubernetes_example.md) file.

### Copy Job Specification

The **Copy Job Specification** is a JSON encoded document that is fully
described in the [SPECIFICATION.md](./SPECIFICATION.md) file.

### Features

The application inspects the _updated_time_ of both the target secret and every
source secret to determine whether an update of the target secret is necessary.

The application supports using Vault's Kubernetes Authentication Method to
obtain a valid Vault token.

The **Copy Job Specification** has sensible defaults allowing smaller
specification files (see the [SPECIFICATION.md](./SPECIFICATION.md) file for
details).

### Use Cases

This section outlines a variety of use cases that this application can easily accommodate and provides a sample Copy Job Specification to show how the use case can be handled.

1. Simple exact copies of secrets from one source:
```json
{
  "target": {
    "address": "https://target.vault:8200",
    "login": {
      "token": "${TARGET_VAULT_TOKEN}"
    }
  },
  "sources": {
    "source-vault": {
      "address": "https://source.vault:8200",
      "login": {
        "token": "${SOURCE_VAULT_TOKEN}"
      }
    }
  },
  "copies": [
    {
      "path": "my-service/my-secret",
      "secret": {
        "source": "source-vault"        
      }
    }
  ]
}
```
2. Copy values from different source secrets into single target secret:
```json
{
  "target": {
    "address": "https://target.vault:8200",
    "login": {
      "token": "${TARGET_VAULT_TOKEN}"      
    }
  },
  "sources": {
    "source-vault": {
      "address": "https://source.vault:8200",
      "login": {
        "token": "${SOURCE_VAULT_TOKEN}"
      }
    }
  },
  "copies": [
    {
      "path": "my-service/my-secret",
      "values": {
        "key1": {
          "source": "source-vault",
          "path": "first/secret",
          "key": "some-key"
        },
        "key2": {
          "source": "source-vault",
          "path": "second/secret",
          "key": "other-key"
        }
      }
    }
  ]
}
```

## Development

This section provides details on how this project is developed. 

### Build

The application is built using a GitHub Actions [workflow](.github/workflows/release.yaml), but it can be compiled locally using the following command:

```
$ go build -o hvc ./cmd/hvc
```

### Testing

The **go test** tool is primarily used to test this project.  To run the unit
tests with code coverage measurements, use the following command:

```
$ go test -coverprofile=coverage.out ./...
```

To view the coverage information, run this command:

```
$ go tool cover -html=coverage.out
```

A browser window will open with the source code marked up with coverage
information.

A GitHub Actions workflow is triggered on every push to a GitHub Pull Request
against the *main* branch. This workflow runs a lint checker and the unit tests.

#### Integration Testing

The also exists an integration test suite in the *tests/* directory that
allows testing use cases with real Vault servers within a Kubernetes cluster.
The test suite launches a target Vault server and a source Vault server in the
provided Kubernetes cluster. It then creates Kubernetes Job resources to run
**hvc** for each test case.

The test suite then uses two Terraform projects: one create the Kubernetes
resources (in the *tests/setup/* directory) and another to configure resources
in each Vault server (in the *tests/configure/* directory).

The Terraform projects are applied and later on destroyed by a Go test function.
Once the Go test has applied all of the Terraform configuration, it then uses
the Kubernetes Go client to create the Kubernetes Job and waits for them to
complete. Once the job completes successfully, the test makes API calls directly
to the target Vault server to verify that test case's secrets were correctly
copied.

This integration test is normally skipped unless an environment variable named
**TEST_INTEGRATION** is set to an non-empty value.  So to run only the
integration test:
```
$ cd tests/
$ TEST_INTEGRATION=1 go test ./
```

To run this integration test in conjunction with the unit tests, simply run this
command from the root directory of this repository:
```
$ TEST_INTEGRATION=1 go test ./...
```
