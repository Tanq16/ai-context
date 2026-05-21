package aicontext

import (
	"path/filepath"
	"strings"
)

type PathFilter struct {
	defaultExcludes []string
	includePatterns []string
	excludePatterns []string
}

func newPathFilter(includePatterns []string, excludePatterns []string) *PathFilter {
	parsedIncludes := make([]string, 0)
	for _, p := range includePatterns {
		if p != "" {
			parsedIncludes = append(parsedIncludes, p)
		}
	}

	parsedExcludes := make([]string, len(excludePatterns))
	for i, pattern := range excludePatterns {
		parsedExcludes[i] = "*/" + pattern
	}

	return &PathFilter{
		defaultExcludes: defaultIgnores,
		includePatterns: parsedIncludes,
		excludePatterns: parsedExcludes,
	}
}

func (pf *PathFilter) shouldInclude(path string, isDir bool) bool {
	for _, pattern := range pf.defaultExcludes {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return false
		}
	}

	for _, pattern := range pf.excludePatterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return false
		}
		if matched, _ := filepath.Match(strings.TrimPrefix(pattern, "*/"), filepath.Base(path)); matched {
			return false
		}
	}

	if len(pf.includePatterns) > 0 {
		if isDir {
			return true
		}
		for _, pattern := range pf.includePatterns {
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				return true
			}
			if matched, _ := filepath.Match(pattern, path); matched {
				return true
			}
		}
		return false
	}

	return true
}

func isBinary(content []byte) bool {
	nullCount := 0
	nonPrintable := 0
	checkSize := min(len(content), 512)
	for i := range checkSize {
		if content[i] == 0 {
			nullCount++
		} else if content[i] < 32 && content[i] != '\n' && content[i] != '\r' && content[i] != '\t' {
			nonPrintable++
		}
	}
	return nullCount > 0 || float64(nonPrintable)/float64(checkSize) > 0.3
}

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
