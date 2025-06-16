{
  description = "APIRight - simple net/http Framework wrapper";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            sqlc
            goose
            air
            golangci-lint
            gotools
            gopls
            delve
          ];

          shellHook = ''
            echo "🚀 APIRight Development Environment"
            echo "Go version: $(go version)"
            echo "SQLC version: $(sqlc version)"
            echo ""
            echo "Available tools:"
            echo "  - go: Go programming language"
            echo "  - sqlc: Generate type-safe code from SQL"
            echo "  - air: Live reload for Go apps"
            echo "  - golangci-lint: Go linter"
            echo ""
            echo "To get started:"
            echo "  1. Initialize your Go module: go mod init"
            echo "  2. Set up SQLC configuration"
            echo "  3. Start developing your API framework!"
          '';
        };

        # packages.default = pkgs.buildGoModule {
        #   pname = "apiright";
        #   version = "0.1.0";
        #   src = ./.;
        #   vendorHash = null;
        # };
      });
}
