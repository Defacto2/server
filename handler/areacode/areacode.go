package areacode

import (
	"fmt"
	"html/template"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// NANP
// https://defacto2.net/f/ac1cb6a North American Pirate-Phreak Association from 1990
// FINDAC.ZIP         4681 08/12/91 Find Area Code v1.0  [1/1]  TWA
// (c) 1991 by MPM Enterprises
// https://en.wikipedia.org/wiki/Area_codes_602,_480,_and_623

type NANPCode uint

// Valid returns true if the NANP code is a valid area code.
func (c NANPCode) Valid() bool {
	ac := AreaCodes()
	return slices.Contains(ac, c)
}

// HTML returns the NANP code as an HTML span element.
//
// Example:
//
//	212 - New York (NY)
func (c NANPCode) HTML() template.HTML {
	empty := template.HTML(``)
	if !c.Valid() {
		return empty
	}
	ts := TerritoryByCode(c)
	if len(ts) == 0 {
		return empty
	}
	html := "<span>"
	for i, t := range ts {
		ac := t.Alpha
		s := ""
		if i == 0 {
			s = fmt.Sprintf(`%d - %s`, c, t.Name)
			if len(ac) > 0 {
				s += fmt.Sprintf(` (%s)`, ac)
			}
			html += s
			continue
		}
		html += fmt.Sprintf(" + %s", t.Name)
		if len(ac) > 0 {
			html += fmt.Sprintf(` (%s)`, ac)
		}
	}
	if note, ok := notes[c]; ok {
		html += fmt.Sprintf(" <small><em>%s</em></small>", note)
	}
	return template.HTML(html + "</span><br>")
}

// AlphaCode represents a two-letter alphabetic code for a territory in the North American Numbering Plan.
type AlphaCode string

func (a AlphaCode) HTML() template.HTML {
	t := TerritoryByAlpha(a)
	html := "<span>"
	html += string(t.Alpha) + " (" + t.Name + ")"
	if len(t.AreaCode) == 0 {
		html += " - n/a</span><br>"
		return template.HTML(html)
	}
	html += " - "
	for i, ac := range t.AreaCode {
		if i > 0 {
			html += ", "
		}
		html += strconv.Itoa(int(ac))
	}
	return template.HTML(html + "</span><br>")
}

// Territory represents a territory in the North American Numbering Plan.
type Territory struct {
	Name     string     // Name of the state, province, or territory.
	Alpha    AlphaCode  // Two-letter alphabetic code.
	AreaCode []NANPCode // Three-digit NANP code used for telephone area codes.
}

func (t Territory) HTML() template.HTML {
	html := "<span>" + t.Name
	if len(t.Alpha) > 0 {
		html += " (" + string(t.Alpha) + ") "
	}
	// join area codes with commas
	if len(t.AreaCode) == 0 {
		html += "- n/a</span><br>"
		return template.HTML(html)
	}
	html += " - "
	for i, ac := range t.AreaCode {
		if i > 0 {
			html += ", "
		}
		html += strconv.Itoa(int(ac))
	}
	html += "</span><br>"
	return template.HTML(html)
}

const (
	njsplit = "however in June 1991, the 201 and the 609 area codes were split to create the 908 area code."
	bronx   = "In July 1992, the Bronx was moved to the 718 area code."
)

var notes = map[NANPCode]string{
	// see: https://www.nytimes.com/1991/06/02/nyregion/201-609-and-now-oh-my-908.html
	201: "Northern NJ, " + njsplit,
	609: "Southern NJ, " + njsplit,
	908: "Central NJ, however it came online on June 1991.",
	206: "Western Washington state, including Seattle.",
	509: "Eastern Washington state.",
	// see: https://www.cpuc.ca.gov/industries-and-topics/internet-and-phone/area-codes-and-numbering
	// 209, 213, 408, 415, 510, 619, 707, 714, 805, 818, 909, 916
	209: "Central California, including Fresno and Stockton.",                                              //1947
	213: "Parts of Los Angeles County.",                                                                    //1947
	916: "Sacramento.",                                                                                     //1947
	714: "Orange County.",                                                                                  //1951
	805: "Central California, including Santa Barbara and Ventura Counties.",                               //1957
	415: "San Francisco Bay Area, until September 1991 when East Bay and Oakland moved to 510.",            //1958
	408: "Silicon Valley including San Benito, Santa Cruz, Santa Clara.",                                   //1959
	707: "Northern California, including Napa and Sonoma.",                                                 //1959
	619: "Southern California, including San Diego, however it came online in 1982.",                       //1982
	818: "San Fernando Valley within Los Angeles County, however it came online in 1984.",                  //1984
	510: "East Bay and Oakland, however it came online in September 1991.",                                 //1991
	310: "Northern Los Angeles County, however it came online in November 1991.",                           //1991
	909: "Parts of Los Angeles County and San Bernardino County, however it came online in November 1992.", //1992
	// see: https://www.nytimes.com/1984/12/29/nyregion/shift-from-212-to-718-code-pains-3-boroughs.html
	// 212, 315, 516, 518, 607, 716, 718, 914
	212: "New York City, but after December 1984 it split to 212 (Manhattan and the Bronx) and 718.",                                  // 1947
	315: "Northern New York, including Syracuse.",                                                                                     // 1947
	516: "Long Island, including Nassau County.",                                                                                      // 1951
	518: "Eastern New York, including Albany.",                                                                                        // 1947
	607: "Southern New York, including Binghamton.",                                                                                   // 1954
	716: "Western New York, including Buffalo.",                                                                                       // 1954
	718: "New York City, created on December 1984 to serve Brooklyn, Queens, and Staten Island." + bronx,                              // 1984
	914: "Southern New York, including Westchester County.",                                                                           // 1947
	917: "New York City, it was the first overlay code and exclusively used for cellphones, however it came online in February 1992.", // 1992
	// 210, 214, 409, 512, 713, 806, 817, 903, 915
	214: "Northern Texas, including Dallas, however after November 1990 it split to form 903.",   // 1947
	409: "Southeastern Texas, including Beaumont and Galveston, however it came online in 1983.", // 1983
	512: "Central Texas, including Austin, however in November 1992 it split to form 210.",       // 1947
	713: "Southern Texas, including Houston, however in 1983 it split to from 409.",              // 1947
	806: "Northern Texas, including Amarillo.",                                                   // 1957
	817: "Western Texas, including Fort Worth.",                                                  // 1953
	903: "Northeastern Texas, including Tyler, however it came online in November 1990." + // 1990
		" However, in until October 1980 this area code was used for northwestern Mexico.",
	210: "Southern Texas, including San Antonio, however it came online in November 1992.", // 1992
	915: "Western Texas, including El Paso.",                                               // 1953
	// 215, 412, 610, 717, 814
	215: "Philadelphia tri-state area, however in 1994 it split to form 610.",
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
	305: "Southern Florida, including Miami. However, Central Florida was split in 1988 to form the 407 area code.", //1947
	407: "Central Florida, including Orlando and Palm Beach, however it came online in 1988.",                       //1988
	813: "Western Florida, including Tampa City.",                                                                   //1953
	904: "Northern Florida, including Jacksonville.",                                                                //1965
	// 313, 517, 616, 906
	313: "Detroit and Flint, however in December 1993 it split to form 810.",
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
	// https://areacode416homes.com/history-of-the-416-area-code
	// 416, 519
	416: "Toronto, however in October 1993 the suburbs surrounding Toronto City split to form the 905 area code.", // 1947
	519: "Southwestern Ontario, including Windsor and London.",                                                    // 1953
	613: "Eastern Ontario, including Ottawa.",
	705: "Northern Ontario, including Sudbury.",
	807: "Northwestern Ontario, including Thunder Bay.",
	905: "Suburbs surrounding Toronto City, however it came online in October 1993." +
		"However, from the 1960s until February 1991 the 905 area code was used by NANP users to call Mexico City.", // 1993
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

// discontinued area codes (Feb 1991)
// 903 north west mexico
// 905 mexico city

// territories is a list of territories in the North American Numbering Plan.
// These can be checked against official lists to ensure accuracy.
var territories = []Territory{
	// Caribbean Islands including non-US territories
	// note, these changed in 1994-99
	{"Caribbean Islands", "", []NANPCode{809}},
	{"United States Government", "", []NANPCode{710}},
	// Canada
	{"Alberta", "AB", []NANPCode{403}},
	{"British Columbia", "BC", []NANPCode{604}},
	{"Manitoba", "MB", []NANPCode{204}},
	{"New Brunswick", "NB", []NANPCode{506}},
	{"Newfoundland and Labrador", "NL", []NANPCode{709}},
	{"Northwest Territories", "NT", []NANPCode{819}},
	{"Nova Scotia", "NS", []NANPCode{902}},
	{"Nunavut", "NU", []NANPCode{}}, // became a territory in 1999
	{"Ontario", "ON", []NANPCode{416, 519, 613, 705, 807, 905}},
	{"Prince Edward Island", "PE", []NANPCode{902}},
	{"Quebec", "QC", []NANPCode{418, 514, 819}},
	{"Saskatchewan", "SK", []NANPCode{306}},
	{"Yukon", "YT", []NANPCode{403}},
	// United States
	{"Alabama", "AL", []NANPCode{205}},
	{"Alaska", "AK", []NANPCode{907}},
	{"Arizona", "AZ", []NANPCode{602}},
	{"Arkansas", "AR", []NANPCode{501}},
	{"California", "CA", []NANPCode{209, 213, 408, 415, 510, 619, 707, 714, 805, 818, 909, 916}},
	{"Colorado", "CO", []NANPCode{303, 719}},
	{"Connecticut", "CT", []NANPCode{203}},
	{"Delaware", "DE", []NANPCode{302}},
	{"District of Columbia", "DC", []NANPCode{202}},
	{"Florida", "FL", []NANPCode{305, 407, 813, 904}},
	{"Georgia", "GA", []NANPCode{404, 706, 912}},
	{"Hawaii", "HI", []NANPCode{808}},
	{"Idaho", "ID", []NANPCode{208}},
	{"Illinois", "IL", []NANPCode{217, 309, 312, 618, 708, 815}},
	{"Indiana", "IN", []NANPCode{219, 317, 812}},
	{"Iowa", "IA", []NANPCode{319, 515, 712}},
	{"Kansas", "KS", []NANPCode{316, 913}},
	{"Kentucky", "KY", []NANPCode{502, 606}},
	{"Louisiana", "LA", []NANPCode{318, 504}},
	{"Maine", "ME", []NANPCode{207}},
	{"Maryland", "MD", []NANPCode{301, 410}},
	{"Massachusetts", "MA", []NANPCode{413, 508, 617}},
	{"Michigan", "MI", []NANPCode{313, 517, 616, 810, 906}},
	{"Minnesota", "MN", []NANPCode{218, 507, 612}},
	{"Mississippi", "MS", []NANPCode{601}},
	{"Missouri", "MO", []NANPCode{314, 417, 816}},
	{"Montana", "MT", []NANPCode{406}},
	{"Nebraska", "NE", []NANPCode{308, 402}},
	{"Nevada", "NV", []NANPCode{702}},
	{"New Hampshire", "NH", []NANPCode{603}},
	{"New Jersey", "NJ", []NANPCode{201, 609, 908}},
	{"New Mexico", "NM", []NANPCode{505}},
	{"New York", "NY", []NANPCode{212, 315, 516, 518, 607, 716, 718, 914, 917}},
	{"North Carolina", "NC", []NANPCode{704, 919}},
	{"North Dakota", "ND", []NANPCode{701}},
	{"Ohio", "OH", []NANPCode{216, 419, 513, 614}},
	{"Oklahoma", "OK", []NANPCode{405, 918}},
	{"Oregon", "OR", []NANPCode{503}},
	{"Pennsylvania", "PA", []NANPCode{215, 412, 610, 717, 814}},
	{"Rhode Island", "RI", []NANPCode{401}},
	{"South Carolina", "SC", []NANPCode{803}},
	{"South Dakota", "SD", []NANPCode{605}},
	{"Tennessee", "TN", []NANPCode{615, 901}},
	{"Texas", "TX", []NANPCode{210, 214, 409, 512, 713, 806, 817, 903, 915}},
	{"Utah", "UT", []NANPCode{801}},
	{"Vermont", "VT", []NANPCode{802}},
	{"Virginia", "VA", []NANPCode{703, 804}},
	{"Washington", "WA", []NANPCode{206, 509}},
	{"West Virginia", "WV", []NANPCode{304}},
	{"Wisconsin", "WI", []NANPCode{414, 608, 715}},
	{"Wyoming", "WY", []NANPCode{307}},
}

// Territories returns a list of all territories in the North American Numbering Plan
// sorted by name in ascending order.
func Territories() []Territory {
	terr := territories
	sort.Slice(terr, func(i, j int) bool {
		return territories[i].Name < territories[j].Name
	})
	return terr
}

// Lookup returns a list of territories that match the given input.
// The input can be a string, integer, or NANPCode.
// If the input is a string, it will match against territory names and alpha codes.
// If the input is an integer, it will match against NANP codes.
func Lookup(a any) []Territory {
	switch v := a.(type) {
	case string:
		if len(v) == 2 {
			return []Territory{TerritoryByAlpha(AlphaCode(v))}
		}
		return TerritoryMatch(v)
	case int, uint:
		return TerritoryByCode(NANPCode(v.(int)))
	case NANPCode:
		return TerritoryByCode(v)
	default:
		return nil
	}
}

// Lookup returns a list of territories that match the given inputs.
//
// See Lookup for more information.
func Lookups(a ...any) []Territory {
	var t []Territory
	for _, query := range a {
		finds := Lookup(query)
		if len(finds) == 0 {
			continue
		}
		for _, find := range finds {
			if Contains(find, t...) {
				continue
			}
			t = append(t, find)
		}
	}
	return t
}

// Contains returns true if the territory is in the list of territories.
func Contains(t Territory, ts ...Territory) bool {
	for _, x := range ts {
		if t.Name == x.Name {
			return true
		}
	}
	return false
}

// AlphaCodes returns a list of all two-letter alphabetic codes for territories
// in the North American Numbering Plan sorted in ascending order.
func AlphaCodes() []AlphaCode {
	var codes []AlphaCode
	for _, t := range territories {
		if t.Alpha == "" {
			continue
		}
		codes = append(codes, t.Alpha)
	}
	slices.Sort(codes)
	codes = slices.Compact(codes) // remove empty strings
	return codes
}

// AreaCodes returns a list of all NANP area codes sorted in ascending order.
func AreaCodes() []NANPCode {
	var codes []NANPCode
	for _, t := range territories {
		codes = append(codes, t.AreaCode...)
	}
	slices.Sort(codes)
	codes = slices.Compact(codes)
	return codes
}

// TerritoryByAlpha returns the territory with the given two-letter alphabetic code.
func TerritoryByAlpha(ac AlphaCode) Territory {
	for _, t := range territories {
		if strings.EqualFold(string(t.Alpha), string(ac)) {
			return t
		}
	}
	return Territory{}
}

// TerritoryByCode returns the territories for the given NANP code.
//
// Generally, this will return a single territory, but it is possible for
// a NANP code to be used in multiple territories, such as provinces in Canada.
func TerritoryByCode(code NANPCode) []Territory {
	if !code.Valid() {
		return nil
	}
	var finds []Territory
	for _, t := range territories {
		for _, ac := range t.AreaCode {
			if ac == code {
				finds = append(finds, t)
			}
		}
	}
	return finds
}

// TerritoryByName returns the territory with the given name.
// The name can be a US state, Canadian province, or other territory.
func TerritoryByName(name string) Territory {
	for _, t := range territories {
		if strings.EqualFold(t.Name, name) {
			return t
		}
	}
	return Territory{}
}

// TerritoryMatch returns a list of territories that contain the given name.
func TerritoryMatch(name string) []Territory {
	m := []Territory{}
	for _, t := range territories {
		if strings.Contains(strings.ToLower(t.Name), strings.ToLower(name)) {
			m = append(m, t)
		}
	}
	return m
}
