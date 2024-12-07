{
  languages.go.enable = true;

  pre-commit.hooks = {
    # Go
    gotest.enable = true;
    golangci-lint.enable = true;

    # Nix
    alejandra.enable = true;
    deadnix.enable = true;
    statix.enable = true;
  };
}
