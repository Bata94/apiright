{
  description = "APIRight - simple net/http Framework wrapper";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            just

            # go
            go_1_25
            air

            golangci-lint
            gotools
            gopls
            delve
            vegeta

            sqlc
            goose

            templ
            tailwindcss_4
          ];
        };
      }
    );
}
