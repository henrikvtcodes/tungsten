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
        config,
        self',
        inputs',
        pkgs,
        system,
        ...
      }: {
        packages = rec {
          default = pkgs.callPackage ./tungsten.nix {
            inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
          };
          tungsten = default;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            gomod2nix.packages.${system}.default
          ];
        };
        formatter = pkgs.alejandra;
      };
    };
}