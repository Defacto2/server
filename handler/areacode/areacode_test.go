package areacode_test

import (
	"fmt"

	"github.com/Defacto2/server/handler/areacode"
)

func ExampleNANPCode_Valid() {
	fmt.Println(areacode.NANPCode(212).Valid())
	fmt.Println(areacode.NANPCode(999999).Valid())
	// Output: true
	// false
}

func ExampleTerritories() {
	t := areacode.Territories()
	name := t[0].Name
	alpha := t[0].AlphaCode
	area := t[0].AreaCode
	fmt.Printf("%s %s %d\n", name, alpha, area)
	fmt.Println(len(t), "territories")
	// Output: Alabama AL [205]
	// 65 territories
}

func ExampleAlphaCodes() {
	codes := areacode.AlphaCodes()
	fmt.Println(codes[0])
	// Output: AB
}

func ExampleTerritoryByAlpha() {
	t := areacode.TerritoryByAlpha("CT")
	fmt.Println(t.Name, t.AlphaCode, t.AreaCode)
	// Output: Connecticut CT [203]
}

func ExampleTerritoryByCode() {
	t := areacode.TerritoryByCode(212)
	fmt.Println(t[0].Name, t[0].AlphaCode, t[0].AreaCode)

	t = areacode.TerritoryByCode(902)
	for _, v := range t {
		fmt.Println(v.Name, v.AlphaCode, v.AreaCode)
	}
	// Output: New York NY [212 315 516 518 607 716 718 914]
	// Nova Scotia NS [902]
	// Prince Edward Island PE [902]
}

func ExampleTerritoryByName() {
	t := areacode.TerritoryByName("ontario")
	fmt.Println(t.AreaCode)
	// Output: [416 519]
}

func ExampleTerritoryMatch() {
	t := areacode.TerritoryMatch("south")
	for _, v := range t {
		fmt.Println(v)
	}
	// Output: {South Carolina SC [803]}
	// {South Dakota SD [605]}
}

func ExampleLookup() {
	t := areacode.Lookup("texas")
	fmt.Println(t[0])

	t = areacode.Lookup("tx")
	fmt.Println(t[0])

	t = areacode.Lookup(214)
	fmt.Println(t[0])
	// Output: {Texas TX [214 409 512 713 806 817 903 915]}
	// {Texas TX [214 409 512 713 806 817 903 915]}
	// {Texas TX [214 409 512 713 806 817 903 915]}
}
