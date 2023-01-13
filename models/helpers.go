package models

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bengarrett/df2023/helpers"
	"github.com/volatiletech/null/v8"
)

// TODOs: https://github.com/bengarrett/Defacto2-2020/blob/78665d391f40bba14f44b6f9220dfe620c60650d/ROOT/views/html3/helpers.cfm
// https://github.com/bengarrett/Defacto2-2020/blob/78665d391f40bba14f44b6f9220dfe620c60650d/ROOT/views/html3/_pubedition.cfm
// https://github.com/bengarrett/Defacto2-2020/blob/78665d391f40bba14f44b6f9220dfe620c60650d/ROOT/views/html3/listfile.cfm
//
// <cfset fileText = Trim("#truncate(LCase(filenameLessExtension(records.fileName)),19,'.')#.#Left(fileExtension(records.fileName),3)#")>
// <cfset spacer = Val(23-Len(fileText))>

// todo: move to helpers/helpers.go
// https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/

func ByteCount(i null.Int64) string {
	if !i.Valid {
		return ""
	}
	b := i.Int64
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d%s", b, strings.Repeat(" ", 2))
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %c",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func DateFmt(t null.Time) string {
	if !t.Valid {
		return ""
	}
	d := t.Time.Day()
	m := AbbrMonth(int(t.Time.Month()))
	y := t.Time.Year()
	return fmt.Sprintf("%02d-%s-%d", d, m, y)
}

func DatePublish(y, m, d null.Int16) string {
	const (
		yx = "????"
		mx = "???"
		dx = "??"
		sp = " "
	)
	ys, ms, ds := yx, mx, dx
	if y.Valid {
		if i := int(y.Int16); IsYear(i) {
			ys = strconv.Itoa(i)
		}
	}
	if m.Valid {
		if s := AbbrMonth(int(m.Int16)); s != "" {
			ms = s
		}
	}
	if d.Valid {
		if i := int(d.Int16); IsDay(i) {
			ds = strconv.Itoa(i)
		}
	}
	if isYearOnly := ys != yx && ms == mx && ds == dx; isYearOnly {
		return fmt.Sprintf("%s%s", strings.Repeat(sp, 7), ys)
	}
	if isInvalidDay := ys != yx && ms != mx && ds == dx; isInvalidDay {
		return fmt.Sprintf("%s%s-%s", strings.Repeat(sp, 3), ms, ys)
	}
	return fmt.Sprintf("%02d-%s-%s", int(d.Int16), ms, ys)
}

func AbbrMonth(i int) string {
	const abbreviated = 3
	s := fmt.Sprint(time.Month(i))
	if len(s) >= abbreviated {
		return s[0:abbreviated]
	}
	return ""
}

func IsDay(i int) bool {
	const maxDay = 31
	if i > 0 && i <= maxDay {
		return true
	}
	return false
}

func IsYear(i int) bool {
	const unix = 1970
	now := time.Now().Year()
	if i >= unix && i <= now {
		return true
	}
	return false
}

func Icon(name null.String) string {
	const error = "unknown"
	if !name.Valid {
		return error
	}
	n := strings.ToLower(filepath.Ext(name.String))
	switch {
	case IsApp(n):
		return "comp2"
	case IsArchive(n):
		return "compressed"
	case IsImage(n):
		return "image2"
	case IsDocument(n):
		return "doc"
	case IsHTML(n):
		return "generic"
	case IsAudio(n):
		return "sound2"
	case IsTune(n):
		return "sound2"
	case IsVideo(n):
		return "movie"
	}
	return error
}

// /*
//  * Generates an icon determined by the file's extension
//  */
// function displayIcon(string filename) {
// 	var ext = ListLast(arguments.filename,".")
// 	if(ListFindNoCase(get(myapp).acceptedArchives,ext)) return "compressed";
// 	if(ListFindNoCase(get(myapp).acceptedAudio,ext)) return "sound2";
// 	if(ListFindNoCase(get(myapp).acceptedChiptunes,ext)) return "sound2";
// 	if(ListFindNoCase(get(myapp).acceptedDocuments,ext)) return "text";
// 	if(ListFindNoCase(get(myapp).acceptedGraphics,ext)) return "image2";
// 	if(ListFindNoCase(get(myapp).acceptedNoPreviews,ext)) return "text";
// 	if(ListFindNoCase(get(myapp).acceptedPrograms,ext)) return "comp2";
// 	if(ListFindNoCase(get(myapp).acceptedVideos,ext)) return "movie";
// 	return "unknown";
// }

// 7z,arc,ark,arj,cab,gz,lha,lzh,rar,tar,tar.gz,zip"
// loc.myapp.acceptedArchives		= "7z,arc,ark,arj,cab,gz,lha,lzh,rar,tar,tar.gz,zip"
// loc.myapp.acceptedDirChrs		= "[^a-z0-9\-\,\& ]"
// loc.myapp.acceptedAudio			= "au,flac,m1a,m2a,mid,midi,mp1,mp2,mp3,mpa,mpga,mpeg,ogg,snd,wav,wave,wma"
// loc.myapp.acceptedChiptunes		= "it,mod,s3m,xm"
// loc.myapp.acceptedDocuments		= "1st,asc,ans,cap,diz,doc,dox,me,nfo,pcb,txt,unp"
// loc.myapp.acceptedGraphics		= "bmp,gif,ico,jpg,jpeg,pdf,png,pcx"
// loc.myapp.acceptedNoPreviews	= ""
// loc.myapp.acceptedPrograms		= "exe,com"
// loc.myapp.acceptedVideos		= "avi,divx,flv,gt,mov,m4a,m4v,mp4,swf,rm,ram,wmv,xvid"
// // blacklistedExt notes: dbm is ColdFusion server, lex is Lucee extension archive
// loc.myapp.blacklistedExt		= "cfm,cfml,cfc,cgi,dbm,lex,lucee,jsp,php,shtml"

func IsApp(name string) bool {
	s := []string{".exe", ".com"}
	return IsValidExt(name, s...)
}

func IsArchive(name string) bool {
	s := []string{".7z", ".arc", ".ark", ".arj", ".cab", ".gz", ".lha", ".lzh", ".rar", ".tar", ".tar.gz", ".zip"}
	return IsValidExt(name, s...)
}

func IsDocument(name string) bool {
	s := []string{".1st", ".asc", ".ans", ".cap", ".diz", ".doc", ".dox", ".me", ".nfo", ".pcb", ".pdf", ".txt", ".unp"}
	return IsValidExt(name, s...)
}

func IsImage(name string) bool {
	s := []string{".bmp", ".gif", ".ico", ".iff", ".jpg", ".jpeg", ".lbm", ".png", ".pcx"}
	return IsValidExt(name, s...)
}

func IsHTML(name string) bool {
	s := []string{".htm", ".html"}
	return IsValidExt(name, s...)
}

func IsAudio(name string) bool {
	s := []string{".au", ".flac", ".m1a", ".m2a", ".mid", ".midi", ".mp1", ".mp2", ".mp3",
		".mpa", ".mpga", ".mpeg", ".ogg", ".snd", ".wav", ".wave", ".wma"}
	return IsValidExt(name, s...)
}

func IsTune(name string) bool {
	s := []string{".it", ".mod", ".s3m", ".xm"}
	return IsValidExt(name, s...)
}

func IsVideo(name string) bool {
	s := []string{".avi", ".divx", ".flv", ".gt", ".mov", ".m4a", ".m4v", ".mp4", ".swf", ".rm", ".ram", ".wmv", ".xvid"}
	return IsValidExt(name, s...)
}

func IsValidExt(name string, valid ...string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return helpers.IsValid(name, ext)
}
