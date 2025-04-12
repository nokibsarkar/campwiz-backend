package consts

import (
	"strings"

	"golang.org/x/net/html"
)

// HtmlToPlainText converts HTML content to plain text.
// It removes HTML tags and extracts the text content, preserving line breaks
// and formatting where appropriate.
// The function handles common HTML elements and ignores script, style, and metadata tags.
// It returns the plain text representation of the HTML content.
// If an error occurs during parsing, it returns an empty string and the error.
// Example usage:
// htmlStr := `<html><body><h1>Hello</h1><p>This is a <b>paragraph</b>.</p></body></html>`
// plainText, err := HtmlToPlainText(htmlStr)
//
//	if err != nil {
//	    log.Fatal(err)
//	}
func HtmlToPlainText(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return htmlStr
	}

	var b strings.Builder

	var extractText func(*html.Node)
	extractText = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Skip non-visible or metadata tags
			switch n.Data {
			case "script", "style", "head", "noscript":
				return
			case "br", "p", "div", "li", "h1", "h2", "h3", "h4", "h5", "h6":
				b.WriteString("\n")
			}
		}

		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				b.WriteString(text + " ")
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}

		// Add line break after closing block tags
		if n.Type == html.ElementNode {
			switch n.Data {
			case "p", "div", "li", "h1", "h2", "h3", "h4", "h5", "h6":
				b.WriteString("\n")
			}
		}
	}

	extractText(doc)

	// Normalize whitespace and multiple line breaks
	lines := strings.Split(b.String(), "\n")
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	return strings.Join(cleaned, "\n")
}
