package aicontext

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func makeWebRequest(urlStr string) (*http.Response, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	return resp, nil
}

func ProcessWebContent(urlStr, outputPath string) error {
	jinaURL := "https://r.jina.ai/" + urlStr
	resp, err := makeWebRequest(jinaURL)
	if err != nil {
		return fmt.Errorf("failed to fetch URL from jina: %w", err)
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	final := fmt.Sprintf("# Webpage Context\n\nSource: %s\n\n%s", urlStr, string(content))
	err = os.WriteFile(outputPath, []byte(final), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	return nil
}
