# TheDraw `.COM` Handler (`internal/thedraw`)

[thedraw.go](file:///home/ben/github/server/internal/dump/thedraw.go) identifies and decodes uncompressed **TheDraw `.COM` file Screen Saves** into CP437 ANSI-escaped text (`.ANS`) for preview image rendering (`ansilove` / `ansimate`) and in-browser ANSI display (`ansibump`).

No third-party source code was copied or ported for this implementation. The decoder was written from direct binary and 8086 stub inspection of sample files, combined with existing Defacto2 repository conventions and the background references documented below.

---

## 1. Primary Empirical Sources (Sample Binary Analysis)

Inspection of sample `.COM` files established that TheDraw produced at least two distinct uncompressed `.COM` screen-save layouts sharing the same initial header:

### Common Header (`0x00`–`0x3E`, 63 bytes)

| Offset        | Size                  | Value / Field                        | Meaning                                                               |
| ------------- | --------------------- | ------------------------------------ | --------------------------------------------------------------------- |
| `0x00`        | 3 bytes               | `EB 3D 90`                           | `JMP short +0x3D; NOP` jumping to entry point `CS:013Fh` (`0x3F`)     |
| `0x03`        | 1 byte                | CGA wait flag                        | Rewritten to `0` at runtime for monochrome mode `7`                   |
| `0x04`–`0x06` | 3 bytes               | Variant-specific                     | See Variant A vs. Variant B below                                     |
| `0x07`–`0x08` | 2 bytes (`uint16` LE) | Start VRAM offset                    | Initial destination offset (`0x0000`)                                 |
| `0x09`–`0x25` | 29 bytes              | `"TheDraw COM file Screen Save\x1A"` | Format signature + DOS `Ctrl+Z` EOF stop                              |
| `0x26`–`0x3E` | 25 bytes              | `"Unsupported Video Mode\r\n$"`      | `INT 21h / AH=09h` error string if video mode is not `2`, `3`, or `7` |

---

### Variant A — 176-byte (`0xB0`) Raw VRAM Copy Stub

Observed in local artifact `/f/ba442c` (`!README!.COM`, `3,856` bytes):

- **Offset `0x04`:** `rows` (`uint8`, `0x17` = `23` rows)
- **Offset `0x05`–`0x06`:** `cols` (`uint16` LE, `0x0050` = `80` columns)
- **Stub (`0x3F`–`0xAF`, 113 bytes):** Checks BIOS video mode (`INT 10h / AH=0Fh`), selects `B800h` (color) or `B000h` (mono), loads `LEA SI, [01B0h]` (`8D 36 B0 01` at `0x79`), and copies `rows × cols` raw `(char, attr)` pairs per row (`REP MOVSW`) into text VRAM starting at file offset `0xB0` (`176 + 23 × 80 × 2 = 3,856` bytes).

---

### Variant B — 240-byte (`0xF0`) RLE Decoder Stub

Observed in `!README!.COM` (`1,310` bytes) and `!YOW.COM` (`2,237` bytes):

- **Offset `0x04`–`0x05`:** `payloadLen` (`uint16` LE: `0x042E` = `1,070` bytes in `!README!.COM`, `240 + 1070 = 1310`; `0x07CD` = `1,997` bytes in `!YOW.COM`, `240 + 1997 = 2237`)
- **Offset `0x06`:** `0x02`
- **Stub (`0x3F`–`0xEF`, 177 bytes):** Reads compressed byte stream from `DS:01F0h` (file offset `0xF0`) using `[0107h]` as `line_start` and decoding control opcodes:

| Opcode byte   | Arguments             | Action                                                                    |
| ------------- | --------------------- | ------------------------------------------------------------------------- |
| `0x00`–`0x0F` | —                     | Set foreground color `0..15` (`attr = (attr & 0xF0) \| b`)                |
| `0x10`–`0x17` | —                     | Set background color `0..7` (`attr = (attr & 0x8F) \| ((b - 0x10) << 4)`) |
| `0x18`        | —                     | Newline (`line_start += 160; DI = line_start`)                            |
| `0x19`        | `n` (`1` byte)        | Repeat space (`0x20`) **`n + 1`** times (`INC CX` at `CS:01CBh`)          |
| `0x1A`        | `n`, `ch` (`2` bytes) | Repeat CP437 byte `ch` **`n + 1`** times (`INC CX` at `CS:01CBh`)         |
| `0x1B`        | —                     | Toggle blink/high-background bit (`attr ^= 0x80`)                         |
| `>= 0x1C`     | —                     | Emit 1 literal character + current attribute                              |

---

### Related Samples Inspected (Scope Boundaries)

- **`$KURUPT.COM` (`2,604` bytes) & `$YOUWIN$.COM` (`1,747` bytes):** Both are wrapped with **PKLITE v1.12 (`0C 01`)** (`"PKLITE Copr. 1991 PKWARE Inc. All Rights Reserved"` at `0x2E`). `$YOUWIN$.COM` visibly exposes `"TheDraw COM file Screen ..."` and trailing RLE sequence `19 4F 18` at the start/end of its compressed payload. Unpacking PKLITE executables is outside `thedraw.go`'s scope.
- **`2FAST4YA.COM` (`5,881` bytes):** Multi-stage encrypted/anti-debug MS-DOS `MZ` executable renamed to `.COM` (uses IRQ masking on port `0x21`, a word decryptor over `0x13A5`–`0x16F9`, and `INT 1`/`INT 3` IVT decryption); unrelated to TheDraw.
- **`SPYHUNT.TPG` (`4,096` bytes):** Raw 80×25 CGA text-mode hardware page dump (`4,000` bytes visible VRAM + `96` bytes of `20 07` padding), identified by `magicnumber.CGAVideoDump` and rendered via `BinTextImager` / `binbump` rather than `thedraw.go`.

---

## 2. Codebase & IBM PC Conventions

- **[dump.go](file:///home/ben/github/server/internal/dump/dump.go):** Established the `internal/dump` package conventions and the `65,536`-byte DOS `.COM` size ceiling (`comMaxBytes`).
- **`magicnumber` (`magicnumber.TheDrawCOM`):** Provides upstream format identification (`"TheDraw COM file Screen Save"` at offset `9`) in `magicnumber.Find`.
- **IBM PC CGA/VGA Text-Mode Attributes → ANSI SGR:**
  - Bits `0..2` encode foreground in **BGR** bit order and bits `4..6` encode background in **BGR** bit order, mapped via `[8]int{0, 4, 2, 6, 1, 5, 3, 7}` to ANSI **RGB** (`30–37` foreground, `40–47` background).
  - Bit `3` (`0x08`) maps to SGR `1` (bold/high-intensity foreground).
  - Bit `7` (`0x80`) maps to SGR `5` (blink / iCE bright background).
  - Lines reaching column `80` omit trailing `\r\n` because 80-column DOS/ANSI renderers (`ansilove`, `ansimate`, `ansibump`) wrap onto the next line automatically; adding `\r\n` after column 80 would insert spurious blank lines.

---

## 3. Background Search References

These external references were consulted for format history, header verification, and variant context (no source code from these projects was read or copied):

1. **`file.org` (TheDraw COM Screen Save Signature Entries)**
   - Referenced for external verification of the offset-`0` jump prefix (`EB 3D 90`) and embedded offset-`9` ASCII marker (`"TheDraw COM file Screen Save\x1A"`).
   - _Note:_ Summary descriptions on general extension directories (`file.org`) describe TheDraw `.COM` saves broadly as raw text-mode screen buffers appended to a loader stub; direct sample disassembly showed that both raw `(char, attr)` buffers (`0xB0` stub) and custom RLE streams (`0xF0` stub) exist.

2. **TrID / Marco Pontello's File Identifier (`mark0.net`, `TrIDDefs.TRD`)**
   - Referenced for sub-variant taxonomy: TrID tracks multiple signature definitions for TheDraw `.COM` screen saves (e.g., generic, `type 1`, and `type 2`) and documents alternate stub jump offsets such as `EB 3D` and `EB 3F` across releases.

3. **Deark (`github.com/jsummers/deark`, Jason Summers)**
   - Consulted `formats.txt` reference summaries noting the `thedraw_com` module's scope: executable character-graphics screen saves generated by TheDraw as well as similar third-party _P-Screen_ `.COM` saves. _(Only the format catalog description was checked; Deark's C source implementation was not read.)_

4. **TheDraw Historical & Release Documentation (`shikadi.net` ModdingWiki, `winworldpc.com`, Wikipedia)**
   - Referenced for authorship and release lineage: **TheDraw** by Ian E. Davis (TheSoft Programming Services), spanning `v3.x` through the final official release **`v4.63`** (October 1993).
   - Confirmed native TheDraw export targets: `.ANS` (ANSI), `.BIN` (raw screen memory), `.BSV` (BASIC `BSAVE`), `.PCB` (PCBoard), `.BBS` (Wildcat!), `.COM` (self-displaying screen saves), and source-code exports (`.H`, `.PAS`, `.ASM`) — never native `.EXE` files (any TheDraw-like `.EXE` comes from wrapping a `.COM` or compiling exported source data).

5. **PKLITE Format References (for packed `.COM` context: Jason Summers' _Notes on PKLITE format_ & `shikadi.net` ModdingWiki)**
   - Consulted when inspecting `$KURUPT.COM` and `$YOUWIN$.COM` to verify PKLITE's non-ZIP custom LZSS + fixed lookup-table algorithm, version history (`v1.00`–`v2.01`), Small vs. Large compression modes, and PKLITE Professional `-e` stub scrambling (`v1.14+`).
