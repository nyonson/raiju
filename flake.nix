{
  description = "Raiju";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }: 
    let
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      # Helper function to pull packages for certain system.
      pkgsFor = system: nixpkgs.legacyPackages.${system};
    in {
      # Basic executable.
      packages = forAllSystems (system: 
        let pkgs = pkgsFor system;
        in {
          default = pkgs.buildGoModule {
            pname = "raiju";
            version = "0.11.1";
            src = ./.;
            vendorHash = "sha256-sNCEZjR+7xWVKLAOOunvNyned4c9VRebt96PhAxiByk=";
          };
        }
      );

      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/raiju";
        };
      });

      # Development shell hooked up to direnv.
      devShells = forAllSystems (system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              gotools
            ];
          };
        }
      );

      # NixOS module for systemd configuration.
      nixosModules.raiju = { config, lib, pkgs, ... }:
        let
          # Option validators.
          hostType = lib.types.strMatching "([a-zA-Z0-9.-]+):([0-9]{1,5})";
          twoIntArray = with lib.types; addCheck (listOf int) (x: 
            builtins.length x == 2 && builtins.all (i: builtins.isInt i) x
          );
          threeIntArray = with lib.types; addCheck (listOf int) (x: 
            builtins.length x == 3 && builtins.all (i: builtins.isInt i) x
          );
          # Convert list of ints to comma-separated string.
          intListToString = intList: builtins.concatStringsSep "," (map toString intList);
        in
        {
          options.services.raiju = {
            enable = lib.mkEnableOption "Raiju Daemon";
            rpcHost = lib.mkOption {
              type = hostType;
              default = "localhost:10009";
              description = "The address and port of the LND instance's RPC interface.";
            };
            macaroonFile = lib.mkOption {
              type = lib.types.path;
              description = "Path to a macaroon of the LND instance with admin priviledges.";
            };
            tlsCertificateFile = lib.mkOption {
              type = lib.types.path;
              description = "Path to the TLS certificate of the LND instance.";
            };
            liquidityFees = lib.mkOption {
              type = threeIntArray;
              default = [1 50 2500];
              description = ''
                The feerates (PPM) raiju applies to channels. The first option is applied to channels with too much local liquidity. 
                The second is for balanced channels (in between the two threshold values). The third option is applied to 
                channels with too little local liquidity.
              '';
            };
            liquidityThresholds = lib.mkOption {
              type = twoIntArray;
              default = [80 20];
              description = ''
                Channel liquidity thresholds defined as the percent of a channel's capacity which is local liquidity. Channels
                which have a local liquidity percent higher than the first value are considered "too much local" and  have the 
                minimum fee applied. Channels with a local liquidity percent in between the two values are considered balanced 
                and have the middle fee applied. Channels with a local liquidity percent below the second value are 
                considered "too little local" and have the maximum fee applied.
              '';
            };
            liquidityStickiness = lib.mkOption {
              type = lib.types.int;
              default = 10;
              description = ''
                The percentage of channel capacity to wait before returning a channel's fee to balanced. For example, if
                a channel goes above the local liquidity maximum threshold, but then sinks back below it, the fees won't
                change back to balanced until it is 10% (the default stickiness setting in this case) below the
                maximum threshold in order to avoid fee configuration thrashing.
              '';
            };
          };

          config = lib.mkIf config.services.raiju.enable {
            systemd.services.raiju = lib.mkIf config.systemd.enable {
              description = "Raiju";
              wantedBy = [ "multi-user.target" ];
              # *Requires* establishes the dependency, while *after* establishes the order.
              # Using the strict requirement of requires so that raiju only starts
              # if LND does. The *wants* dependency is a loose requirement which will still attempt to
              # start raiju even if LND fails, which would be useless.
              requires = [ "lnd.service" ];
              after = [ "lnd.service" ];
              serviceConfig = {
                ExecStart = "${self.packages.${pkgs.system}.default}/bin/raiju daemon";
                Restart = "always";
              };
              environment = {
                RAIJU_HOST = "${config.services.raiju.rpcHost}";
                RAIJU_MAC_PATH = "${config.services.raiju.macaroonFile}";
                RAIJU_TLS_PATH = "${config.services.raiju.tlsCertificateFile}";
                RAIJU_LIQUIDITY_FEES = "${intListToString config.services.raiju.liquidityFees}";
                RAIJU_LIQUIDITY_STICKINESS = "${toString config.services.raiju.liquidityStickiness}";
                RAIJU_LIQUIDITY_THRESHOLDS = "${intListToString config.services.raiju.liquidityThresholds}";
              };
            };

            environment = {
              systemPackages = [ self.packages.${pkgs.system}.default ];
              # Set system variables so raiju command can be called easily.
              variables = {
                RAIJU_HOST = "${config.services.raiju.rpcHost}";
                RAIJU_MAC_PATH = "${config.services.raiju.macaroonFile}";
                RAIJU_TLS_PATH = "${config.services.raiju.tlsCertificateFile}";
                RAIJU_LIQUIDITY_FEES = "${intListToString config.services.raiju.liquidityFees}";
                RAIJU_LIQUIDITY_STICKINESS = "${toString config.services.raiju.liquidityStickiness}";
                RAIJU_LIQUIDITY_THRESHOLDS = "${intListToString config.services.raiju.liquidityThresholds}";
              };
            };
          };
        };

      # Expose the NixOS module
      nixosModules.default = self.nixosModules.raiju;
    };
}
