package releaser_test

import (
	"fmt"

	"github.com/Defacto2/server/handler/releaser"
)

func ExampleCell() {
	s := "  Defacto2  demo  group."
	fmt.Println(releaser.Cell(s))
	// Output: DEFACTO2 DEMO GROUP
}

func ExampleClean() {
	s := "  Defacto2  demo  group."
	fmt.Println(releaser.Clean(s))
	// Output: Defacto2 Demo Group
}

func ExampleHumanize() {
	path := "razor-1911-demo"
	fmt.Println(releaser.Humanize(path))
	// Output: Razor 1911 Demo
}

func ExampleLink() {
	path := "class*paradigm*razor-1911-demo"
	fmt.Println(releaser.Link(path))
	// Output: Class + Paradigm + Razor 1911 Demo
}

func ExampleObfuscate() {
	s := "Defacto2 Demo Group."
	fmt.Println(releaser.Obfuscate(s))
	// Output: defacto2-demo-group
}

func ExampleIndex() {
	fmt.Println(releaser.Index("united-software-association*fairlight"))
	fmt.Println(releaser.Index("class*paradigm*razor-1911"))
	fmt.Println(releaser.Index("coop"))
	// Output: UNITED SOFTWARE ASSOCIATION, FAIRLIGHT
	// CLASS, PARADIGM, RAZOR 1911
	// COOP
}
