# let_blog/flake.nix
{
  description = "A development shell for the Full-Stack Blog project";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShell = pkgs.mkShell {
          packages = [
            # The Docker CLI and daemon
            pkgs.docker
          ];
          
          shellHook = ''
            echo "---"
            echo "Welcome to the Deployment Shell."
            echo "---"
            PS1="Docker $PS1"
          '';
        };
      }
    );
}
