package readme_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/readme"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

var testLog = logs.Discard()

func TestNormalize(t *testing.T) {
	t.Parallel()

	t.Run("mask and crlf", func(t *testing.T) {
		t.Parallel()

		nulls := bytes.Repeat([]byte{0x00}, 50)
		text := []byte("hello world.")
		crlf := []byte("\r\n")
		eof := []byte{0x1a}
		serial := []byte(" Serial!! B014-56789A-BCDEFGH-G45X example << ")

		s := [][]byte{text, serial, text, crlf, eof, nulls}
		got := readme.Normalize(bytes.Join(s, []byte("")))

		wants := `hello world. Sxxxxx!! B014-5$$89A-BCDEFGH-G45X example << hello world.` +
			"\n\x1a" +
			strings.Repeat(` `, 50)
		be.Equal(t, string(got), wants)
	})

	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{"empty", nil, nil},
		{"whitespace only", []byte(" \t\r\n "), nil},
		{"eof only", []byte{0x1a, 0x1a}, nil},
		{"trailing eof", []byte("hello\x1a\x1a"), []byte("hello")},
		{"trailing egrave eof", []byte("hello\x8a\x1a"), []byte("hello")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := readme.Normalize(bytes.Clone(tt.input))
			be.Equal(t, got, tt.want)
		})
	}
}

func TestBuffersEmpty(t *testing.T) {
	t.Parallel()

	txt := readme.Text{}
	unused := new(bytes.Buffer)
	got0, got1, err := txt.Buffers(testLog)
	be.Err(t, err, nil)
	be.Equal(t, got0, unused)
	be.Equal(t, got1, unused)

	got0, got1, err = txt.Buffers(nil)
	be.Err(t, err, nil)
	be.Equal(t, got0, unused)
	be.Equal(t, got1, unused)
}

func TestBuffersPlainText(t *testing.T) {
	t.Parallel()

	const unid = "d9c025c8-44be-4ec5-a9ec-40463a39a722"
	testdata := filepath.Join(testutil.Testdata(), "uuid")

	txt := readme.Text{
		Download: dir.Directory(testdata),
		Extra:    dir.Directory(testdata),
		Filename: "testbuffers1.zip",
		MaxSize:  1024 * 100,
		UUID:     unid,
	}

	gotStr, gotRune, err := txt.Buffers(testLog)
	be.Err(t, err, nil)

	const diz = ":DIZ BEGIN:\n\n  <- TAB\n  A placeholder FILE_ID.DIZ\n\n: DIZ END :\n\n"
	const body = ":TXT BEGIN:\n\nHELLO WORLD!\n\n: TXT END :\n"
	const help = "\n:HLP BEGIN:\n\nhelper text...\n\n: HLP END :"
	const long = "skipped, text is too long"

	wants := diz + body + help
	got := gotStr.String()
	be.Equal(t, got, wants)

	got = gotRune.String()
	be.Equal(t, got, wants)

	// test the too long feedback
	txt.MaxSize = 1 // 1 byte
	gotStr, gotRune, err = txt.Buffers(testLog)
	be.Err(t, err, nil)

	got = gotStr.String()
	be.Equal(t, got, long)

	got = gotRune.String()
	be.Equal(t, got, long)
}

func TestBuffersANSI(t *testing.T) {
	t.Parallel()

	const unid = "3e54b770-e71c-4e95-aa8c-f898e6b18f77"
	testdata := filepath.Join(testutil.Testdata(), "uuid")

	txt := readme.Text{
		Download: dir.Directory(testdata),
		Extra:    dir.Directory(testdata),
		Filename: "testbuffers2.zip",
		MaxSize:  1024 * 100,
		UUID:     unid,
	}
	got0, got1, err := txt.Buffers(testLog)
	be.Err(t, err, nil)

	const wants = `<div style="color:#aaa;background-color:#000;"><span style="color:#aaa;">:TXT CSI BEGIN:</span>` +
		"\n\n" +
		`<span style="color:#fff;">hello world</span>` +
		"\n\n" +
		`<span style="color:#aaa;">: TXT CSI END :</span>` +
		`</div>`
	got := got0.String()
	be.Equal(t, got, wants)

	got = got1.String()
	be.Equal(t, got, "")
}

func TestBuffersDOSBinary(t *testing.T) {
	t.Parallel()

	const unid = "3e54b770-e71c-4e95-aa8c-f898e6b18f77"
	const body = "hello world!!!"
	dest, err := testutil.MkDOSBIN(t, unid+".txt", body)
	be.Err(t, err, nil)
	temp := filepath.Dir(dest)

	txt := readme.Text{
		Download: dir.Directory(temp),
		Extra:    dir.Directory(temp),
		Filename: "hello.bin",
		MaxSize:  1024 * 100,
		Sign:     magicnumber.Unknown,
		UUID:     unid,
	}
	got0, _, err := txt.Buffers(testLog)
	be.Err(t, err, nil)
	got := got0.String()
	be.True(t, len(got) > 0)

	const pre = `<div><span style="color:#fff;background-color:#000;">`
	const charsPerLine = 80
	padding := bytes.Repeat([]byte{0x20}, charsPerLine-len(body))

	wants := pre + body + string(padding)
	be.True(t, strings.HasPrefix(got, wants))

	const suf = "</span>\n</div>"
	be.True(t, strings.HasSuffix(got, suf))
}

func TestBuffersSAUCEBinaryText(t *testing.T) {
	t.Parallel()

	makeSAUCEBin := func(halfCols, rows, fillChar, fillAttr byte) []byte {
		payloadLen := uint32(halfCols) * 2 * uint32(rows) * 2
		const sauceLen = 128
		out := make([]byte, payloadLen+1+sauceLen)
		for i := uint32(0); i < payloadLen; i += 2 {
			out[i] = fillChar
			out[i+1] = fillAttr
		}
		// Place a cell with 0x00 attr, a 0x0D char + 0x0A attr (CRLF pair), a 0x00 char (NUL blank), and 'Z'.
		out[0], out[1] = 'A', 0x00
		out[2], out[3] = 0x0D, 0x0A
		out[4], out[5] = 0x00, 0x07
		out[6], out[7] = 'Z', 0x0F

		out[payloadLen] = 0x1A

		sauce := out[payloadLen+1:]
		copy(sauce[:7], "SAUCE00")
		copy(sauce[82:90], "19960101")
		binary.LittleEndian.PutUint32(sauce[90:94], payloadLen)
		sauce[94] = 5 // BinaryText
		sauce[95] = halfCols

		return out
	}

	for _, halfCols := range []byte{80, 160} {
		cols := int(halfCols) * 2
		t.Run(strconv.Itoa(cols)+"cols", func(t *testing.T) {
			t.Parallel()

			data := makeSAUCEBin(halfCols, 2, ' ', 0x07)
			dlDir := t.TempDir()
			extraDir := t.TempDir()
			const unid = "a1b2c3d4-0000-0000-0000-000000000001"
			be.Err(t, os.WriteFile(filepath.Join(dlDir, unid), data, 0o600), nil)

			txt := readme.Text{
				Download: dir.Directory(dlDir),
				Extra:    dir.Directory(extraDir),
				Filename: "test.bin",
				Platform: "ansi",
				Year:     1996,
				MaxSize:  1024 * 1024,
				UUID:     unid,
			}
			got0, got1, err := txt.Buffers(testLog)
			be.Err(t, err, nil)
			be.Equal(t, got1.Len(), 0)

			html := got0.String()
			// Must preserve 0x00 black-on-black attribute on 'A', convert 0x00 char to space (no \x00), and keep 'Z' at cell 4.
			be.True(t, !strings.Contains(html, "\x00"))
			be.True(t, strings.Contains(html, `<span style="color:#000;background-color:#000;">A</span>`))
			be.True(t, strings.Contains(html, `<span style="color:#aaa;background-color:#000;"> </span>`))
			be.True(t, strings.Contains(html, `<span style="color:#fff;background-color:#000;">Z`))
			// Exactly 2 rows -> 2 newlines inside <div>...</div>.
			be.Equal(t, strings.Count(html, "\n"), 2)
			// First row must contain all cols characters (no CRLF collapse or wrong wrap).
			firstLine, _, ok := strings.Cut(html, "\n")
			be.True(t, ok)
			be.True(t, strings.HasSuffix(firstLine, strings.Repeat(" ", cols-4)+"</span>"))
		})
	}
}

func TestBuffersICEDraw(t *testing.T) {
	t.Parallel()

	const (
		hdrLen  = 12
		fontLen = 4096
		palLen  = 48
	)
	body := []byte{'I', 0x03, 'D', 0x0F, 'F', 0x0F, 0x00, 0x00}
	idf := make([]byte, hdrLen+len(body)+fontLen+palLen)
	copy(idf[:4], "\x041.4")
	binary.LittleEndian.PutUint16(idf[8:10], 3) // width = 4
	copy(idf[hdrLen:], body)
	palOff := hdrLen + len(body) + fontLen
	idf[palOff+3*3], idf[palOff+3*3+1], idf[palOff+3*3+2] = 33, 19, 1
	idf[palOff+15*3], idf[palOff+15*3+1], idf[palOff+15*3+2] = 63, 63, 63

	dlDir := t.TempDir()
	extraDir := t.TempDir()
	const unid = "c1d2e3f4-0000-0000-0000-000000000001"
	be.Err(t, os.WriteFile(filepath.Join(dlDir, unid), idf, 0o600), nil)

	txt := readme.Text{
		Download: dir.Directory(dlDir),
		Extra:    dir.Directory(extraDir),
		Filename: "LS-ICE.IDF",
		Platform: "ansi",
		MaxSize:  1024 * 1024,
		UUID:     unid,
	}
	got0, got1, err := txt.Buffers(testLog)
	be.Err(t, err, nil)
	be.Equal(t, got1.Len(), 0)
	be.Equal(t, txt.Sign, magicnumber.ICEDraw)
	be.True(t, txt.CustomFont)

	html := got0.String()
	be.True(t, strings.Contains(html, `<span style="color:#864d04;background-color:#000000;">I</span>`))
	be.True(t, strings.Contains(html, `<span style="color:#ffffff;background-color:#000000;">DF</span>`))
}

func TestBuffersXBIN(t *testing.T) {
	t.Parallel()

	xb := []byte{
		'X', 'B', 'I', 'N', 0x1A,
		0x02, 0x00, // width = 2
		0x01, 0x00, // height = 1
		16,   // fontHeight
		0x0C, // Compress(4) | NonBlink(8)
		0x01, 'X', 0x9F, 'B', 0x0F,
	}

	dlDir := t.TempDir()
	extraDir := t.TempDir()
	const unid = "d1e2f3a4-0000-0000-0000-000000000001"
	be.Err(t, os.WriteFile(filepath.Join(dlDir, unid), xb, 0o600), nil)

	txt := readme.Text{
		Download: dir.Directory(dlDir),
		Extra:    dir.Directory(extraDir),
		Filename: "CT-XBIN.XB",
		Platform: "ansi",
		MaxSize:  1024 * 1024,
		UUID:     unid,
	}
	got0, got1, err := txt.Buffers(testLog)
	be.Err(t, err, nil)
	be.Equal(t, got1.Len(), 0)
	be.Equal(t, txt.Sign, magicnumber.XBinaryText)
	be.True(t, !txt.CustomFont)

	html := got0.String()
	be.True(t, strings.Contains(html, `<span style="color:#ffffff;background-color:#5555ff;">X</span>`))
	be.True(t, strings.Contains(html, `<span style="color:#ffffff;background-color:#000000;">B</span>`))
}

func TestBuffersBSAVE(t *testing.T) {
	t.Parallel()

	const (
		headerLen  = 7
		payloadLen = 4000 // 80x25 CGA video memory dump
	)
	bsave := make([]byte, headerLen+payloadLen)
	bsave[0] = 0xFD
	binary.LittleEndian.PutUint16(bsave[1:3], 0xB800)
	binary.LittleEndian.PutUint16(bsave[5:7], payloadLen)
	for i := headerLen; i < len(bsave); i += 2 {
		bsave[i] = ' '
		bsave[i+1] = 0x07
	}
	bsave[headerLen], bsave[headerLen+1] = 'C', 0x0F
	bsave[headerLen+2], bsave[headerLen+3] = 'G', 0x0F
	bsave[headerLen+4], bsave[headerLen+5] = 'A', 0x0F

	dlDir := t.TempDir()
	extraDir := t.TempDir()
	const unid = "e1e2e3e4-0000-0000-0000-000000000001"
	be.Err(t, os.WriteFile(filepath.Join(dlDir, unid), bsave, 0o600), nil)

	txt := readme.Text{
		Download: dir.Directory(dlDir),
		Extra:    dir.Directory(extraDir),
		Filename: "screen.bin",
		Platform: "ansi",
		Year:     1988,
		MaxSize:  1024 * 1024,
		UUID:     unid,
	}
	got0, got1, err := txt.Buffers(testLog)
	be.Err(t, err, nil)
	be.Equal(t, got1.Len(), 0)
	be.Equal(t, txt.Sign, magicnumber.CGAVideoDump)
	be.True(t, strings.Contains(got0.String(), "CGA"))
}

func TestBuffersSkippedImage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []byte
		want magicnumber.Signature
	}{
		{
			name: "gif",
			data: []byte("GIF89a\x01\x01\x01\x01\x80\x01\x01\xff\xff\xff\x01\x01\x01;"),
			want: magicnumber.GraphicsInterchangeFormat,
		},
		{
			name: "png",
			data: []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n', 0, 0, 0, 13, 'I', 'H', 'D', 'R'},
			want: magicnumber.PortableNetworkGraphics,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dlDir := t.TempDir()
			extraDir := t.TempDir()
			const unid = "f1f2f3f4-0000-0000-0000-000000000001"
			be.Err(t, os.WriteFile(filepath.Join(dlDir, unid), tt.data, 0o600), nil)

			txt := readme.Text{
				Download: dir.Directory(dlDir),
				Extra:    dir.Directory(extraDir),
				Filename: "fake.txt",
				Platform: "text",
				MaxSize:  1024 * 1024,
				UUID:     unid,
			}
			got0, got1, err := txt.Buffers(testLog)
			be.Err(t, err, nil)
			be.Equal(t, got0.Len(), 0)
			be.Equal(t, got1.Len(), 0)
			be.Equal(t, txt.Sign, tt.want)
		})
	}
}

func TestBuffersAmigaANSI(t *testing.T) {
	t.Parallel()

	body := []byte("\x1b[31mAmiga\x1b[0m")
	const sauceLen = 128
	data := make([]byte, len(body)+1+sauceLen)
	copy(data, body)
	data[len(body)] = 0x1A
	s := data[len(body)+1:]
	copy(s[:7], "SAUCE00")
	copy(s[82:90], "19960101")
	binary.LittleEndian.PutUint32(s[90:94], uint32(len(body)))
	s[94] = 1                                   // Character
	s[95] = 1                                   // ANSi
	binary.LittleEndian.PutUint16(s[96:98], 80) // TInfo1: character width
	s[105] = 1                                  // TFlags: iCE color non-blink

	dlDir := t.TempDir()
	extraDir := t.TempDir()
	const unid = "11223344-0000-0000-0000-000000000001"
	be.Err(t, os.WriteFile(filepath.Join(dlDir, unid), data, 0o600), nil)

	txt := readme.Text{
		Download: dir.Directory(dlDir),
		Extra:    dir.Directory(extraDir),
		Filename: "art.ans",
		Platform: "textamiga",
		MaxSize:  1024 * 1024,
		UUID:     unid,
	}
	got0, got1, err := txt.Buffers(testLog)
	be.Err(t, err, nil)
	be.Equal(t, got1.Len(), 0)
	be.Equal(t, txt.Record.ID, "SAUCE")
	be.True(t, strings.Contains(got0.String(), "Amiga"))
}

func TestBuffersUnknownSAUCE(t *testing.T) {
	t.Parallel()

	const sauceLen = 128
	data := make([]byte, 32+1+sauceLen)
	copy(data[:4], []byte{'P', 'K', 0x03, 0x04})
	data[32] = 0x1A
	s := data[33:]
	copy(s[:7], "SAUCE00")
	copy(s[7:42], "Demo Title")
	copy(s[82:90], "19950601")

	dlDir := t.TempDir()
	extraDir := t.TempDir()
	const unid = "22334455-0000-0000-0000-000000000001"
	be.Err(t, os.WriteFile(filepath.Join(dlDir, unid), data, 0o600), nil)

	txt := readme.Text{
		Download: dir.Directory(dlDir),
		Extra:    dir.Directory(extraDir),
		Filename: "demo.zip",
		Platform: "dos",
		MaxSize:  1024 * 1024,
		UUID:     unid,
	}
	got0, got1, err := txt.Buffers(testLog)
	be.Err(t, err, nil)
	be.Equal(t, got0.Len(), 0)
	be.Equal(t, got1.Len(), 0)
	be.Equal(t, txt.Record.ID, "SAUCE")
	be.Equal(t, strings.TrimSpace(txt.Record.Title), "Demo Title")
}
