{
  lib,
  buildGoModule,
  buildGoApplication,
  pkgs,
  ...
}:
buildGoApplication {
  pname = "tungsten";
  version = "0.0.0";

  src = ./.;
  pwd = ./.;
  modules = ./gomod2nix.toml;
}