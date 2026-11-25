//nolint:exhaustruct_v5
package command

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/Defacto2/magicnumber"
	"github.com/bengarrett/bbs"
	"github.com/bengarrett/sauce"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	bbsMaxCols = 80
	bbsMaxRows = 50
	bbsGlyphW  = 8
	bbsGlyphH  = 16
)

// bbsImager checks whether src contains non-ANSI BBS color codes (such as PCBoard
// @X codes or Renegade pipes) and, if so, renders a static 8x16 VGA image via
// [bbs.BBS.Image] and optimizes the resulting PNG preview, WebP preview, and WebP thumbnail.
func (ds Dirs) bbsImager(ctx context.Context, sl *slog.Logger, src, unid string) (bool, error) {
	raw, err := os.ReadFile(filepath.Clean(src))
	if err != nil {
		return false, fmt.Errorf("read source: %w", err)
	}

	body := prepareBBSBytes(raw)
	if len(body) == 0 || magicnumber.CSI(bytes.NewReader(body)) {
		return false, nil
	}

	kind := bbs.Find(bytes.NewReader(body))
	if !kind.Valid() || kind == bbs.ANSI {
		return false, nil
	}

	img, err := kind.Image(body)
	if err != nil {
		return false, fmt.Errorf("render bbs image: %w", err)
	}
	frame := normalizeBBSFrame(img)

	framePNG, err := writeFramePNG(sl, frame)
	if err != nil {
		return false, fmt.Errorf("write frame png: %w", err)
	}
	if err := ds.optiAnsilove(ctx, sl, unid, framePNG); err != nil {
		return false, fmt.Errorf("optimize frame: %w", err)
	}
	return true, nil
}

func normalizeBBSFrame(img image.Image) image.Image {
	const alphaOpaque = 0xff
	targetW := bbsMaxCols * bbsGlyphW
	maxH := bbsMaxRows * bbsGlyphH

	bounds := img.Bounds()
	targetH := min(bounds.Dy(), maxH)
	if bounds.Dx() == targetW && bounds.Dy() == targetH {
		return img
	}

	dst := image.NewNRGBA(image.Rect(0, 0, targetW, targetH))
	bg := color.NRGBA{R: 0, G: 0, B: 0, A: alphaOpaque}
	draw.Draw(dst, dst.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(0, 0, min(bounds.Dx(), targetW), targetH), img, bounds.Min, draw.Src)
	return dst
}

func prepareBBSBytes(raw []byte) []byte {
	const (
		asciiNUL = 0x00
		asciiEOF = 0x1a
	)

	body := sauce.Trim(raw)
	body = bytes.TrimRight(body, string([]byte{asciiEOF}))
	if len(body) == 0 {
		return nil
	}

	if hasHighBitByte(body) && utf8.Valid(body) {
		enc := encoding.ReplaceUnsupported(charmap.CodePage437.NewEncoder())
		if cp437, _, err := transform.Bytes(enc, body); err == nil {
			body = cp437
		}
	}

	out := bytes.Clone(body)
	for i, b := range out {
		if b == asciiNUL {
			out[i] = ' '
		}
	}
	return out
}

func hasHighBitByte(b []byte) bool {
	for _, c := range b {
		if c >= utf8.RuneSelf {
			return true
		}
	}
	return false
}
