{ pkgs }:

pkgs.buildGoModule {
  pname = "mclaude-otel";
  version = "0.1.0";
  src = pkgs.lib.cleanSource ../.;

  vendorHash = "sha256-2M/v9Ppm4MImCER582Y1GSmUZfxwRZ7rIlHQr0Untok=";

  subPackages = [ "cmd/mclaude-otel" ];

  # Enable CGO for libsql support
  env.CGO_ENABLED = "1";

  # Build dependencies for libsql
  nativeBuildInputs = with pkgs; [
    pkg-config
  ];

  buildInputs = with pkgs; [
    sqlite
  ];

  meta = with pkgs.lib; {
    description = "OpenTelemetry receiver for mclaude";
    homepage = "https://github.com/emiliopalmerini/mclaude";
    license = licenses.mit;
    maintainers = [ ];
  };
}
