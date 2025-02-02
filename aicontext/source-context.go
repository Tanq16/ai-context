package aicontext

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	log "github.com/rs/zerolog/log"
)

// FileEntry holds the file path, content, and language
type FileEntry struct {
	Path     string
	Content  string
	Language string
}

// Output holds the output data
type Output struct {
	GenerationDate string
	FileCount      int
	TotalSize      int64
	DirectoryTree  string
	Files          []FileEntry
}

// ProcessorConfig holds the configuration for the processor
type ProcessorConfig struct {
	OutputPath        string
	AdditionalIgnores []string
}

// Processor processes the source code context
type Processor struct {
	config         ProcessorConfig
	ignorePatterns *IgnorePatterns
}

// Markdown template for the output
const markdownTemplate = `# Source Code Context

Generated on: {{.GenerationDate}}

## Repository Overview
- Total Files: {{.FileCount}}
- Total Size: {{.TotalSize}} bytes

## Directory Structure
` + "```" + `
{{.DirectoryTree}}
` + "```" + `

## File Contents

{{range .Files}}
### File: {{.Path}}

` + "```{{.Language}}" + `
{{.Content}}
` + "```" + `




{{end}}`

// NewProcessor creates a new Processor instance
func NewProcessor(config ProcessorConfig) *Processor {
	return &Processor{
		config:         config,
		ignorePatterns: newIgnorePatterns(config.AdditionalIgnores),
	}
}

// ProcessDirectory processes the directory
func (p *Processor) ProcessDirectory(path string) error {
	output, err := p.processDirectory(path)
	if err != nil {
		return fmt.Errorf("failed to process directory: %w", err)
	}
	return p.writeOutput(output)
}

// ProcessGitHubURL processes the GitHub URL
func (p *Processor) ProcessGitHubURL(url string) error {
	tempDir, err := os.MkdirTemp("", "aicontext-clone-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	cloneOpts := &git.CloneOptions{
		URL:      url,
		Progress: nil,
		Depth:    1,
	}
	if token := os.Getenv("GH_TOKEN"); token != "" {
		log.Debug().Msg("using GitHub token for authentication")
		cloneOpts.Auth = &http.BasicAuth{
			Username: "git", // can be anything but not empty
			Password: token,
		}
	}
	_, err = git.PlainClone(tempDir, false, cloneOpts)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	log.Debug().Str("path", tempDir).Msg("cloned repository")
	return p.ProcessDirectory(tempDir)
}

// processGitRepository processes the Git repository
func (p *Processor) processDirectory(root string) (*Output, error) {
	output := &Output{
		GenerationDate: time.Now().Format(time.RFC3339),
		Files:          make([]FileEntry, 0),
	}
	log.Debug().Str("path", root).Msg("processing directory")
	var totalSize int64
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if p.ignorePatterns.shouldIgnore(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if isBinary(content) {
			return nil
		}
		totalSize += info.Size()
		output.Files = append(output.Files, FileEntry{
			Path:     relPath,
			Content:  string(content),
			Language: detectLanguage(relPath),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	output.FileCount = len(output.Files)
	output.TotalSize = totalSize
	output.DirectoryTree = p.generateDirectoryTree(root)
	log.Debug().Int("fileCount", output.FileCount).Int64("totalSize", output.TotalSize).Msg("processed directory")
	return output, nil
}

// writeOutput writes the output to the file
func (p *Processor) writeOutput(output *Output) error {
	tmpl, err := template.New("markdown").Parse(markdownTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	file, err := os.Create(p.config.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()
	if err := tmpl.Execute(file, output); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	log.Debug().Str("path", p.config.OutputPath).Msg("wrote output")
	return nil
}

// detectLanguage detects the language based on the file extension
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".c":
		return "c"
	case ".cpp":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".swift":
		return "swift"
	case ".rs":
		return "rust"
	case ".sh":
		return "bash"
	case ".yml", ".yaml":
		return "yaml"
	case ".json":
		return "json"
	case ".md":
		return "markdown"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".sql":
		return "sql"
	case ".dockerfile":
		return "dockerfile"
	default:
		return ""
	}
}

// generateDirectoryTree generates the directory tree
func (p *Processor) generateDirectoryTree(root string) string {
	var tree strings.Builder
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if p.ignorePatterns.shouldIgnore(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if relPath == "." {
			return nil
		}
		indent := strings.Repeat("  ", strings.Count(relPath, string(filepath.Separator)))
		if info.IsDir() {
			fmt.Fprintf(&tree, "%s%s/\n", indent, info.Name())
		} else {
			fmt.Fprintf(&tree, "%s%s\n", indent, info.Name())
		}
		return nil
	})
	if err != nil {
		return fmt.Sprintf("Error generating tree: %v", err)
	}
	log.Debug().Msg("generated directory tree")
	return tree.String()
}

// helper function to find the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
