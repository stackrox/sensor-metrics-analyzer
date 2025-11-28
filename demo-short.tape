# VHS Demo Script - Short Version (for README badge/preview)
# Run with: vhs demo-short.tape

Output demo-short.gif
Set FontSize 18
Set Width 1200
Set Height 700
Set Theme "Dracula"
Set Padding 15
Set Framerate 24
Set PlaybackSpeed 1.0
Set WindowBar Colorful
Set BorderRadius 8

# Launch TUI directly
Type "./bin/metrics-analyzer analyze --format tui --rules ./automated-rules ../metrics.txt"
Enter
Sleep 2s

# Quick navigation
Down Sleep 300ms
Down Sleep 300ms
Down Sleep 300ms

# View details
Enter
Sleep 2s

# Back and filter
Escape
Sleep 500ms
Type "2"
Sleep 1.5s

# Search
Type "/"
Type "heap"
Enter
Sleep 1.5s

# Help
Type "?"
Sleep 2s

# Quit
Escape
Sleep 300ms
Type "q"

