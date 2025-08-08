{
  description = "A declarative DNS server";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/release-25.05";

    flake-parts.url = "github:hercules-ci/flake-parts";

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs @ {
    flake-parts,
    nixpkgs,
    gomod2nix,
    ...
  }:
    flake-parts.lib.mkFlake {inherit inputs;} {
      systems = ["x86_64-linux" "aarch64-darwin" "x86_64-darwin" "aarch64-linux"];
      perSystem = {
        pkgs,
        system,
        self',
        ...
      }: rec {
        packages = rec {
          default = tungsten;
          tungsten = pkgs.callPackage ./tungsten.nix {
            inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
            self = self';
          };

          tungsten-full = pkgs.callPackage ./tungsten-full.nix {
            inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
          };

          pkl-go =
            pkgs.callPackage ./pkl-gen-go.nix {
            };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            gomod2nix.packages.${system}.default
            unbound.lib
            unbound
            packages.pkl-go
            pkl
          ];

          TUNGSTEN_DEV_MODE = 1;
          TUNGSTEN_LOG_LEVEL = "debug";
          TUNGSTEN_LOG_FORMAT = "pretty";
        };
        formatter = pkgs.alejandra;
      };
    };
}
