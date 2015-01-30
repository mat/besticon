package xml

import (
	"errors"
	"fmt"
	"github.com/moovweb/gokogiri/help"
	"github.com/moovweb/gokogiri/xpath"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func badOutput(actual string, expected string) {
	fmt.Printf("Got:\n[%v]\n", actual)
	fmt.Printf("Expected:\n[%v]\n", expected)
}

func RunTest(t *testing.T, suite string, name string, specificLogic func(t *testing.T, doc *XmlDocument), extraAssertions ...func(doc *XmlDocument) (string, string, string)) {
	defer CheckXmlMemoryLeaks(t)

	//println("Initiating test:" + suite + ":" + name)

	input, output, error := getTestData(filepath.Join("tests", suite, name))

	if len(error) > 0 {
		t.Errorf("Error gathering test data for %v:\n%v\n", name, error)
		t.FailNow()
	}

	expected := string(output)

	//println("Got raw input/output")

	doc, err := parseInput(input)

	if err != nil {
		t.Error(err.Error())
	}

	//println("parsed input")

	if specificLogic != nil {
		specificLogic(t, doc)
	}
	if doc.String() != expected {
		badOutput(doc.String(), expected)
		t.Error("the output of the xml doc does not match")
	}
	for _, extraAssertion := range extraAssertions {
		actual, expected, message := extraAssertion(doc)

		if actual != expected {
			badOutput(actual, expected)
			t.Error(message)
		}
	}

	doc.Free()
}

func RunBenchmark(b *testing.B, suite string, name string, specificLogic func(b *testing.B, doc *XmlDocument)) {
	b.StopTimer()

	//	defer CheckXmlMemoryLeaks(b)

	input, _, error := getTestData(filepath.Join("tests", suite, name))

	if len(error) > 0 {
		panic(fmt.Sprintf("Error gathering test data for %v:\n%v\n", name, error))
	}

	doc, err := parseInput(input)

	if err != nil {
		panic("Error:" + err.Error())
	}

	b.StartTimer()

	if specificLogic != nil {
		specificLogic(b, doc)
	}

	doc.Free()

	//	println("----------- END OF BENCHMARK -----------")
}

func parseInput(input interface{}) (*XmlDocument, error) {
	var realInput []byte

	switch thisInput := input.(type) {
	case []byte:
		realInput = thisInput
	case string:
		realInput = []byte(thisInput)
	default:
		return nil, errors.New("Unrecognized parsing input!")
	}

	doc, err := Parse(realInput, DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("parsing error:%v\n", err.Error()))
	}

	return doc, nil
}

func getTestData(name string) (input []byte, output []byte, error string) {
	var errorMessage string
	offset := "\t"
	inputFile := filepath.Join(name, "input.txt")

	input, err := ioutil.ReadFile(inputFile)

	if err != nil {
		errorMessage += fmt.Sprintf("%vCouldn't read test (%v) input:\n%v\n", offset, name, offset+err.Error())
	}

	output, err = ioutil.ReadFile(filepath.Join(name, "output.txt"))

	if err != nil {
		errorMessage += fmt.Sprintf("%vCouldn't read test (%v) output:\n%v\n", offset, name, offset+err.Error())
	}

	return input, output, errorMessage
}

func collectTests(suite string) (names []string, error string) {
	testPath := filepath.Join("tests", suite)
	entries, err := ioutil.ReadDir(testPath)

	if err != nil {
		return nil, fmt.Sprintf("Couldn't read tests:\n%v\n", err.Error())
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "_") || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if entry.IsDir() {
			names = append(names, filepath.Join(testPath, entry.Name()))
		}
	}

	return
}

func CheckXmlMemoryLeaks(t *testing.T) {
	// LibxmlCleanUpParser() should only be called once during the lifetime of the
	// program, but because there's no way to know when the last test of the suite
	// runs in go, we can't accurately call it strictly once, so just avoid calling
	// it for now because it's known to cause crashes if called multiple times.
	//help.LibxmlCleanUpParser()

	if !help.LibxmlCheckMemoryLeak() {
		t.Errorf("Memory leaks: %d!!!", help.LibxmlGetMemoryAllocation())
		help.LibxmlReportMemoryLeak()
	}
}

// This is a simple test implementation of the VariableScope interface.
// Currently it's os simple it ignores the namespace argument.
type SimpleVariableScope struct {
	variables map[string]interface{}
	functions map[string]xpath.XPathFunction
}

func (s *SimpleVariableScope) ResolveVariable(name, ns string) interface{} {
	v, _ := s.variables[name]
	return v
}

func (s *SimpleVariableScope) IsFunctionRegistered(name, ns string) bool {
	_, ok := s.functions[name]
	return ok
}
func (s *SimpleVariableScope) ResolveFunction(name, ns string) xpath.XPathFunction {
	f, _ := s.functions[name]
	return f
}

func newSimpleVariableScope() *SimpleVariableScope {
	s := &SimpleVariableScope{make(map[string]interface{}), make(map[string]xpath.XPathFunction)}
	return s
}
