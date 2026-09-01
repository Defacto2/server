package lism_test

import (
	"fmt"
	"slices"

	"github.com/Defacto2/server/handler/releaser/lism"
)

func ExampleFind() {
	fmt.Println(lism.Find("Firm"))
	// Output: [the-firm]
}

func ExampleInitialism() {
	fmt.Println(lism.Initialism("defacto2"))
	// Output: [DF2 DF]
}

func ExampleCopy() {
	const want = "USA"
	for path, names := range lism.Copy() {
		if slices.Contains(names, want) {
			fmt.Printf("The uri for %v: %v\n", want, path)
		}
	}
	// Output: The uri for USA: united-software-association*fairlight
}

func ExampleExist() {
	fmt.Println(lism.Exist("defacto2"))
	// Output: true
}

func ExampleString() {
	fmt.Println(lism.String("the-firm")) // FiRM, FRM

	fmt.Println(lism.String("united-software-association*fairlight")) // USA
	// Output: FiRM, FRM
	// USA/Fairlight, USA/FLT, USA
}
