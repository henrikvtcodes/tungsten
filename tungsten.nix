{
  buildGoApplication,
  pkgs,
  self,
  ...
}:
buildGoApplication rec {
  pname = "tungsten";
  version = "0.0.0";

  src = ./.;
  pwd = ./.;
  modules = ./gomod2nix.toml;

  CGO_ENABLED = 0;

  # preBuild = ''
  #   export XDG_CACHE_HOME="$TMPDIR/xdg-cache"
  #   export XDG_DATA_HOME="$TMPDIR/xdg-data"
  #   export JAVA_HOME="$TMPDIR/java"
  #   export GRAALVM_HOME="$TMPDIR/graalvm"
  #   export PKL_HOME="$TMPDIR/pkl-home"
  #   mkdir -p "$PKL_HOME" "$XDG_DATA_HOME" "$XDG_CACHE_HOME" "$JAVA_HOME" "$GRAALVM_HOME" "$TMPDIR/java-tmp" # Ensure the directory exists

  #   export _JAVA_OPTIONS="-Djava.io.tmpdir=$TMPDIR/java-tmp"

  #   ls -l $TMPDIR

  #   export PKL_EXEC=${pkgs.pkl}/bin/pkl
  #   ${pkl-go}/bin/pkl-gen-go --cache-dir $PKL_HOME  --generator-settings generator-settings.pkl --base-path github.com/henrikvtcodes/tungsten config/Server.pkl
  # '';

  preBuild = ''
  ${pkgs.gnused}/bin/sed -i "s|@version-dev@|${version}|g" util/version.go
  ${pkgs.gnused}/bin/sed -i "s|@sha-dev@|$(echo ${self.rev} || cut -c1-7)|g" util/version.go
  '';
}
