{
  description = "APIRight - A framework to take sqlc structs to a ready API";

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
            go_1_23
            gopls
            gotools
            go-tools
            delve
            sqlite
            postgresql
            sqlc
            air # for hot reloading
          ];

          shellHook = ''
            echo "🚀 APIRight development environment"
            echo "Go version: $(go version)"
            echo "Available tools: go, gopls, sqlc, air, sqlite3, psql"
            echo ""
            echo "Quick start:"
            echo "  go mod tidy    # Install dependencies"
            echo "  go run examples/basic/main.go    # Run example"
            echo "  air           # Hot reload development"
            echo ""
          '';

          CGO_ENABLED = "1";
        };

        packages.default = pkgs.buildGoModule {
          pname = "apiright";
          version = "0.1.0";
          src = ./.;
          vendorHash = null;
          
          buildInputs = with pkgs; [
            sqlite
          ];

          meta = with pkgs.lib; {
            description = "A framework to take sqlc structs to a ready API";
            homepage = "https://github.com/bata94/apiright";
            license = licenses.mit;
            maintainers = [ ];
          };
        };
      });
}