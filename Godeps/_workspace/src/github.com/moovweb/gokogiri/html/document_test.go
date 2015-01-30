package html

import "testing"
import "fmt"

func TestParseDocument(t *testing.T) {
	expected :=
		`<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.0 Transitional//EN" "http://www.w3.org/TR/REC-html40/loose.dtd">
<html><body><div><h1></h1></div></body></html>
`
	expected_xml :=
		`<?xml version="1.0" encoding="utf-8" standalone="yes"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.0 Transitional//EN" "http://www.w3.org/TR/REC-html40/loose.dtd">
<html>
  <body>
    <div>
      <h1/>
    </div>
  </body>
</html>
`
	doc, err := Parse([]byte("<html><body><div><h1></div>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}

	if doc.String() != expected {
		println("got:\n", doc.String())
		println("expected:\n", expected)
		t.Error("the output of the html doc does not match")
	}

	s, _ := doc.ToXml(nil, nil)
	if string(s) != expected_xml {
		println("got:\n", string(s))
		println("expected:\n", expected_xml)
		t.Error("the xml output of the html doc does not match")
	}

	doc.Free()
	CheckXmlMemoryLeaks(t)
}

func TestEmptyDocument(t *testing.T) {
	expected :=
		`<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.0 Transitional//EN" "http://www.w3.org/TR/REC-html40/loose.dtd">

`
	doc, err := Parse(nil, DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}

	if doc.String() != expected {
		println(doc.String())
		t.Error("the output of the html doc does not match the empty xml")
	}
	doc.Free()
	CheckXmlMemoryLeaks(t)
}

func TestNodeById(t *testing.T) {
	html := "<html><head></head><body><div id='yup'>success</div><div id='nope'>fail</div></body></html>"
	doc, _ := Parse([]byte(html), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	p := doc.NodeById("yup")
	if p == nil {
		t.Errorf("Did not find node by ID!")
		return
	}
	output := fmt.Sprintf("%v", p.Content())
	if output != "success" {
		t.Errorf("Incorrect node selected by ID!")
	}
}

/*
func TestHTMLFragmentEncoding(t *testing.T) {
	defer CheckXmlMemoryLeaks(t)

	input, output, error := getTestData(filepath.Join("tests", "document", "html_fragment_encoding"))

	if len(error) > 0 {
		t.Errorf("Error gathering test data for %v:\n%v\n", "html_fragment_encoding", error)
		t.FailNow()
	}

	expected := string(output)

	inputEncodingBytes := []byte("utf-8")

	buffer := make([]byte, 100)
	fragment, err := ParseFragment([]byte(input), inputEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes, buffer)

	if err != nil {
		println("WHAT")
		t.Error(err.Error())
	}

	if fragment.String() != expected {
		badOutput(fragment.String(), expected)
		t.Error("the output of the xml doc does not match")
	}

	fragment.Node.MyDocument().Free()
}
*/
