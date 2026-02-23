// Package areacode provides information about North American Numbering Plan telephone area codes.
package areacode

import (
	"cmp"
	"fmt"
	"html/template"
	"slices"
	"strconv"
	"strings"
)

const limit = 999

// NAN represents a North American Numbering Plan area code.
type NAN int

// Valid returns true if the NANP code is a valid area code.
func (c NAN) Valid() bool {
	ac := AreaCodes()
	return slices.Contains(ac, c)
}

// HTML returns the NANP code as an HTML span element.
//
// Example:
//
//	212 - New York (NY)
func (c NAN) HTML() template.HTML {
	empty := template.HTML(``)
	if !c.Valid() {
		return empty
	}
	territories := TerritoryByCode(c)
	if len(territories) == 0 {
		return empty
	}
	var html strings.Builder
	html.WriteString("<span>")
	for i, val := range territories {
		abbr := val.Abbreviation
		if i == 0 {
			s := fmt.Sprintf(`%d - %s`, c, val.Name)
			if len(abbr) > 0 {
				s += fmt.Sprintf(` (%s)`, abbr)
			}
			html.WriteString(s)
			continue
		}
		html.WriteString(" + " + val.Name)
		if len(abbr) > 0 {
			fmt.Fprintf(&html, ` (%s)`, abbr)
		}
	}
	if note, ok := Notes()[c]; ok {
		fmt.Fprintf(&html, " <small><em>%s</em></small>", note)
	}
	html.WriteString("</span><br>")
	return template.HTML(html.String())
}

// Abbreviation represents a two-letter abbreviation for a territory in the North American Numbering Plan.
type Abbreviation string

func (a Abbreviation) HTML() template.HTML {
	t := TerritoryByAbbr(a)
	var html strings.Builder
	html.WriteString("<span>")
	html.WriteString(string(t.Abbreviation) + " (" + t.Name + ")")
	if len(t.AreaCodes) == 0 {
		html.WriteString(" - n/a</span><br>")
		return template.HTML(html.String())
	}
	html.WriteString(" - ")
	for i, ac := range t.AreaCodes {
		if i > 0 {
			html.WriteString(", ")
		}
		html.WriteString(strconv.Itoa(int(ac)))
	}
	html.WriteString("</span><br>")
	return template.HTML(html.String())
}

// Territory represents a territory in the North American Numbering Plan.
type Territory struct {
	Name         string       // Name of the state, province, or territory.
	Abbreviation Abbreviation // Two-letter abbreviation.
	AreaCodes    []NAN        // Three-digit NAN code used for telephone area codes.
}

func (t Territory) HTML() template.HTML {
	var html strings.Builder
	html.WriteString("<span>" + t.Name)
	if len(t.Abbreviation) > 0 {
		html.WriteString(" (" + string(t.Abbreviation) + ") ")
	}
	// join area codes with commas
	if len(t.AreaCodes) == 0 {
		html.WriteString("- n/a</span><br>")
		return template.HTML(html.String())
	}
	html.WriteString(" - ")
	for i, ac := range t.AreaCodes {
		if i > 0 {
			html.WriteString(", ")
		}
		html.WriteString(strconv.Itoa(int(ac)))
	}
	html.WriteString("</span><br>")
	return template.HTML(html.String())
}

const (
	njsplit = "however in June 1991, the 201 and the 609 area codes were split to create the 908 area code."
	bronx   = "On July 1992, the Bronx moved from 212 to 718."
)

// Notes returns a map of North American Numbering Plan area codes notes that provide additional information,
// such as when an area code was split or created and the regions it serves.
func Notes() map[NAN]string { //nolint:funlen
	return map[NAN]string{
		201: "Northern NJ, including Newark, " + njsplit,
		609: "Southern NJ, including Trenton, " + njsplit,
		908: "Central NJ, including Union and Somerset counties, however it came online on June 1991.",
		206: "Western Washington state, including Seattle.",
		509: "Eastern Washington state, including Spokane and the Tri-Cities.",
		// 209, 213, 310, 408, 415, 510, 619, 707, 714, 805, 818, 909, 916
		209: "Central California, including Fresno and Stockton.",                                                      // 1947
		213: "Parts of Los Angeles County.",                                                                            // 1947
		916: "Sacramento.",                                                                                             // 1947
		714: "Orange County.",                                                                                          // 1951
		805: "Central California, including Santa Barbara and Ventura Counties.",                                       // 1957
		415: "San Francisco Bay Area, however in September 1991 East Bay and Oakland split to from the 510 area code.", // 1958
		408: "Silicon Valley including San Benito, Santa Cruz, Santa Clara.",                                           // 1959
		707: "Northern California, including Napa and Sonoma.",                                                         // 1959
		619: "Southern California, including San Diego, however it came online in 1982.",                               // 1982
		818: "San Fernando Valley within Los Angeles County, however it came online in 1984.",                          // 1984
		510: "East Bay and Oakland, however it came online in September 1991.",                                         // 1991
		310: "Northern Los Angeles County, however it came online in November 1991.",                                   // 1991
		909: "Parts of Los Angeles County and San Bernardino County, however it came online in November 1992.",         // 1992
		// 212, 315, 516, 518, 607, 716, 718, 914
		212: "New York City, but after December 1984 it split to 212 (Manhattan and the Bronx) and 718 (Brooklyn, Queens, and Staten Island). " + bronx, // 1947
		315: "Northern New York, including Syracuse.",                                                                                                   // 1947
		516: "Long Island, including Nassau County.",                                                                                                    // 1951
		518: "Eastern New York, including Albany.",                                                                                                      // 1947
		607: "Southern New York, including Binghamton.",                                                                                                 // 1954
		716: "Western New York, including Buffalo.",                                                                                                     // 1954
		718: "New York City, created on December 1984 to serve Brooklyn, Queens, and Staten Island. " + bronx,                                           // 1984
		914: "Southern New York, including Westchester County.",                                                                                         // 1947
		917: "New York City, coming online in February 1992 and the first overlay area code, however it is exclusive for cellphones.",                   // 1992
		// 210, 214, 409, 512, 713, 806, 817, 903, 915
		214: "Northern Texas, including Dallas, however after November 1990 it split to form 903.",                                     // 1947
		409: "Southeastern Texas, including Beaumont and Galveston, however it came online in 1983.",                                   // 1983
		512: "Central Texas, including Austin, however in November 1992, the Southern region including San Antonio split to form 210.", // 1947
		713: "Southern Texas, including Houston, however in 1983 it split to from the 409 area code.",                                  // 1947
		806: "Northern Texas, including Amarillo.",                                                                                     // 1957
		817: "Western Texas, including Fort Worth.",                                                                                    // 1953
		903: "Northeastern Texas, including Tyler, coming online in November 1990. " + // 1990
			"However, up until October 1980 this area code was used for calling northwestern Mexico.",
		210: "Southern Texas, including San Antonio, however it came online in November 1992.", // 1992
		915: "Western Texas, including El Paso.",                                               // 1953
		// 215, 412, 610, 717, 814
		215: "Philadelphia tri-state area, however in 1994 the Western region split to form the 610 area code.",
		412: "Western Pennsylvania, including Pittsburgh.",
		717: "Central Pennsylvania, including Harrisburg.",
		814: "Western Pennsylvania, including Erie.",
		610: "Western suburbs of Philadelphia, Berks County and Lehigh Valley, however it came online in January 1994.", // 1994
		// 216, 419, 513, 614
		216: "Northern Ohio, including Cleveland.",
		419: "Northwestern Ohio, including Toledo.",
		513: "Southwestern Ohio, including Cincinnati.",
		614: "Central Ohio, including Columbus.",
		// 217, 309, 312, 618, 708, 815
		217: "Central Illinois, including Springfield.",
		309: "Central Illinois, including Peoria.",
		312: "Chicago metropolitan area aka Chicagoland (except for the 815 suburbs). " +
			"However in November 1989 it split to form 708 and 312 became exclusive to Chicago City.",
		618: "Southern Illinois, including Carbondale.",
		708: "Chicago suburbs, however it came online in November 1989.",
		815: "Northern Illinois, including Rockford.",
		// 218, 507, 612
		218: "Northern Minnesota, including Duluth.",
		507: "Southern Minnesota, including Rochester.",
		612: "Minneapolis and some suburbs.",
		// 219, 317, 812
		219: "Northwestern Indiana aka South Bend, including Gary.",
		317: "Central Indiana, including Indianapolis.",
		812: "Southern Indiana, including Evansville.",
		// 301, 410
		301: "Maryland. However in November 1991, Balimore and the Eastern Shore of Maryland were split to create the 410 area code.",
		410: "Balimore and the Eastern Shore of Maryland, however it came online in November 1991.",
		// 303, 719
		303: "Central Colorado, including Denver.",
		719: "Southern Colorado, including Colorado Springs, however it came online in March 1988.",
		// 305, 407, 813, 904
		305: "Southern Florida, including Miami. However, Central Florida was split in 1988 to form the 407 area code.", // 1947
		407: "Central Florida, including Orlando and Palm Beach, however it came online in 1988.",                       // 1988
		813: "Western Florida, including Tampa City.",                                                                   // 1953
		904: "Northern Florida, including Jacksonville.",                                                                // 1965
		// 313, 517, 616, 906
		313: "Detroit and Flint, however in December 1993 the Northern region split to form 810.",
		517: "Central Michigan, including Lansing.",
		616: "Western Michigan, including Grand Rapids.",
		810: "Northern Detroit and Flint, however it came online in December 1993.",
		906: "Upper Peninsula of Michigan, including Marquette.",
		// 314, 417, 816
		314: "Eastern Missouri, including St. Louis.",
		417: "Southwestern Missouri, including Springfield.",
		816: "Western Missouri, including Kansas City.",
		// 316, 913
		316: "Southern Kansas, including Wichita.",
		913: "Northern Kansas, including Kansas City.",
		// 318, 504
		318: "Northern Louisiana, including Shreveport.",
		504: "Southern Louisiana, including New Orleans.",
		// 319, 515, 712
		319: "Eastern Iowa, including Cedar Rapids.",
		515: "Central Iowa, including Des Moines.",
		712: "Western Iowa, including Sioux City.",
		// 308, 402
		308: "Central Nebraska, including Grand Island.",
		402: "Eastern Nebraska, including Omaha and Lincoln.",
		// 404, 706, 912
		404: "Northern Georgia, including Atlanta. However in May 1992, all the areas outside of Atlanta were split to form the 706 area code.", // 1947
		706: "Northern Georgia, including Columbus, however it came online in May 1992.",                                                        // 1992
		912: "Southern Georgia, including Savannah.",                                                                                            // 1954
		// 405, 918
		405: "Central Oklahoma, including Oklahoma City.",
		918: "Eastern Oklahoma, including Tulsa.",
		// 414, 608, 715
		414: "Southeastern Wisconsin, including Milwaukee.",
		608: "Southwestern Wisconsin, including Madison.",
		715: "Northern Wisconsin, including Eau Claire.",
		// 416, 519
		416: "Toronto, however in October 1993 the suburbs surrounding Toronto City split to form the 905 area code.", // 1947
		519: "Southwestern Ontario, including Windsor and London.",                                                    // 1953
		613: "Eastern Ontario, including Ottawa.",
		705: "Northern Ontario, including Sudbury.",
		807: "Northwestern Ontario, including Thunder Bay.",
		905: "Suburbs surrounding Toronto City, coming online in October 1993. " +
			"However, up until February 1991 the 905 area code was used to call Mexico City.", // 1993
		// 418, 514, 819
		418: "Eastern Quebec, including Quebec City.",
		514: "Montreal.",
		819: "Western Quebec and the Northwest Territories, including Gatineau.",
		// 502, 606
		502: "Northern Kentucky, including Louisville.",
		606: "Eastern Kentucky, including Ashland.",
		// 413, 508, 617
		413: "Western Massachusetts, including Springfield.",
		508: "Southeastern Massachusetts, including Worcester, however it came online in July 1988.",
		617: "Eastern Massachusetts, including Boston, however the surrounding regions were split in July 1988 to form the 508 area code.", // 1988
		// 615, 901
		615: "Central Tennessee, including Nashville.",
		901: "Western Tennessee, including Memphis.",
		// 703, 804
		703: "Northern Virginia, including Arlington.",
		804: "Southern Virginia, including Richmond.",
		// 704, 919
		704: "Western North Carolina, including Charlotte.",                                                             // 1947
		919: "Eastern North Carolina, including Raleigh, however the 910 area code split in November 1993.",             // 1954
		910: "Southern North Carolina, including Fayetteville and Wilmington, however it came online in November 1993.", // 1993
		// 710
		710: "Federal Government of the United States emergency services, however it came online in 1983.",
		// 809
		809: "Bermuda and the Caribbean Islands. However in November 1994 Bermuda received the 441 area code, " +
			"and the other islands obtained their own area codes from 1995 onwards.",
	}
}

// territories is a list of territories in the North American Numbering Plan.
// These can be checked against official lists to ensure accuracy.
func territories() []Territory {
	return []Territory{
		// Miscellaneous
		{"Caribbean Islands", "", []NAN{809}},
		{"United States Government", "", []NAN{710}},
		// Canada
		{"Alberta", "AB", []NAN{403}},
		{"British Columbia", "BC", []NAN{604}},
		{"Manitoba", "MB", []NAN{204}},
		{"New Brunswick", "NB", []NAN{506}},
		{"Newfoundland and Labrador", "NL", []NAN{709}},
		{"Northwest Territories", "NT", []NAN{819}},
		{"Nova Scotia", "NS", []NAN{902}},
		{"Nunavut", "NU", []NAN{}}, // became a territory in 1999
		{"Ontario", "ON", []NAN{416, 519, 613, 705, 807, 905}},
		{"Prince Edward Island", "PE", []NAN{902}},
		{"Quebec", "QC", []NAN{418, 514, 819}},
		{"Saskatchewan", "SK", []NAN{306}},
		{"Yukon", "YT", []NAN{403}},
		// United States
		{"Alabama", "AL", []NAN{205}},
		{"Alaska", "AK", []NAN{907}},
		{"Arizona", "AZ", []NAN{602}},
		{"Arkansas", "AR", []NAN{501}},
		{"California", "CA", []NAN{209, 213, 310, 408, 415, 510, 619, 707, 714, 805, 818, 909, 916}},
		{"Colorado", "CO", []NAN{303, 719}},
		{"Connecticut", "CT", []NAN{203}},
		{"Delaware", "DE", []NAN{302}},
		{"District of Columbia", "DC", []NAN{202}},
		{"Florida", "FL", []NAN{305, 407, 813, 904}},
		{"Georgia", "GA", []NAN{404, 706, 912}},
		{"Hawaii", "HI", []NAN{808}},
		{"Idaho", "ID", []NAN{208}},
		{"Illinois", "IL", []NAN{217, 309, 312, 618, 708, 815}},
		{"Indiana", "IN", []NAN{219, 317, 812}},
		{"Iowa", "IA", []NAN{319, 515, 712}},
		{"Kansas", "KS", []NAN{316, 913}},
		{"Kentucky", "KY", []NAN{502, 606}},
		{"Louisiana", "LA", []NAN{318, 504}},
		{"Maine", "ME", []NAN{207}},
		{"Maryland", "MD", []NAN{301, 410}},
		{"Massachusetts", "MA", []NAN{413, 508, 617}},
		{"Michigan", "MI", []NAN{313, 517, 616, 810, 906}},
		{"Minnesota", "MN", []NAN{218, 507, 612}},
		{"Mississippi", "MS", []NAN{601}},
		{"Missouri", "MO", []NAN{314, 417, 816}},
		{"Montana", "MT", []NAN{406}},
		{"Nebraska", "NE", []NAN{308, 402}},
		{"Nevada", "NV", []NAN{702}},
		{"New Hampshire", "NH", []NAN{603}},
		{"New Jersey", "NJ", []NAN{201, 609, 908}},
		{"New Mexico", "NM", []NAN{505}},
		{"New York", "NY", []NAN{212, 315, 516, 518, 607, 716, 718, 914, 917}},
		{"North Carolina", "NC", []NAN{704, 910, 919}},
		{"North Dakota", "ND", []NAN{701}},
		{"Ohio", "OH", []NAN{216, 419, 513, 614}},
		{"Oklahoma", "OK", []NAN{405, 918}},
		{"Oregon", "OR", []NAN{503}},
		{"Pennsylvania", "PA", []NAN{215, 412, 610, 717, 814}},
		{"Rhode Island", "RI", []NAN{401}},
		{"South Carolina", "SC", []NAN{803}},
		{"South Dakota", "SD", []NAN{605}},
		{"Tennessee", "TN", []NAN{615, 901}},
		{"Texas", "TX", []NAN{210, 214, 409, 512, 713, 806, 817, 903, 915}},
		{"Utah", "UT", []NAN{801}},
		{"Vermont", "VT", []NAN{802}},
		{"Virginia", "VA", []NAN{703, 804}},
		{"Washington", "WA", []NAN{206, 509}},
		{"West Virginia", "WV", []NAN{304}},
		{"Wisconsin", "WI", []NAN{414, 608, 715}},
		{"Wyoming", "WY", []NAN{307}},
	}
}

// Territories returns a list of all territories in the North American Numbering Plan
// sorted by name in ascending order.
func Territories() []Territory {
	terr := territories()
	slices.SortFunc(terr, func(i, j Territory) int {
		if n := strings.Compare(i.Name, j.Name); n != 0 {
			return n
		}
		return cmp.Compare(i.Abbreviation, j.Abbreviation)
	})
	return terr
}

// Lookup returns a list of territories that match the given input.
// The input can be a string, integer, or NAN.
// If the input is a string, it will match against territory names and abbreviations.
// If the input is an integer, it will match against NANP codes.
func Lookup(a any) []Territory {
	const abbreviation = 2
	switch v := a.(type) {
	case string:
		switch len(v) {
		case 0, 1:
			return nil
		case abbreviation:
			return []Territory{TerritoryByAbbr(Abbreviation(v))}
		}
		return TerritoryContains(v)
	case int, uint:
		if c, ok := a.(int); ok {
			return TerritoryByCode(NAN(c))
		}
		if c, ok := a.(uint); ok {
			if c > limit {
				return nil
			}
			return TerritoryByCode(NAN(int(c)))
		}
		return nil
	case NAN:
		return TerritoryByCode(v)
	default:
		return nil
	}
}

// Lookups returns a list of territories that match the given inputs.
//
// See Lookup for more information.
func Lookups(a ...any) []Territory {
	var t []Territory
	for _, query := range a {
		finds := Lookup(query)
		if len(finds) == 0 {
			continue
		}
		for find := range slices.Values(finds) {
			if Contains(find, t...) {
				continue
			}
			t = append(t, find)
		}
	}
	return t
}

// Result represents the result of a query, which can be an area code or a list of territories.
type Result struct {
	Terr     []Territory
	AreaCode NAN
}

// Query returns the result of a query from a form input.
func Query(a any) Result {
	switch val := a.(type) {
	case string:
		if c, err := strconv.Atoi(val); err == nil {
			ac := NAN(c)
			return Result{AreaCode: ac}
		}
		return Result{Terr: Lookup(val)}
	case int, uint:
		if c, ok := a.(int); ok {
			return Result{AreaCode: NAN(c)}
		}
		if c, ok := a.(uint); ok {
			if c > limit {
				return Result{}
			}
			return Result{AreaCode: NAN(int(c))}
		}
		return Result{}
	case NAN:
		return Result{AreaCode: val}
	default:
		return Result{}
	}
}

// Queries returns a list of results for multiple queries from a form input.
func Queries(s ...string) []Result {
	const maximum = 99
	queries := make([]any, len(s))
	for i, val := range s {
		if i > maximum {
			break
		}
		val = strings.TrimSpace(val)
		if x, err := strconv.Atoi(val); err == nil {
			queries[i] = x
			continue
		}
		queries[i] = val
	}
	r := make([]Result, 0, len(queries))
	for _, query := range queries {
		find := Query(query)
		if !find.AreaCode.Valid() {
			none := len(find.Terr) == 0
			if none {
				continue
			}
			none = len(find.Terr) == 1 && find.Terr[0].Name == ""
			if none {
				continue
			}
		}
		r = append(r, find)
	}
	return r
}

// Contains returns true if the territory is in the list of territories.
func Contains(t Territory, ts ...Territory) bool {
	for val := range slices.Values(ts) {
		if t.Name == val.Name {
			return true
		}
	}
	return false
}

// Abbreviations returns a list of all two-letter abbreviations for territories
// in the North American Numbering Plan sorted in ascending order.
func Abbreviations() []Abbreviation {
	abbr := make([]Abbreviation, 0, len(territories()))
	for val := range slices.Values(territories()) {
		if val.Abbreviation == "" {
			continue
		}
		abbr = append(abbr, val.Abbreviation)
	}
	slices.Sort(abbr)
	abbr = slices.Compact(abbr) // remove empty strings
	return abbr
}

// AreaCodes returns a list of all NANP area codes sorted in ascending order.
func AreaCodes() []NAN {
	codes := make([]NAN, 0, len(territories()))
	for val := range slices.Values(territories()) {
		codes = append(codes, val.AreaCodes...)
	}
	slices.Sort(codes)
	codes = slices.Compact(codes)
	return codes
}

// TerritoryByAbbr returns the territory with the given two-letter abbreviation.
func TerritoryByAbbr(abbr Abbreviation) Territory {
	if len(abbr) == 0 {
		return Territory{}
	}
	for val := range slices.Values(territories()) {
		if strings.EqualFold(string(val.Abbreviation), string(abbr)) {
			return val
		}
	}
	return Territory{}
}

// TerritoryByCode returns the territories for the given North American Numbering code.
//
// Generally, this will return a single territory, but it is possible for
// a NAN code to be used in multiple territories, such as provinces in Canada.
func TerritoryByCode(code NAN) []Territory {
	if !code.Valid() {
		return nil
	}
	var finds []Territory
	for val := range slices.Values(territories()) {
		for ac := range slices.Values(val.AreaCodes) {
			if ac == code {
				finds = append(finds, val)
			}
		}
	}
	return finds
}

// TerritoryByName returns the territory with the given name.
// The name can be a US state, Canadian province, or other territory.
func TerritoryByName(name string) Territory {
	for val := range slices.Values(territories()) {
		if strings.EqualFold(val.Name, name) {
			return val
		}
	}
	return Territory{}
}

// TerritoryContains returns a list of territories with names that contain the given string.
func TerritoryContains(s string) []Territory {
	vals := []Territory{}
	for val := range slices.Values(territories()) {
		substr := strings.ToLower(s)
		if strings.Contains(strings.ToLower(val.Name), substr) {
			vals = append(vals, val)
		}
	}
	return vals
}
