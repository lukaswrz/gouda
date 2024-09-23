{
  description = "Go HTTP file upload and download utilities";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    devenv.url = "github:cachix/devenv";
    treefmt.url = "github:numtide/treefmt-nix";
  };

  nixConfig = {
    extra-trusted-public-keys = "devenv.cachix.org-1:w1cLUi8dv3hnoSPGAuibQv+f9TZLr6cv/Hm9XgU50cw=";
    extra-substituters = "https://devenv.cachix.org";
  };

  outputs = {
    self,
    nixpkgs,
    flake-parts,
    devenv,
    treefmt,
  } @ inputs:
    flake-parts.lib.mkFlake {inherit inputs;} {
      imports = [devenv.flakeModule];

      systems = nixpkgs.lib.systems.flakeExposed;

      perSystem = {
        pkgs,
        system,
        ...
      }: let
        treefmtEval = treefmt.lib.evalModule pkgs ./treefmt.nix;
      in {
        devenv.shells.default.imports = [./devenv.nix];

        formatter = treefmtEval.config.build.wrapper;

        checks.formatting = treefmtEval.${pkgs.system}.config.build.check self;
      };
    };
}
