package open

import "github.com/meigma/blob"

// ArchiveLoadedMsg is sent when the archive has been loaded successfully.
type ArchiveLoadedMsg struct {
	Index   *blob.IndexView
	Archive *blob.Archive
}

// ArchiveErrorMsg is sent when loading the archive fails.
type ArchiveErrorMsg struct {
	Err error
}

// FileContentMsg is sent when file content has been loaded.
type FileContentMsg struct {
	Path     string
	Content  []byte
	IsBinary bool
}

// FileErrorMsg is sent when loading a file fails.
type FileErrorMsg struct {
	Path string
	Err  error
}

// CopyCompleteMsg is sent when a file copy completes successfully.
type CopyCompleteMsg struct {
	SourcePath string
	DestPath   string
}

// CopyErrorMsg is sent when a file copy fails.
type CopyErrorMsg struct {
	SourcePath string
	DestPath   string
	Err        error
}
