package html

/*
#cgo pkg-config: libxml-2.0

#include <libxml/HTMLtree.h>
#include <libxml/HTMLparser.h>
#include "helper.h"
*/
import "C"

import (
	"errors"
	"github.com/moovweb/gokogiri/help"
	. "github.com/moovweb/gokogiri/util"
	"github.com/moovweb/gokogiri/xml"
	//"runtime"
	"unsafe"
)

//xml parse option
const (
	HTML_PARSE_RECOVER   xml.ParseOption = 1 << 0  /* Relaxed parsing */
	HTML_PARSE_NODEFDTD  xml.ParseOption = 1 << 2  /* do not default a doctype if not found */
	HTML_PARSE_NOERROR   xml.ParseOption = 1 << 5  /* suppress error reports */
	HTML_PARSE_NOWARNING xml.ParseOption = 1 << 6  /* suppress warning reports */
	HTML_PARSE_PEDANTIC  xml.ParseOption = 1 << 7  /* pedantic error reporting */
	HTML_PARSE_NOBLANKS  xml.ParseOption = 1 << 8  /* remove blank nodes */
	HTML_PARSE_NONET     xml.ParseOption = 1 << 11 /* Forbid network access */
	HTML_PARSE_NOIMPLIED xml.ParseOption = 1 << 13 /* Do not add implied html/body... elements */
	HTML_PARSE_COMPACT   xml.ParseOption = 1 << 16 /* compact small text nodes */
)

const EmptyHtmlDoc = ""

//default parsing option: relax parsing
var DefaultParseOption xml.ParseOption = HTML_PARSE_RECOVER |
	HTML_PARSE_NONET |
	HTML_PARSE_NOERROR |
	HTML_PARSE_NOWARNING

type HtmlDocument struct {
	*xml.XmlDocument
}

//default encoding in byte slice
var DefaultEncodingBytes = []byte(xml.DefaultEncoding)
var emptyHtmlDocBytes = []byte(EmptyHtmlDoc)

var ErrSetMetaEncoding = errors.New("Set Meta Encoding failed")
var ERR_FAILED_TO_PARSE_HTML = errors.New("failed to parse html input")
var emptyStringBytes = []byte{0}

//create a document
func NewDocument(p unsafe.Pointer, contentLen int, inEncoding, outEncoding []byte) (doc *HtmlDocument) {
	doc = &HtmlDocument{}
	doc.XmlDocument = xml.NewDocument(p, contentLen, inEncoding, outEncoding)
	doc.Me = doc
	node := doc.Node.(*xml.XmlNode)
	node.Document = doc
	//runtime.SetFinalizer(doc, (*HtmlDocument).Free)
	return
}

//parse a string to document
func Parse(content, inEncoding, url []byte, options xml.ParseOption, outEncoding []byte) (doc *HtmlDocument, err error) {
	inEncoding = AppendCStringTerminator(inEncoding)
	outEncoding = AppendCStringTerminator(outEncoding)

	var docPtr *C.xmlDoc
	contentLen := len(content)

	if contentLen > 0 {
		var contentPtr, urlPtr, encodingPtr unsafe.Pointer

		contentPtr = unsafe.Pointer(&content[0])
		if len(url) > 0 {
			url = AppendCStringTerminator(url)
			urlPtr = unsafe.Pointer(&url[0])
		}
		if len(inEncoding) > 0 {
			encodingPtr = unsafe.Pointer(&inEncoding[0])
		}

		docPtr = C.htmlParse(contentPtr, C.int(contentLen), urlPtr, encodingPtr, C.int(options), nil, 0)

		if docPtr == nil {
			err = ERR_FAILED_TO_PARSE_HTML
		} else {
			doc = NewDocument(unsafe.Pointer(docPtr), contentLen, inEncoding, outEncoding)
		}
	}
	if docPtr == nil {
		doc = CreateEmptyDocument(inEncoding, outEncoding)
	}
	return
}

func CreateEmptyDocument(inEncoding, outEncoding []byte) (doc *HtmlDocument) {
	help.LibxmlInitParser()
	docPtr := C.htmlNewDoc(nil, nil)
	doc = NewDocument(unsafe.Pointer(docPtr), 0, inEncoding, outEncoding)
	return
}

func (document *HtmlDocument) ParseFragment(input, url []byte, options xml.ParseOption) (fragment *xml.DocumentFragment, err error) {
	root := document.Root()
	if root == nil {
		fragment, err = parsefragment(document, nil, input, url, options)
	} else {
		fragment, err = parsefragment(document, root.XmlNode, input, url, options)
	}
	return
}

func (doc *HtmlDocument) MetaEncoding() string {
	metaEncodingXmlCharPtr := C.htmlGetMetaEncoding((*C.xmlDoc)(doc.DocPtr()))
	return C.GoString((*C.char)(unsafe.Pointer(metaEncodingXmlCharPtr)))
}

func (doc *HtmlDocument) SetMetaEncoding(encoding string) (err error) {
	var encodingPtr unsafe.Pointer = nil
	if len(encoding) > 0 {
		encodingBytes := AppendCStringTerminator([]byte(encoding))
		encodingPtr = unsafe.Pointer(&encodingBytes[0])
	}
	ret := int(C.htmlSetMetaEncoding((*C.xmlDoc)(doc.DocPtr()), (*C.xmlChar)(encodingPtr)))
	if ret == -1 {
		err = ErrSetMetaEncoding
	}
	return
}
