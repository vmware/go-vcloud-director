package util

import (
	"testing"
)

func TestSanitizedName(t *testing.T) {
	tables := []struct {
		badPath  string
		goodPath string
	}{
		{"\\..\\1.txt", "1.txt"},
		{"///foo/bar", "foo/bar"},
		{"C:/loo/bar2", "loo/bar2"},
		{"C:\\loo\\bar2", "loo\\bar2"},
		{"../../foo../../ba..r", "foo../ba..r"},
	}
	for _, table := range tables {
		fixedPath := sanitizedName(table.badPath)
		if fixedPath != table.goodPath {
			t.Errorf("expected and fixedPath didn't match - %s : %s", table.goodPath, fixedPath)
		}
	}
}
