# Tungsten

Tungsten aims to be a highly flexible and programmable DNS server.

> [!WARNING]
> This is very much under development. It is far from stable, and I do not recommend that others use this.

## Developing

### Prerequisites

This project uses Pkl Codegen, and thus certain tools must be installed for this to work.

```sh
go install github.com/apple/pkl-go/cmd/pkl-gen-go@v0.10.0 # This provides the `pkl-gen-go` command
```

### Build

In this case, `generate` references the config files in `config` to run `pkl-gen-go` and create the corresponding go structs/interfaces.

```sh
go generate && go build
```
