package cmd

import (
	"bufio"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tanq16/ai-context/aicontext"
	u "github.com/tanq16/ai-context/utils"
)

var cmdFlags struct {
	threads    int
	url        string
	listFile   string
	ignoreList []string
}

var AIContextVersion = "dev"

var rootCmd = &cobra.Command{
	Use:     "ai-context",
	Short:   "Produce AI context-file for GitHub project, directory, or YouTube video.",
	Version: AIContextVersion,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && cmdFlags.listFile == "" {
			cmdFlags.url = args[0]
		} else if len(args) == 0 && cmdFlags.listFile == "" {
			u.PrintError("no URL argument or list file provided")
			os.Exit(1)
		} else if len(args) > 0 && cmdFlags.listFile != "" {
			u.PrintError("received both URL argument and list file")
			os.Exit(1)
		}

		// Input URL processing
		var urls []string
		if cmdFlags.listFile == "" {
			urls = append(urls, cmdFlags.url)
		} else {
			file, err := os.Open(cmdFlags.listFile)
			if err != nil {
				u.PrintError("failed to open list file")
				os.Exit(1)
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
				u.PrintError("failed to read list file")
				os.Exit(1)
			}
		}
		aicontext.Handler(urls, cmdFlags.ignoreList, cmdFlags.threads)
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
	rootCmd.Flags().StringVarP(&cmdFlags.listFile, "file", "f", "", "File with list of URLs to process")
	rootCmd.Flags().IntVarP(&cmdFlags.threads, "threads", "t", 10, "Number of threads to use for processing")
	rootCmd.Flags().StringSliceVarP(&cmdFlags.ignoreList, "ignore", "i", []string{}, "Additional patterns to ignore (e.g., 'tests,docs'); helpful with GitHub or local directories")
}
