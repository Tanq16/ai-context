package aicontext

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	log "github.com/rs/zerolog/log"
)

var urlRegex = map[string]string{
	"gh":        "https://github.com/.+/.+",
	"yt":        "https://youtu.be/.+",
	"yt1":       "https://www.youtube.com/watch\\?v=.+",
	"yt2":       "https://youtube.com/watch\\?v=.+",
	"generic":   "https?://.+",
	"directory": "\\.?/?.+",
}

var filename = "context.md"
var imgfolder = "images"

type processResult struct {
	url string
	err error
}

// DirectoryHandler processes a single directory
func DirectoryHandler(directory, output string, ignoreList []string) {
	processor := NewProcessor(ProcessorConfig{
		OutputPath:        output,
		AdditionalIgnores: ignoreList,
	})
	err := processor.ProcessDirectory(directory)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to process source")
	}
	log.Info().Str("output", output).Msg("successfully generated context")
}

// GitHubHandler processes GitHub URLs (input <- channel, output -> channel)
func GitHubHandler(urlChan chan string, resultChan chan processResult, output string, ignoreList []string) {
	for url := range urlChan {
		processor := NewProcessor(ProcessorConfig{
			OutputPath:        output,
			AdditionalIgnores: ignoreList,
		})
		err := processor.ProcessGitHubURL(url)
		if err != nil {
			log.Error().Err(err).Str("url", url).Msg("failed to process GitHub repository")
			resultChan <- processResult{url: url, err: err}
			continue
		}
		log.Info().Str("url", url).Str("output", output).Msg("successfully processed GitHub repository")
		resultChan <- processResult{url: url, err: nil}
	}
}

// YouTubeHandler processes YouTube URLs (input <- channel, output -> channel)
func YouTubeHandler(urlChan chan string, resultChan chan processResult, output string) {
	for url := range urlChan {
		segments, err := DownloadTranscript(url)
		if err != nil {
			log.Error().Err(err).Str("url", url).Msg("failed to get transcript")
			resultChan <- processResult{url: url, err: err}
			continue
		}
		var content strings.Builder
		content.WriteString("# Video Transcript\n\n")
		for _, segment := range segments {
			content.WriteString(fmt.Sprintf("[%s] %s\n\n", segment.StartTime, segment.Text))
		}
		if err := os.WriteFile(output, []byte(content.String()), 0644); err != nil {
			log.Error().Err(err).Str("url", url).Msg("failed to write transcript")
			resultChan <- processResult{url: url, err: err}
			continue
		}
		log.Info().Str("url", url).Str("output", output).Msg("successfully generated transcript")
		resultChan <- processResult{url: url, err: nil}
	}
}

// TODO: See if anything else is needed for collecting directory paths
// TODO: Move logic to cmd/ai-context.go
// readURLsFromFile gets all URLs from a file
func readURLsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" {
			urls = append(urls, url)
		}
	}
	return urls, scanner.Err()
}

// Handler processes tasks
func Handler(directory, url, listFile, output string, ignoreList []string) {
	// Handle directory separately as it doesn't need goroutines
	// TODO: Implement directory list with goroutines and absorb next if block into the following block
	if directory != "" {
		if url != "" || listFile != "" {
			log.Fatal().Msg("cannot specify directory with url or list file")
		}
		DirectoryHandler(directory, output, ignoreList)
		return
	}
	if url == "" && listFile == "" {
		log.Fatal().Msg("either url or list of urls must be specified")
	}
	if url != "" && listFile != "" {
		log.Fatal().Msg("cannot specify both url and list file")
	}

	// TODO: Move logic to cmd/ai-context.go
	// Input URL processing
	var urls []string
	var err error
	if listFile != "" {
		urls, err = readURLsFromFile(listFile)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read URLs from file")
		}
	} else {
		urls = []string{url}
	}

	githubURLChan := make(chan string)
	youtubeURLChan := make(chan string)
	resultChan := make(chan processResult)
	var wg sync.WaitGroup

	// Start handler goroutines
	wg.Add(2)
	go func() {
		defer wg.Done()
		GitHubHandler(githubURLChan, resultChan, output, ignoreList)
	}()
	go func() {
		defer wg.Done()
		YouTubeHandler(youtubeURLChan, resultChan, output)
	}()

	// Start result collector
	errors := make([]error, 0)
	go func() {
		for result := range resultChan {
			if result.err != nil {
				errors = append(errors, fmt.Errorf("failed to process %s: %v", result.url, result.err))
			}
		}
	}()

	// Process URLs
	for _, u := range urls {
		matched := false
		// Check YouTube URL Regex
		if isYT, _ := regexp.MatchString(urlRegex["yt"], u); isYT {
			youtubeURLChan <- u
			matched = true
			log.Debug().Str("url", u).Msg("processing YouTube URL")
		} else if isYT, _ := regexp.MatchString(urlRegex["yt1"], u); isYT {
			youtubeURLChan <- u
			matched = true
			log.Debug().Str("url", u).Msg("processing YouTube URL")
		} else if isYT, _ := regexp.MatchString(urlRegex["yt2"], u); isYT {
			youtubeURLChan <- u
			matched = true
			log.Debug().Str("url", u).Msg("processing YouTube URL")
			// Check GitHub URL Regex
		} else if isGitHub, _ := regexp.MatchString(urlRegex["gh"], u); isGitHub {
			githubURLChan <- u
			matched = true
			log.Debug().Str("url", u).Msg("processing GitHub URL")
			// Check Generic URL Regex
		} else if isGeneric, _ := regexp.MatchString(urlRegex["generic"], u); isGeneric {
			// TODO: Implement generic URL processing and uncomment the following line
			// matched = true
			log.Error().Str("url", u).Msg("NOT IMPLEMENTED: generic URL processing")
			errors = append(errors, fmt.Errorf("NOT IMPLEMENTED: generic URL processing: %s", u))
		} else if isDirectory, _ := regexp.MatchString(urlRegex["directory"], u); isDirectory {
			// TODO: Implement concurrent directory processing
			// matched = true
		}
		// TODO: Implement directory list with goroutines and remove the next if block and add to else above
		if !matched {
			log.Error().Str("url", u).Msg("invalid URL format")
			errors = append(errors, fmt.Errorf("invalid URL format: %s", u))
		}
	}

	close(githubURLChan)
	close(youtubeURLChan)
	wg.Wait()
	close(resultChan)

	// Error Collation
	if len(errors) > 0 {
		for _, err := range errors {
			log.Error().Err(err).Msg("processing error")
		}
		log.Fatal().Msg("one or more URLs failed to process")
	}
}
