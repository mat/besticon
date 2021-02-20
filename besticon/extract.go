// Package besticon includes functions
// finding icons for a given web site.
package besticon

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var iconPaths = []string{
	"/favicon.ico",
	"/apple-touch-icon.png",
	"/apple-touch-icon-precomposed.png",
}

const (
	favIcon                   = "icon"
	appleTouchIcon            = "apple-touch-icon"
	appleTouchIconPrecomposed = "apple-touch-icon-precomposed"
)

type empty struct{}

// Find all icons in this html. We use siteURL as the base url unless we detect
// another base url in <head>
func findIconLinks(siteURL *url.URL, html []byte) ([]string, error) {
	doc, e := docFromHTML(html)
	if e != nil {
		return nil, e
	}

	baseURL := determineBaseURL(siteURL, doc)

	// Use a map to avoid dups
	links := make(map[string]empty)

	// Add common, hard coded icon paths
	for _, path := range iconPaths {
		links[urlFromBase(baseURL, path)] = empty{}
	}

	// Add icons found in page
	urls := extractIconTags(doc)
	for _, u := range urls {
		absoluteURL, e := absoluteURL(baseURL, u)
		if e == nil {
			links[absoluteURL] = empty{}
		}
	}

	// Turn unique keys into array
	var result []string
	for u := range links {
		result = append(result, u)
	}
	sort.Strings(result)

	return result, nil
}

// What is the baseURL for this doc?
func determineBaseURL(siteURL *url.URL, doc *goquery.Document) *url.URL {
	baseTagHref := extractBaseTag(doc)
	if baseTagHref != "" {
		baseTagURL, e := url.Parse(baseTagHref)
		if e != nil {
			return siteURL
		}
		return baseTagURL
	}

	return siteURL
}

// Convert bytes => doc
func docFromHTML(html []byte) (*goquery.Document, error) {
	doc, e := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if e != nil || doc == nil {
		return nil, errParseHTML
	}
	return doc, nil
}

// List of css selectors for icons
var csspaths = strings.Join([]string{
	"link[rel='icon']",
	"link[rel='shortcut icon']",
	"link[rel='apple-touch-icon']",
	"link[rel='apple-touch-icon-precomposed']",

	// Capitalized variants, TODO: refactor
	"link[rel='ICON']",
	"link[rel='SHORTCUT ICON']",
	"link[rel='APPLE-TOUCH-ICON']",
	"link[rel='APPLE-TOUCH-ICON-PRECOMPOSED']",
}, ", ")

var errParseHTML = errors.New("besticon: could not parse html")

// Find <head><base href="xxx">
func extractBaseTag(doc *goquery.Document) string {
	href := ""
	doc.Find("head base[href]").First().Each(func(i int, s *goquery.Selection) {
		href, _ = s.Attr("href")
	})
	return href
}

var (
	iconTypes   = []string{favIcon, appleTouchIcon, appleTouchIconPrecomposed}
	iconTypesRe = regexp.MustCompile(fmt.Sprintf(`\b(%s)\b`, strings.Join(iconTypes, "|")))
)

// Find icons from doc using csspaths
func extractIconTags(doc *goquery.Document) []string {
	var hits []string
	doc.Find("link[href][rel]").Each(func(i int, s *goquery.Selection) {
		href := extractIconTag(s)
		if href != "" {
			hits = append(hits, href)
		}
	})
	return hits
}

func extractIconTag(s *goquery.Selection) string {
	// What sort of iconType is in this <rel>?
	rel, _ := s.Attr("rel")
	if rel == "" {
		return ""
	}
	rel = strings.ToLower(rel)

	var iconType string
	for _, i := range strings.Split(rel, " ") {
		if iconTypesRe.MatchString(i) {
			iconType = i
			break
		}
	}
	if iconType == "" {
		return ""
	}

	href, _ := s.Attr("href")
	if href == "" {
		return ""
	}

	return href
}
