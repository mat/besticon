package css

import (
	"io/ioutil"
	"strings"
	"testing"
)

func read(filename string) string {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		panic("css2xpath test could not open a test file")
	}
	return string(contents)
}

func TestSelectors(t *testing.T) {
	cssSelectors := strings.Split(string(read("./test/inputs")), "\n")
	localXPaths := strings.Split(string(read("./test/outputs-local")), "\n")
	globalXPaths := strings.Split(string(read("./test/outputs-global")), "\n")

	for i, css := range cssSelectors {
		xpathG := strings.TrimSpace(Convert(css, GLOBAL))
		xpathL := strings.TrimSpace(Convert(css, LOCAL))
		if xpathG != strings.TrimSpace(globalXPaths[i]) {
			t.Errorf("IN:\t%s <GLOBAL>\nOUT:\t%s\nEXPECTED:\t%s\n", css, xpathG, globalXPaths[i])
		}
		if xpathL != strings.TrimSpace(localXPaths[i]) {
			t.Errorf("IN:\t%s <LOCAL>\nOUT:\t%s\nEXPECT:\t%s\n", css, xpathL, localXPaths[i])
		}
	}
}
