{ pkgs }:

pkgs.buildGoModule {
  pname = "mclaude";
  version = "0.2.0";
  src = pkgs.lib.cleanSourceWith {
    src = ../.;
    filter = path: type:
      let baseName = baseNameOf path;
      in !(baseName == "vendor" && type == "directory");
  };

  vendorHash = "sha256-W0RbtVOVKWh5cgXpEtuTCzAInnLpIYG6CKKC5FmUnrY=";

  # Use proxy mode to preserve native library files in go-libsql
  proxyVendor = true;

  subPackages = [ "cmd/mclaude" ];

  # Enable CGO for libsql support
  env.CGO_ENABLED = "1";

  # Build dependencies
  nativeBuildInputs = with pkgs; [
    pkg-config
  ];

  buildInputs = with pkgs; [
  ] ++ pkgs.lib.optionals pkgs.stdenv.hostPlatform.isDarwin [
    pkgs.apple-sdk_15
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
