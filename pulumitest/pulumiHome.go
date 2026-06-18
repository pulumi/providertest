package pulumitest

import (
	"os"
	"path/filepath"
)

// isolatedPulumiHome returns a per-test PULUMI_HOME directory.
//
// Tests in a package run in parallel against the same PULUMI_HOME, so they share
// Pulumi's on-disk schema cache (PULUMI_HOME/schemas). That cache is keyed by the
// requested provider version, so an operation that resolves a provider to a
// different version than the program declares (e.g. provider-upgrade testing) can
// persist a schema under a mismatched key. A later test that reads it then fails
// with "could not find package version information". Giving each test its own
// PULUMI_HOME/schemas keeps that cache private so one test cannot poison another.
// See pulumi/home#4816.
//
// The plugins directory is symlinked from the ambient PULUMI_HOME so plugin
// binaries are not re-downloaded per test; only the schema cache is per-test.
// (Pulumi's plugin discovery skips symlinked plugin *entries*, so we symlink the
// whole plugins directory, which it follows transparently, rather than each
// plugin.)
func isolatedPulumiHome(t PT) string {
	t.Helper()
	home := filepath.Join(t.TempDir(), "pulumi-home")
	if err := os.MkdirAll(home, 0o700); err != nil {
		ptFatalF(t, "failed to create isolated PULUMI_HOME: %v", err)
	}
	// Share plugin binaries and login/config from the ambient home so plugins
	// aren't re-downloaded and ambient-backend consumers keep working. The
	// schemas directory is deliberately NOT shared: it is the per-test cache that
	// keeps schema entries from poisoning each other.
	if ambient := ambientPulumiHome(); ambient != "" {
		for _, entry := range []string{"plugins", "credentials.json", "config.json"} {
			src := filepath.Join(ambient, entry)
			if _, err := os.Lstat(src); err != nil {
				continue // not present in the ambient home; nothing to share
			}
			if err := os.Symlink(src, filepath.Join(home, entry)); err != nil {
				ptFatalF(t, "failed to share %q into isolated PULUMI_HOME: %v", entry, err)
			}
		}
	}
	return home
}

// withPulumiHome returns an environment for exec.Cmd that sets PULUMI_HOME to the
// given directory. If home is empty it returns nil so the command inherits the
// ambient environment unchanged.
func withPulumiHome(home string) []string {
	if home == "" {
		return nil
	}
	return append(os.Environ(), "PULUMI_HOME="+home)
}

// ambientPulumiHome resolves the PULUMI_HOME that would be used without
// isolation, so its plugin binaries can be shared.
func ambientPulumiHome() string {
	if h := os.Getenv("PULUMI_HOME"); h != "" {
		return h
	}
	if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, ".pulumi")
	}
	return ""
}
