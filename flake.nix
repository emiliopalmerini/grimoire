{
  description = "Grimoire - CLI spellbook for developer productivity";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in {
      packages = forAllSystems (system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in {
          default = pkgs.buildGoModule {
            pname = "grimoire";
            version = "0.2.0";
            src = ./.;
            vendorHash = "sha256-Ia1+Xg3CT6xMNJEVCNB6orOC926/nM8dj/suAtlonRU=";
          };
        });
    };
}
