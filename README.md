# Tungsten

Tungsten aims to be a highly flexible and programmable DNS server. I am creating this to act as my internal DNS solution, resolving private or internal names. I am not sure if this will ever be suitable for public, high volume DNS resolution.

> [!WARNING]
> This is very much under development. It is far from stable, and I do not recommend that others use this.

## Features/Goals

- [x] Easily add all sorts of records using structured, typesafe configuration instead of RFC 1035 syntax
- [x] Autopopulate names from a Tailscale IPN/Tailnet
- [x] Configuration hot-reloading
- [ ] Forward DNS queries to different places depending on what zone answers for them
- [ ] Fully recursive resolution (likely with libunbound)
- [ ] SkyDNS-like serving from etcd

## Usage

### Configuration

**Enviroment Variables**
Environment variables are used for a few select things; mostly related to logging at the moment.

| Variable              | Default    | Description                                                                                                                                                                                   |
|-----------------------|------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `TUNGSTEN_LOG_FORMAT` | `json`     | Can be `json` or `pretty`; determines what log output looks like.                                                                                                                             |
| `TUNGSTEN_LOG_LEVEL`  | `2` (warn) | This value gets passed into [`zerolog.ParseLevel`](https://pkg.go.dev/github.com/rs/zerolog@v1.34.0#ParseLevel). See zerolog level docs [here](https://github.com/rs/zerolog#leveled-logging) |
| `TUNGSTEN_DEV_MODE`   | `false`    | Any truthy or falsy value (ie `0`, `1`, `true`, `false`, etc)                                                                                                                                 |

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

## Credits & Inspiration
- [SDNS](https://github.com/semihalev/sdns) (Go patterns for starting/stopping stuff, and DoH/DoQ things)
- [damomurf/coredns-tailscale](https://github.com/damomurf/coredns-tailscale) (Pulling information from tailscale for self-hosted MagicDNS)
- [CoreDNS `bind` plugin](https://github.com/coredns/coredns/blob/abb0a52c5ffcff1421098effd3a58e1c9c01fbbe/plugin/bind/setup.go) (Translating iface names to bindable addresses)