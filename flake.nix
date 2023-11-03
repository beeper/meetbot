{
  description = "Google Meet Matrix Bot";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    (flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs { system = system; };
        in
        rec {
          packages.meetbot = pkgs.buildGoModule rec {
            pname = "meetbot";
            version = "unstable-2023-11-03";
            src = self;

            subPackages = [ "cmd/meetbot" ];

            propagatedBuildInputs = [ pkgs.olm ];

            vendorSha256 = "sha256-UFLL86hahs2AE2VrO4oKO73KtAbpwO5+l2Km25MRKk0=";
          };
          defaultPackage = packages.meetbot;
          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              go_1_19
              olm
              pre-commit
              gotools
              gopls
            ];
          };
        }
      ));
}
