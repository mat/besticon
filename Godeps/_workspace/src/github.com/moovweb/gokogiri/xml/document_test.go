package xml

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDocuments(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	tests, err := collectTests("document")

	if len(err) > 0 {
		t.Errorf(err)
	}

	errors := make([]string, 0)

	print("\nTesting: Basic Parsing [")

	for _, test := range tests {
		error := RunDocumentParseTest(t, test)

		if error != nil {
			errors = append(errors, fmt.Sprintf("Test %v failed:\n%v\n", test, *error))
			print("F")
		} else {
			print(".")
		}
	}

	println("]")

	if len(errors) > 0 {
		errorMessage := "\t" + strings.Join(strings.Split(strings.Join(errors, "\n\n"), "\n"), "\n\t")
		t.Errorf("\nSome tests failed! (%d passed / %d total) :\n%v", len(tests)-len(errors), len(tests), errorMessage)
	} else {
		fmt.Printf("\nAll (%d) tests passed!\n", len(tests))
	}
}

func TestBufferedDocuments(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	tests, err := collectTests("document")

	if len(err) > 0 {
		t.Errorf(err)
	}

	errors := make([]string, 0)

	print("\nTesting: Buffered Parsing [")

	for _, test := range tests {
		error := RunParseDocumentWithBufferTest(t, test)

		if error != nil {
			errors = append(errors, fmt.Sprintf("Test %v failed:\n%v\n", test, *error))
			print("F")
		} else {
			print(".")
		}
	}

	println("]")

	if len(errors) > 0 {
		errorMessage := "\t" + strings.Join(strings.Split(strings.Join(errors, "\n\n"), "\n"), "\n\t")
		t.Errorf("\nSome tests failed! (%d passed / %d total) :\n%v", len(tests)-len(errors), len(tests), errorMessage)
	} else {
		fmt.Printf("\nAll (%d) tests passed!\n", len(tests))
	}
}

func RunParseDocumentWithBufferTest(t *testing.T, name string) (error *string) {
	var errorMessage string
	offset := "\t"

	defer CheckXmlMemoryLeaks(t)

	input, output, dataError := getTestData(name)

	if len(dataError) > 0 {
		errorMessage += dataError
	}

	buffer := make([]byte, 500000)

	doc, err := Parse(input, DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	if err != nil {
		errorMessage = fmt.Sprintf("parsing error:%v\n", err)
	}

	if string(doc.ToBuffer(buffer)) != string(output) {
		formattedOutput := offset + strings.Join(strings.Split("["+doc.String()+"]", "\n"), "\n"+offset)
		formattedExpectedOutput := offset + strings.Join(strings.Split("["+string(output)+"]", "\n"), "\n"+offset)
		errorMessage = fmt.Sprintf("%v-- Got --\n%v\n%v-- Expected --\n%v\n", offset, formattedOutput, offset, formattedExpectedOutput)
	}
	doc.Free()

	if len(errorMessage) > 0 {
		return &errorMessage
	}
	return nil

}

func RunDocumentParseTest(t *testing.T, name string) (error *string) {

	var errorMessage string
	offset := "\t"

	defer CheckXmlMemoryLeaks(t)

	input, output, dataError := getTestData(name)

	if len(dataError) > 0 {
		errorMessage += dataError
	}

	doc, err := Parse(input, DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	if err != nil {
		errorMessage = fmt.Sprintf("parsing error:%v\n", err)
	}

	if doc.String() != string(output) {
		formattedOutput := offset + strings.Join(strings.Split("["+doc.String()+"]", "\n"), "\n"+offset)
		formattedExpectedOutput := offset + strings.Join(strings.Split("["+string(output)+"]", "\n"), "\n"+offset)
		errorMessage = fmt.Sprintf("%v-- Got --\n%v\n%v-- Expected --\n%v\n", offset, formattedOutput, offset, formattedExpectedOutput)
		testOutput := filepath.Join(name, "test_output.txt")
		ioutil.WriteFile(testOutput, []byte(doc.String()), os.FileMode(0666))
		errorMessage += fmt.Sprintf("%v Output test output to: %v\n", offset, testOutput)
	}
	doc.Free()

	if len(errorMessage) > 0 {
		return &errorMessage
	}
	return nil

}

func BenchmarkDocOutput(b *testing.B) {
	b.StopTimer()

	tests, err := collectTests("document")

	if len(err) > 0 {
		fmt.Printf(err)
		return
	}

	docs := make([]*XmlDocument, 0)

	for _, testName := range tests {

		input, _, dataError := getTestData(testName)

		if len(dataError) > 0 {
			fmt.Printf(dataError)
			return
		}
		doc, err := Parse(input, DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

		if err != nil {
			fmt.Printf("parsing error:%v\n", err)
			return
		}
		docs = append(docs, doc)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		for index, _ := range tests {
			_ = docs[index].String()
		}
	}

}

func BenchmarkDocOutputToBuffer(b *testing.B) {
	b.StopTimer()

	tests, err := collectTests("document")

	if len(err) > 0 {
		fmt.Printf(err)
		return
	}

	docs := make([]*XmlDocument, 0)

	for _, testName := range tests {

		input, _, dataError := getTestData(testName)

		if len(dataError) > 0 {
			fmt.Printf(dataError)
			return
		}
		doc, err := Parse(input, DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

		if err != nil {
			fmt.Printf("parsing error:%v\n", err)
			return
		}
		docs = append(docs, doc)
	}

	buffer := make([]byte, 500*1024)

	b.StartTimer()

	for i := 0; i < b.N; i++ {

		for index, _ := range tests {

			_ = docs[index].ToBuffer(buffer)

		}
	}

}

func TestRemoveNamespaces(t *testing.T) {
	xml := "<SOAP-ENV:Envelope xmlns:SOAP-ENV=\"http://schemas.xmlsoap.org/soap/envelope/\"><SOAP-ENV:Body><m:setPresence xmlns:m=\"http://schemas.microsoft.com/winrtc/2002/11/sip\"><m:presentity m:uri=\"test\"><m:availability m:aggregate=\"300\" m:description=\"online\"/><m:activity m:aggregate=\"400\" m:description=\"Active\"/><deviceName xmlns=\"http://schemas.microsoft.com/2002/09/sip/client/presence\" name=\"WIN-0DDABKC1UI8\"/></m:presentity></m:setPresence></SOAP-ENV:Body></SOAP-ENV:Envelope>"
	xml_no_namespace := "<Envelope><Body><setPresence><presentity uri=\"test\"><availability aggregate=\"300\" description=\"online\"/><activity aggregate=\"400\" description=\"Active\"/><deviceName name=\"WIN-0DDABKC1UI8\"/></presentity></setPresence></Body></Envelope>"

	doc, _ := Parse([]byte(xml), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	doc.Root().RecursivelyRemoveNamespaces()
	doc2, _ := Parse([]byte(xml_no_namespace), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	output := fmt.Sprintf("%v", doc)
	output_no_namespace := fmt.Sprintf("%v", doc2)
	if output != output_no_namespace {
		t.Errorf("Xml namespaces not removed!")
	}
}

func TestRemoveDefaultNamespace(t *testing.T) {
	xml := `
<body xmlns="http://jabber.org/protocol/httpbind" xmlns:stream="http://etherx.jabber.org/streams" to="127.0.0.1" rid="3" sid="0acad5262d995374">
  <iq id="2" type="get" from="">
    <query xmlns="jabber:iq:auth">
      <username>xyz</username>
    </query>
  </iq>
</body>
`

	xml_no_namespace := `
<body xmlns:stream="http://etherx.jabber.org/streams" to="127.0.0.1" rid="3" sid="0acad5262d995374">
  <iq id="2" type="get" from="">
    <query xmlns="jabber:iq:auth">
      <username>xyz</username>
    </query>
  </iq>
</body>
`
	doc, _ := Parse([]byte(xml), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	doc.Root().RemoveDefaultNamespace()
	doc2, _ := Parse([]byte(xml_no_namespace), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	output := fmt.Sprintf("%v", doc)
	output_no_namespace := fmt.Sprintf("%v", doc2)
	if output != output_no_namespace {
		t.Errorf("Default namespace not removed!")
	}
}

func TestNodeById(t *testing.T) {
	xml := "<!DOCTYPE doc [\n<!ELEMENT para (#PCDATA)>\n<!ATTLIST para label ID #IMPLIED>\n]>\n<doc><para>Failed</para><para label=\"W11\">Success</para></doc>"

	doc, _ := Parse([]byte(xml), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	p := doc.NodeById("W11")

	if p == nil {
		t.Errorf("Did not find node by ID!")
		return
	}

	output := fmt.Sprintf("%v", p.Content())
	if output != "Success" {
		t.Errorf("Incorrect node selected by ID!")
	}
}

func TestUnparsedEntityURI(t *testing.T) {
	xml := "<!DOCTYPE doc [\n<!ELEMENT doc EMPTY>\n<!ATTLIST doc attr ENTITY #REQUIRED>\n<!NOTATION JPEG SYSTEM \"jpeg.exe\">\n<!ENTITY test SYSTEM \"test.jpg\" NDATA JPEG>\n]>\n<doc attr='test'/>"
	doc, _ := Parse([]byte(xml), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	expected := "test.jpg"
	actual := doc.UnparsedEntityURI("test")

	if actual == "" {
		t.Errorf("Did not find unparsed entity 'test'")
		return
	}

	if actual != expected {
		t.Errorf("Expected '%s', but got '%s' calling doc.UnparsedEntityURI", expected, actual)
	}
}
