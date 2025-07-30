{
  lib,
  buildGoModule,
  pkgs,
  fetchFromGitHub,
  ...
}:
buildGoModule rec {
  pname = "pkl-go";
  version = "0.10.0";

  subPackages = ["cmd/pkl-gen-go"];

  src = fetchFromGitHub {
    owner = "apple";
    repo = "pkl-go";
    rev = "v${version}";
    sha256 = "sha256-XjcQApsEBzaWdFK/QS+g0t2CO1zW9t7er4xiH8MnDO8=";
  };
  vendorHash = "sha256-YySJhQCboZJXwSJ9fTBkiIouErHMlwYcT8qHdtRyMQI=";

  preBuild = with pkgs; ''
    ${gnused}/bin/sed -i "s|var Version = \"development\"|var Version = \"${version}\"|" cmd/pkl-gen-go/pkl-gen-go.go
  '';
}
