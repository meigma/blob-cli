package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meigma/blob-cli/internal/archive"
)

func TestTreeCmd_NilConfig(t *testing.T) {
	viper.Reset()

	ctx := context.Background()

	treeCmd.SetContext(ctx)
	err := treeCmd.RunE(treeCmd, []string{"ghcr.io/test:v1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not loaded")
}

func TestTreeText(t *testing.T) {
	root := &archive.DirEntry{
		Name:  ".",
		IsDir: true,
		Children: []*archive.DirEntry{
			{
				Name:  "config",
				IsDir: true,
				Children: []*archive.DirEntry{
					{Name: "app.yaml", IsDir: false},
				},
			},
			{Name: "README.md", IsDir: false},
		},
	}
	flags := treeFlags{}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := treeText(root, flags)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)
	output := buf.String()

	// Check tree structure
	assert.Contains(t, output, "./")
	assert.Contains(t, output, "config/")
	assert.Contains(t, output, "app.yaml")
	assert.Contains(t, output, "README.md")

	// Check summary line
	assert.Contains(t, output, "1 directory")
	assert.Contains(t, output, "2 files")
}

func TestTreeText_DirsFirst(t *testing.T) {
	root := &archive.DirEntry{
		Name:  ".",
		IsDir: true,
		Children: []*archive.DirEntry{
			{Name: "README.md", IsDir: false},
			{Name: "config", IsDir: true},
			{Name: "Makefile", IsDir: false},
		},
	}
	flags := treeFlags{dirsFirst: true}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := treeText(root, flags)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)

	// config/ should appear before files
	configIdx := bytes.Index(buf.Bytes(), []byte("config/"))
	readmeIdx := bytes.Index(buf.Bytes(), []byte("README.md"))
	makefileIdx := bytes.Index(buf.Bytes(), []byte("Makefile"))

	assert.Less(t, configIdx, readmeIdx)
	assert.Less(t, configIdx, makefileIdx)
}

func TestTreeText_Empty(t *testing.T) {
	root := &archive.DirEntry{
		Name:  ".",
		IsDir: true,
	}
	flags := treeFlags{}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := treeText(root, flags)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)

	assert.Contains(t, buf.String(), "./")
	assert.Contains(t, buf.String(), "0 directories")
	assert.Contains(t, buf.String(), "0 files")
}

func TestTreeJSON(t *testing.T) {
	root := &archive.DirEntry{
		Name:  ".",
		Path:  "",
		IsDir: true,
		Children: []*archive.DirEntry{
			{
				Name:  "config",
				Path:  "config",
				IsDir: true,
				Children: []*archive.DirEntry{
					{Name: "app.yaml", Path: "config/app.yaml", IsDir: false},
				},
			},
			{Name: "README.md", Path: "README.md", IsDir: false},
		},
	}
	flags := treeFlags{}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := treeJSON("ghcr.io/test:v1", "/", root, flags)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)

	var got treeResult
	err = json.Unmarshal(buf.Bytes(), &got)
	require.NoError(t, err)

	assert.Equal(t, "ghcr.io/test:v1", got.Ref)
	assert.Equal(t, "/", got.Path)
	assert.Equal(t, 1, got.DirCount)
	assert.Equal(t, 2, got.FileCount)

	assert.NotNil(t, got.Root)
	assert.Equal(t, ".", got.Root.Name)
	assert.True(t, got.Root.IsDir)
	require.Len(t, got.Root.Children, 2)

	// Check config directory
	assert.Equal(t, "config", got.Root.Children[0].Name)
	assert.True(t, got.Root.Children[0].IsDir)
	require.Len(t, got.Root.Children[0].Children, 1)
	assert.Equal(t, "app.yaml", got.Root.Children[0].Children[0].Name)
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		n        int
		singular string
		plural   string
		want     string
	}{
		{0, "file", "files", "0 files"},
		{1, "file", "files", "1 file"},
		{2, "file", "files", "2 files"},
		{1, "directory", "directories", "1 directory"},
		{5, "directory", "directories", "5 directories"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := pluralize(tt.n, tt.singular, tt.plural)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertToTreeNode(t *testing.T) {
	entry := &archive.DirEntry{
		Name:  "root",
		Path:  "/",
		IsDir: true,
		Children: []*archive.DirEntry{
			{Name: "file.txt", Path: "/file.txt", IsDir: false},
			{Name: "dir", Path: "/dir", IsDir: true},
		},
	}

	node := convertToTreeNode(entry, false)

	assert.Equal(t, "root", node.Name)
	assert.Equal(t, "/", node.Path)
	assert.True(t, node.IsDir)
	require.Len(t, node.Children, 2)
	assert.Equal(t, "file.txt", node.Children[0].Name)
	assert.Equal(t, "dir", node.Children[1].Name)
}

func TestConvertToTreeNode_DirsFirst(t *testing.T) {
	entry := &archive.DirEntry{
		Name:  "root",
		Path:  "/",
		IsDir: true,
		Children: []*archive.DirEntry{
			{Name: "file.txt", Path: "/file.txt", IsDir: false},
			{Name: "dir", Path: "/dir", IsDir: true},
		},
	}

	node := convertToTreeNode(entry, true)

	require.Len(t, node.Children, 2)
	// With dirsFirst, directory should come first
	assert.Equal(t, "dir", node.Children[0].Name)
	assert.True(t, node.Children[0].IsDir)
	assert.Equal(t, "file.txt", node.Children[1].Name)
	assert.False(t, node.Children[1].IsDir)
}
