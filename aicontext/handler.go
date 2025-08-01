package aicontext

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/tanq16/ai-context/utils"
)

var URLRegex = map[string]string{
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

func GetOutFileName(input string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	reReplace := regexp.MustCompile(`https?_|www_|youtube_com_|github_com_|watch_v_|__|com_`)
	res := strings.ToLower(re.ReplaceAllString(input, "_"))
	res = reReplace.ReplaceAllString(res, "")
	res = strings.Trim(res, "_")
	return res + ".md"
}

func cleanURL(rawURL string) (string, error) {
	if after, ok := strings.CutPrefix(rawURL, "github/"); ok {
		rawURL = "https://github.com/" + after
	}
	if match, _ := regexp.MatchString(`^\.?\.?\/.*`, rawURL); match {
		return rawURL, nil
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}
	// preserve video id for yt
	if strings.Contains(parsedURL.Host, "youtube.com") {
		videoID := parsedURL.Query().Get("v")
		if videoID != "" {
			query := url.Values{}
			query.Set("v", videoID)
			parsedURL.RawQuery = query.Encode()
		} else {
			parsedURL.RawQuery = ""
		}
	} else {
		parsedURL.RawQuery = ""
	}
	parsedURL.Fragment = ""
	return parsedURL.String(), nil
}

func handlerWorker(toProcess input, resultChan chan result, ignoreList []string, consoleMgr *utils.Console) {
	switch toProcess.urlType {
	case "gh":
		output := path.Join("context", "gh-"+GetOutFileName(toProcess.url))
		codeProcessor := NewProcessor(ProcessorConfig{
			OutputPath:        output,
			AdditionalIgnores: ignoreList,
		})
		consoleMgr.Log("discovered as type gh; starting collection", false)
		err := codeProcessor.ProcessGitHubURL(toProcess.url)
		if err != nil {
			resultChan <- result{url: toProcess.url, err: err}
			return
		}
		resultChan <- result{url: toProcess.url, err: nil}
	case "dir":
		output := path.Join("context", "dir-"+GetOutFileName(toProcess.url))
		codeProcessor := NewProcessor(ProcessorConfig{
			OutputPath:        output,
			AdditionalIgnores: ignoreList,
		})
		err := codeProcessor.ProcessDirectory(toProcess.url)
		if err != nil {
			resultChan <- result{url: toProcess.url, err: err}
			return
		}
		resultChan <- result{url: toProcess.url, err: nil}
	case "yt":
		segments, err := DownloadTranscript(toProcess.url)
		output := path.Join("context", "yt-"+GetOutFileName(toProcess.url))
		if err != nil {
			resultChan <- result{url: toProcess.url, err: err}
			return
		}
		var content strings.Builder
		content.WriteString("# Video Transcript\n\n")
		for _, segment := range segments {
			content.WriteString(fmt.Sprintf("[%s] %s\n\n", segment.StartTime, segment.Text))
		}
		if err := os.WriteFile(output, []byte(content.String()), 0644); err != nil {
			resultChan <- result{url: toProcess.url, err: err}
			return
		}
		resultChan <- result{url: toProcess.url, err: nil}
	case "generic":
		output := path.Join("context", "web-"+GetOutFileName(toProcess.url))
		err := ProcessWebContent(toProcess.url, output)
		if err != nil {
			resultChan <- result{url: toProcess.url, err: err}
			return
		}
		resultChan <- result{url: toProcess.url, err: nil}
	}
}

func Handler(urls []string, ignoreList []string, threads int, detailLog bool) {
	// Pre-filter URLs
	var cleanedUrls []string
	for _, u := range urls {
		cleaned, err := cleanURL(u)
		if err != nil {
			continue
		}
		cleanedUrls = append(cleanedUrls, cleaned)
	}
	urls = cleanedUrls

	outputMgr := utils.NewManager()
	consoleMgr := utils.NewConsole()
	outputMgr.StartDisplay()
	consoleMgr.Start()
	defer outputMgr.StopDisplay()
	defer consoleMgr.Stop()
	if detailLog {
		outputMgr.Disable()
	} else {
		consoleMgr.Disable()
	}
	outputMgr.SetMessage("Creating file structure")
	consoleMgr.Log("Creating file structure", false)

	// Create output directories if they doesn't exist
	if err := os.MkdirAll("context", 0755); err != nil {
		outputMgr.Complete("", fmt.Errorf("couldn't create context directory: %w", err))
		return
	}
	if err := os.MkdirAll(path.Join("context", "images"), 0755); err != nil {
		outputMgr.Complete("", fmt.Errorf("couldn't create images directory: %w", err))
		return
	}
	totUrls := len(urls)
	pluralS := "s"
	if totUrls == 1 {
		pluralS = ""
	}
	outputMgr.SetMessage(fmt.Sprintf("Gathering %d context file%s", totUrls, pluralS))
	consoleMgr.Log(fmt.Sprintf("Gathering %d context file%s", totUrls, pluralS), false)

	inputURLChan := make(chan input)
	resultChan := make(chan result)
	progressChan := make(chan int64)
	var outerWG sync.WaitGroup

	// Start progress reporter
	outerWG.Add(1)
	go func(totUrls int64) {
		defer outerWG.Done()
		totCompleted := int64(0)
		for completed := range progressChan {
			totCompleted += completed
			outputMgr.ReportProgress(totCompleted, totUrls, fmt.Sprintf("%d finished", totCompleted))
			consoleMgr.Log("finished", false, "total completed", totCompleted, "total URLs", totUrls)
		}
	}(int64(totUrls))

	// Start handler goroutines
	outerWG.Add(1)
	go func(progCh chan<- int64) {
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, threads)
		defer outerWG.Done()
		defer close(resultChan)
		defer close(progressChan)
		defer close(semaphore)
		defer wg.Wait()
		for toProcess := range inputURLChan {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(toProcess input) {
				defer wg.Done()
				defer func() { <-semaphore }()
				handlerWorker(toProcess, resultChan, ignoreList, consoleMgr)
				progressChan <- 1
			}(toProcess)
		}
	}(progressChan)

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
			for ut, reg := range URLRegex {
				if isMatch, _ := regexp.MatchString(reg, u); isMatch {
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
				// log.Error().Str("url", u).Msg("invalid URL format")
			} else {
				// log.Debug().Str("url", u).Msg("processing input")
			}
		}
	}(urls)

	outerWG.Wait()

	// Error Collation
	var errMsg string
	if len(errors) > 0 {
		for _, err := range errors {
			errMsg += (err.Error() + "\n")
		}
	}
	// Remove images directory if it's empty
	files, err := os.ReadDir(path.Join("context", "images"))
	if err != nil {
		outputMgr.SetMessage("Operations Completed, error in cleanup")
		errMsg += (fmt.Errorf("couldn't read images directory: %w", err)).Error() + "\n"
	} else if len(files) == 0 {
		err := os.RemoveAll(path.Join("context", "images"))
		if err != nil {
			outputMgr.SetMessage("Operations Completed, error in cleanup")
			errMsg += (fmt.Errorf("failed to delete empty images directory: %w", err)).Error() + "\n"
		}
	}
	// Complete output manager
	if errMsg != "" {
		outputMgr.Complete("", fmt.Errorf("%s", errMsg))
		consoleMgr.Log("Completed all operations", true, "errors", errMsg)
	} else {
		outputMgr.Complete("All operations completed", nil)
		consoleMgr.Log("Completed all operations", false)
	}
}
