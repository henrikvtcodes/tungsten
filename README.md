# Tungsten

Tungsten aims to be a highly flexible and programmable DNS server. I am creating this to act as my internal DNS solution, resolving private or internal names. I am not sure if this will ever be suitable for public, high volume DNS resolution.

> [!WARNING]  
> **Use at your own risk:**  
> I cannot guarantee that this DNS server is fully stable. I use this in my homelab, where some downtime is okay.

## Features/Goals

- [x] Easily add all sorts of records using structured, typesafe configuration instead of RFC 1035 syntax
- [x] Autopopulate names from a Tailscale IPN/Tailnet
- [x] Configuration hot-reloading
- [x] Forward DNS queries to different places depending on what zone answers for them
- [x] Fully recursive resolution with libunbound
- [ ] Serving from etcd
- [ ] Allow zones to individually bind to specific interfaces and addresses
- [ ] Shortcut syntax for certain DNS-SD services, as well as simpler SRV record syntax

_Please note: this is not an exhaustive list, and will be updated in the future to reflect the current standing of this project._

## Usage

### Configuration

Configuration is done via a combination of environment variables (for settings that are not likely to change) and the [Pkl](https://pkl-lang.org) configuration file.

#### Enviroment Variables

Environment variables are used for a few select things; mostly related to logging at the moment.

| Variable              | Default    | Description                                                                                                                                                                                   |
| --------------------- | ---------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `TUNGSTEN_LOG_FORMAT` | `json`     | Can be `json` or `pretty`; determines what log output looks like.                                                                                                                             |
| `TUNGSTEN_LOG_LEVEL`  | `2` (warn) | This value gets passed into [`zerolog.ParseLevel`](https://pkg.go.dev/github.com/rs/zerolog@v1.34.0#ParseLevel). See zerolog level docs [here](https://github.com/rs/zerolog#leveled-logging) |
| `TUNGSTEN_DEV_MODE`   | `false`    | Any truthy or falsy value (ie `0`, `1`, `true`, `false`, etc)                                                                                                                                 |

#### Main Config

You should check out the `example.pkl` file for a general idea of how it's structured, but here's the general gist:

- `amends` tells Pkl what file to fetch for the type definitions (the url must return the text content of the template file)
- It is highly recommended to have the Pkl language extension installed, as that will hint the type definitions and make it way easier to write your own config

### Enabling Recursive Resolution

By default, binaries are not built with support for recursive resolution.

To build with libunbound, add the `unbound` tag to the build command like so:

```go
go build -tags unbound
```

#### Nix Pkgs

The `tungsten-full` package includes unbound, whereas the regular `tungsten` or default package do not.

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
