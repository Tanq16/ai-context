package aicontext

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/rs/zerolog/log"
)

var urlRegex = map[string]string{
	"gh":     "https://github.com/.+/.+",
	"yt":     "https://youtu.be/.+",
	"ytfull": "https://www.youtube.com/watch\\?v=.+",
}

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

func GitHubHandler(url, output string, ignoreList []string) {
	processor := NewProcessor(ProcessorConfig{
		OutputPath:        output,
		AdditionalIgnores: ignoreList,
	})
	err := processor.ProcessGitHubURL(url)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to process source")
	}
	log.Info().Str("output", output).Msg("successfully generated context")
}

func YouTubeHandler(url, output string) {
	segments, err := DownloadTranscript(url)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get transcript")
	}
	var content strings.Builder
	content.WriteString("# Video Transcript\n\n")
	for _, segment := range segments {
		content.WriteString(fmt.Sprintf("[%s] %s\n\n", segment.StartTime, segment.Text))
	}
	if err := os.WriteFile(output, []byte(content.String()), 0644); err != nil {
		log.Fatal().Err(err).Msg("failed to write transcript")
	}
	log.Info().Str("output", output).Msg("successfully generated transcript")
}

// Handler processes tasks
func Handler(directory, url, listFile, output string, ignoreList []string) {
	// Placeholder function for Go module
	if directory == "" && url == "" && listFile == "" {
		log.Fatal().Err(fmt.Errorf("either directory or url or list of urls must be specified")).Msg("failed to process source")
	}
	if (directory != "" && url != "") || (directory != "" && listFile != "") || (url != "" && listFile != "") {
		log.Fatal().Err(fmt.Errorf("cannot specify these combos; provide file or directory or url")).Msg("failed to process source")
	}
	// Check type of URL
	if url != "" {
		urlType := ""
		matched := false
		for regexType, regex := range urlRegex {
			if matched, _ = regexp.MatchString(regex, url); matched {
				urlType = regexType
				break
			}
		}
		if !matched {
			log.Fatal().Str("url", url).Msg("invalid URL")
		}
		if urlType == "yt" || urlType == "ytfull" {
			YouTubeHandler(url, output)
		} else {
			GitHubHandler(url, output, ignoreList)
		}
	} else if listFile != "" {
		// Process list of URLs
		// urls, err := ReadURLs(listFile)
		// if err != nil {
		// 	log.Fatal().Err(err).Msg("failed to read URLs")
		// }
	} else {
		DirectoryHandler(directory, output, ignoreList)
	}
}
