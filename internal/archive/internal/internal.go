// Package internal contains the internal functions for the archive package.
package internal

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// ArjItem returns true if the string is a row from the [arj program] list command.
//
// [arj program]: https://arj.sourceforge.net/
func ARJItem(s string) bool {
	const minLen = 6
	if len(s) < minLen {
		return false
	}
	if s[3:4] != ")" {
		return false
	}
	x := s[:3]
	if _, err := strconv.Atoi(x); err != nil {
		return false
	}
	return true
}

// MagicLHA returns true if the LHA file type is matched in the magic string.
func MagicLHA(magic string) bool {
	s := strings.Split(magic, " ")
	const lha, lharc = "lha", "lharc"
	if s[0] == lharc {
		return true
	}
	if s[0] != lha {
		return false
	}
	if len(s) < len(lha) {
		return false
	}
	if strings.Join(s[0:3], " ") == "lha archive data" {
		return true
	}
	if strings.Join(s[2:4], " ") == "archive data" {
		return true
	}
	return false
}

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
