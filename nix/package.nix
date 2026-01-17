{ pkgs }:

pkgs.buildGoModule {
  pname = "claude-watcher";
  version = "0.1.0";
  src = pkgs.lib.cleanSource ../.;

  vendorHash = "sha256-RtVSv2eNHOTyPSDmuKqFTuwT4O7Kvk/SGvwf2O4k72E=";

  subPackages = [ "cmd/claude-watcher" ];

  # Enable CGO for libsql support
  env.CGO_ENABLED = "1";

  # Build dependencies for libsql
  nativeBuildInputs = with pkgs; [
    pkg-config
  ];

  buildInputs = with pkgs; [
    sqlite
  ];

  # Copy migrations for embedded use
  postInstall = ''
    mkdir -p $out/share/claude-watcher
    cp -r ${../.}/migrations $out/share/claude-watcher/
  '';

  meta = with pkgs.lib; {
    description = "Analytics and experimentation platform for Claude Code";
    homepage = "https://github.com/emiliopalmerini/claude-watcher";
    license = licenses.mit;
    maintainers = [ ];
  };
}
