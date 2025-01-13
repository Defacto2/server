package areacode_test

import (
	"fmt"

	"github.com/Defacto2/server/handler/areacode"
)

func ExampleQuery() {
	fmt.Println(areacode.Query("az").Terr)
	fmt.Println(areacode.Query("arizona").Terr[0].HTML())
	fmt.Println(areacode.Query("ut").Terr[0].HTML())
	fmt.Println(areacode.Query("602").AreaCode.HTML())
	// Output: [{Arizona AZ [602]}]
	// <span>Arizona (AZ)  - 602</span><br>
	// <span>Utah (UT)  - 801</span><br>
	// <span>602 - Arizona (AZ)</span><br>

}

func ExampleQueries() {
	q := areacode.Queries("az", "arizona", "ut", "602")
	for _, result := range q {
		s := result.AreaCode.HTML()
		if s != "" {
			fmt.Println(s)
		}
		for _, t := range result.Terr {
			fmt.Println(t.HTML())
		}
	}
	// Output: <span>Arizona (AZ)  - 602</span><br>
	// <span>Arizona (AZ)  - 602</span><br>
	// <span>Utah (UT)  - 801</span><br>
	// <span>602 - Arizona (AZ)</span><br>
}

func ExampleNAN_Valid() {
	fmt.Println(areacode.NAN(212).Valid())
	fmt.Println(areacode.NAN(999999).Valid())
	// Output: true
	// false
}

func ExampleTerritories() {
	t := areacode.Territories()
	name := t[0].Name
	alpha := t[0].Abbreviation
	area := t[0].AreaCodes
	fmt.Printf("%s %s %d\n", name, alpha, area)
	fmt.Println(len(t), "territories")
	// Output: Alabama AL [205]
	// 66 territories
}

func ExampleAbbreviations() {
	codes := areacode.Abbreviations()
	fmt.Println(codes[0])
	// Output: AB
}

func ExampleTerritoryByAbbr() {
	t := areacode.TerritoryByAbbr("CT")
	fmt.Println(t.Name, t.Abbreviation, t.AreaCodes)
	// Output: Connecticut CT [203]
}

func ExampleTerritoryByCode() {
	t := areacode.TerritoryByCode(212)
	fmt.Println(t[0].Name, t[0].Abbreviation, t[0].AreaCodes)

	t = areacode.TerritoryByCode(902)
	for _, v := range t {
		fmt.Println(v.Name, v.Abbreviation, v.AreaCodes)
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

func ExampleTerritoryContains() {
	t := areacode.TerritoryContains("south")
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

func ExampleAbbreviation_HTML() {
	fmt.Println(areacode.Abbreviation("AB").HTML())
	// Output: <span>AB (Alberta) - 403</span><br>
}

func ExampleTerritory_HTML() {
	t := areacode.TerritoryByCode(710)
	fmt.Println(t[0].HTML())
	// Output: <span>United States Government - 710</span><br>
}
