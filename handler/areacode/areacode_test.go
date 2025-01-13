package areacode_test

import (
	"fmt"

	"github.com/Defacto2/server/handler/areacode"
)

func ExampleNAN_Valid() {
	fmt.Println(areacode.NAN(212).Valid())
	fmt.Println(areacode.NAN(999999).Valid())
	// Output: true
	// false
}

func ExampleTerritories() {
	t := areacode.Territories()
	name := t[0].Name
	alpha := t[0].AlphaCode
	area := t[0].AreaCodes
	fmt.Printf("%s %s %d\n", name, alpha, area)
	fmt.Println(len(t), "territories")
	// Output: Alabama AL [205]
	// 66 territories
}

func ExampleAlphaCodes() {
	codes := areacode.AlphaCodes()
	fmt.Println(codes[0])
	// Output: AB
}

func ExampleTerritoryByAlpha() {
	t := areacode.TerritoryByAlpha("CT")
	fmt.Println(t.Name, t.AlphaCode, t.AreaCodes)
	// Output: Connecticut CT [203]
}

func ExampleTerritoryByCode() {
	t := areacode.TerritoryByCode(212)
	fmt.Println(t[0].Name, t[0].AlphaCode, t[0].AreaCodes)

	t = areacode.TerritoryByCode(902)
	for _, v := range t {
		fmt.Println(v.Name, v.AlphaCode, v.AreaCodes)
	}
	// Output: New York NY [212 315 516 518 607 716 718 914 917]
	// Nova Scotia NS [902]
	// Prince Edward Island PE [902]
}

func ExampleTerritoryByName() {
	t := areacode.TerritoryByName("ontario")
	fmt.Println(t.AreaCodes)
	// Output: [416 519 613 705 807 905]
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
	// Output: {Texas TX [210 214 409 512 713 806 817 903 915]}
	// {Texas TX [210 214 409 512 713 806 817 903 915]}
	// {Texas TX [210 214 409 512 713 806 817 903 915]}
}

func ExampleLookups() {
	t := areacode.Lookups(817, "iowa", 202)
	for _, v := range t {
		fmt.Println(v)
	}
	// Output: {Texas TX [210 214 409 512 713 806 817 903 915]}
	// {Iowa IA [319 515 712]}
	// {District of Columbia DC [202]}
}

func ExampleNAN_HTML() {
	fmt.Println(areacode.NAN(403).HTML())
	// Output: <span>403 - Alberta (AB) + Yukon (YT)</span><br>
}

func ExampleAC_HTML() {
	fmt.Println(areacode.AC("AB").HTML())
	// Output: <span>AB (Alberta) - 403</span><br>
}

func ExampleTerritory_HTML() {
	t := areacode.TerritoryByCode(710)
	fmt.Println(t[0].HTML())
	// Output: <span>United States Government - 710</span><br>
}
