package router

import (
	"fmt"
	"html/template"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/bengarrett/df2023/helpers"
	"github.com/bengarrett/df2023/models"
	"github.com/volatiletech/null/v8"
)

const (
	maxPad  = 80
	padding = " "
	noValue = "-"
)

var TemplateFuncMap = template.FuncMap{
	"leading":  Leading,
	"leadInt":  LeadInt,
	"leadStr":  LeadStr,
	"iconFmt":  models.Icon,
	"linkFile": FileName,
	"linkHref": FileHref,
	"linkPad":  FileLinkPad,
	"datePub":  LeadPub,
	"datePost": LeadPost,
	"byteFmt":  LeadFS,
	"descript": Description,
}

// <cfif Right(records.record_title,1) is '.'>
// #Trim(Left(records.record_title,Len(records.record_title)-1))#<cfelse>
// #Trim(records.record_title)#</cfif><cfif Len(records.group_brand_for)>
//
// <cfif Len(records.record_title)> from <cfelse>From </cfif>#records.group_brand_for#</cfif>
// <cfif ListFindNoCase("Dos,Java,Linux,Windows,Mac10",records.platform)>
// for #Trim(Replacenocase(getPlatformName(records.platform),'Apps. ',''))#</cfif>
// </cfif></cfoutput>

func Description(w int, section, platform, brand, title null.String) string {
	category := strings.TrimSpace(section.String)
	if category == "magazine" {
		return fmt.Sprintf("%s issue %s.", brand.String, title.String)
	}
	desc := ""
	if t := helpers.TrimPunct(title.String); t == "" {
		desc = "From "
	} else {
		desc = fmt.Sprintf("%s from ", t)
	}
	desc += brand.String
	if IsApp(platform) {
		desc += FmtApp(platform)
	}
	return fmt.Sprintf("%s.", desc)
}

func IsApp(platform null.String) bool {
	s := []string{"dos", "java", "linux", "windows", "mac10"}
	p := strings.TrimSpace(strings.ToLower(platform.String))
	return helpers.IsValid(p, s...)
}

func FmtApp(platform null.String) string {
	s := ""
	p := strings.TrimSpace(strings.ToLower(platform.String))
	switch p {
	case "dos":
		s = "DOS"
	case "java":
		s = "Java"
	case "linux":
		s = "Linux"
	case "windows":
		s = "Windows"
	case "mac10":
		s = "macOS"
	}
	if s == "" {
		return ""
	}
	return fmt.Sprintf(" for %s", s)
}

func FileLink() {}

func FileName(w int, name null.String) string {
	return helpers.TruncFilename(w, name.String)
}

func FileLinkPad(w int, name null.String) string {
	if !name.Valid {
		return Leading(w)
	}
	s := helpers.TruncFilename(w, name.String)
	if len(s) < w {
		return LeadStr(w, s)
	}
	return ""
}

func FileHref(id int64) string {
	href, err := url.JoinPath("/", "d", helpers.ObfuscateParam(strconv.Itoa(int(id))))
	if err != nil {
		log.Println(err) //TODO: log to file
	}
	return href
}

func LeadFileLink(w int, id int64, name null.String) string {
	//<a href="/d/a228dd"></a>
	if !name.Valid {
		return Leading(w)
	}
	href, err := url.JoinPath("/", "d", helpers.ObfuscateParam(strconv.Itoa(int(id))))
	if err != nil {
		log.Println(err) //TODO: log to file
	}
	s := helpers.TruncFilename(w, name.String)
	html := fmt.Sprintf("<a href=\"%s\">%s</a>", href, s)
	if len(s) < w {
		return html + LeadStr(w, s)
	}
	return html
}

func LeadPub(w int, y, m, d null.Int16) string {
	s := models.DatePublish(y, m, d)
	if len(s) < w {
		return LeadStr(w, s) + s
	}
	return s
}

func LeadPost(w int, t null.Time) string {
	s := models.DateFmt(t)
	if len(s) < w {
		return LeadStr(w, s) + s
	}
	return s
}

func LeadFS(w int, size null.Int64) string {
	if !size.Valid {
		return Leading(w)
	}
	s := helpers.ByteCount(size.Int64)
	l := len(s)
	return Leading(w-l) + s
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

func Leading(w int) string {
	if w < 1 {
		return ""
	}
	return strings.Repeat(padding, w)
}
