package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"golang.org/x/net/html"
)

func CrawlSite(startURL string, maxPages int) (string, error) {
	visited := map[string]bool{}
	var allText strings.Builder
	queue := []string{startURL}
	base, err := url.Parse(startURL)
	if err != nil {
		return "", fmt.Errorf("ошибка разбора url: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for len(queue) > 0 && len(visited) < maxPages {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		resp, err := client.Get(current)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(io.LimitReader(resp.Body, 1_000_000)) 
		if err != nil {
			continue
		}
		doc, err := html.Parse(strings.NewReader(string(body)))
if err != nil {
	continue
}

bodyNode := findBodyNode(doc)
if bodyNode == nil {
	continue
}

text := extractText(bodyNode)

		allText.WriteString(text)
		allText.WriteString("\n")

		links := extractLinks(doc, base)
		for _, link := range links {
			if !visited[link] {
				queue = append(queue, link)
			}
		}
	}

	return allText.String(), nil
}

func extractText(n *html.Node) string {
	var result strings.Builder

	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if len(text) > 1 {
				result.WriteString(text)
				result.WriteString(" ")
			}
		}
		if n.Type == html.ElementNode && n.Data != "script" && n.Data != "style" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				visit(c)
			}
		}
	}
	visit(n)

	return result.String()
}

func extractLinks(n *html.Node, base *url.URL) []string {
	var links []string

	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href := attr.Val
					link, err := base.Parse(href)
					if err == nil && link.Host == base.Host {
						clean := link.Scheme + "://" + link.Host + link.Path
						links = append(links, clean)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(n)
	return links
}

func findBodyNode(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "body" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findBodyNode(c); result != nil {
			return result
		}
	}
	return nil
}
