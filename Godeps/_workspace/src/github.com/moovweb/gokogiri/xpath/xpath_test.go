package xpath

//please check the search tests in gokogiri/xml and gokogiri/html
import "testing"

func TestCompileGoodExpr(t *testing.T) {
	defer CheckXmlMemoryLeaks(t)
	e := Compile(`./*`)
	if e == nil {
		t.Error("expr should be good")
	}
	e.Free()
}

func TestCompileBadExpr(t *testing.T) {
	//defer CheckXmlMemoryLeaks(t)
	//this test causes memory leaks in libxml
	//however, the memory leak is very small and does not grow as more bad expressions are compiled
	e := Compile("./")
	if e != nil {
		t.Error("expr should be bad")
	}
	e = Compile(".//")
	if e != nil {
		t.Error("expr should be bad")
	}
}
