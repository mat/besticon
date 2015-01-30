package besticon

import "fmt"

func ExampleFetchBestIcon() {
	i, err := FetchBestIcon("github.com")
	if err != nil {
		fmt.Println("could not fetch icon: ", err)
	} else {
		fmt.Println("found an icon at ", i.URL)
	}
}
