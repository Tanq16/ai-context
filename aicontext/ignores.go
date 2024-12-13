package aicontext

import (
	"path/filepath"
	"strings"
)

type IgnorePatterns struct {
	defaultPatterns []string
	customPatterns  []string
}

func newIgnorePatterns(additionalPatterns []string) *IgnorePatterns {
	customPatterns := make([]string, len(additionalPatterns))
	for i, pattern := range additionalPatterns {
		customPatterns[i] = "*/" + pattern
	}
	return &IgnorePatterns{
		defaultPatterns: defaultIgnores,
		customPatterns:  customPatterns,
	}
}

func (ip *IgnorePatterns) shouldIgnore(path string) bool {
	// Check default patterns
	for _, pattern := range ip.defaultPatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
	}
	// Check custom patterns
	for _, pattern := range ip.customPatterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
		// Also check if the base name matches
		if matched, _ := filepath.Match(strings.TrimPrefix(pattern, "*/"), filepath.Base(path)); matched {
			return true
		}
	}
	return false
}

// Check if the content is binary
// funny heuristic, but it works against a very limited sample set
func isBinary(content []byte) bool {
	nullCount := 0
	nonPrintable := 0
	checkSize := min(len(content), 512)
	for i := 0; i < checkSize; i++ {
		if content[i] == 0 {
			nullCount++
		} else if content[i] < 32 && content[i] != '\n' && content[i] != '\r' && content[i] != '\t' {
			nonPrintable++
		}
	}
	return nullCount > 0 || float64(nonPrintable)/float64(checkSize) > 0.3
}

// Default ignored patterns
var defaultIgnores = []string{
	".git",
	".gitignore",
	".gitmodules",
	".gitattributes",
	"node_modules",
	"*.gz",
	"*.bz2",
	"*.zip",
	"*.tar",
	"*.tgz",
	"*.xz",
	"*.rar",
	"*.7z",
	"vendor",
	"*.exe",
	"*.dll",
	"*.so",
	"*.dylib",
	"*.tar.gz",
	"*.jpg",
	"*.jpeg",
	"*.png",
	"*.gif",
	"*.ico",
	"*.tif",
	"*.tiff",
	"*.bmp",
	"*.svg",
	"*.webp",
	"*.mpg",
	"*.mp2",
	"*.mpeg",
	"*.ogg",
	"*.mp3",
	"*.mp4",
	"*.avi",
	"*.pdf",
	"*.doc",
	"*.docx",
	"*.class",
	"*.pyc",
	"*.o",
	"poetry.lock",
	"yarn.lock",
	"package-lock.json",
	"composer.lock",
	"pytest_cache",
	"pypy_cache",
	"pyproject.toml",
	"poetry.toml",
	"bin",
	"LICENSE",
	"AUTHORS",
	"CONTRIBUTORS",
	"OWNERS",
	"CONTRIBUTING.md",
	"CHANGELOG.md",
	"go.sum",
	"go.mod",
	".obsidian",
	".vscode",
	".idea",
	".DS_Store",
	"*.apk",
	"*.ipa",
	"*.dmg",
	"*.iso",
	"*.msi",
	"*.deb",
	"*.rpm",
	"*.jar",
	"*.war",
	"*.ttf",
	"*.woff",
	"*.woff2",
	"*.otf",
}
