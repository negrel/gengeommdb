{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    let
      outputsWithoutSystem = { };
      outputsWithSystem = flake-utils.lib.eachDefaultSystem (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          devShells = {
            default = pkgs.mkShell {
              buildInputs = with pkgs; [
                go
                gopls
              ];
            };
          };
          packages = {
            default = pkgs.buildGoModule {
              name = "gengeommdb";
              version = "0.1.0";
              vendorHash = "sha256-YLFSLxKRIMWJSPc3i98HycsNqC+TycEFO9Ba3dLTMyU=";

              src = ./.;
            };
          };
        }
      );
    in
    outputsWithSystem // outputsWithoutSystem;
}
