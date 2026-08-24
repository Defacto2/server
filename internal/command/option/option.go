package option

// TODO: create a custom test image, maybe a screenshot of the website,
// then test all the arguments and dump the output in the home directory,
// so they can be manually checked:
// - name the files after the Test names
// - have the tests only run manually, search for maybe a param to auto disable unless using -run ?

// Opts represents the command line options to use with a terminal application.
//
// Each option and value in the slice must use separate strings.
type Opts []string

func Join(arg ...string) Opts {
	if len(arg) == 0 {
		return nil
	}
	return Opts(arg)
}

// AnsiAmiga appends the command line arguments for the [ansilove command]
// to transform a Commodore Amiga ANSI text file into a PNG image.
//
// [ansilove command]: https://github.com/ansilove/ansilove
func (o *Opts) AnsiAmiga(src, tmp string) {
	if o == nil {
		return
	}

	// SYNOPSIS
	//    ansilove [-dhiqrsSv] [-b bits] [-c columns] [-f font] [-m mode] [-o file]
	//             [-R factor] [-t type] file
	*o = append(*o,
		"-f", "topaz+", // use Topaz plus font
		"-m", "workbench", // use an Amiga palette
		"-S", // use sauce metadata
	)
	if tmp != "" {
		*o = append(*o,
			"-o", tmp)
	}
	if src != "" {
		*o = append(*o, src)
	}
}

// AnsiMsDos appends the command line arguments for the [ansilove command] to
// transform an ANSI text file into a PNG image.
//
// [ansilove command]: https://github.com/ansilove/ansilove
func (o *Opts) AnsiDOS(src, tmp string) {
	if o == nil {
		return
	}

	// common font values: 80x25, 80x50, "spleen", "terminus"
	*o = append(*o,
		"-d",          // apply DOS aspect-radio
		"-f", "80x50", // use VGA hires font
		"-i", // use iCE colors
		"-S", // use sauce metadata
	)
	if tmp != "" {
		*o = append(*o,
			"-o", tmp)
	}
	if src != "" {
		*o = append(*o, src)
	}
}

// Pixelate appends ImageMagick v6 flags to
// downscale and upscale an image, producing a blocky pixellation effect.
// TODO: update flags to version v7.
func (o *Opts) Pixelate(path string) {
	if o == nil {
		return
	}

	*o = append(*o,
		path,
		"-scale", "5%", // first downscale
		"-scale", "2000%", // then upscale
		path,
	)
}

const (
	gravity     = "-gravity"
	trim        = "-trim"
	extent      = "-extent"
	size400x400 = "400x400" // 400 x 400 pixel image size

)

// ThumbAlignment appends the command line arguments for the magick command to
// transform an image into a 400 x 400 pixel image using the gravity alignment.
func (o *Opts) ThumbAlignment(src, tmp string, align int) {
	if o == nil {
		return
	}

	if src != "" {
		*o = append(*o, src)
	}
	switch align {
	case 0: // top
		*o = append(*o, gravity, "North")
	case 1: // middle
		*o = append(*o, gravity, "center")
	case 2: // bottom
		*o = append(*o, gravity, "South")
	case 3: // left
		*o = append(*o, gravity, "West")
	case 4: // right
		*o = append(*o, gravity, "East")
	default:
	}
	*o = append(*o,
		trim, extent,
		size400x400,
	)
	if tmp != "" {
		*o = append(*o, tmp)
	}
}

func (o *Opts) CropAlignment(src, tmp string, crop int) {
	if o == nil {
		return
	}

	if src != "" {
		*o = append(*o, src)
	}
	*o = append(*o, gravity, "North")
	switch crop {
	case 0: // 1:1
		*o = append(*o, extent, "1:1")
	case 1: // 4:3
		*o = append(*o, extent, "4:3")
	case 2: // 1:2
		*o = append(*o, extent, "1:2")
	default:
	}
	if tmp != "" {
		*o = append(*o, tmp)
	}
}

// Gif2webp appends the command line arguments for the [gif2webp command] to transform args GIF image into args webp image.
//
// [gif2webp command]: https://developers.google.com/speed/webp/docs/gif2webp
func (o *Opts) Gif2webp(src, tmp string) {
	if o == nil {
		return
	}

	if src != "" {
		*o = append(*o, src)
	}
	*o = append(*o,
		"-q", "100", // compression factor for RBG values: 0-100
		"-mt", // use multi-threading
	)
	if tmp != "" {
		*o = append(*o,
			"-o", tmp,
		)
	}
}

// PNGPixel appends the command line arguments for the magick command to transform
// a screenshot or image with text to a PNG image.
func (o *Opts) PNGPixel(thumbnail bool, src, tmp string) {
	if o == nil {
		return
	}

	if src != "" {
		*o = append(*o, src)
	}
	// define compression options, that replace the "-quality" options
	const define = "-define"
	*o = append(*o,
		define, "png:compression-filter=5",
		define, "png:compression-level=9",
		define, "png:compression-strategy=1",
		define, "png:exclude-chunk=all",
	)
	*o = append(*o,
		"-flatten",
		"-strip",            // strip the src image of any profiles, comments, etc
		"-posterize", "136", // reduce the image to 136 color levels per channel
	)
	if thumbnail {
		o.Thumb()
	}
	if tmp != "" {
		*o = append(*o, tmp)
	}
}

// JPGPhoto appends the command line arguments for the magick command to transform
// a photo to a JPEG image.
func (o *Opts) JPGPhoto(thumbnail bool, src, tmp string) {
	if o == nil {
		return
	}
	if src != "" {
		*o = append(*o, src)
	}
	// previously the gaussian blur flag was used to improve compression,
	// however it caused some images to break and it shouldn't be used.
	*o = append(*o,
		"-strip",         // strip the image of any profiles or comments
		"-quality", "75", // lossy compression quality
	)
	if thumbnail {
		o.Thumb()
	}
	if tmp != "" {
		*o = append(*o, tmp)
	}
}

// WebpPhoto appends the command line arguments for the [cwebp command] to transform
// a photo to the WebP format.
//
// [cwebp command]: https://developers.google.com/speed/webp/docs/cwebp
func (o *Opts) WebpPhoto(src, tmp string) {
	if o == nil {
		return
	}
	*o = append(*o,
		"-lossless", // use lossless compression
		"-exact",    // preserve RGB values
		"-mt",       // use multi-threading
	)
	// input
	if src != "" {
		*o = append(*o, src)
	}
	// output
	if tmp != "" {
		*o = append(*o,
			"-o", tmp,
		)
	}
}

// WebpText appends the command line arguments for the [cwebp command] to transform
// a screenshot or image with text to the WebP format.
//
// [cwebp command]: https://developers.google.com/speed/webp/docs/cwebp
func (o *Opts) WebpPixel(src, tmp string) {
	if o == nil {
		return
	}
	*o = append(*o,
		"-preset", "text",
		"-z", "6", // lossless compression level
		"-mt", // use multi-threading
	)
	// input
	if src != "" {
		*o = append(*o, src)
	}
	// output
	if tmp != "" {
		*o = append(*o,
			"-o", tmp,
		)
	}
}

// Thumbnail appends the command line arguments for magick to transform
// an image into a squared, 400 x 400 pixel thumbnail.
func (o *Opts) Thumb() {
	*o = append(*o,
		"-filter", "Triangle", // resize filter
		"-thumbnail",
		size400x400, // the use of thumbnail is more performant than other flags
		"-background", "#999",
		gravity, "center",
		extent, size400x400, // image size and offset
	)
}
