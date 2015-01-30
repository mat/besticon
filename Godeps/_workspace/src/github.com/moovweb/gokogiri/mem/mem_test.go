package mem

import "testing"

const EXPECTED_VERSION = "2.7.8"

func TestLibxml(t *testing.T) {
	if LIBXML_VERSION != EXPECTED_VERSION {
		t.Fatal("Invalid libxml version got:", LIBXML_VERSION, "expected", EXPECTED_VERSION)
	}
	if AllocSize() != 0 {
		t.Fatal(AllocSize(), "remaining allocations")
	}
}
