package cmd

import (
	"bufio"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"log"

	"github.com/tanq16/ai-context/internal/aicontext"
	"github.com/tanq16/ai-context/utils"
)

var cmdFlags struct {
	threads    int
	url        string
	listFile   string
	ignoreList []string
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
			log.Fatalf("ERROR [ai-context] no URL argument or list file provided")
		} else if len(args) > 0 && cmdFlags.listFile != "" {
			log.Fatalf("ERROR [ai-context] received both URL argument and list file")
		}

		// Input URL processing
		var urls []string
		if cmdFlags.listFile == "" {
			urls = append(urls, cmdFlags.url)
		} else {
			file, err := os.Open(cmdFlags.listFile)
			if err != nil {
				log.Fatalf("ERROR [ai-context] failed to open list file: %v", err)
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
				log.Fatalf("ERROR [ai-context] failed to read list file: %v", scanner.Err())
			}
		}
		aicontext.Handler(urls, cmdFlags.ignoreList, cmdFlags.threads, false)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func setupLogs() {
	if debugFlag {
		utils.GlobalDebugFlag = true
	}
	if forAIFlag {
		utils.GlobalForAIFlag = true
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
	rootCmd.Flags().StringSliceVarP(&cmdFlags.ignoreList, "ignore", "i", []string{}, "Additional patterns to ignore (e.g., 'tests,docs'); helpful with GitHub or local directories")
}
