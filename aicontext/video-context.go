package aicontext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	log "github.com/rs/zerolog/log"
)

// TranscriptParams holds the parameters for the transcript request
type TranscriptParams struct {
	Params string `json:"params"`
}

// TranscriptSegment holds the transcript segment data
type TranscriptSegment struct {
	StartTime string `json:"startTime"`
	Text      string `json:"text"`
}

// InnerTubeRequest holds the request data for the InnerTube API
type InnerTubeRequest struct {
	Context struct {
		Client struct {
			ClientName    string `json:"clientName"`
			ClientVersion string `json:"clientVersion"`
		} `json:"client"`
	} `json:"context"`
}

// ExtractVideoID extracts video ID from various YouTube URL formats
func ExtractVideoID(urlStr string) (string, error) {
	if strings.Contains(urlStr, "youtu.be/") {
		parts := strings.Split(urlStr, "/")
		return parts[len(parts)-1], nil
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	if strings.Contains(u.Host, "youtube.com") {
		if v := u.Query().Get("v"); v != "" {
			return v, nil
		}
	}
	return "", fmt.Errorf("could not extract video ID from URL")
}

// getTranscriptParams extracts transcript parameters from video data
func getTranscriptParams(videoData map[string]interface{}) (string, error) {
	panels, ok := videoData["engagementPanels"].([]interface{})
	if !ok {
		return "", fmt.Errorf("no engagement panels found")
	}
	for _, panel := range panels {
		if p, ok := panel.(map[string]interface{}); ok {
			if section, ok := p["engagementPanelSectionListRenderer"].(map[string]interface{}); ok {
				if section["panelIdentifier"] == "engagement-panel-searchable-transcript" {
					if content, ok := section["content"].(map[string]interface{}); ok {
						if renderer, ok := content["continuationItemRenderer"].(map[string]interface{}); ok {
							if endpoint, ok := renderer["continuationEndpoint"].(map[string]interface{}); ok {
								if transcriptEndpoint, ok := endpoint["getTranscriptEndpoint"].(map[string]interface{}); ok {
									if params, ok := transcriptEndpoint["params"].(string); ok {
										return params, nil
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return "", fmt.Errorf("transcript parameters not found")
}

// formatTranscriptSegments extracts and formats transcript segments
func formatTranscriptSegments(transcriptData map[string]interface{}) ([]TranscriptSegment, error) {
	var segments []TranscriptSegment
	actions, ok := transcriptData["actions"].([]interface{})
	if !ok || len(actions) == 0 {
		return nil, fmt.Errorf("no transcript actions found")
	}
	updateAction, ok := actions[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid action format")
	}
	panelAction, ok := updateAction["updateEngagementPanelAction"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid panel action")
	}
	content, ok := panelAction["content"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid content")
	}

	if renderer, ok := content["transcriptRenderer"].(map[string]interface{}); ok {
		if body, ok := renderer["content"].(map[string]interface{}); ok {
			if searchPanel, ok := body["transcriptSearchPanelRenderer"].(map[string]interface{}); ok {
				if bodyContent, ok := searchPanel["body"].(map[string]interface{}); ok {
					if segmentList, ok := bodyContent["transcriptSegmentListRenderer"].(map[string]interface{}); ok {
						if initialSegments, ok := segmentList["initialSegments"].([]interface{}); ok {
							for _, segment := range initialSegments {
								if segRenderer, ok := segment.(map[string]interface{}); ok {
									if renderer, ok := segRenderer["transcriptSegmentRenderer"].(map[string]interface{}); ok {
										startTime := renderer["startTimeText"].(map[string]interface{})["simpleText"].(string)
										text := renderer["snippet"].(map[string]interface{})["runs"].([]interface{})[0].(map[string]interface{})["text"].(string)
										segments = append(segments, TranscriptSegment{
											StartTime: startTime,
											Text:      text,
										})
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return segments, nil
}

// DownloadTranscript downloads and formats transcript from a YouTube video URL
func DownloadTranscript(videoURL string) ([]TranscriptSegment, error) {
	videoID, err := ExtractVideoID(videoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract video ID: %v", err)
	}
	// Extract the API key from the page
	apiKey, err := extractInnertubeKey(videoID)
	log.Debug().Str("apiKey", apiKey).Msg("extracted API key")
	if err != nil {
		// Fall back to default key if extraction fails
		apiKey = "AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"
		log.Debug().Msg("falling back to publicly known Web Client API key")
	}
	// Create InnerTube request for video data
	nextReq := InnerTubeRequest{}
	nextReq.Context.Client.ClientName = "WEB"
	nextReq.Context.Client.ClientVersion = "2.20240105.01.00"
	// First request to get video data
	reqBody := map[string]interface{}{
		"videoId": videoID,
		"context": nextReq.Context,
	}
	videoData, err := makeInnerTubeRequest("next", reqBody, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get video data: %v", err)
	}
	log.Debug().Msg("got video data")
	// Extract transcript parameters
	params, err := getTranscriptParams(videoData)
	if err != nil {
		return nil, fmt.Errorf("failed to get transcript params: %v", err)
	}
	log.Debug().Str("params", params).Msg("got transcript params")
	transcriptReq := map[string]interface{}{
		"params":  params,
		"context": nextReq.Context,
	}
	transcriptData, err := makeInnerTubeRequest("get_transcript", transcriptReq, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get transcript: %v", err)
	}
	log.Debug().Msg("got transcript data")
	// Format transcript segments
	segments, err := formatTranscriptSegments(transcriptData)
	if err != nil {
		return nil, fmt.Errorf("failed to format transcript: %v", err)
	}
	log.Debug().Msg("formatted transcript")
	return segments, nil
}

// makeInnerTubeRequest makes a request to the InnerTube API
func makeInnerTubeRequest(endpoint string, reqBody interface{}, apiKey string) (map[string]interface{}, error) {
	baseURL := "https://www.youtube.com/youtubei/v1"
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/%s?key=%s", baseURL, endpoint, apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.157 Safari/537.36")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// extractInnertubeKey extracts the INNERTUBE_API_KEY from the video page
func extractInnertubeKey(videoID string) (string, error) {
	resp, err := http.Get("https://www.youtube.com/watch?v=" + videoID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch video page: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read page content: %v", err)
	}
	pattern := regexp.MustCompile(`"INNERTUBE_API_KEY":"([^"]+)"`)
	matches := pattern.FindSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find INNERTUBE_API_KEY in page")
	}
	return string(matches[1]), nil
}
