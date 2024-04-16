package pulumitest

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

func tempDirWithoutCleanupOnFailedTest(t PT, desc string) string {
	t.TempDir()
	c := getOrCreateTempDirState(t)

	// Use a single parent directory for all the temporary directories
	// created by a test, each numbered sequentially.
	c.tempDirMu.Lock()
	var nonExistent bool
	if c.tempDir == "" { // Usually the case with js/wasm
		nonExistent = true
	} else {
		_, err := os.Stat(c.tempDir)
		nonExistent = os.IsNotExist(err)
		if err != nil && !nonExistent {
			ptFatalF(t, "TempDir: %v", err)
		}
	}

	if nonExistent {
		t.Helper()

		// Drop unusual characters (such as path separators or
		// characters interacting with globs) from the directory name to
		// avoid surprising os.MkdirTemp behavior.
		mapper := func(r rune) rune {
			if r < utf8.RuneSelf {
				const allowed = "!#$%&()+,-.=@^_{}~ "
				if '0' <= r && r <= '9' ||
					'a' <= r && r <= 'z' ||
					'A' <= r && r <= 'Z' {
					return r
				}
				if strings.ContainsRune(allowed, r) {
					return r
				}
			} else if unicode.IsLetter(r) || unicode.IsNumber(r) {
				return r
			}
			return -1
		}
		pattern := strings.Map(mapper, t.Name())
		c.tempDir, c.tempDirErr = os.MkdirTemp("", pattern)
		if c.tempDirErr == nil {
			t.Cleanup(func() {
				if ptFailed(t) && !runningInCI() {
					ptErrorF(t, "TempDir leaving %s to help debugging: %q", desc, c.tempDir)
				} else if err := os.RemoveAll(c.tempDir); err != nil {
					ptErrorF(t, "TempDir RemoveAll cleanup: %v", err)
				}
			})
		}
	}

	if c.tempDirErr == nil {
		c.tempDirSeq++
	}
	seq := c.tempDirSeq
	c.tempDirMu.Unlock()

	if c.tempDirErr != nil {
		ptFatalF(t, "TempDir: %v", c.tempDirErr)
	}

	dir := fmt.Sprintf("%s%c%03d", c.tempDir, os.PathSeparator, seq)
	if err := os.Mkdir(dir, 0777); err != nil {
		ptFatalF(t, "TempDir: %v", err)
	}
	return dir
}

type tempDirState struct {
	tempDir    string
	tempDirMu  sync.Mutex
	tempDirSeq int
	tempDirErr error
}

var tempDirStates sync.Map

func getOrCreateTempDirState(pointer any) *tempDirState {
	fresh := &tempDirState{}
	st, _ := tempDirStates.LoadOrStore(pointer, fresh)
	return st.(*tempDirState)
}
