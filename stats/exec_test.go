package stats_test

import (
	"bufio"
	"os"
	"testing"

	"github.com/lesomnus/tegra-exporter/stats"
)

func TestExecuteIn(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("chroot needs root")
	}

	// `pwd` shows the working directory is moved into the root.
	open := stats.ExecuteIn("/", "/bin/sh", "-c", "pwd")
	r, err := open(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	sc := bufio.NewScanner(r)
	if !sc.Scan() {
		t.Fatal("no output")
	}
	if sc.Text() != "/" {
		t.Fatalf("unexpected working directory: %q", sc.Text())
	}
	// Close cancels before waiting, so its error is not about the command.
	r.Close()
}

func TestValidateRoot(t *testing.T) {
	file := t.TempDir() + "/file"
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		desc  string
		root  string
		name  string
		valid bool
	}{
		{"valid", t.TempDir(), "/usr/bin/tegrastats", true},
		{"relative root", "hostfs", "/usr/bin/tegrastats", false},
		{"relative command", t.TempDir(), "tegrastats", false},
		{"missing root", t.TempDir() + "/missing", "/usr/bin/tegrastats", false},
		{"root is a file", file, "/usr/bin/tegrastats", false},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			err := stats.ValidateRoot(tc.root, tc.name)
			if tc.valid && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
