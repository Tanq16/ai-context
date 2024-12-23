package aicontext

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/google/uuid"
	log "github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

// ProcessWebContent converts a webpage to markdown and downloads associated images
func ProcessWebContent(urlStr, outputPath string) error {
	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}
	resp, err := http.Get(urlStr)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// parse HTML
	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}
	log.Debug().Msg("processing images")
	err = processImages(doc, baseURL)
	if err != nil {
		return fmt.Errorf("failed to process images: %w", err)
	}

	// convert to markdown
	markdown, err := htmltomarkdown.ConvertString(renderNode(doc))
	if err != nil {
		return fmt.Errorf("failed to convert to markdown: %w", err)
	}
	re := regexp.MustCompile(`\[!\[(.*?)\]\((.*?)\)\]\(.*?\)`)
	markdown = re.ReplaceAllString(markdown, "![$1]($2)")
	final := fmt.Sprintf("# Webpage Context: %s\n\nSource: %s\n\n%s", baseURL.Host, urlStr, markdown)

	// write output
	err = os.WriteFile(outputPath, []byte(final), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	return nil
}

// processImages downloads images and updates their src attributes
func processImages(doc *html.Node, baseURL *url.URL) error {
	var processNode func(*html.Node)
	processNode = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for i, attr := range n.Attr {
				if attr.Key == "src" {
					imgURL, err := url.Parse(attr.Val)
					if err != nil {
						log.Warn().Err(err).Str("url", attr.Val).Msg("failed to parse image URL")
						continue
					}
					if !imgURL.IsAbs() {
						imgURL = baseURL.ResolveReference(imgURL)
					}

					// Download image and name it a uuid
					ext := path.Ext(imgURL.Path)
					filename := uuid.New().String() + ext
					localPath := path.Join("context", "images", filename)
					err = downloadImage(imgURL.String(), localPath)
					if err != nil {
						log.Warn().Err(err).Str("url", imgURL.String()).Msg("failed to download image")
						continue
					}
					// make src point to local image
					n.Attr[i].Val = path.Join("images", filename)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processNode(c)
		}
	}
	processNode(doc)
	return nil
}

// downloadImage downloads an image from URL to the specified local path
func downloadImage(imgURL string, localPath string) error {
	resp, err := http.Get(imgURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

// renderNode renders an HTML node back to string
func renderNode(n *html.Node) string {
	var buf strings.Builder
	err := html.Render(&buf, n)
	if err != nil {
		return ""
	}
	return buf.String()
}
