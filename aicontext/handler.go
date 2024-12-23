package aicontext

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"

	log "github.com/rs/zerolog/log"
)

var urlRegex = map[string]string{
	"gh":  "https://github.com/.+",
	"yt":  "https://youtu.be/.+",
	"yt1": "https://www.youtube.com/watch\\?v=.+",
	"yt2": "https://youtube.com/watch\\?v=.+",
	"dir": "^\\.?\\./.+|^/.*",
}

type result struct {
	url string
	err error
}

type input struct {
	url     string
	urlType string
}

// getOutFileName remove all special characters, path, and URL artifacts to produce a unique filename
func getOutFileName(input string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	reReplace := regexp.MustCompile(`https?_|www_|youtube_com_|github_com_|watch_v_|__`)
	res := strings.ToLower(re.ReplaceAllString(input, "_"))
	res = reReplace.ReplaceAllString(res, "")
	res = strings.TrimPrefix(res, "_")
	// return res + "-" + time.Now().Format("150405") + ".md"
	return res + ".md"
}

// handlerWorker processes the input and sends the result to the result channel
func handlerWorker(toProcess input, resultChan chan result, ignoreList []string) {
	switch toProcess.urlType {
	case "gh":
		output := path.Join("context", "gh-"+getOutFileName(toProcess.url))
		processor := NewProcessor(ProcessorConfig{
			OutputPath:        output,
			AdditionalIgnores: ignoreList,
		})
		err := processor.ProcessGitHubURL(toProcess.url)
		if err != nil {
			log.Error().Err(err).Str("url", toProcess.url).Msg("failed to process GitHub URL")
			resultChan <- result{url: toProcess.url, err: err}
			return
		}
		log.Debug().Str("url", toProcess.url).Str("output", output).Msg("successfully processed GtHub URL")
		resultChan <- result{url: toProcess.url, err: nil}
	case "dir":
		output := path.Join("context", "dir-"+getOutFileName(toProcess.url))
		processor := NewProcessor(ProcessorConfig{
			OutputPath:        output,
			AdditionalIgnores: ignoreList,
		})
		err := processor.ProcessDirectory(toProcess.url)
		if err != nil {
			log.Error().Err(err).Str("url", toProcess.url).Msg("failed to process directory")
			resultChan <- result{url: toProcess.url, err: err}
			return
		}
		log.Debug().Str("url", toProcess.url).Str("output", output).Msg("successfully processed directory")
		resultChan <- result{url: toProcess.url, err: nil}
	case "yt":
		segments, err := DownloadTranscript(toProcess.url)
		output := path.Join("context", "yt-"+getOutFileName(toProcess.url))
		if err != nil {
			log.Error().Err(err).Str("url", toProcess.url).Msg("failed to get transcript")
			resultChan <- result{url: toProcess.url, err: err}
			return
		}
		var content strings.Builder
		content.WriteString("# Video Transcript\n\n")
		for _, segment := range segments {
			content.WriteString(fmt.Sprintf("[%s] %s\n\n", segment.StartTime, segment.Text))
		}
		if err := os.WriteFile(output, []byte(content.String()), 0644); err != nil {
			log.Error().Err(err).Str("url", toProcess.url).Msg("failed to write transcript")
			resultChan <- result{url: toProcess.url, err: err}
			return
		}
		log.Debug().Str("url", toProcess.url).Str("output", output).Msg("successfully generated transcript")
		resultChan <- result{url: toProcess.url, err: nil}
	case "generic":
		// TODO: Implement generic URL processing
		return
	}
}

// Handler processes tasks
func Handler(urls []string, ignoreList []string, threads int) {
	// Create output directories if they doesn't exist
	if err := os.MkdirAll("context", 0755); err != nil {
		log.Fatal().Err(err).Msg("failed to create 'context' directory")
	}
	if err := os.MkdirAll(path.Join("context", "images"), 0755); err != nil {
		log.Fatal().Err(err).Msg("failed to create images directory in 'context'")
	}

	inputURLChan := make(chan input)
	resultChan := make(chan result)
	var outerWG sync.WaitGroup

	// Start handler goroutines
	outerWG.Add(1)
	go func() {
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, threads)
		defer outerWG.Done()
		defer close(resultChan)
		defer close(semaphore)
		defer wg.Wait()
		for toProcess := range inputURLChan {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(toProcess input) {
				defer wg.Done()
				defer func() { <-semaphore }()
				handlerWorker(toProcess, resultChan, ignoreList)
			}(toProcess)
		}
	}()

	// Start result collector
	outerWG.Add(1)
	errors := make([]error, 0)
	go func() {
		defer outerWG.Done()
		for result := range resultChan {
			if result.err != nil {
				errors = append(errors, fmt.Errorf("failed to process %s: %v", result.url, result.err))
			}
		}
	}()

	// Process URLs
	outerWG.Add(1)
	go func(urls []string) {
		defer outerWG.Done()
		defer close(inputURLChan)
		for _, u := range urls {
			matched := false
			for ut, reg := range urlRegex {
				if isMatch, _ := regexp.MatchString(reg, u); isMatch {
					fmt.Printf("URL: %s\nType: %s\n", u, ut)
					if ut == "yt" || ut == "yt1" || ut == "yt2" {
						inputURLChan <- input{url: u, urlType: "yt"}
					} else {
						inputURLChan <- input{url: u, urlType: ut}
					}
					matched = true
					break
				}
			}
			if !matched && strings.HasPrefix(u, "http") {
				inputURLChan <- input{url: u, urlType: "generic"}
				matched = true
			}
			if !matched {
				log.Error().Str("url", u).Msg("invalid URL format")
			} else {
				log.Debug().Str("url", u).Msg("processing input")
			}
		}
	}(urls)

	fmt.Println("Here")
	outerWG.Wait()
	fmt.Println("Here1")

	// Remove images directory if it's empty
	if files, err := os.ReadDir(path.Join("context", "images")); err != nil || len(files) == 0 {
		if err := os.RemoveAll(path.Join("context", "images")); err != nil {
			log.Error().Err(err).Msg("failed to remove 'images' directory")
		}
	}

	// Error Collation
	if len(errors) > 0 {
		for _, err := range errors {
			log.Error().Err(err).Msg("processing error")
		}
		log.Fatal().Msg("one or more URLs failed to process")
	}
}
