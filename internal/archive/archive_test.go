package archive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "dot", input: ".", want: ""},
		{name: "root", input: "/", want: ""},
		{name: "simple", input: "foo", want: "foo"},
		{name: "leading_slash", input: "/foo", want: "foo"},
		{name: "trailing_slash", input: "foo/", want: "foo"},
		{name: "both_slashes", input: "/foo/", want: "foo"},
		{name: "nested", input: "/foo/bar", want: "foo/bar"},
		{name: "nested_trailing", input: "/foo/bar/", want: "foo/bar"},
		{name: "double_slash", input: "//foo//bar//", want: "foo/bar"},
		{name: "dot_path", input: "./foo", want: "foo"},
		{name: "dot_dot_path", input: "../foo", want: "../foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := normalizePath(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSortDirsFirst(t *testing.T) {
	t.Parallel()

	entries := []*DirEntry{
		{Name: "zebra.txt", IsDir: false},
		{Name: "alpha", IsDir: true},
		{Name: "beta.txt", IsDir: false},
		{Name: "gamma", IsDir: true},
		{Name: "apple.txt", IsDir: false},
	}

	SortDirsFirst(entries)

	// Directories should come first, alphabetically
	assert.Equal(t, "alpha", entries[0].Name)
	assert.True(t, entries[0].IsDir)
	assert.Equal(t, "gamma", entries[1].Name)
	assert.True(t, entries[1].IsDir)

	// Then files, alphabetically
	assert.Equal(t, "apple.txt", entries[2].Name)
	assert.False(t, entries[2].IsDir)
	assert.Equal(t, "beta.txt", entries[3].Name)
	assert.False(t, entries[3].IsDir)
	assert.Equal(t, "zebra.txt", entries[4].Name)
	assert.False(t, entries[4].IsDir)
}

func TestSortDirsFirst_EmptySlice(t *testing.T) {
	t.Parallel()

	var entries []*DirEntry
	SortDirsFirst(entries)
	assert.Empty(t, entries)
}

func TestSortDirsFirst_OnlyDirs(t *testing.T) {
	t.Parallel()

	entries := []*DirEntry{
		{Name: "charlie", IsDir: true},
		{Name: "alpha", IsDir: true},
		{Name: "bravo", IsDir: true},
	}

	SortDirsFirst(entries)

	assert.Equal(t, "alpha", entries[0].Name)
	assert.Equal(t, "bravo", entries[1].Name)
	assert.Equal(t, "charlie", entries[2].Name)
}

func TestSortDirsFirst_OnlyFiles(t *testing.T) {
	t.Parallel()

	entries := []*DirEntry{
		{Name: "charlie.txt", IsDir: false},
		{Name: "alpha.txt", IsDir: false},
		{Name: "bravo.txt", IsDir: false},
	}

	SortDirsFirst(entries)

	assert.Equal(t, "alpha.txt", entries[0].Name)
	assert.Equal(t, "bravo.txt", entries[1].Name)
	assert.Equal(t, "charlie.txt", entries[2].Name)
}
