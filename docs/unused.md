# Unused

These funcs are not used in the codebase. They are kept here for reference.

> "internal/archive/archive.go"

```go
// Replace the filename file extension with the ext string.
// Leaving ext empty returns the filename without a file extension.
func Replace(ext, filename string) string {
	const sep = "."
	s := strings.Split(filename, sep)
	if ext == "" && len(s) == 1 {
		return filename
	}
	if ext == "" {
		return strings.Join(s[:len(s)-1], sep)
	}
	if len(s) == 1 {
		s = append(s, ".tmp")
	}
	s[len(s)-1] = strings.Join(strings.Split(ext, sep), "")
	return strings.Join(s, sep)
}
```

> "internal/archive/internal/internal.go"

```go
// ExitArj returns the exit status of the arj command error.
func ExitArj(err error) string {
	if err == nil {
		return ""
	}
	statuses := map[int]string{
		0:  "success",
		1:  "warning",
		2:  "fatal error",
		3:  "crc error (header, file or bad password)",
		4:  "arj-security error",
		5:  "disk full or write error",
		6:  "cannot open archive or file",
		7:  "user error, bad command line parameters",
		8:  "not enough memory",
		9:  "not an arj archive",
		10: "MS-DOS XMS memory error",
		11: "user control break",
		12: "too many chapters (over 250)",
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		if waitStatus, statusExists := exitError.Sys().(syscall.WaitStatus); statusExists {
			if status, exitStatus := statuses[waitStatus.ExitStatus()]; exitStatus {
				return status
			}
		}
	}
	return err.Error()
}

// ExitUnRar returns the exit status of the unrar command error.
func ExitUnRar(err error) string {
	if err == nil {
		return ""
	}
	statuses := map[int]string{
		0:   "success",
		1:   "success with warning",
		2:   "fatal error",
		3:   "invalid checksum, data damage",
		4:   "attempt to modify a locked archive",
		5:   "write error",
		6:   "file open error",
		7:   "wrong command line option",
		8:   "not enough memory",
		9:   "file create error",
		10:  "no files matching the specified mask and options were found",
		11:  "incorrect password",
		255: "user stopped the process with control-C",
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		if waitStatus, statusExists := exitError.Sys().(syscall.WaitStatus); statusExists {
			if status, exitStatus := statuses[waitStatus.ExitStatus()]; exitStatus {
				return status
			}
		}
	}
	return err.Error()
}

// ExitUnzip returns the exit status of the unzip command error.
func ExitUnzip(err error) string {
	if err == nil {
		return ""
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		if waitStatus, statusExists := exitError.Sys().(syscall.WaitStatus); statusExists {
			statuses := map[int]string{
				0:  "success",
				1:  "success with warning",
				2:  "generic error in the zipfile format",
				3:  "severe error in zipfile format",
				4:  "unable to allocate memory for buffers",
				5:  "unable to allocate memory or tty to read decryption password",
				6:  "unable to allocate memory during decompression to disk",
				7:  "unable to allocate memory during in-memory decompression",
				8:  "unused",
				9:  "the specified zip file was not found",
				10: "invalid command arguments",
				11: "no matching files were found",
				12: "possible zip-bomb detected, aborting",
				50: "the disk is full during extraction",
				51: "the end of the zip archive was encountered prematurely",
				80: "user stopped the process with control-C",
				81: "testing or extraction of one or more files failed due to " +
					"unsupported compression methods or unsupported decryption",
				82: "no files were found due to bad decryption password",
			}
			if status, exitStatus := statuses[waitStatus.ExitStatus()]; exitStatus {
				return status
			}
		}
	}
	return err.Error()
}
```

> "model/exists.go"

```go
// FileExists returns true if the file record exists in the database.
// This function will also return true for records that have been marked as deleted.
func FileExists(ctx context.Context, exec boil.ContextExecutor, id int64) (bool, error) {
	if exec == nil {
		return false, ErrDB
	}
	ok, err := models.Files(models.FileWhere.ID.EQ(id), qm.WithDeleted()).Exists(ctx, exec)
	if err != nil {
		return false, fmt.Errorf("models file exist %d: %w", id, err)
	}
	return ok, nil
}
```

> "model/update.go"

```go
// UpdateNoReadme updates the retrotxt_no_readme column value with val.
// It returns nil if the update was successful.
// Id is the database id of the record.
func UpdateNoReadme(db *sql.DB, id int64, val bool) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("update no readme: %w", err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("find file: %w", err)
	}
	i := int16(0)
	if val {
		i = 1
	}
	f.RetrotxtNoReadme = null.NewInt16(i, true)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("f.update: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("tx.commit: %w", err)
	}
	return nil
}

// art.RetrotxtNoReadme.Int16
```

> "internal/render/render.go"

```go
// UTF16 returns true if the byte slice is embedded with a UTF-16 BOM (byte order mark).
func UTF16(r io.Reader) bool {
	if r == nil {
		return false
	}
	const minimum = 2
	p := make([]byte, minimum)
	if _, err := io.ReadFull(r, p); err != nil {
		return false
	}
	if len(p) < minimum {
		return false
	}
	const y, thorn = 0xff, 0xfe
	littleEndian := p[0] == y && p[1] == thorn
	if littleEndian {
		return true
	}
	bigEndian := p[0] == thorn && p[1] == y
	return bigEndian
}

func TestUTF16(t *testing.T) {
	t.Parallel()
	assert.False(t, render.UTF16(nil))

	r := bytes.NewReader(nil)
	assert.False(t, render.UTF16(r))

	b := []byte{0xff, 0xfe, 0x00, 0x00, 0x00, 0x00}
	r = bytes.NewReader(b)
	assert.True(t, render.UTF16(r))

	b = []byte{0x00, 0x00, 0xfe, 0xff, 0x00, 0x00}
	r = bytes.NewReader(b)
	assert.False(t, render.UTF16(r))

	b = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	r = bytes.NewReader(b)
	assert.False(t, render.UTF16(r))

	s := "😀 some unicode text 😀"
	u := stringToUTF16(s)
	u = append([]uint16{0xFEFF}, u...)
	b = uint16ArrayToByteArray(u)
	r = bytes.NewReader(b)
	assert.True(t, render.UTF16(r))
}
```