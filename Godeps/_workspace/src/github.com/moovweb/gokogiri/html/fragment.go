package html

//#include "helper.h"
import "C"
import (
	"bytes"
	"errors"
	. "github.com/moovweb/gokogiri/util"
	"github.com/moovweb/gokogiri/xml"
	"unsafe"
)

var fragmentWrapperStart = []byte("<div>")
var fragmentWrapperEnd = []byte("</div>")
var fragmentWrapper = []byte("<html><body>")
var bodySigBytes = []byte("<body")

var ErrFailParseFragment = errors.New("failed to parse html fragment")
var ErrEmptyFragment = errors.New("empty html fragment")

const initChildrenNumber = 4

func parsefragment(document xml.Document, node *xml.XmlNode, content, url []byte, options xml.ParseOption) (fragment *xml.DocumentFragment, err error) {
	//set up pointers before calling the C function
	var contentPtr, urlPtr unsafe.Pointer
	if len(url) > 0 {
		urlPtr = unsafe.Pointer(&url[0])
	}

	var root xml.Node
	if node == nil {
		containBody := (bytes.Index(content, bodySigBytes) >= 0)

		content = append(fragmentWrapper, content...)
		contentPtr = unsafe.Pointer(&content[0])
		contentLen := len(content)

		inEncoding := document.InputEncoding()
		var encodingPtr unsafe.Pointer
		if len(inEncoding) > 0 {
			encodingPtr = unsafe.Pointer(&inEncoding[0])
		}
		htmlPtr := C.htmlParseFragmentAsDoc(document.DocPtr(), contentPtr, C.int(contentLen), urlPtr, encodingPtr, C.int(options), nil, 0)

		//Note we've parsed the fragment within the given document
		//the root is not the root of the document; rather it's the root of the subtree from the fragment
		html := xml.NewNode(unsafe.Pointer(htmlPtr), document)

		if html == nil {
			err = ErrFailParseFragment
			return
		}
		root = html

		if !containBody {
			root = html.FirstChild()
			html.AddPreviousSibling(root)
			html.Remove() //remove html otherwise it's leaked
		}
	} else {
		//wrap the content
		newContent := append(fragmentWrapperStart, content...)
		newContent = append(newContent, fragmentWrapperEnd...)
		contentPtr = unsafe.Pointer(&newContent[0])
		contentLen := len(newContent)
		rootElementPtr := C.htmlParseFragment(node.NodePtr(), contentPtr, C.int(contentLen), urlPtr, C.int(options), nil, 0)
		if rootElementPtr == nil {
			//try to parse it as a doc
			fragment, err = parsefragment(document, nil, content, url, options)
			return
		}
		if rootElementPtr == nil {
			err = ErrFailParseFragment
			return
		}
		root = xml.NewNode(unsafe.Pointer(rootElementPtr), document)
	}

	fragment = &xml.DocumentFragment{}
	fragment.Node = root
	fragment.InEncoding = document.InputEncoding()
	fragment.OutEncoding = document.OutputEncoding()

	document.BookkeepFragment(fragment)
	return
}

func ParseFragment(content, inEncoding, url []byte, options xml.ParseOption, outEncoding []byte) (fragment *xml.DocumentFragment, err error) {
	inEncoding = AppendCStringTerminator(inEncoding)
	outEncoding = AppendCStringTerminator(outEncoding)
	document := CreateEmptyDocument(inEncoding, outEncoding)
	fragment, err = parsefragment(document, nil, content, url, options)
	return
}
