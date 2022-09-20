{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/998f0f7924198b2460458728de59fe738997f28e.tar.gz") {}
}:

pkgs.mkShell {
  packages = [ pkgs.go_1_19 ];
}
