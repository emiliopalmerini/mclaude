{
  description = "Claude Watcher - Analytics and experimentation platform for Claude Code";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          claude-watcher = pkgs.callPackage ./nix/package.nix {};
          default = self.packages.${system}.claude-watcher;
        };

        devShells.default = pkgs.callPackage ./nix/devShell.nix {};
      }
    );
}
