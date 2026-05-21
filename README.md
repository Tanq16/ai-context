<div align="center">
  <img src=".github/assets/logo.png" alt="AI Context Logo" width="200">
  <h1>AI Context</h1>

  <a href="https://github.com/tanq16/ai-context/actions/workflows/release.yaml"><img alt="Build Workflow" src="https://github.com/tanq16/ai-context/actions/workflows/release.yaml/badge.svg"></a>&nbsp;<a href="https://github.com/tanq16/ai-context/releases"><img alt="GitHub Release" src="https://img.shields.io/github/v/release/tanq16/ai-context"></a><br><br>
  <a href="#capabilities">Capabilities</a> &bull; <a href="#installation">Installation</a> &bull; <a href="#usage">Usage</a> &bull; <a href="#tips-and-notes">Tips & Notes</a>
</div>

---

Generate AI-friendly markdown files from GitHub repos, local code, YouTube videos, or webpages using a multi-arch, multi-OS CLI tool to make your interactions with LLMs (like ChatGPT, Claude, etc.) easy.

## Capabilities

| Category | Commands | Description |
|----------|----------|-------------|
| Processing | `ai-context [url/path]` | Process local directories, GitHub repos, YouTube videos, or webpages |
| Batch | `ai-context -f [file]` | Process multiple sources concurrently from a list file |

## Installation

### Binary

Download from [releases](https://github.com/tanq16/ai-context/releases):

```bash
# Linux/macOS
ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH=amd64; [ "$ARCH" = "aarch64" ] && ARCH=arm64
curl -sL https://github.com/tanq16/ai-context/releases/latest/download/ai-context-$(uname -s | tr '[:upper:]' '[:lower:]')-$ARCH -o ai-context
chmod +x ai-context
sudo mv ai-context /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/tanq16/ai-context.git
cd ai-context
make build
```

## Usage

### Processing

Generate context from a single source.

```bash
# Process a single path (local directory) with additional exclude patterns
ai-context /path/to/directory  -e "tests,docs,*doc.*"

# Only include specific file types with max size limit
ai-context /path/to/directory -i "*.go,*.md" -s 5242880

# Process one URL (GitHub repo or YouTube Video or Webpage URL)
ai-context https://www.youtube.com/watch?v=video_id

# Process private GitHub repository
GH_TOKEN=$(cat /secrets/GH.PAT) ai-context https://github.com/ORG/REPO
```

**Flags:**
- `--include, -i` - Include files matching globs (e.g., '*.go,*.md')
- `--exclude, -e` - Exclude files matching globs (e.g., 'tests,docs')
- `--max-size, -s` - Maximum file size in bytes to include (default 10MB)
- `--debug` - Enable debug logging
- `--for-ai` - AI-friendly output (plain text, piped input)

### Batch Processing

Generate context from multiple sources listed in a file.

```bash
# Make a list of paths
cat << EOF > listfile
../notif
/working/cybernest
https://github.com/assetnote/h2csmuggler
https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html
EOF

# Process URL list concurrently
ai-context -f listfile
```

**Flags:**
- `--file, -f` - File with list of URLs to process
- `--threads, -t` - Number of threads to use for processing (default: 10)

## Tips and Notes

- For directory path (in URL or listfile mode), the path should either start with `/` (absolute) or with `./` or `../` (relative). For current directory, always use `./` for correct regex matching.
- Do a `head -n 200 context/FILE.md` (or 500 lines) to view the content tree of the processed code base or directory to see what's been included. Then refine your `-e` flag arguments to exclude additional patterns.
- The `--for-ai` flag produces plain text without ANSI colors, which is easier for AI agents to parse.

## Scenarios

### The "Learn & Implement" Scenario
`ai-context -f listfile`
*Where `listfile` contains your local `internal/` directory and a Medium article URL explaining an architectural pattern.*
**Use Case:** The user wants an LLM to refactor their existing `internal/` packages using a specific pattern described in a Medium article. AI-Context will output markdown of both the codebase and the article content, ready to pipe to your LLM prompt.

### The "Bug Hunt" Scenario
`ai-context https://github.com/org/repo -i "pkg/math/*.go"`
**Use Case:** Pulls down only the specific logic paths from a GitHub repository to save token limits when asking an LLM to find a bug.

## Acknowledgments

This project takes inspiration from, uses, or references:

- [repomix](https://github.com/yamadashy/repomix): inspiration for turning code into context
- [innertube](https://github.com/tombulled/innertube): inspiration for code to get transcript from YouTube video
- [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown/v2): used to convert HTML to MD
- [go-git](https://github.com/go-git/go-git/tree/main): git operations in Go
