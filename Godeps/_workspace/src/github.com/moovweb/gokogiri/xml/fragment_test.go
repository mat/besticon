package xml

import "testing"

func TestParseDocumentFragmentBasic(t *testing.T) {
	defer CheckXmlMemoryLeaks(t)

	doc, err := Parse(nil, DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		t.Error("parsing error:", err.Error())
		return
	}
	root := doc.Root()
	if root != nil {
		println("root:", root.String())
	}
	docFragment, err := doc.ParseFragment([]byte("hi"), nil, DefaultParseOption)
	if err != nil {
		t.Error(err.Error())
		doc.Free()
		return
	}
	if len(docFragment.Children()) != 1 {
		t.Error("the number of children from the fragment does not match")
	}
	doc.Free()
}

func TestParseDocumentFragment(t *testing.T) {
	defer CheckXmlMemoryLeaks(t)

	doc, err := Parse(nil, DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		t.Error("parsing error:", err.Error())
		return
	}
	docFragment, err := doc.ParseFragment([]byte("<foo></foo><!-- comment here --><bar>fun</bar>"), nil, DefaultParseOption)
	if err != nil {
		t.Error(err.Error())
		doc.Free()
		return
	}
	if docFragment.String() != "<foo/><!-- comment here --><bar>fun</bar>" {
		t.Error("fragment output is wrong\n")
		doc.Free()
		return
	}
	if len(docFragment.Children()) != 3 {
		t.Error("the number of children from the fragment does not match")
	}
	doc.Free()
}

func TestSearchDocumentFragment(t *testing.T) {
	defer CheckXmlMemoryLeaks(t)

	doc, err := Parse([]byte("<moovweb><z/><s/></moovweb>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		t.Error("parsing error:", err.Error())
		return
	}
	docFragment, err := doc.ParseFragment([]byte("<foo></foo><!-- comment here --><bar>fun</bar>"), nil, DefaultParseOption)
	if err != nil {
		t.Error(err.Error())
		doc.Free()
		return
	}
	nodes, err := docFragment.Search(".//*")
	if err != nil {
		t.Error("fragment search has error")
		doc.Free()
		return
	}
	if len(nodes) != 2 {
		t.Error("the number of children from the fragment does not match")
	}
	nodes, err = docFragment.Search("//*")

	if err != nil {
		t.Error("fragment search has error")
		doc.Free()
		return
	}

	if len(nodes) != 3 {
		t.Error("the number of children from the fragment's document does not match")
	}

	doc.Free()
}

func TestSearchDocumentFragmentWithEmptyDoc(t *testing.T) {
	defer CheckXmlMemoryLeaks(t)

	doc, err := Parse(nil, DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		t.Error("parsing error:", err.Error())
		return
	}
	docFragment, err := doc.ParseFragment([]byte("<foo></foo><!-- comment here --><bar>fun</bar>"), nil, DefaultParseOption)
	if err != nil {
		t.Error(err.Error())
		doc.Free()
		return
	}
	nodes, err := docFragment.Search(".//*")
	if err != nil {
		t.Error("fragment search has error")
		doc.Free()
		return
	}
	if len(nodes) != 2 {
		t.Error("the number of children from the fragment does not match")
	}
	nodes, err = docFragment.Search("//*")

	if err != nil {
		t.Error("fragment search has error")
		doc.Free()
		return
	}

	if len(nodes) != 0 {
		t.Error("the number of children from the fragment's document does not match")
	}

	doc.Free()
}
