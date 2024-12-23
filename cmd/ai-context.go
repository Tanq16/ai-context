package cmd

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tanq16/ai-context/aicontext"
)

var cmdFlags struct {
	threads    int
	url        string
	listFile   string
	ignoreList []string
}

var rootCmd = &cobra.Command{
	Use:   "ai-context",
	Short: "Produce AI context-file for GitHub project, directory, or YouTube video.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug, _ := cmd.Flags().GetBool("debug"); !debug {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if cmdFlags.url == "" && cmdFlags.listFile == "" {
			log.Fatal().Msg("either url or list of urls must be specified")
		}
		if cmdFlags.url != "" && cmdFlags.listFile != "" {
			log.Fatal().Msg("cannot specify both url and list file")
		}

		// Input URL processing
		var urls []string
		if cmdFlags.listFile == "" {
			urls = append(urls, cmdFlags.url)
		} else {
			file, err := os.Open(cmdFlags.listFile)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to open list file")
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				url := strings.TrimSpace(scanner.Text())
				if url != "" {
					urls = append(urls, url)
				}
			}
			if scanner.Err() != nil {
				log.Fatal().Err(scanner.Err()).Msg("failed to read list file")
			}
		}
		aicontext.Handler(urls, cmdFlags.ignoreList, cmdFlags.threads)
		log.Info().Msg("All Operations Completed!")
	},
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Configure global logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
		NoColor:    false, // Enable color output
	}
	// Set global logger with console writer
	log.Logger = zerolog.New(output).With().Timestamp().Logger()
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug mode")

	rootCmd.Flags().StringVarP(&cmdFlags.url, "url", "u", "", "URL to process (GitHub, YouTube)")
	rootCmd.Flags().StringVarP(&cmdFlags.listFile, "file", "f", "", "File with list of URLs to process")
	rootCmd.Flags().IntVarP(&cmdFlags.threads, "threads", "t", 5, "Number of threads to use for processing (default: 5)")
	rootCmd.Flags().StringSliceVarP(&cmdFlags.ignoreList, "ignore", "i", []string{}, "Additional patterns to ignore (e.g., 'tests,docs'); helpful with GitHub or local directories")
}
