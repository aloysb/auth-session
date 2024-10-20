{
  description = "A Nix-flake-based Go development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
    templ.url = "github:a-h/templ";
  };

  outputs = { self, nixpkgs, flake-utils, templ }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells = {
          default = pkgs.mkShell {

            buildInputs = with pkgs; [
              go_1_22
              gotools
              golangci-lint
              gopls
              go-outline
              gopkgs
              templ.packages.${system}.templ
            ];

            preBuild = ''
              ${templ.packages.${system}.templ}/bin/templ generate
            '';

            shellHook = ''
              ${pkgs.go_1_22}/bin/go version
            '';
          };
        };
      });
}

