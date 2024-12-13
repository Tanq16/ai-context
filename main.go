// AI Context is a command-line tool for generating structured context from various sources.
// It helps in processing local directories, GitHub repositories, and YouTube video transcripts
// into a markdown format that's optimized for AI language models.
//
// Key Features:
//
// Local Directory Processing:
//   - Analyzes local codebases
//   - Creates structured file hierarchy
//   - Includes file contents with syntax highlighting
//
// GitHub Repository Processing:
//   - Clones and processes public repositories
//   - Generates comprehensive project overviews
//   - Maintains repository structure
//
// YouTube Transcript Processing:
//   - Downloads video transcripts
//   - Formats with timestamps
//   - Creates AI-friendly text output
//
// Example Usage:
//
//	# Process a local directory
//	ai-context -d /path/to/directory -o output.md
//
//	# Process a GitHub repository
//	ai-context -u https://github.com/username/repo -o output.md
//
//	# Download a YouTube transcript
//	ai-context -v https://www.youtube.com/watch?v=video_id -o output.md
package main

import "github.com/tanq16/ai-context/cmd"

func main() {
	cmd.Execute()
}
