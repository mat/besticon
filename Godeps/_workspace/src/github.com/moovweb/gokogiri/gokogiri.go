/*
The gokogiri package provides a Go interface to the libxml2 library.

It is inspired by the ruby-based Nokogiri API, and allows one to parse, manipulate, and create HTML and XML
documents. Nodes can be selected using either CSS selectors (in much the same fashion as jQuery) or XPath 1.0 expressions,
and a simple DOM-like inteface allows for building up documents from scratch.
*/
package gokogiri

import (
	"github.com/moovweb/gokogiri/html"
	"github.com/moovweb/gokogiri/xml"
)

/*
ParseHtml parses an UTF-8 encoded byte array and returns an html.HtmlDocument. It uses parsing default options that ignore
errors or warnings, making it suitable for the poorly-formed 'tag soup' often found on the web.

If the content is not UTF-8 encoded or you want to customize the parsing options, you should call html.Parse directly.
*/
func ParseHtml(content []byte) (doc *html.HtmlDocument, err error) {
	return html.Parse(content, html.DefaultEncodingBytes, nil, html.DefaultParseOption, html.DefaultEncodingBytes)
}

/*
ParseXml parses an UTF-8 encoded byte array and returns an xml.XmlDocument. By default the parsing options ignore validation
and suppress errors and warnings. This allows one to liberal in accepting badly-formed documents, but is not standards-compliant.

If the content is not UTF-8 encoded or you want to customize the parsing options, you should call the Parse or ReadFile functions
found in the github.com/moovweb/gokogiri/xml package. The xml.StrictParsingOption is conveniently provided for standards-compliant
behaviour.
*/
func ParseXml(content []byte) (doc *xml.XmlDocument, err error) {
	return xml.Parse(content, xml.DefaultEncodingBytes, nil, xml.DefaultParseOption, xml.DefaultEncodingBytes)
}
