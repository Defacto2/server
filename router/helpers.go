package router

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/bengarrett/df2023/models"
)

const (
	maxPad  = 80
	padding = " "
	noValue = "-"
)

var TemplateFuncMap = template.FuncMap{
	"leadInt": LeadInt,
	"leadStr": LeadStr,
	"dateStr": models.DatePub,
	"dateFmt": models.DateFmt,
	"byteFmt": models.ByteCount,
}

// LeadInt takes an int and returns it as a string, w characters wide with whitespace padding.
func LeadInt(w, i int) string {
	s := noValue
	if i > 0 {
		s = strconv.Itoa(i)
	}
	l := len(s)
	if l >= w {
		return s
	}
	count := w - l
	if count > maxPad {
		count = maxPad
	}
	return fmt.Sprintf("%s%s", strings.Repeat(padding, count), s)
}

// LeadStr takes a string and returns the leading whitespace padding, w characters wide.
// the value of string is note returned.
func LeadStr(w int, s string) string {
	l := len(s)
	if l >= w {
		return ""
	}
	return strings.Repeat(padding, w-l)
}
