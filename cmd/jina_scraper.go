package cmd

import (
	"bufio"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tanq16/ai-context/internal/aicontext"
	"github.com/tanq16/ai-context/utils"
)

var jinaCmdFlags struct {
	threads      int
	url          string
	listFile     string
	includeGlobs []string
	excludeGlobs []string
	maxSize      int64
}

var jinaCmd = &cobra.Command{
	Use:   "jina-scraper",
	Short: "Produce AI context-file using jina.ai for web fetching.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && jinaCmdFlags.listFile == "" {
			jinaCmdFlags.url = args[0]
		} else if len(args) == 0 && jinaCmdFlags.listFile == "" {
			utils.PrintFatal("no URL argument or list file provided", nil)
		} else if len(args) > 0 && jinaCmdFlags.listFile != "" {
			utils.PrintFatal("received both URL argument and list file", nil)
		}

		var urls []string
		if jinaCmdFlags.listFile == "" {
			urls = append(urls, jinaCmdFlags.url)
		} else {
			file, err := os.Open(jinaCmdFlags.listFile)
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
		aicontext.Handler(urls, jinaCmdFlags.includeGlobs, jinaCmdFlags.excludeGlobs, jinaCmdFlags.maxSize, jinaCmdFlags.threads, false, true)
	},
}

func init() {
	rootCmd.AddCommand(jinaCmd)

	jinaCmd.Flags().StringVarP(&jinaCmdFlags.listFile, "file", "f", "", "File with list of URLs to process")
	jinaCmd.Flags().IntVarP(&jinaCmdFlags.threads, "threads", "t", 10, "Number of threads to use for processing")
	jinaCmd.Flags().StringSliceVarP(&jinaCmdFlags.includeGlobs, "include", "i", []string{}, "Include files matching globs (e.g., '*.go,*.md')")
	jinaCmd.Flags().StringSliceVarP(&jinaCmdFlags.excludeGlobs, "exclude", "e", []string{}, "Exclude files matching globs (e.g., 'tests,docs')")
	jinaCmd.Flags().Int64VarP(&jinaCmdFlags.maxSize, "max-size", "s", 10485760, "Maximum file size in bytes to include (default 10MB)")
}
