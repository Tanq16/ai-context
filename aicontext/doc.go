// Package aicontext provides functionality to generate AI-friendly context from various sources.
// It supports processing of local directories, GitHub repositories, and YouTube video transcripts,
// outputting the content in a markdown format optimized for AI interactions.
//
// The package offers three main functionalities:
//   - Local Directory Processing: Analyzes and creates structured representation of local codebases
//   - GitHub Repository Processing: Clones and processes public GitHub repositories
//   - YouTube Transcript Processing: Downloads and formats video transcripts
//   - Webpage Processing: Downloads and processes webpages including images
//
// Usage:
//
//	processor := aicontext.NewProcessor(aicontext.ProcessorConfig{
//	    OutputPath: "output.md",
//	})
//	os.MkdirAll(path.Join("context", "images"), os.ModePerm) // only for web processing
//
//	// Process a directory
//	err := processor.ProcessDirectory("/path/to/dir")
//
//	// Or process a GitHub URL
//	err := processor.ProcessGitHubURL("https://github.com/user/repo")
//
//	// Or download a transcript
//	segments, err := aicontext.DownloadTranscript("https://youtube.com/watch?v=...")
//
//	// Or process a webpage
//	err := aicontext.ProcessWebContent("https://example.com", "output.md")
package aicontext
