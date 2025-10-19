{
  description = "Budget CLI tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      packages.default = pkgs.buildGoModule {
        pname = "bud";
        version = "unstable-2025-10-19";
        src = ./.;
        vendorHash = "sha256-9RXwv7xwmrty5HBMbxzonyiARXRi4vA6+rxDYV5gVSU=";
        ldflags = [
          "-X main.version=unstable-2025-10-19"
          "-X main.commit=03ad9bcbccbf1d82337416a30fe8a3986e567eac"
        ];
      };
    });
}
