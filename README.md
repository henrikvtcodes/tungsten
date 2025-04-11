# Tungsten

Tungsten aims to be a highly flexible and programmable DNS server. I am creating this to act as my internal DNS solution, resolving private or internal names. I am not sure if this will ever be suitable for public, high volume DNS resolution.

> [!WARNING]
> This is very much under development. It is far from stable, and I do not recommend that others use this.

## Features/Goals

- [ ] Easily add all sorts of records using structured, typesafe configuration instead of RFC 1035 syntax
- [ ] Autopopulate names from a Tailscale IPN/Tailnet
- [ ] Forward DNS queries to different places depending on what zone answers for them
- [ ] Fully recursive resolution (likely with libunbound)
- [ ] Caching with `bbolt`
- [ ] SkyDNS-like serving from etcd

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
