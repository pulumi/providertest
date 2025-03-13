package pulumitest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

func tempDirWithoutCleanupOnFailedTest(t PT, desc, tempDir string) string {
	if tempDir != "" { // If a tempDir is provided, create on first test and don't worry about cleanup.
		if !filepath.IsAbs(tempDir) {
			absTempDir, err := filepath.Abs(tempDir)
			if err != nil {
				ptFatalF(t, "TempDir: %v", err)
			}
			tempDir = absTempDir
		}
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			if err := os.Mkdir(tempDir, 0755); err != nil {
				ptFatalF(t, "TempDir: %v", err)
			}
		}
	}
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
		c.tempDir, c.tempDirErr = os.MkdirTemp(tempDir, pattern)
		if c.tempDirErr == nil {
			t.Cleanup(func() {
				t.Helper()
				if ptFailed(t) && shouldRetainFilesOnFailure() {
					ptLogF(t, "Skipping removal of %s temp directories on failures: %q", desc, c.tempDir)
					t.Log("To remove these directories on failures, set PULUMITEST_RETAIN_FILES_ON_FAILURE=false")
					return
				}
				err := os.RemoveAll(c.tempDir)
				t.Log("Removed temp directories. To retain these, set PULUMITEST_RETAIN_FILES_ON_FAILURE=true")
				if err != nil {
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
	if err := os.Mkdir(dir, 0755); err != nil {
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
