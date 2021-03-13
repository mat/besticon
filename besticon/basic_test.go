package besticon

import (
	"fmt"
	"testing"
)

type testBasic struct {
	url   string
	icons []testBasicIcon
}

type testBasicIcon struct {
	url    string
	width  int
	format string
}

//
// test runner
//

func TestDomains(t *testing.T) {
	tests := []testBasic{
		// aol - has one pixel gifs
		{"http://aol.com", []testBasicIcon{
			{"http://www.aol.com/favicon.ico", 32, "ico"},
			{"http://www.aol.com/favicon.ico?v=2", 32, "ico"},
		}},

		// aws.amazon.com - this one has a base url
		{"http://aws.amazon.com", []testBasicIcon{
			{"http://a0.awsstatic.com/main/images/site/touch-icon-ipad-144-precomposed.png", 144, "png"},
			{"http://a0.awsstatic.com/main/images/site/touch-icon-iphone-114-precomposed.png", 114, "png"},
			{"http://a0.awsstatic.com/main/images/site/favicon.ico", 16, "ico"},
			{"http://aws.amazon.com/favicon.ico", 16, "ico"},
		}},

		// daringfireball
		{"http://daringfireball.net", []testBasicIcon{
			{"http://daringfireball.net/graphics/apple-touch-icon.png", 314, "png"},
			{"http://daringfireball.net/favicon.ico", 32, "ico"},
			{"http://daringfireball.net/graphics/favicon.ico?v=005", 32, "ico"},
		}},

		// github
		{"http://github.com", []testBasicIcon{
			// later - for svg
			// {"https://assets-cdn.github.com/pinned-octocat.svg", 9999, "svg" },
			{"https://github.com/apple-touch-icon-144.png", 144, "png"},
			{"https://github.com/apple-touch-icon.png", 120, "png"},
			{"https://github.com/apple-touch-icon-114.png", 114, "png"},
			{"https://github.com/apple-touch-icon-precomposed.png", 57, "png"},
			{"https://assets-cdn.github.com/favicon.ico", 32, "ico"},
			{"https://github.com/favicon.ico", 32, "ico"},
		}},

		// kicktipp.de
		{"http://kicktipp.de", []testBasicIcon{
			{"http://info.kicktipp.de/assets/img/jar_cb333387130/assets/img/logos/apple-touch-icon-57x57-precomposed.png", 57, "png"},
			{"http://www.kicktipp.de/apple-touch-icon-precomposed.png", 57, "png"},
			{"http://www.kicktipp.de/apple-touch-icon.png", 57, "png"},
			{"http://www.kicktipp.de/favicon.ico", 32, "gif"},
			{"http://info.kicktipp.de/assets/img/jar_cb1652512069/assets/img/logos/favicon.png", 16, "png"},
		}},

		// netflix - has cookie redirects
		{"http://netflix.com", []testBasicIcon{
			{"https://assets.nflxext.com/us/ffe/siteui/common/icons/nficon2016.png", 64, "png"},
			{"https://assets.nflxext.com/us/ffe/siteui/common/icons/nficon2016.ico", 64, "ico"},
			{"https://www.netflix.com/favicon.ico", 64, "ico"},
		}},
	}

	for _, test := range tests {
		fmt.Println("===========================================")
		fmt.Printf("= %s \n", test.url)
		fmt.Println("===========================================")

		// no errors expected
		actualIcons, _, err := fetchIconsWithVCR2(test.url)
		assertEquals(t, nil, err)

		// now compare icons
		assertEquals(t, len(test.icons), len(actualIcons))
		for i := range test.icons {
			assertEquals(t, test.icons[i].url, actualIcons[i].URL)
			assertEquals(t, test.icons[i].width, actualIcons[i].Width)
			assertEquals(t, test.icons[i].format, actualIcons[i].Format)
		}
	}
}
