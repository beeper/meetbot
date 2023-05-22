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
        {
          packages.meetbot = pkgs.buildGoModule rec {
            pname = "meetbot";
            version = "unstable-2023-05-20";

            src = ./.;

            subPackages = [ "cmd/meetbot" ];

            propagatedBuildInputs = [ pkgs.olm ];

            vendorSha256 = "sha256-fCD9JDa2Xk8WyshKVLAcIjmIjdoKrq/hBaGN2pMXjpU=";
          };
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
