{
  description = "Google Meet Matrix Bot";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    (flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { system = system; };
      in rec {
        packages.meetbot = pkgs.buildGoModule {
          pname = "meetbot";
          version = "unstable-2024-07-10";
          src = self;
          subPackages = [ "cmd/meetbot" ];

          tags = [ "goolm" ];

          vendorHash = "sha256-JLUCKKfIBlyBu/SK2u/rMt/CZXQfe5o9jIq4GOu2wC8=";
        };
        defaultPackage = packages.meetbot;

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go pre-commit gotools go-tools ];
        };
      }));
}
