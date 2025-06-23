package main

import (
	"fmt"

	"github.com/m-m-f/gowiki"
)

type K struct {
}

func (k K) Get(page gowiki.WikiLink) (string, error) {
	return "", nil
}
func main() {
	t := K{}
	art, err := gowiki.ParseArticle("Test", `
== Test ==
* Hello World
* [[Test2]]
* [[Test3|Test 3]]
* [[Test4|Test 4]]
* {{Test5}}
- 8988
<img jhjs>Test</img>
- [[Portal:Test|Portal]]
{|
|-
| Test
| [[File:Test.jpg|thumb|Test]]
|- Test2
|}
	`, t)
	if err != nil {
		panic(err)
	}
	fmt.Println("Title:", art.Title)
	fmt.Println("Abstract:", art.GetText())
}
