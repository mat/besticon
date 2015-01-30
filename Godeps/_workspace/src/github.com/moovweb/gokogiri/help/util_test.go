package help

import "testing"

func CheckXmlMemoryLeaks(t *testing.T) {
	// LibxmlCleanUpParser() should only be called once during the lifetime of the
	// program, but because there's no way to know when the last test of the suite
	// runs in go, we can't accurately call it strictly once, so just avoid calling
	// it for now because it's known to cause crashes if called multiple times.
	//LibxmlCleanUpParser()

	if !LibxmlCheckMemoryLeak() {
		t.Errorf("Memory leaks: %d!!!", LibxmlGetMemoryAllocation())
		LibxmlReportMemoryLeak()
	}
}
