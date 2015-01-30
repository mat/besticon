package html

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestParseDocument_CP1252(t *testing.T) {
	input, err := ioutil.ReadFile("./tests/document/encoding/input.html")
	if err != nil {
		t.Error("err:", err.Error())
		return
	}
	doc, err := Parse(input, []byte("windows-1252"), nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		t.Error("err:", err.Error())
		return
	}
	out := doc.String()
	if index := bytes.IndexByte([]byte(out), byte(146)); index >= 0 {
		t.Error("the output is not properly encoded")
	}
	doc.Free()
	CheckXmlMemoryLeaks(t)
}

func TestParseDocumentWithInOutEncodings(t *testing.T) {
	println("Starting to read input file.")
	input, err := ioutil.ReadFile("./tests/document/encoding/input.html")
	if err != nil {
		t.Error("err:", err.Error())
		return
	}
	println("Succesfully read input file, beginning parsing.")
	doc, err := Parse(input, []byte("windows-1252"), nil, DefaultParseOption, []byte("windows-1252"))
	if err != nil {
		t.Error("err:", err.Error())
		return
	}
	println("Successfully parsed, getting document as a string...")
	out := doc.String()
	if index := bytes.IndexByte([]byte(out), byte(146)); index < 0 {
		t.Error("the output is not properly encoded")
	}

	println("Test complete, about to free document.")
	doc.Free()
	println("Successfully freed document, checking for memory leaks...")
	CheckXmlMemoryLeaks(t)
	println("Finished checking for leaks.")
}
