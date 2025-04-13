package cmd

import (
	"os"
	"strings"

	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/internal/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var envFile string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run sync",
	Run: func(cmd *cobra.Command, args []string) {
		readEnvFile()

		if zerolog.GlobalLevel() == zerolog.DebugLevel {
			dict := zerolog.Dict()
			for _, env := range os.Environ() {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					dict = dict.Str(parts[0], parts[1])
				}
			}

			log.Info().Dict("env", dict).Msg("Environment variables")
		}

		service, err := service.Init()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize service")
		}

		if err = service.Run(); err != nil {
			log.Fatal().Err(err).Msg("Sync failed")
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVar(&envFile, "env-file", "", "Read env from `.env` file")
}

func readEnvFile() {
	if envFile == "" {
		return
	}

	if err := config.LoadEnvFile(envFile); err != nil {
		log.Fatal().Err(err).Msg("Failed to load env file")
	}
}
