# hvc - Hashicorp Vault Copier


<img align="left" alt="Vault Copying" src="./copy.png" width="67" height="80" style="margin:5px"/>

This project builds the **hvc** application that facilitates copying secrets
from one or more Vault server(s) to a single target Vault server.

It features the ability to copy secrets to a Vault server using source secrets
from one or more source Vault servers.


## Usage

This section explains how to use this application.  The application uses a
**Copy Job Specification** file to model which secrets are copied to a target
Vault.

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

The also exists an integration test suite in the *testing/* directory that
allows testing use cases with real Vault servers. The test suite launches a
target Vault server that listens at the address `http://localhost:8200` and a
source Vault server that listens at the address `http://localhost:8300`.

The test suite then uses a Terraform project to configure resources in each
Vault server. It uses the *testing/target.tf* file to configure resources in the
target Vault server and the *testing/source.tf* file to configure resources in
the source Vault server.

Once both Vault servers are configured, the test suite iterates over each Copy
Job Specification file found in the *testing/successul/* directory and executes
the **copy** command with each file and marks the test case as passed if the
application exits with a success status (exit code `0`) or failed otherwise.
