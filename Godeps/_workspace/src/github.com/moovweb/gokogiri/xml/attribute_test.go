package xml

import "testing"
import "fmt"

func TestSetValue(t *testing.T) {
	defer CheckXmlMemoryLeaks(t)
	doc, err := Parse([]byte("<foo id=\"a\" myname=\"ff\"><bar class=\"shine\"/></foo>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}
	root := doc.Root()
	attributes := root.Attributes()
	if len(attributes) != 2 || attributes["myname"].String() != "ff" {
		fmt.Printf("%v, %q\n", attributes, attributes["myname"].String())
		t.Error("root's attributes do not match")
	}
	child := root.FirstChild()
	childAttributes := child.Attributes()
	if len(childAttributes) != 1 || childAttributes["class"].String() != "shine" {
		t.Error("child's attributes do not match")
	}
	attributes["myname"].SetValue("new")
	expected :=
		`<foo id="a" myname="new">
  <bar class="shine"/>
</foo>`
	if root.String() != expected {
		println("got:\n", root.String())
		println("expected:\n", expected)
		t.Error("root's new attr do not match")
	}
	attributes["id"].Remove()
	expected =
		`<foo myname="new">
  <bar class="shine"/>
</foo>`

	if root.String() != expected {
		println("got:\n", root.String())
		println("expected:\n", expected)
		t.Error("root's remove attr do not match")
	}
	doc.Free()
}

func TestSetAttribute(t *testing.T) {
	defer CheckXmlMemoryLeaks(t)
	doc, err := Parse([]byte("<foo id=\"a\" myname=\"ff\"><bar class=\"shine\"/></foo>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}
	root := doc.Root()
	attributes := root.Attributes()
	if len(attributes) != 2 || attributes["myname"].String() != "ff" {
		fmt.Printf("%v, %q\n", attributes, attributes["myname"].String())
		t.Error("root's attributes do not match")
	}

	root.SetAttr("id", "cooler")
	root.SetAttr("id2", "hot")
	root.SetAttr("id3", "")
	expected :=
		`<foo id="cooler" myname="ff" id2="hot" id3="">
  <bar class="shine"/>
</foo>`
	if root.String() != expected {
		println("got:\n", root.String())
		println("expected:\n", expected)
		t.Error("root's new attr do not match")
	}
	if root.Attr("id3") != "" {
		println("got:\n", root.Attr("id3"))
		println("expected:\n", "")
		t.Error("root's attr should have empty val")
	}
	if root.Attribute("id3") == nil {
		t.Error("root's attr should not be nil")
	}
	doc.Free()
}

func TestSetEmptyAttribute(t *testing.T) {
	defer CheckXmlMemoryLeaks(t)
	doc, err := Parse([]byte("<foo id=\"a\" myname=\"ff\"><bar class=\"shine\"/></foo>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}
	root := doc.Root()
	attributes := root.Attributes()
	if len(attributes) != 2 || attributes["myname"].String() != "ff" {
		fmt.Printf("%v, %q\n", attributes, attributes["myname"].String())
		t.Error("root's attributes do not match")
	}

	root.SetAttr("", "cool")
	expected :=
		`<foo id="a" myname="ff" ="cool">
  <bar class="shine"/>
</foo>`
	if root.String() != expected {
		println("got:\n", root.String())
		println("expected:\n", expected)
		t.Error("root's new attr do not match")
	}

	root.SetAttr("", "")
	expected =
		`<foo id="a" myname="ff" ="">
  <bar class="shine"/>
</foo>`
	if root.String() != expected {
		println("got:\n", root.String())
		println("expected:\n", expected)
		t.Error("root's new attr do not match")
	}
	doc.Free()
}
