# hvc - Hashicorp Vault Copier

This project builds an application that facilitates copying secrets from one or
more Vault server(s) to a single target Vault server.

## Usage

This section explains how to use this application.  The application uses a
**Copy Job Specification** file to model which secrets are copied to a target Vault.

### Copy Job Specification

The **Copy Job Specification** is a JSON encoded document that is fully described in the [SPECIFICATION.md](./SPECIFICATION.md) file.

### Features

The application inspects the _updated_time_ of both the target secret and every
source secret to determine whether an update of the target secret is necessary.

## Development

This section provides details on how this project is developed. 

### Build

The application is built using the **go build** tool.  To compile and generate
the binary, run the following command:

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

