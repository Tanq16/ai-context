<div align="center">

<img src=".github/assets/logo.png" alt="AI Context Logo" width="200"/>

<h1>AI Context</h1>

[![Release Build](https://github.com/tanq16/ai-context/actions/workflows/build-release.yml/badge.svg)](https://github.com/tanq16/ai-context/actions/workflows/build-release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/tanq16/ai-context)](https://goreportcard.com/report/github.com/tanq16/ai-context)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/tanq16/ai-context?status.svg)](https://godoc.org/github.com/tanq16/ai-context)

Generate AI-friendly context from repositories or videos.

</div>

A command-line tool designed to produce a context file from various sources, to make interactions with LLM apps (like ChatGPT, Claude, etc.) easy. Currently, it processes local directories (for code), GitHub repositories (for code), and YouTube videos (for transcript), outputting the content in a markdown format optimized for AI models.

## Features

* **Local Directory Processing**: mainly for locally available code bases (directories or git repos); includes directory structure and file contents within context
* **GitHub Repository Processing**: clones and processes provided github link and does the same as *Local Directory Processing*
* **YouTube Transcript Processing**: downloads transcripts for given YouTube video link and preserves time segments

## Installation

You can either download built binaries directly or build locally.

To download the latest release for your platform and OS, visit the [releases page](https://github.com/tanq16/ai-context/releases). Binaries are available for Windows, MacOS, and Linux for both AMD64 (x86_64) and ARM64 (incl. Apple Silicon) architectures.

To build locally, dothe following:

```bash
# Clone
git clone https://github.com/tanq16/ai-context.git
cd ai-context

# Build
go build .
```

## Usage

```bash
# Process a local directory (default output file ai-context.md)
ai-context -d /path/to/directory

# Process a GitHub repository (manually specified output file)
ai-context -u https://github.com/username/repo -o output.md

# Download a YouTube video transcript
ai-context -v https://www.youtube.com/watch?v=video_id

# Ignore specific patterns while processing
ai-context -d ../massive-codebase -i "tests,docs"
```

### Command Line Options

* `-d, --directory`: provide a local directory (can be relative) to process
* `-u, --url`: provide a GitHub repository URL to process
* `-v, --video`: provide a YouTube video URL to process
* `-o, --output`: set the output file path (default: ai-context.md)
* `-i, --ignore`: add additional patterns to ignore during processing (comma-separated)
* `--debug`: verbose logging

> [!TIP]
> Do a `head -n 200 ai-context.md` (or 500 lines) to view the content tree of the processed code base to see what's been included. Then you can refine your `-i` flag arguments to ignore additional patterns.

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

## Development

AI Context is written in Go and has these dependencies:

- `go-git` for repository cloning
- `cobra` for command-line interface
- `zerolog` for structured logging

## Contributing

Feel free to submit issues and pull requests.

## Acknowledgments

This project takes inspiration from and references:

- repomix (https://github.com/yamadashy/repomix)
- innertube (https://github.com/tombulled/innertube)
