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
	"golang.org/x/net/html"
)

var ignoreHTMLTags = map[string]bool{
	"script":   true,
	"style":    true,
	"noscript": true,
	"header":   true,
	"footer":   true,
	"aside":    true,
	"nav":      true,
	"form":     true,
	"iframe":   true,
}

var ignoerAttributes = regexp.MustCompile(`(?i)comment|meta|footnote|masthead|related|shoutbox|sponsor|ad-break|agegate|pagination|pager|popup|tweet|twitter|social|nav|menu|authors|newsletter`)

func makeWebRequest(urlStr string) (*http.Response, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	// google bot simulate
	// req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	// req.Header.Set("Referer", "https://t.co/")        // twitter refer
	// req.Header.Set("X-Forwarded-For", "66.249.66.10") // google bot IP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	return resp, nil
}

func ProcessWebContent(urlStr, outputPath string) error {
	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}
	resp, err := makeWebRequest(urlStr)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}
	cleanHTML(doc)
	_ = processImages(doc, baseURL) // ignore errors
	markdown, err := htmltomarkdown.ConvertString(renderNode(doc))
	if err != nil {
		return fmt.Errorf("failed to convert to markdown: %w", err)
	}
	// double wrap
	// re := regexp.MustCompile(`\[!\[(.*?)\]\((.*?)\)\]\(.*?\)`)
	// markdown = re.ReplaceAllString(markdown, "![$1]($2)")
	title := findTitle(doc)
	if title == "" {
		title = baseURL.Host
	}
	final := fmt.Sprintf("# Webpage Context: %s\n\nSource: %s\n\n%s", title, urlStr, markdown)
	err = os.WriteFile(outputPath, []byte(final), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	return nil
}

func cleanHTML(n *html.Node) {
	var next *html.Node
	for c := n.FirstChild; c != nil; c = next {
		next = c.NextSibling
		if ignoreHTMLTags[c.Data] {
			n.RemoveChild(c)
			continue
		}
		if c.Type == html.ElementNode {
			shouldRemove := false
			for _, attr := range c.Attr {
				if attr.Key == "id" || attr.Key == "class" {
					if ignoerAttributes.MatchString(attr.Val) {
						shouldRemove = true
						break
					}
				}
			}
			if shouldRemove {
				n.RemoveChild(c)
				continue
			}
		}
		// recursive for node that wasn't cleaned
		cleanHTML(c)
	}
}

func findTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		if n.FirstChild != nil {
			return strings.TrimSpace(n.FirstChild.Data)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := findTitle(c); title != "" {
			return title
		}
	}
	return ""
}

func processImages(doc *html.Node, baseURL *url.URL) error {
	var processNode func(*html.Node)
	processNode = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for i, attr := range n.Attr {
				if attr.Key == "src" {
					imgURL, err := url.Parse(attr.Val)
					if err != nil {
						continue
					}
					if !imgURL.IsAbs() {
						imgURL = baseURL.ResolveReference(imgURL)
					}

					// Download image and name it a uuid
					ext := path.Ext(imgURL.Path)
					if ext == "" {
						ext = ".jpg" // default extension
					}
					filename := uuid.New().String() + ext
					localPath := path.Join("context", "images", filename)
					err = downloadImage(imgURL.String(), localPath)
					if err != nil {
						continue // just skip image
					}
					// make src point to local image
					n.Attr[i].Val = path.Join(".", "images", filename)
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

func downloadImage(imgURL string, localPath string) error {
	resp, err := http.Get(imgURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func renderNode(n *html.Node) string {
	var buf strings.Builder
	err := html.Render(&buf, n)
	if err != nil {
		return ""
	}
	return buf.String()
}
