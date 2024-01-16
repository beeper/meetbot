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
          version = "unstable-2023-01-16";
          src = self;

          tags = [ "goolm" ];

          subPackages = [ "cmd/meetbot" ];

          vendorHash = "sha256-3a4H2wDy13EwApNVttMu5SaXILGcGYm51XppIDNYk9k=";
        };
        defaultPackage = packages.meetbot;

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go_1_21 pre-commit gotools go-tools ];
        };
      }));
}
