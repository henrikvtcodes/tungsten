{
  buildGoApplication,
  pkgs,
  pkl-go,
  ...
}:
buildGoApplication {
  pname = "tungsten";
  version = "0.0.0";

  src = ./.;
  pwd = ./.;
  modules = ./gomod2nix.toml;

  preBuild = ''
  export PATH="$PATH:${pkgs.pkl}/bin"
  ${pkl-go}/bin/pkl-gen-go config/Server.pkl
  '';

  buildInputs = with pkgs; [
  unbound
  unbound.lib

  ];
}