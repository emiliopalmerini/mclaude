{
  description = "Claude Watcher - Dashboard for tracking Claude Code session usage";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in {
      packages = forAllSystems (system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in {
          default = pkgs.buildGoModule {
            pname = "claude-watcher";
            version = "0.2.0";
            src = ./.;
            vendorHash = "sha256-gODQwQB52Qes2zmjWHZoX+SCF9or4cE3S/MKdKD3RIg=";
            proxyVendor = true;

            nativeBuildInputs = [ pkgs.templ pkgs.sqlc ];

            preBuild = ''
              sqlc generate
              templ generate
            '';

            subPackages = [ "cmd" "cmd/session-tracker" ];

            postInstall = ''
              mv $out/bin/cmd $out/bin/claude-watcher
              mkdir -p $out/share/claude-watcher
              cp -r static $out/share/claude-watcher/
            '';
          };
        });

      devShells = forAllSystems (system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in {
          default = pkgs.mkShell {
            packages = [ pkgs.go_1_25 pkgs.templ pkgs.sqlc ];
          };
        });
    };
}
