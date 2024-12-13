package cmd

import (
	"fmt"
	"os"
	"strings"
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
	ignoreList []string
	videoURL   string
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
		if cmdFlags.videoURL != "" {
			segments, err := aicontext.DownloadTranscript(cmdFlags.videoURL)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get transcript")
			}
			var content strings.Builder
			content.WriteString("# Video Transcript\n\n")
			for _, segment := range segments {
				content.WriteString(fmt.Sprintf("[%s] %s\n\n", segment.StartTime, segment.Text))
			}
			if err := os.WriteFile(cmdFlags.output, []byte(content.String()), 0644); err != nil {
				log.Fatal().Err(err).Msg("failed to write transcript")
			}
			log.Info().Str("output", cmdFlags.output).Msg("successfully generated transcript")
			return
		}
		if cmdFlags.directory == "" && cmdFlags.url == "" {
			log.Fatal().Err(fmt.Errorf("either directory or url must be specified")).Msg("failed to process source")
		}
		if cmdFlags.directory != "" && cmdFlags.url != "" {
			log.Fatal().Err(fmt.Errorf("cannot specify both directory and url")).Msg("failed to process source")
		}
		// Create the processor
		processor := aicontext.NewProcessor(aicontext.ProcessorConfig{
			OutputPath:        cmdFlags.output,
			AdditionalIgnores: cmdFlags.ignoreList,
		})
		// Process based on input type
		var err error
		if cmdFlags.directory != "" {
			err = processor.ProcessDirectory(cmdFlags.directory)
		} else {
			err = processor.ProcessGitHubURL(cmdFlags.url)
		}
		if err != nil {
			log.Fatal().Err(err).Msg("failed to process source")
		}
		log.Info().Str("output", cmdFlags.output).Msg("successfully generated context")
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
	rootCmd.Flags().StringVarP(&cmdFlags.url, "url", "u", "", "GitHub URL to process")
	rootCmd.Flags().StringVarP(&cmdFlags.videoURL, "video", "v", "", "YouTube video URL to process")
	rootCmd.Flags().StringVarP(&cmdFlags.output, "output", "o", "ai-context.md", "Output file path (default: ai-context.md)")
	rootCmd.Flags().StringSliceVarP(&cmdFlags.ignoreList, "ignore", "i", []string{}, "Additional patterns to ignore (e.g., 'tests,docs')")
}
