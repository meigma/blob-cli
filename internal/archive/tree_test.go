package archive

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTreePrinter_Print(t *testing.T) {
	t.Parallel()

	root := &DirEntry{
		Name:  ".",
		IsDir: true,
		Children: []*DirEntry{
			{
				Name:  "config",
				IsDir: true,
				Children: []*DirEntry{
					{Name: "app.yaml", IsDir: false},
					{Name: "db.yaml", IsDir: false},
				},
			},
			{Name: "README.md", IsDir: false},
		},
	}

	var buf bytes.Buffer
	printer := &TreePrinter{Writer: &buf}
	printer.Print(root)

	expected := `./
├── config/
│   ├── app.yaml
│   └── db.yaml
└── README.md
`
	assert.Equal(t, expected, buf.String())
}

func TestTreePrinter_Print_DirsFirst(t *testing.T) {
	t.Parallel()

	root := &DirEntry{
		Name:  ".",
		IsDir: true,
		Children: []*DirEntry{
			{Name: "README.md", IsDir: false},
			{
				Name:  "config",
				IsDir: true,
				Children: []*DirEntry{
					{Name: "app.yaml", IsDir: false},
				},
			},
			{Name: "Makefile", IsDir: false},
		},
	}

	var buf bytes.Buffer
	printer := &TreePrinter{Writer: &buf, DirsFirst: true}
	printer.Print(root)

	expected := `./
├── config/
│   └── app.yaml
├── Makefile
└── README.md
`
	assert.Equal(t, expected, buf.String())
}

func TestTreePrinter_Print_Empty(t *testing.T) {
	t.Parallel()

	root := &DirEntry{
		Name:  ".",
		IsDir: true,
	}

	var buf bytes.Buffer
	printer := &TreePrinter{Writer: &buf}
	printer.Print(root)

	expected := "./\n"
	assert.Equal(t, expected, buf.String())
}

func TestTreePrinter_Print_DeepNesting(t *testing.T) {
	t.Parallel()

	root := &DirEntry{
		Name:  ".",
		IsDir: true,
		Children: []*DirEntry{
			{
				Name:  "a",
				IsDir: true,
				Children: []*DirEntry{
					{
						Name:  "b",
						IsDir: true,
						Children: []*DirEntry{
							{
								Name:  "c",
								IsDir: true,
								Children: []*DirEntry{
									{Name: "file.txt", IsDir: false},
								},
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	printer := &TreePrinter{Writer: &buf}
	printer.Print(root)

	expected := `./
└── a/
    └── b/
        └── c/
            └── file.txt
`
	assert.Equal(t, expected, buf.String())
}

func TestCounts(t *testing.T) {
	t.Parallel()

	root := &DirEntry{
		Name:  ".",
		IsDir: true,
		Children: []*DirEntry{
			{
				Name:  "config",
				IsDir: true,
				Children: []*DirEntry{
					{Name: "app.yaml", IsDir: false},
					{Name: "db.yaml", IsDir: false},
				},
			},
			{
				Name:  "scripts",
				IsDir: true,
				Children: []*DirEntry{
					{Name: "setup.sh", IsDir: false},
				},
			},
			{Name: "README.md", IsDir: false},
		},
	}

	dirs, files := Counts(root)
	assert.Equal(t, 2, dirs)
	assert.Equal(t, 4, files)
}

func TestCounts_Empty(t *testing.T) {
	t.Parallel()

	root := &DirEntry{
		Name:  ".",
		IsDir: true,
	}

	dirs, files := Counts(root)
	assert.Equal(t, 0, dirs)
	assert.Equal(t, 0, files)
}

func TestCounts_OnlyFiles(t *testing.T) {
	t.Parallel()

	root := &DirEntry{
		Name:  ".",
		IsDir: true,
		Children: []*DirEntry{
			{Name: "a.txt", IsDir: false},
			{Name: "b.txt", IsDir: false},
			{Name: "c.txt", IsDir: false},
		},
	}

	dirs, files := Counts(root)
	assert.Equal(t, 0, dirs)
	assert.Equal(t, 3, files)
}

func TestCounts_OnlyDirs(t *testing.T) {
	t.Parallel()

	root := &DirEntry{
		Name:  ".",
		IsDir: true,
		Children: []*DirEntry{
			{Name: "a", IsDir: true},
			{Name: "b", IsDir: true},
		},
	}

	dirs, files := Counts(root)
	assert.Equal(t, 2, dirs)
	assert.Equal(t, 0, files)
}
