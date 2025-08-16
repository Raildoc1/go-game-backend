package service

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestValidateAndCleanPath(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("eval symlinks root: %v", err)
	}
	root, err = filepath.Abs(root)
	if err != nil {
		t.Fatalf("abs root: %v", err)
	}

	// Create some files/dirs under root
	inRoot := filepath.Join(root, "config.yaml")
	if err := os.WriteFile(inRoot, []byte("ok"), 0o600); err != nil {
		t.Fatalf("write inRoot: %v", err)
	}

	subdir := filepath.Join(root, "conf.d")
	if err := os.Mkdir(subdir, 0o750); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}
	inSub := filepath.Join(subdir, "a.yaml")
	if err := os.WriteFile(inSub, []byte("ok"), 0o600); err != nil {
		t.Fatalf("write inSub: %v", err)
	}

	// Outside root
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "secrets.yaml")
	if err := os.WriteFile(outsideFile, []byte("nope"), 0o600); err != nil {
		t.Fatalf("write outsideFile: %v", err)
	}

	// Try to create symlinks; if not permitted, we’ll skip symlink cases later.
	var (
		linkInsideToInside string
		linkInsideToOut    string
		symlinkSupported   = true
	)
	linkInsideToInside = filepath.Join(root, "link-inside.yaml")
	if err := os.Symlink(inSub, linkInsideToInside); err != nil {
		// Windows without admin or some FSs: symlink not supported
		symlinkSupported = false
	}
	linkInsideToOut = filepath.Join(root, "link-outside.yaml")
	if symlinkSupported {
		if err := os.Symlink(outsideFile, linkInsideToOut); err != nil {
			symlinkSupported = false
		}
	}

	type tc struct {
		name            string
		path            string
		wantOK          bool
		skipIfNoSymlink bool
	}

	tests := []tc{
		{
			name:   "exact root allowed",
			path:   root,
			wantOK: true,
		},
		{
			name:   "file in root allowed",
			path:   inRoot,
			wantOK: true,
		},
		{
			name:   "file in subdir allowed",
			path:   inSub,
			wantOK: true,
		},
		{
			name: "cleaned traversal that stays inside allowed",
			// root/./conf.d/../config.yaml -> root/config.yaml
			path:   filepath.Join(root, ".", "conf.d", "..", "config.yaml"),
			wantOK: true,
		},
		{
			name: "traversal outside rejected",
			// ../../outsideDir/secrets.yaml should be rejected
			path:   filepath.Join(root, "..", "..", outsideDir, "secrets.yaml"),
			wantOK: false,
		},
		{
			name:            "symlink inside->inside allowed",
			path:            linkInsideToInside,
			wantOK:          true,
			skipIfNoSymlink: true,
		},
		{
			name:            "symlink inside->outside rejected",
			path:            linkInsideToOut,
			wantOK:          false,
			skipIfNoSymlink: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIfNoSymlink && !symlinkSupported {
				if runtime.GOOS == "windows" {
					t.Skip("symlinks not supported (Windows without dev mode/admin) — skipping symlink test")
				}
				t.Skip("symlinks not supported on this FS — skipping symlink test")
			}

			got, err := validateAndCleanPath(root, tt.path)
			if tt.wantOK {
				if err != nil {
					t.Fatalf("ValidatePath(%q, %q) error: %v", root, tt.path, err)
				}
				// Returned path should be absolute and inside root.
				if !isUnder(got, root) && got != root {
					t.Fatalf("validated path %q not under root %q", got, root)
				}
			} else {
				if err == nil {
					t.Fatalf("ValidatePath(%q, %q) expected error, got OK (%q)", root, tt.path, got)
				}
			}
		})
	}
}

// isUnder reports whether p is within root (prefix check with path separator).
func isUnder(p, root string) bool {
	p = filepath.Clean(p)
	root = filepath.Clean(root)
	if p == root {
		return true
	}
	sep := string(os.PathSeparator)
	return len(p) > len(root) && p[:len(root)] == root && p[len(root):len(root)+1] == sep
}
