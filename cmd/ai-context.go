package cmd

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tanq16/ai-context/aicontext"
)

var cmdFlags struct {
	directory  string
	url        string
	output     string
	listFile   string
	ignoreList []string
}

var urlRegex = map[string]string{
	"gh":     "https://github.com/.+/.+",
	"yt":     "https://youtu.be/.+",
	"ytfull": "https://www.youtube.com/watch\\?v=.+",
}

var rootCmd = &cobra.Command{
	Use: "ai-context",
	// REFERENCES:
	// - https://github.com/yamadashy/repomix
	// - https://github.com/tombulled/innertube
	Short: "Produce AI context-file for GitHub project, directory, or YouTube video.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug, _ := cmd.Flags().GetBool("debug"); !debug {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		aicontext.Handler(cmdFlags.directory, cmdFlags.url, cmdFlags.listFile, cmdFlags.output, cmdFlags.ignoreList)
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

	rootCmd.Flags().StringVarP(&cmdFlags.directory, "directory", "d", "", "Local directory to process")
	rootCmd.Flags().StringVarP(&cmdFlags.url, "url", "u", "", "URL to process (GitHub, YouTube)")
	rootCmd.Flags().StringVarP(&cmdFlags.listFile, "file", "f", "", "File with list of URLs to process")
	rootCmd.Flags().StringVarP(&cmdFlags.output, "output", "o", "ai-context.md", "Output file path (default: ai-context.md)")
	rootCmd.Flags().StringSliceVarP(&cmdFlags.ignoreList, "ignore", "i", []string{}, "Additional patterns to ignore (e.g., 'tests,docs'); helpful with GitHub or local directories")
}
