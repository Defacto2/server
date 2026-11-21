# iCEDraw `.IDF` Handler (`internal/icedraw`)

[icedraw.go](file:///home/ben/github/server/internal/icedraw/icedraw.go) decodes **iCEDraw v1.4 (`.IDF`)** character-graphics files into styled HTML (`<div>...</div>` with `<span style="color:#rrggbb;background-color:#rrggbb;">...</span>` runs) for the in-browser readme viewer (`#precontentBinary`).

No third-party source code was copied or ported for this implementation. The decoder was written from direct binary inspection of the sample artifact, combined with existing Defacto2 repository conventions and the reference documentation listed below.

---

## 1. Primary Empirical Source (`ab229bf` — `LS-ICE.IDF`)

Inspection of local artifact `/f/ab229bf` (`LS-ICE.IDF`, `29,698` bytes, Insane Creators Enterprise, 1995) established the binary layout and several format details required for accurate browser rendering:

### File Layout

| Offset | Size | Field / Value | Notes |
| --- | --- | --- | --- |
| `0x00`–`0x03` | 4 bytes | `"\x041.4"` (`04 31 2E 34`) | iCEDraw v1.4 format signature |
| `0x04`–`0x05` | 2 bytes (`uint16` LE) | `x1` (`0`) | Left column bound (`0`-based) |
| `0x06`–`0x07` | 2 bytes (`uint16` LE) | `y1` (`0`) | Top row bound (`0`-based) |
| `0x08`–`0x09` | 2 bytes (`uint16` LE) | `x2` (`79`) | Right column bound (`width = x2 - x1 + 1 = 80`) |
| `0x0A`–`0x0B` | 2 bytes (`uint16` LE) | `y2` (`21`) | Viewport bottom row bound |
| `0x0C`..`EOF - 4144` | Variable | RLE-compressed `[char, attr]` stream | `25,542` bytes in `LS-ICE.IDF` |
| `EOF - 4144`..`EOF - 49` | 4,096 bytes | Custom 8×16 VGA bitmap font | 256 glyphs × 16 scanlines |
| `EOF - 48`..`EOF - 1` | 48 bytes | Custom 6-bit VGA DAC palette | 16 RGB triplets, each channel `0x00`–`0x3F` (`0`–`63`) |

*(If a file also carries an optional trailing `0x1A` + `COMNT` + `SAUCE00` record, `sauce.Trim` strips it before reading the trailing 4,096 + 48 bytes.)*

### RLE Stream Decoding

The compressed body is read as `uint16` little-endian words (`[char, attr]`):

- **Word `0x0001` (`01 00`):** Run-length escape, followed by `count` (`uint16` LE) and a 2-byte `[ch, attr]` cell repeated `count` times.
- **Any other word:** Literal 1-cell `[ch, attr]` pair (`ch = word & 0xFF`, `attr = word >> 8`).

### Key Empirical Findings from `LS-ICE.IDF`

1. **Artwork exceeds `y2`:** Although `y1..y2` in the header is `0..21`, the RLE stream decodes to **`16,000` cells** (`80 × 200` rows), containing **`182` non-blank rows** (`1..182`). Cropping to `y2 - y1 + 1` would discard ~88% of the artwork; trimming trailing blank rows (matching `thedraw.trimBlank`) preserves the full piece while dropping unused canvas rows.
2. **Non-standard VGA palette:** The trailing 48-byte palette in `LS-ICE.IDF` differs from the default IBM CGA 16-color palette (for example, color `3` is `(33, 19, 1)` brown rather than cyan, color `5` is `(40, 40, 50)`, and color `12` is `(56, 26, 26)`). Scaling each 6-bit DAC value (`0..63`) to 8-bit RGB via `(v << 2) | (v >> 4)` produces exact `#rrggbb` colors in CSS without `ansibump`'s fixed 16-color palette limitation.
3. **Why raw fallback failed before:** Running `readme.Normalize()` on `.IDF` bytes replaced `0x00` bytes with spaces (`0x20`), causing `magicnumber.Find` to misclassify the file as `PlainText` and render raw binary bytes (` 1.4 O K ▄ ▌ ...`) inside `#precontentCP437`.

---

## 2. Codebase & Rendering Conventions

- **[magicnumber](file:///home/ben/github/magicnumber/text.go) (`magicnumber.ICEDraw`):** Identifies `.IDF` files by `\x041.4`, `x1 <= x2 && y1 <= y2`, minimum `12 + 4096 + 48` payload length (after optional SAUCE/COMNT), and all 48 palette bytes `<= 0x3F`.
- **[thedraw.go](file:///home/ben/github/server/internal/thedraw/thedraw.go):** Established the internal decoder package pattern, trailing-blank-row trimming (`trimBlank`), and `handler/readme` integration before `Normalize()`.
- **[readme.go](file:///home/ben/github/server/handler/readme/readme.go):** Skips `Normalize()` for `magicnumber.ICEDraw` and routes `Buffers()` to `icedraw.Render` (returning HTML into `textBuf` with an empty `runeBuf` for `#precontentBinary`).
- **[dirs.go](file:///home/ben/github/server/handler/app/dirs.go):** Keeps `magicnumber.ICEDraw.Title()` on the 80-column `font-dos font-large render` path with zoom/16px/8px controls (`contentBinarySwappers = true`).
- **[filerecord.go](file:///home/ben/github/server/handler/internal/filerecord/filerecord.go):** Includes `magicnumber.ICEDraw` in `Entry.isText()` for archive inspection.
- **Browser HTML Cell Rules:**
  - `0x00` character bytes are mapped to `' '` (`0x20`) because browsers collapse `U+0000` to zero width in HTML.
  - CP437 control-range bytes (`0x01`–`0x1F`, `0x7F`) are mapped to printable Unicode equivalents (rather than raw ASCII control codes), and `0xFF` maps to `\u00A0`.
  - **Scope limitation:** `<pre id="precontentBinary">` renders using the web font (`Ac437_IBM_VGA_8x16` / `Ac437_IBM_EGA_8x8`), so the embedded 4,096-byte custom font is validated/skipped by `icedraw.Render` while custom glyphs remain visible in the server-side `ansilove` PNG/WebP preview image.

---

## 3. Background & External References

These external resources were consulted for format verification and ecosystem context (no source code was copied):

1. **Archive Team — _Just Solve the File Format Problem_ (`fileformats.archiveteam.org`)**
   - Referenced for iCEDraw (`.IDF`) format documentation, `\x041.4` version identifier, and legacy text-mode graphics preservation notes.
2. **AnsiLove / `libansilove` (`ansilove.org`, `github.com/ansilove/libansilove`, `src/idf.c`)**
   - Referenced to confirm `.IDF` loader conventions (separate `src/idf.c` loader handling header bounds, RLE stream, 4,096-byte 8×16 font, and 48-byte VGA palette) and how `server`'s preview pipeline already renders `.IDF` via `ansilove`.
3. **`file.org` & FFmpeg (`ffmpeg.org`) iCEDraw / `.IDF` Format Summaries**
   - Referenced for cross-checking the `\x041.4` signature, 4,096-byte font block (`256 × 16`), and 48-byte 6-bit VGA palette structure.
4. **ACiD SAUCE Specification (via [github.com/bengarrett/sauce](file:///home/ben/github/sauce/README.md))**
   - Referenced for optional `SAUCE00` / `COMNT` trailer stripping (`sauce.Trim`) and `DataType` distinctions between `BinaryText` (`5`, `.BIN`) and `Character` / `iCEDraw` (`1`/`5`).
