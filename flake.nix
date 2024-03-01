{
  inputs = {
    nixpkgs.url = "nixpkgs";
    flake-utils.url = "flake-utils";
  };

  outputs = { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        go = pkgs.go_1_19;
      in
      {
        devShells.default = pkgs.mkShell {
          packages = [
            go
          ];
        };

        checks.go_test = pkgs.stdenv.mkDerivation {
          name = "xcontext-go-test";
          src = ./.;
          __impure = true;

          nativeBuildInputs = [
            pkgs.cacert
            go
          ];

          buildPhase = ''
            runHook preBuild

            HOME="$(mktemp -d)"
            go test -mod=readonly -race -v ./...

            runHook postBuild
          '';

          installPhase = ''
            runHook preInstall
            touch "$out"
            runHook postInstall
          '';
        };
      }
    );
}
