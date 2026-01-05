{
  description = "Grimorio - CLI spellbook for developer productivity";

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
            pname = "grimorio";
            version = "0.2.0";
            src = ./.;
            vendorHash = "sha256-F1VJ3HbnQJxLIe/Gr3tR3BivMLHYG9WBtP2J0I5ZjmU=";
          };
        });
    };
}
