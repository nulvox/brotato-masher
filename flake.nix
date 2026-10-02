{
  description = "Brotato save editor development environment";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in {
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [ go nodejs chromium playwright-test jq curl ];
          shellHook = ''
            export GO111MODULE=on
            export PLAYWRIGHT_BROWSERS_PATH="${pkgs.chromium}/bin"
            echo "brotato-masher dev shell: Go $(go version | awk '{print $3}') + Chromium"
            echo "Browser fallback: chromium --headless --no-sandbox --dump-dom"
          '';
        };
      });
    };
}
