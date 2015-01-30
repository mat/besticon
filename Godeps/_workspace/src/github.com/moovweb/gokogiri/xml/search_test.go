package xml

import "testing"

func TestSearch(t *testing.T) {

	testLogic := func(t *testing.T, doc *XmlDocument) {
		root := doc.Root()
		result, _ := root.Search(".//*[@class]")
		if len(result) != 2 {
			t.Error("search at root does not match")
		}
		result, _ = root.Search("//*[@class]")
		if len(result) != 3 {
			t.Error("search at root does not match")
		}
		result, _ = doc.Search(".//*[@class]")
		if len(result) != 3 {
			t.Error("search at doc does not match")
		}
		result, _ = doc.Search(".//*[@class='shine']")
		if len(result) != 2 {
			t.Error("search with value at doc does not match")
		}
	}

	RunTest(t, "node", "search", testLogic)
}

func BenchmarkSearch(b *testing.B) {

	benchmarkLogic := func(b *testing.B, doc *XmlDocument) {
		root := doc.Root()

		for i := 0; i < b.N; i++ {
			root.Search(".//*[@class]")
		}
	}

	RunBenchmark(b, "node", "search", benchmarkLogic)
}

func BenchmarkBigDocDeepSearchyTagName(b *testing.B) {

	benchmarkLogic := func(b *testing.B, doc *XmlDocument) {

		for i := 0; i < b.N; i++ {
			doc.Search("//div")
		}
	}

	RunBenchmark(b, "document", "big_un", benchmarkLogic)
}

func BenchmarkBigDocPunctuatedDeepSearch(b *testing.B) {

	benchmarkLogic := func(b *testing.B, doc *XmlDocument) {

		for i := 0; i < b.N; i++ {
			doc.Search("//*[@class='filters']//div")
		}
	}

	RunBenchmark(b, "document", "big_un", benchmarkLogic)
}

func BenchmarkBigDocDeepSearchByID(b *testing.B) {

	benchmarkLogic := func(b *testing.B, doc *XmlDocument) {

		for i := 0; i < b.N; i++ {
			doc.Search("//*[@id='ppp']")
			//nodes, _ := doc.Search("//*[@id='ppp']")
			//fmt.Printf("%v\t", len(nodes))
		}
	}

	RunBenchmark(b, "document", "big_un", benchmarkLogic)
}

func BenchmarkBigDocDeepSearchByClass(b *testing.B) {

	benchmarkLogic := func(b *testing.B, doc *XmlDocument) {

		for i := 0; i < b.N; i++ {
			doc.Search("//*[@class]")
			//nodes, _ := doc.Search("//*[@class]")
			//fmt.Printf("%v\t", len(nodes))
		}
	}

	RunBenchmark(b, "document", "big_un", benchmarkLogic)
}

func BenchmarkBigDocDeepSearchByClassContains(b *testing.B) {

	benchmarkLogic := func(b *testing.B, doc *XmlDocument) {

		for i := 0; i < b.N; i++ {
			doc.Search("//*[contains(@class, 'header')]")
		}
	}

	RunBenchmark(b, "document", "big_un", benchmarkLogic)
}

func BenchmarkBigDocDeepSearchBySemanticClass(b *testing.B) {

	benchmarkLogic := func(b *testing.B, doc *XmlDocument) {

		for i := 0; i < b.N; i++ {
			doc.Search("//*[contains(concat(concat(' ', @class), ' '), concat(concat(' ','header'), ' '))]")
		}
	}

	RunBenchmark(b, "document", "big_un", benchmarkLogic)
}

func BenchmarkBigDocDeepSearchByText(b *testing.B) {

	benchmarkLogic := func(b *testing.B, doc *XmlDocument) {

		for i := 0; i < b.N; i++ {
			doc.Search("//*[text()='hey']")
		}
	}

	RunBenchmark(b, "document", "big_un", benchmarkLogic)
}

func BenchmarkBigDocDeepSearchByTextContains(b *testing.B) {

	benchmarkLogic := func(b *testing.B, doc *XmlDocument) {

		for i := 0; i < b.N; i++ {
			doc.Search("//*[contains(text(),'hey')]")
		}
	}

	RunBenchmark(b, "document", "big_un", benchmarkLogic)
}

func BenchmarkBigDocSearchAncestorAxes(b *testing.B) {

	benchmarkLogic := func(b *testing.B, doc *XmlDocument) {
		elem, _ := doc.Search("//*[@id='ppp']")
		for i := 0; i < b.N; i++ {
			elem[0].Search("ancestor::html")
		}
	}

	RunBenchmark(b, "document", "big_un", benchmarkLogic)
}

func BenchmarkBigDocSearchLongTraverseUpToRoot(b *testing.B) {

	benchmarkLogic := func(b *testing.B, doc *XmlDocument) {
		elem, _ := doc.Search("//*[@id='ppp']")

		for i := 0; i < b.N; i++ {
			elem[0].Search("../../../../../../../../..")
		}
	}

	RunBenchmark(b, "document", "big_un", benchmarkLogic)
}

func BenchmarkBigDocSearchShortTraverseUpToRoot(b *testing.B) {

	benchmarkLogic := func(b *testing.B, doc *XmlDocument) {
		elem, _ := doc.Search("//*[@id='ppp']")

		for i := 0; i < b.N; i++ {
			elem[0].Search("../../../..")
		}
	}

	RunBenchmark(b, "document", "big_un", benchmarkLogic)
}
