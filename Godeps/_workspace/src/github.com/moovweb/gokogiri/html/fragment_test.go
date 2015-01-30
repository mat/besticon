package html

import "testing"

func TestParseDocumentFragmentText(t *testing.T) {
	doc, err := Parse(nil, []byte("iso-8859-1"), nil, DefaultParseOption, []byte("iso-8859-1"))
	if err != nil {
		println(err.Error())
	}
	docFragment, err := doc.ParseFragment([]byte("ok\r\n"), nil, DefaultParseOption)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(docFragment.Children()) != 1 || docFragment.Children()[0].String() != "ok\r\n" {
		println(docFragment.String())
		t.Error("the children from the fragment text do not match")
	}
	doc.Free()
	CheckXmlMemoryLeaks(t)
}

func TestParseDocumentFragment(t *testing.T) {
	doc, err := Parse(nil, DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		println(err.Error())
	}
	docFragment, err := doc.ParseFragment([]byte("<div><h1>"), nil, DefaultParseOption)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(docFragment.Children()) != 1 || docFragment.Children()[0].String() != "<div><h1></h1></div>" {
		t.Error("the of children from the fragment do not match")
	}

	doc.Free()
	CheckXmlMemoryLeaks(t)

}

func TestParseDocumentFragment2(t *testing.T) {
	docStr := `<html>
<head><meta http-equiv="Content-Type" content="text/html; charset=utf-8"></head>
<body>
  </body>
</html>`
	doc, err := Parse([]byte(docStr), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		println(err.Error())
	}
	docFragment, err := doc.ParseFragment([]byte("<script>cool & fun</script>"), nil, DefaultParseOption)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(docFragment.Children()) != 1 || docFragment.Children()[0].String() != "<script>cool & fun</script>" {
		t.Error("the of children from the fragment do not match")
	}

	doc.Free()
	CheckXmlMemoryLeaks(t)
}

func TestSearchDocumentFragment(t *testing.T) {
	doc, err := Parse([]byte("<div class='cool'></div>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		println(err.Error())
	}
	docFragment, err := doc.ParseFragment([]byte("<div class='cool'><h1>"), nil, DefaultParseOption)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(docFragment.Children()) != 1 || docFragment.Children()[0].String() != "<div class=\"cool\"><h1></h1></div>" {
		t.Error("the of children from the fragment do not match")
	}

	nodes, err := docFragment.Search(".//*")
	if err != nil {
		t.Error("fragment search has error")
		return
	}
	if len(nodes) != 2 {
		t.Error("the number of children from the fragment does not match")
	}
	nodes, err = docFragment.Search("//div[@class='cool']")

	if err != nil {
		t.Error("fragment search has error")
		return
	}

	if len(nodes) != 1 {
		println(len(nodes))
		for _, node := range nodes {
			println(node.String())
		}
		t.Error("the number of children from the fragment's document does not match")
	}

	doc.Free()
	CheckXmlMemoryLeaks(t)
}

func TestAddFragmentWithNamespace(t *testing.T) {
	doc, err := Parse([]byte("<div class='cool'></div>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		println(err.Error())
	}
	defer doc.Free()
	docFragment, err := doc.ParseFragment([]byte("<div xmlns='http://www.moovweb.com' class='cool'><h1>"), nil, DefaultParseOption)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if docFragment.String() != `<div xmlns="http://www.moovweb.com" class="cool"><h1></h1></div>` {
		t.Errorf("doc fragment does not match\n")
	}
	doc2, err := Parse([]byte("<div class='not so cool'></div>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		println(err.Error())
		return
	}
	defer doc2.Free()
	body := doc2.Root().FirstChild()
	body.AddChild(docFragment)
	if doc2.String() != `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.0 Transitional//EN" "http://www.w3.org/TR/REC-html40/loose.dtd">
<html><body>
<div class="not so cool"></div>
<div xmlns="http://www.moovweb.com" class="cool"><h1></h1></div>
</body></html>
` 	{
		t.Errorf("document does not match after adding a fragment with namespace\n")
	}
	CheckXmlMemoryLeaks(t)
}
