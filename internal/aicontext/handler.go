package aicontext

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

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

func handlerWorker(ctx context.Context, toProcess input, resultChan chan result, includeGlobs []string, excludeGlobs []string, maxSize int64, useJina bool) {
	select {
	case <-ctx.Done():
		resultChan <- result{url: toProcess.url, err: ctx.Err()}
		return
	default:
	}

	switch toProcess.urlType {
	case "gh":
		output := path.Join("context", "gh-"+GetOutFileName(toProcess.url))
		codeProcessor := NewProcessor(ProcessorConfig{
			OutputPath:   output,
			IncludeGlobs: includeGlobs,
			ExcludeGlobs: excludeGlobs,
			MaxSize:      maxSize,
		})
		utils.PrintIndentedRunning(fmt.Sprintf("%s: starting collection", toProcess.url))
		err := codeProcessor.ProcessGitHubURL(toProcess.url)
		if err != nil {
			resultChan <- result{url: toProcess.url, err: err}
			return
		}
		resultChan <- result{url: toProcess.url, err: nil}
	case "dir":
		output := path.Join("context", "dir-"+GetOutFileName(toProcess.url))
		codeProcessor := NewProcessor(ProcessorConfig{
			OutputPath:   output,
			IncludeGlobs: includeGlobs,
			ExcludeGlobs: excludeGlobs,
			MaxSize:      maxSize,
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
		err := ProcessWebContent(toProcess.url, output, useJina)
		if err != nil {
			resultChan <- result{url: toProcess.url, err: err}
			return
		}
		resultChan <- result{url: toProcess.url, err: nil}
	}
}

func Handler(ctx context.Context, urls []string, includeGlobs []string, excludeGlobs []string, maxSize int64, threads int, detailLog bool, useJina bool) {
	var cleanedUrls []string
	for _, u := range urls {
		cleaned, err := cleanURL(u)
		if err != nil {
			continue
		}
		cleanedUrls = append(cleanedUrls, cleaned)
	}
	urls = cleanedUrls

	utils.PrintRunning("Creating file structure")

	if err := os.MkdirAll("context", 0755); err != nil {
		utils.ClearLines(1)
		utils.PrintFatal("couldn't create context directory", err)
	}
	if err := os.MkdirAll(path.Join("context", "images"), 0755); err != nil {
		utils.ClearLines(1)
		utils.PrintFatal("couldn't create images directory", err)
	}
	utils.ClearLines(1)
	totUrls := len(urls)
	if totUrls == 0 {
		utils.PrintSuccess("Completed all operations successfully")
		return
	}
	pluralS := "s"
	if totUrls == 1 {
		pluralS = ""
	}

	utils.PrintRunning(fmt.Sprintf("Gathering %d context file%s", totUrls, pluralS))

	progressChan := make(chan int64, totUrls)
	var printed atomic.Bool
	
	reporterDone := make(chan struct{})
	go func() {
		defer close(reporterDone)
		totCompleted := int64(0)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		firstTick := true
		for {
			select {
			case <-ctx.Done():
				return
			case completed, ok := <-progressChan:
				if !ok {
					return
				}
				totCompleted += completed
				if !firstTick {
					utils.ClearPreviousLine()
				}
				firstTick = false
				printed.Store(true)
				pct := int(float64(totCompleted) / float64(totUrls) * 100)
				utils.PrintProgress(fmt.Sprintf("%d/%d finished", totCompleted, totUrls), pct)
			case <-ticker.C:
				if !firstTick {
					utils.ClearPreviousLine()
				}
				firstTick = false
				printed.Store(true)
				pct := int(float64(totCompleted) / float64(totUrls) * 100)
				utils.PrintProgress(fmt.Sprintf("%d/%d finished", totCompleted, totUrls), pct)
			}
		}
	}()

	g, groupCtx := errgroup.WithContext(ctx)
	g.SetLimit(threads)

	var errorsMu sync.Mutex
	var errors []error

	for _, u := range urls {
		matched := false
		var toProcess input
		for ut, reg := range URLRegex {
			if isMatch, _ := regexp.MatchString(reg, u); isMatch {
				if ut == "yt" || ut == "yt1" || ut == "yt2" {
					toProcess = input{url: u, urlType: "yt"}
				} else {
					toProcess = input{url: u, urlType: ut}
				}
				matched = true
				break
			}
		}
		if !matched && strings.HasPrefix(u, "http") {
			toProcess = input{url: u, urlType: "generic"}
			matched = true
		}

		if !matched {
			continue
		}

		g.Go(func() error {
			select {
			case <-groupCtx.Done():
				return groupCtx.Err()
			default:
			}

			resultChan := make(chan result, 1)
			handlerWorker(groupCtx, toProcess, resultChan, includeGlobs, excludeGlobs, maxSize, useJina)
			
			res := <-resultChan
			if res.err != nil {
				errorsMu.Lock()
				errors = append(errors, fmt.Errorf("failed to process %s: %v", res.url, res.err))
				errorsMu.Unlock()
			}
			
			select {
			case <-groupCtx.Done():
				return groupCtx.Err()
			case progressChan <- 1:
			}

			return nil
		})
	}

	_ = g.Wait()

	close(progressChan)
	<-reporterDone

	if printed.Load() {
		utils.ClearPreviousLine()
	}
	utils.ClearLines(1)

	var errMsg string
	if len(errors) > 0 {
		for _, err := range errors {
			errMsg += (err.Error() + "\n")
		}
	}

	files, err := os.ReadDir(path.Join("context", "images"))
	if err != nil {
		errMsg += (fmt.Errorf("couldn't read images directory: %w", err)).Error() + "\n"
	} else if len(files) == 0 {
		err := os.RemoveAll(path.Join("context", "images"))
		if err != nil {
			errMsg += (fmt.Errorf("failed to delete empty images directory: %w", err)).Error() + "\n"
		}
	}

	if errMsg != "" {
		utils.PrintError("Completed all operations with errors", fmt.Errorf("%s", errMsg))
		for _, err := range errors {
			utils.PrintIndentedError("Job failed", err)
		}
	} else {
		utils.PrintSuccess("Completed all operations successfully")
	}
}
