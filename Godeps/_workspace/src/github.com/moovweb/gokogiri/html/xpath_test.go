package html

import "testing"

func TestUnfoundFuncInXpath(t *testing.T) {
	defer CheckXmlMemoryLeaks(t)

	doc, err := Parse([]byte("<html><body><div><h1></div>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}

	html := doc.Root().FirstChild()
	results, _ := html.Search("./div[matches(text(), 'foo')]")
	if results != nil {
		t.Error("should return nil because the function is not found")
	}
	doc.Free()
}

func TestXpathEmptyResult(t *testing.T) {
	defer CheckXmlMemoryLeaks(t)

	doc, err := Parse([]byte("<html><body><div><h1></div>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}

	html := doc.Root().FirstChild()
	results, err := html.Search("./div[@calass='cool']")
	if err != nil {
		t.Error("Xpath eval should not return nil")
	}
	if len(results) > 0 {
		t.Error("Xpath should return empty result")
	}
	doc.Free()
}
