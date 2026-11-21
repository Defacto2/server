# XBIN `.XB` Handler (`internal/xbin`)

[xbin.go](file:///home/ben/github/server/internal/xbin/xbin.go) decodes **XBIN (eXtended BINary / `.XB`)** character-graphics files into styled HTML (`<div>...</div>` with `<span style="color:#rrggbb;background-color:#rrggbb;">...</span>` runs) using the shared renderer in [vga.go](file:///home/ben/github/server/internal/vga/vga.go) for the in-browser readme viewer (`#precontentBinary`).

No third-party source code was copied or ported for this implementation. The decoder was written from scratch in this repository from direct binary inspection of sample `.XB` artifacts, existing Defacto2 conventions ([icedraw.go](file:///home/ben/github/server/internal/icedraw/icedraw.go) and [thedraw.go](file:///home/ben/github/server/internal/thedraw/thedraw.go)), and the external format references below.

---

## 1. Primary Empirical Sources (Sample Binary Analysis)

The decoder was verified against the five standalone `.XB` artifacts from `XBIN.zip` (`/f/aa1c8ef`, dated `1996-08-01`) and `magicnumber`'s test fixture:

| Artifact / File | Filename | Size | Dimensions | `fontHeight` | `flags` | Notes |
| --- | --- | ---: | --- | ---: | --- | --- |
| `/f/ab1ccc7` | `US-XBIN.XB` | 6,317 B | 80 × 22 | 16 | `0x0f` | Palette + Font + RLE + iCE |
| `/f/b12db5f` | `UN-ACID.XB` | 5,138 B | 80 × 15 | 16 | `0x0f` | Palette + Font + RLE + iCE |
| `/f/ab1f1ff` | `CT-XBIN.XB` | 6,217 B | 80 × 20 | 16 | `0x0f` | Palette + Font + RLE (`1,933` B → `1,600` cells) + iCE |
| `/f/b02b24f` | `T1-XBIN.XB` | 5,330 B | 80 × 22 | 16 | `0x0f` | Palette + Font + RLE (`1,046` B → `1,760` cells) + iCE |
| `/f/b23006f` | `UN-ACID2.XB` | 5,208 B | 80 × 18 | 16 | `0x0f` | Palette + Font + RLE + iCE |
| [binarytxt.xb](file:///home/ben/github/magicnumber/testdata/binarytxt.xb) | `binarytxt.xb` | 8,413 B | 80 × 25 | 16 | `0x03` | Palette + Font, uncompressed, standard blink |

### XBIN Binary Layout

1. **Header (`0x00`–`0x0A`, 11 bytes):**
   - `0x00`–`0x04`: Magic `"XBIN\x1a"` (`58 42 49 4E 1A`)
   - `0x05`–`0x06`: `width` (`uint16` LE)
   - `0x07`–`0x08`: `height` (`uint16` LE)
   - `0x09`: `fontHeight` (`uint8`, scanlines per glyph; `16` in all inspected samples)
   - `0x0A`: `flags` bitfield:
     - `0x01` (`Palette`): 48-byte 6-bit VGA DAC palette (`16 × 3` bytes, `0x00`–`0x3F`) follows header.
     - `0x02` (`Font`): raw 1-bit-per-pixel font (`fontHeight × 256` bytes, or `fontHeight × 512` if `0x10` is set) follows palette/header.
     - `0x04` (`Compress`): image grid uses XBIN RLE compression instead of raw `[char, attr]` pairs.
     - `0x08` (`NonBlink`): attribute bit 7 selects high-intensity backgrounds (`0..15`, iCE color) instead of blinking.
     - `0x10` (`512Chars`): 512-glyph font mode; attribute bit 3 selects glyph bank `256..511`, restricting foreground to colors `0..7`.
2. **Image Data Stream (`flags & 0x04`):**
   - Uncompressed: `width × height` sequential `[ch, attr]` byte pairs.
   - Compressed: each run starts with control byte `ctrl`, where `mode = (ctrl >> 6) & 0x03` and `count = (ctrl & 0x3F) + 1`:
     - `0`: `count` literal `[ch, attr]` pairs (`2 × count` bytes)
     - `1`: 1 `ch` followed by `count` `attr` bytes
     - `2`: 1 `attr` followed by `count` `ch` bytes
     - `3`: 1 `[ch, attr]` pair repeated `count` times
3. **Optional SAUCE Trailer:** All five ACiD samples end with `0x1A` + 128-byte `SAUCE00` (`DataType = 6`, `FileType = 0`, `TInfo = (width, height, 0, 0)`, `TFlags = 0x00`, `TInfoS = ""`). Per the SAUCE spec, `TFlags` and `TInfoS` are unused for XBIN because XBIN carries its own header, palette, and font. `sauce.Trim` strips the trailer before decoding.

### Key Empirical Finding on Embedded Fonts

Comparing the 4,096-byte embedded fonts across `ab1ccc7`, `b12db5f`, `ab1f1ff`, and `b23006f` showed every file embeds a distinct per-artwork glyph table (pairwise diffs of **72 to 153 glyphs** out of 256), redefining slots such as `0x01`–`0x1F`, `0x7F`, and `0xE2` into custom curves, bevels, and pixel tiles.

Because `<pre id="precontentBinary">` renders with the fixed `Ac437_IBM_VGA_8x16` / `Ac437_IBM_EGA_8x8` webfonts, [xbin.go](file:///home/ben/github/server/internal/xbin/xbin.go) skips the embedded font bytes (`HasFont`) and flags the UI so [artifact.tmpl](file:///home/ben/github/server/view/app/artifact.tmpl) displays `XB (custom fonts cannot be displayed; see the preview image above)` while the server-side `ansilove` PNG/WebP preview displays the exact custom glyphs.

---

## 2. Codebase Integration

- **[vga.go](file:///home/ben/github/server/internal/vga/vga.go):** Shared CP437 cell grid, 6-bit→8-bit VGA DAC palette scaling (`(v << 2) | (v >> 4)`), `StandardCGA()` fallback, `ICEColor` / `Chars512` attribute masking, `CP437Rune` control-range mapping, and HTML `<span>` merging (also used by [icedraw.go](file:///home/ben/github/server/internal/icedraw/icedraw.go)).
- **[magicnumber](file:///home/ben/github/magicnumber/text.go) (`magicnumber.XBinaryText` / `XBin`):** Matches the `"XBIN\x1a"` header.
- **[readme.go](file:///home/ben/github/server/handler/readme/readme.go):** Skips `Normalize()` for `magicnumber.XBinaryText` in `primary()` and routes `Buffers()` to `xbin.Render`, setting `Text.Sign` and `Text.CustomFont` (`xbin.HasFont`).
- **[dirs.go](file:///home/ben/github/server/handler/app/dirs.go) & [artifact.tmpl](file:///home/ben/github/server/view/app/artifact.tmpl):** Sets `contentBinaryLabel = "XB"` and `contentBinaryCustomFont` for the `<h6>` header, keeping 80-column `font-dos` / `font-large` zoom controls (`contentBinarySwappers = true`).
- **[filerecord.go](file:///home/ben/github/server/handler/internal/filerecord/filerecord.go):** Includes `magicnumber.XBinaryText` in `Entry.isText()` for archive inspection.

---

## 3. Background & External References

These external references were consulted for specification verification and context (no source code was copied):

1. **ACiD XBIN Specification (Olivier "Tasmaniac" Reubens, 1996) & ACiD SAUCE Specification (`acid.org`, [sauce/README.md](file:///home/ben/github/sauce/README.md))**
   - Referenced for the 11-byte XBIN header, flag bit assignments (`Palette`, `Font`, `Compress`, `NonBlink`, `512Chars`), RLE control-byte encoding (`00`/`01`/`10`/`11`), and confirmation that `SAUCE` `TFlags`/`TInfoS` are unused for `DataType = 6` (`XBin`).
2. **UK National Archives PRONOM / Format Notes, Archive Team (`fileformats.archiveteam.org`), & `bbcat.dev` XBIN Flag Summaries**
   - Referenced to cross-check header offsets, `fontHeight` range (`1`–`32`), 256/512 character font block sizing, and 6-bit (`0x00`–`0x3F`) RGB palette format.
3. **AnsiLove / `libansilove` (`ansilove.org`) & `textmode-corpus`**
   - Referenced for ecosystem conventions on `.XB` rendering (used by `server`'s preview PNG/WebP pipeline in [dirs.go](file:///home/ben/github/server/internal/command/dirs.go) to render embedded custom fonts).
