{ pkgs }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gopls
    gotools
    go-tools
    sqlc
    templ
    sqlite
    pkg-config
  ];

  shellHook = ''
    export CGO_ENABLED=1
    echo "Claude Watcher dev shell"
    echo "Commands: make build, make test, make sqlc, make templ"
  '';
}
