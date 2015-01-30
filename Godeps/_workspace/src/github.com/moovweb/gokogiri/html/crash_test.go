package html

import "testing"

func TestCrazyMove(t *testing.T) {
	input := `
<html>
<body>
<div id="foo" name="foo1"> 
<div id="bar" name="bar1"></div>
<div id="foo" name="foo2"></div>
</div>
<div id="bar" name="bar2"></div>
</body>
</html>`
	doc, err := Parse([]byte(input), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}

	foos, err := doc.Search("//div[@id='foo']")
	if err != nil {
		t.Error("search has error:", err)
		return
	}
	for _, foo := range foos {
		bars, _ := foo.Search("//div[@id='bar']")
		for _, bar := range bars {
			bar.AddChild(foo)
		}
	}

	doc.Free()
	CheckXmlMemoryLeaks(t)
}
