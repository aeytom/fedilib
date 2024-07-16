package fedilib

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

func StripHtml(reader io.Reader) (string, error) {
	nd, err := html.Parse(reader)
	if err != nil {
		return "", err
	}
	return StripHtmlFromNode(nd), nil
}

func StripHtmlFromString(in string) (string, error) {
	return StripHtml(strings.NewReader(in))
}

func StripHtmlFromNode(nd *html.Node) string {
	out := ""
	for c := nd.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.ElementNode:
			switch c.Data {
			case "ol":
				out += "\n"
			case "ul":
				out += "\n"
			case "li":
				out += "\n- "
			}
			out += StripHtmlFromNode(c)
			switch c.Data {
			case "p":
				out += "\n\n"
			case "div":
				out += "\n"
			case "ol":
				out += "\n"
			case "ul":
				out += "\n"
			case "li":
				out += "\n"
			}
		case html.TextNode:
			out += c.Data
		}
	}
	return out
}
