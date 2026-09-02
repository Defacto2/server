// Package runner is used to build test images for a visual confirmationthat will be outputed to the terminal.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
)

// Runner runs through all the methods and funcs that manipulate photos and images.
// The files are placed in a temporary directory so they can be individually
// checked and confirmed. The directory location is displayed when running.
//
// INFO: To use, run `go run internal/runner/runner.go` from the repo root.

const (
	screenshot = "SCREEN.PNG"
	screenSize = 328_468
	pictImg    = "PICTURE-IMAGER"
)

func printOut(a ...any) {
	if _, err := fmt.Fprintln(os.Stdout, a...); err != nil {
		log.Fatal(a, err)
	}
}

var testdata = filepath.Join("internal", "command", "testdata")

func main() { //nolint:funlen

	sl := slog.Default()
	ctx := context.Background()

	// Application file handlers

	lookup(ctx, sl)

	src, dstDir := initData()
	printOut("Temporary working directory:", dstDir)

	dst := filepath.Join(dstDir, screenshot)
	copyScreen(sl, src, dst)
	printOut("Copied source screenshot:", dst)

	// Image file handlers

	printOut("Running Images Pixelate")
	const timeout = 15 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)

	err := command.ImagesPixelate(ctx, sl, "SCREEN", dstDir)
	if err != nil {
		cancel()
		log.Fatal(err)
	}
	newpath := filepath.Join(dstDir, "SCREEN-PIXELATE.PNG")
	err = os.Rename(dst, newpath)
	if err != nil {
		cancel()
		log.Fatal(err)
	}
	copyScreen(sl, src, dst)

	alignThumbs(ctx, sl, src, dst, dstDir)
	cropImages(ctx, sl, src, dst, dstDir)

	d := command.Dirs{
		Download:  "",
		Preview:   dir.Directory(dstDir),
		Thumbnail: dir.Directory(dstDir),
		Extra:     dir.Directory(dstDir),
	}
	err = d.PictureImager(ctx, sl, src, pictImg)
	if err != nil {
		cancel()
		log.Fatal(err)
	}

	const pixel = "GENERATE_THUMB_PIXEL"
	gthImg := filepath.Join(dstDir, pixel) + ".PNG"
	copyScreen(sl, src, gthImg)
	err = d.Thumbs(ctx, sl, pixel, command.Pixel)
	if err != nil {
		cancel()
		log.Fatal(err)
	}

	const photo = "GENERATE_THUMB_PHOTO"
	gthImg = filepath.Join(dstDir, photo) + ".PNG"
	copyScreen(sl, src, gthImg)
	err = d.Thumbs(ctx, sl, photo, command.Photo)
	if err != nil {
		cancel()
		log.Fatal(err)
	}

	const webp = "OPTIMIZE.PNG"
	srcWebp := filepath.Join(testdata, "TEST.PNG")
	dstWebp := filepath.Join(dstDir, webp)
	err = command.CopyFile(sl, srcWebp, dstWebp)
	if err != nil {
		cancel()
		log.Fatal(err)
	}
	err = command.OptimizePNG(ctx, sl, dstWebp)
	if err != nil {
		cancel()
		log.Fatal(err)
	}

	// Text file handlers

	txt := filepath.Join(testdata, "TEST.ASCII")

	err = d.TextDeferred(ctx, sl, txt, "DEFERRED_TXT")
	if err != nil {
		cancel()
		log.Fatal(err)
	}
	err = d.DizDeferred(sl, txt, "DEFERRED_DIZ")
	if err != nil {
		cancel()
		log.Fatal(err)
	}

	err = d.TextImager(ctx, sl, txt, "ASCII-DOS", false) // TODO: broken output
	if err != nil {
		cancel()
		log.Fatal(err)
	}
	err = d.TextImager(ctx, sl, txt, "ASCII-AMIGA", true) // TODO: broken output
	if err != nil {
		cancel()
		log.Fatal(err)
	}

	t := command.Text{
		UUID:    "SCREEN",
		MaxRows: 1,
		MaxCols: 0,
		UTF8:    false,
	}
	loc, err := t.Crop(sl, txt)
	if err != nil {
		cancel()
		log.Fatal(err)
	}
	newpath = filepath.Join(dstDir, "ASCII-1rows.txt")
	err = os.Rename(loc, newpath)
	if err != nil {
		cancel()
		log.Fatal(err)
	}

	printOut("TEST.ASCII cropped:", newpath)
	cancel()
}

func initData() (src, dstDir string) { //nolint:nonamedreturns
	src = filepath.Join(testdata, screenshot)
	st, err := os.Stat(src)
	if err != nil {
		log.Fatal(err)
	}
	if st.Size() != screenSize {
		log.Fatal("unexpected file size for src")
	}

	dstDir, err = os.MkdirTemp("", "df2app-*")
	if err != nil {
		log.Fatal(err)
	}

	return src, dstDir
}

func lookup(ctx context.Context, sl *slog.Logger) {
	path, err := command.Lookup(command.Ansilove)
	if err != nil {
		sl.Info("missing command", slog.Any("error", err))
		return
	}
	if path != "" {
		printOut("found command ansilove at", path)
	}
	r := command.Runner{
		Timeout:    0,
		Log:        sl,
		WorkingDir: "",
	}
	out, err := r.Run(ctx, command.Ansilove, "-v")
	if err != nil {
		sl.Info("error with command", slog.Any("error", err))
		return
	}
	printOut(string(out))
}

func copyScreen(sl *slog.Logger, src, dst string) {
	err := command.CopyFile(sl, src, dst)
	if err != nil {
		log.Fatal(err)
	}
	st, err := os.Stat(dst)
	if err != nil {
		log.Fatal(err)
	}
	if st.Size() != screenSize {
		log.Fatal("unexpected file size for src:", src, "dst:", dst)
	}
}

func alignThumbs(ctx context.Context, sl *slog.Logger, src, dst, dstDir string) {
	aligns := [...]command.Align{
		command.Top,
		command.Middle,
		command.Bottom,
		command.Left,
		command.Right,
	}

	for _, align := range aligns {
		s := align.String()
		printOut("Running Align Thumbs:", s)
		preview := dir.Directory(testdata)
		thumbnail := dir.Directory(dstDir)
		err := align.Thumbs(ctx, sl, "SCREEN", preview, thumbnail)
		if err != nil {
			log.Fatal(err)
		}
		newpath := filepath.Join(dstDir, "SCREEN-THUMB-ALIGN-"+strings.ToUpper(s)+".PNG")
		err = os.Rename(dst, newpath)
		if err != nil {
			log.Fatal(err)
		}
		copyScreen(sl, src, dst)
	}
}

func cropImages(ctx context.Context, sl *slog.Logger, src, dst, dstDir string) {
	crops := [...]command.Crop{
		command.SquareTop,
		command.FourThree,
		command.OneTwo,
	}

	for _, crop := range crops {
		s := crop.String()
		printOut("Running Crop Images:", s)
		preview := dir.Directory(dstDir)
		err := crop.Images(ctx, sl, "SCREEN", preview)
		if err != nil {
			log.Fatal(err)
		}
		newpath := filepath.Join(dstDir, "SCREEN-CROP-IMAGE-"+strings.ToUpper(s)+".PNG")
		err = os.Rename(dst, newpath)
		if err != nil {
			log.Fatal(err)
		}
		copyScreen(sl, src, dst)
	}
}
