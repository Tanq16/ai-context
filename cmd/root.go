package cmd

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/tanq16/ai-context/internal/aicontext"
	"github.com/tanq16/ai-context/utils"
)

var cmdFlags struct {
	threads      int
	url          string
	listFile     string
	includeGlobs []string
	excludeGlobs []string
	maxSize      int64
	offline      bool
}

var AppVersion = "dev-build"
var debugFlag bool
var forAIFlag bool

var rootCmd = &cobra.Command{
	Use:     "ai-context",
	Short:   "Produce AI context-file for GitHub project, directory, or YouTube video.",
	Version: AppVersion,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && cmdFlags.listFile == "" {
			cmdFlags.url = args[0]
		} else if len(args) == 0 && cmdFlags.listFile == "" {
			utils.PrintFatal("no URL argument or list file provided", nil)
		} else if len(args) > 0 && cmdFlags.listFile != "" {
			utils.PrintFatal("received both URL argument and list file", nil)
		}

		var urls []string
		if cmdFlags.listFile == "" {
			urls = append(urls, cmdFlags.url)
		} else {
			file, err := os.Open(cmdFlags.listFile)
			if err != nil {
				utils.PrintFatal("failed to open list file", err)
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
				utils.PrintFatal("failed to read list file", scanner.Err())
			}
		}
		aicontext.Handler(urls, cmdFlags.includeGlobs, cmdFlags.excludeGlobs, cmdFlags.maxSize, cmdFlags.threads, false, cmdFlags.offline)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func setupLogs() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
		NoColor:    false,
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debugFlag {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = zerolog.New(output).With().Timestamp().Logger()
		utils.GlobalDebugFlag = true
	}
	if forAIFlag {
		utils.GlobalForAIFlag = true
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}
}

func init() {
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&forAIFlag, "for-ai", false, "AI-friendly output (plain text, piped input)")
	rootCmd.MarkFlagsMutuallyExclusive("debug", "for-ai")

	cobra.OnInitialize(setupLogs)

	rootCmd.Flags().StringVarP(&cmdFlags.listFile, "file", "f", "", "File with list of URLs to process")
	rootCmd.Flags().IntVarP(&cmdFlags.threads, "threads", "t", 10, "Number of threads to use for processing")
	rootCmd.Flags().StringSliceVarP(&cmdFlags.includeGlobs, "include", "i", []string{}, "Include files matching globs (e.g., '*.go,*.md')")
	rootCmd.Flags().StringSliceVarP(&cmdFlags.excludeGlobs, "exclude", "e", []string{}, "Exclude files matching globs (e.g., 'tests,docs')")
	rootCmd.Flags().Int64VarP(&cmdFlags.maxSize, "max-size", "s", 10485760, "Maximum file size in bytes to include (default 10MB)")
	rootCmd.Flags().BoolVarP(&cmdFlags.offline, "offline", "o", false, "Disable online fetching (e.g., r.jina.ai)")
}
