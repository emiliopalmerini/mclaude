{ pkgs }:

pkgs.buildGoModule {
  pname = "mclaude";
  version = "0.1.0";
  src = pkgs.lib.cleanSource ../.;

  vendorHash = "sha256-2M/v9Ppm4MImCER582Y1GSmUZfxwRZ7rIlHQr0Untok=";

  subPackages = [ "cmd/mclaude" ];

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
    mkdir -p $out/share/mclaude
    cp -r ${../.}/migrations $out/share/mclaude/
  '';

  meta = with pkgs.lib; {
    description = "Analytics and experimentation platform for Claude Code";
    homepage = "https://github.com/emiliopalmerini/mclaude";
    license = licenses.mit;
    maintainers = [ ];
  };
}
