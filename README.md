<div align="center">

<img src=".github/assets/logo.png" alt="AI Context Logo" width="225"/>

<h1>AI Context</h1>

[![Release Build](https://github.com/tanq16/ai-context/actions/workflows/build-release.yml/badge.svg)](https://github.com/tanq16/ai-context/actions/workflows/build-release.yml)
[![GitHub Release](https://img.shields.io/github/v/release/tanq16/ai-context)](https://github.com/Tanq16/ai-context/releases/latest)

[![Go Report Card](https://goreportcard.com/badge/github.com/tanq16/ai-context)](https://goreportcard.com/report/github.com/tanq16/ai-context)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/tanq16/ai-context?status.svg)](https://godoc.org/github.com/tanq16/ai-context)

Generate AI-friendly markdown context files from repositories or videos or local source code.

</div>

---

A command-line tool designed to produce context files from various sources, to make interactions with LLM apps (like ChatGPT, Claude, etc.) easy. It can process multiple sources to output the context in a markdown format optimized for use by AI models.

## Features

- **Local Directory Processing**
    - this is mainly for locally available code bases (directories or already cloned git repos)
    - the context file includes directory structure and all file contents within context
- **GitHub Repository Processing**
    - this clones and processes provided github link and does the same as *Local Directory Processing*
    - it temporarily clones the repository, so no need for cleanup
- **YouTube Transcript Processing**
    - this downloads transcripts for given YouTube video link and preserves time segments
- **WebPage Processing**
    - this converts the HTML of a webpage to markdown
    - it also downloads all images from the page and stores them locally with UUID names
    - the images in markdown will refer to local paths

## Installation

- **Binary**
    - Download the latest release for your platform and OS from the [releases page](https://github.com/tanq16/ai-context/releases)
    - Binaries are build via GitHub actions for MacOS, Linux, and Windows for both AMD64 (x86_64) and ARM64 (incl. Apple Silicon) architectures
    - Use this to download specific version if needed
- **Go Install**
    - Run the following command (requires `Go v1.22+`):
    ```bash
    go install github.com/tanq16/ai-context@latest
    ```
    - For specific versions, prefer the binaries or local build process as I haven't implemented Go binary versioning for the project
- **Local Build**
    - To build locally, do the following:
    ```bash
    # Clone
    git clone https://github.com/tanq16/ai-context.git && \
    cd ai-context
    ```
    ```bash
    # Build
    go build .
    ```

## Usage

```bash
# Process a single path (local directory) with additional ignore patterns
ai-context -u /path/to/directory  -i "tests,docs,*doc.*"

# Process one URL (GitHub repo or YouTube Video or Webpage URL)
ai-context -u https://www.youtube.com/watch?v=video_id

# Make a list of paths
cat << EOF > listfile
../notif
/working/cybernest
https://github.com/assetnote/h2csmuggler
https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html
EOF

# Process everything concurrently
ai-context -l listfile
```

> [!WARNING]
> For directory path (in URL or listfile mode), the path should either start with `/` (absolute) or with `./` or `../` (relative). For current directory, always use `./` for correct regex matching.

### Output

- The tool creates a local folder called `context` and puts all gathered context into `.md` files in that folder.
- The filenames have the following syntax: `TYPE-PATHNAME.md` (example, `gh-ffuf_ffuf.md`).
- Every single path in the `listfile` mode will result in a new context file.
- All images are named as a UUID and are downloaded to `context/images` directory.

### Command Line Options

- `-u, --url`: provide a path (GitHub repo, YouTube video, WebPage link, or relative/absolute directory path) to process
- `-f, --file`: provide a file with a list of paths (URLs or directory paths) to process
- `-i, --ignore`: add additional patterns to ignore during processing (comma-separated)
- `-t, --threads`: (optional) number of workers for file processing; default value is 5
- `--debug`: verbose logging (helpful if something isn't working as expected)

> [!TIP]
> - Do a `head -n 200 context/FILE.md` (or 500 lines) to view the content tree of the processed code base or directory to see what's been included. Then refine your `-i` flag arguments to ignore additional patterns.
> - When processing a large number of items, it can look stalled due to thread limits and image download times; use `--debug` to enable verbose logs to know it's running.

## Default Ignores

The tool automatically ignores common files and directories that typically don't add value to the context, including:

- Version control files (.git, .gitignore)
- Dependencies (node_modules, vendor)
- Compiled files (*.exe, *.dll)
- Media files (images, videos, audio)
- Documentation files
- Lock files (package-lock.json, yarn.lock)
- Build artifacts and caches

For a full list, see `aicontext/ignores.go`.

## Contributing

Feel free to submit issues or PRs.

## Acknowledgments

This project takes inspiration from and references:

- [repomix](https://github.com/yamadashy/repomix): inspiration for turning code into context
- [innertube](https://github.com/tombulled/innertube): inspiration for code to get transcript from YouTube video
- [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown/v2): used to convert HTML to MD
