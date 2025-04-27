{
  description = "A declarative DNS server";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/release-24.11";

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
        ...
      }: {
        packages = rec {
          default = pkgs.callPackage ./tungsten.nix {
            inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
          };
          tungsten = default;
          pkl-go = pkgs.callPackage ./pkl-gen-go.nix {
            # inherit (gomod2nix.legacyPackages.${system}) buildGoModule;
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            gomod2nix.packages.${system}.default
            # packages.pkl-go
          ];

          TUNGSTEN_DEV_MODE = 1;
          TUNGSTEN_LOG_LEVEL = "debug";
          TUNGSTEN_LOG_FORMAT = "pretty";
        };
        formatter = pkgs.alejandra;
      };
    };
}