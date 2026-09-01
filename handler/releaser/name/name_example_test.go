package name_test

import (
	"fmt"

	"github.com/Defacto2/server/handler/releaser/name"
)

func ExampleHumanize() {
	s, _ := name.Humanize("defacto2")
	fmt.Println(s)

	s, _ = name.Humanize("razor-1911-demo")
	fmt.Println(s)

	s, _ = name.Humanize("razor-1911-demo*trsi")
	fmt.Println(s)
	// Output:
	// defacto2
	// razor 1911 demo
	// razor 1911 demo, trsi
}

func ExampleHumanize_error() {
	_, err := name.Humanize("razor-1911-demo#trsi")
	if err != nil {
		fmt.Println(err)
	}
	// Output:
	// releaser name: path has invalid characters
}

func ExampleCopy() {
	find := name.Path("surprise-productions")
	for key, val := range name.Copy() {
		if key == find {
			fmt.Println(val)
		}
	}
	// Output: Surprise! Productions
}

func ExampleObfuscate() {
	obf := name.Obfuscate("ACiD Productions")
	if !name.Valid(obf) {
		fmt.Println("invalid")
	} else {
		fmt.Println(string(obf))
	}
	// Output: acid-productions
}

func ExampleList() {
	uri := "defacto2net"
	for key, val := range name.Copy() {
		if key == name.Path(uri) {
			fmt.Println(val)
		}
	}
	// Output: Defacto2 website
}

func ExampleUpper() {
	uri := "beer"
	for key, val := range name.Upper() {
		if key == name.Path(uri) {
			fmt.Println(val)
		}
	}
	// Output: BEER
}

func ExampleString() {
	fmt.Println(name.String("acid-productions"))
	// Output: ACiD Productions
}

func ExampleString_unlisted() {
	s := name.String("defacto2")
	fmt.Println(len(s))
	// Output: 0
}

func ExampleValid() {
	fmt.Println(name.Valid("defacto2"))

	fmt.Println(name.Valid("Defacto2"))
	// Output: true
	// false
}
