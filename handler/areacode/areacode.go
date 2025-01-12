package areacode

import (
	"slices"
	"sort"
	"strings"
)

// NANP
// https://defacto2.net/f/ac1cb6a North American Pirate-Phreak Association from 1990
// FINDAC.ZIP         4681 08/12/91 Find Area Code v1.0  [1/1]  TWA
// (c) 1991 by MPM Enterprises

type NANPCode uint

// Valid returns true if the NANP code is a valid area code.
func (c NANPCode) Valid() bool {
	ac := AreaCodes()
	return slices.Contains(ac, c)
}

// Territory represents a territory in the North American Numbering Plan.
type Territory struct {
	Name      string     // Name of the state, province, or territory.
	AlphaCode string     // Two-letter alphabetic code.
	AreaCode  []NANPCode // Three-digit NANP code used for telephone area codes.
}

// territories is a list of territories in the North American Numbering Plan.
// These can be checked against official lists to ensure accuracy.
var territories = []Territory{
	// Caribbean Islands including non-US territories
	// note, these changed in 1994-99
	{"Caribbean Islands", "", []NANPCode{809}},
	// Canada
	{"Alberta", "AB", []NANPCode{403}},
	{"British Columbia", "BC", []NANPCode{604}},
	{"Manitoba", "MB", []NANPCode{204}},
	{"New Brunswick", "NB", []NANPCode{506}},
	{"Newfoundland and Labrador", "NL", []NANPCode{709}},
	{"Northwest Territories", "NT", []NANPCode{}}, // did not have area codes until 1997
	{"Nova Scotia", "NS", []NANPCode{902}},
	{"Nunavut", "NU", []NANPCode{}}, // created in 1999
	{"Ontario", "ON", []NANPCode{416, 519}},
	{"Prince Edward Island", "PE", []NANPCode{902}},
	{"Quebec", "QC", []NANPCode{418, 514, 819}},
	{"Saskatchewan", "SK", []NANPCode{306}},
	{"Yukon", "YT", []NANPCode{403}},
	// United States
	{"Alabama", "AL", []NANPCode{205}},
	{"Alaska", "AK", []NANPCode{907}},
	{"Arizona", "AZ", []NANPCode{602}},
	{"Arkansas", "AR", []NANPCode{501}},
	{"California", "CA", []NANPCode{209, 213, 408, 415, 619, 707, 714, 805, 818, 916}},
	{"Colorado", "CO", []NANPCode{303, 719}},
	{"Connecticut", "CT", []NANPCode{203}},
	{"Delaware", "DE", []NANPCode{302}},
	{"District of Columbia", "DC", []NANPCode{202}},
	{"Florida", "FL", []NANPCode{305, 407, 813, 904}},
	{"Georgia", "GA", []NANPCode{404, 912}},
	{"Hawaii", "HI", []NANPCode{808}},
	{"Idaho", "ID", []NANPCode{208}},
	{"Illinois", "IL", []NANPCode{217, 309, 312, 618, 815}},
	{"Indiana", "IN", []NANPCode{219, 317, 812}},
	{"Iowa", "IA", []NANPCode{319, 515, 712}},
	{"Kansas", "KS", []NANPCode{316, 913}},
	{"Kentucky", "KY", []NANPCode{502, 606}},
	{"Louisiana", "LA", []NANPCode{318, 504}},
	{"Maine", "ME", []NANPCode{207}},
	{"Maryland", "MD", []NANPCode{301}},
	{"Massachusetts", "MA", []NANPCode{413, 508, 617}},
	{"Michigan", "MI", []NANPCode{313, 517, 616, 906}},
	{"Minnesota", "MN", []NANPCode{218, 507, 612}},
	{"Mississippi", "MS", []NANPCode{601}},
	{"Missouri", "MO", []NANPCode{314, 417, 816}},
	{"Montana", "MT", []NANPCode{406}},
	{"Nebraska", "NE", []NANPCode{308, 402}},
	{"Nevada", "NV", []NANPCode{702}},
	{"New Hampshire", "NH", []NANPCode{603}},
	{"New Jersey", "NJ", []NANPCode{201, 609}},
	{"New Mexico", "NM", []NANPCode{505}},
	{"New York", "NY", []NANPCode{212, 315, 516, 518, 607, 716, 718, 914}},
	{"North Carolina", "NC", []NANPCode{704, 919}},
	{"North Dakota", "ND", []NANPCode{701}},
	{"Ohio", "OH", []NANPCode{216, 419, 513, 614}},
	{"Oklahoma", "OK", []NANPCode{405, 918}},
	{"Oregon", "OR", []NANPCode{503}},
	{"Pennsylvania", "PA", []NANPCode{215, 412, 717, 814}},
	{"Rhode Island", "RI", []NANPCode{401}},
	{"South Carolina", "SC", []NANPCode{803}},
	{"South Dakota", "SD", []NANPCode{605}},
	{"Tennessee", "TN", []NANPCode{615, 901}},
	{"Texas", "TX", []NANPCode{214, 409, 512, 713, 806, 817, 903, 915}},
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
// If the input is a string, it will match against territory names.
// If the input is an integer, it will match against NANP codes.
func Lookup(a any) []Territory {
	switch v := a.(type) {
	case string:
		if len(v) == 2 {
			return []Territory{TerritoryByAlpha(v)}
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

// TODO:
func Lookups(a ...any) []Territory {
	var t []Territory

	find := func(s string, x []Territory) bool {
		for _, v := range x {
			if strings.EqualFold(v.Name, s) {
				return true
			}
		}
		return false
	}

	for _, v := range a {
		l := Lookup(v)
		if find(l[0].Name, t) {
			continue
		}
		t = append(t, Lookup(v)...)
	}
	return t
}

// AlphaCodes returns a list of all two-letter alphabetic codes for territories
// in the North American Numbering Plan sorted in ascending order.
func AlphaCodes() []string {
	var codes []string
	for _, t := range territories {
		if t.AlphaCode == "" {
			continue
		}
		codes = append(codes, t.AlphaCode)
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
func TerritoryByAlpha(alpha string) Territory {
	for _, t := range territories {
		if strings.EqualFold(t.AlphaCode, alpha) {
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
