package name_test

import (
	"fmt"

	"github.com/Defacto2/server/handler/releaser/name"
)

func ExampleAll() {
	all := name.All()
	fmt.Println(all["pump"])
	fmt.Println(all["intel"])
	fmt.Println(all["beer"])
	// Output:
	// Pump!
	// intel
	// BEER
}

func ExampleClean() {
	s := name.Clean("Defacto2! 1996", func(b byte) bool {
		return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
	})
	fmt.Println(s)
	// Output: Defacto
}

func ExampleCopy() {
	fmt.Println(name.Copy()["surprise-productions"])
	// Output: Surprise! Productions
}

func ExampleFind() {
	fmt.Println(name.Find("OldSkool"))
	fmt.Println(name.Find("Defacto2"))
	// Output:
	// oldskool
	//
}

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

func ExampleList() {
	list := name.List{
		"defacto2net": "Defacto2 website",
	}
	fmt.Println(list["defacto2net"])
	// Output: Defacto2 website
}

func ExampleLower() {
	fmt.Println(name.Lower()["intel"])
	// Output: intel
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

func ExampleString() {
	fmt.Println(name.String("acid-productions"))
	// Output: ACiD Productions
}

func ExampleString_unlisted() {
	s := name.String("defacto2")
	fmt.Println(len(s))
	// Output: 0
}

func ExampleUpper() {
	fmt.Println(name.Upper()["beer"])
	// Output: BEER
}

func ExampleValid() {
	fmt.Println(name.Valid("defacto2"))
	fmt.Println(name.Valid("Defacto2"))
	// Output:
	// true
	// false
}
