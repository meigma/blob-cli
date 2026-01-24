# blob-cli Demo
# Showcases browsing and extracting files from remote OCI archives
#
# Prerequisites: Build the binary first with `go build -o blob .`
# Run with: vhs demo/blob-demo.tape

Output demo/demo.gif

Set Shell bash
Set FontSize 16
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms
Set Padding 20
Set CursorBlink false
Set WindowBar Colorful
Set BorderRadius 8

# Set up the reference variable (visible, shows the image we're using)
Type "# Browse a remote archive without downloading it"
Sleep 500ms
Enter
Sleep 1s

Type "REF=ghcr.io/meigma/blob-cli/src:latest"
Sleep 300ms
Enter
Sleep 800ms

# Tree view
Type "./blob tree $REF"
Sleep 300ms
Enter
Wait
Sleep 2.5s

# Clear and show ls
Type "clear"
Enter
Sleep 300ms

Type "# List files with sizes"
Sleep 300ms
Enter
Sleep 500ms

Type "./blob ls -lh $REF"
Sleep 300ms
Enter
Wait
Sleep 2.5s

# Clear and show cat - the magic moment
Type "clear"
Enter
Sleep 300ms

Type "# View a file directly — no full download!"
Sleep 300ms
Enter
Sleep 500ms

Type@30ms "./blob cat $REF main.go"
Sleep 300ms
Enter
Wait
Sleep 3s

# Clear and show cp
Type "clear"
Enter
Sleep 300ms

Type "# Extract just what you need"
Sleep 300ms
Enter
Sleep 500ms

Type "./blob cp $REF:/go.mod ."
Sleep 300ms
Enter
Wait
Sleep 800ms

Type "head -5 go.mod"
Sleep 300ms
Enter
Sleep 2s

# Clear and show TUI
Type "clear"
Enter
Sleep 300ms

Type "# Or browse interactively"
Sleep 300ms
Enter
Sleep 500ms

Type "./blob open $REF"
Sleep 300ms
Enter
Sleep 2.5s

# Navigate in TUI
Down
Sleep 350ms
Down
Sleep 350ms
Down
Sleep 350ms
Enter
Sleep 2s

# Exit TUI
Type "q"
Sleep 1.5s
