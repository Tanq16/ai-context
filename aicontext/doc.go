// Package aicontext provides functionality to generate AI-friendly context from various sources.
// It supports processing of local directories, GitHub repositories, and YouTube video transcripts,
// outputting the content in a markdown format optimized for AI interactions.
//
// The package offers three main functionalities:
//   - Local Directory Processing: Analyzes and creates structured representation of local codebases
//   - GitHub Repository Processing: Clones and processes public GitHub repositories
//   - YouTube Transcript Processing: Downloads and formats video transcripts
//
// Usage:
//
//	processor := aicontext.NewProcessor(aicontext.ProcessorConfig{
//	    OutputPath: "output.md",
//	})
//
//	// Process a directory
//	err := processor.ProcessDirectory("/path/to/dir")
//
//	// Or process a GitHub URL
//	err := processor.ProcessGitHubURL("https://github.com/user/repo")
//
//	// Or download a transcript
//	segments, err := aicontext.DownloadTranscript("https://youtube.com/watch?v=...")
package aicontext
