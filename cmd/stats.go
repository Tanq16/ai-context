package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/tanq16/ai-context/internal/aicontext"
	"github.com/tanq16/ai-context/utils"
)

var statsCmd = &cobra.Command{
	Use:   "stats [file]",
	Short: "Show file statistics including estimated LLM token count.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		stats, err := aicontext.CalculateFileStats(filePath)
		if err != nil {
			utils.PrintFatal("failed to calculate stats", err)
		}

		if utils.GlobalDebugFlag {
			log.Info().
				Str("package", "stats").
				Str("path", filePath).
				Int("lines", stats.Lines).
				Int("words", stats.Words).
				Int("chars", stats.Characters).
				Int64("bytes", stats.Bytes).
				Int("tokens", stats.EstimatedTokens).
				Msg("file stats")
		} else if utils.GlobalForAIFlag {
			utils.PrintGeneric(fmt.Sprintf("[INFO] lines=%d words=%d chars=%d bytes=%d tokens=%d",
				stats.Lines, stats.Words, stats.Characters, stats.Bytes, stats.EstimatedTokens))
		} else {
			utils.PrintInfo(fmt.Sprintf("File: %s", filePath))
			utils.PrintGeneric(fmt.Sprintf("  Lines:       %d", stats.Lines))
			utils.PrintGeneric(fmt.Sprintf("  Words:       %d", stats.Words))
			utils.PrintGeneric(fmt.Sprintf("  Characters:  %d", stats.Characters))
			utils.PrintGeneric(fmt.Sprintf("  Size:        %s (%d bytes)", stats.HumanSize, stats.Bytes))
			utils.PrintGeneric(fmt.Sprintf("  Est. Tokens: ~%d", stats.EstimatedTokens))
		}
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
