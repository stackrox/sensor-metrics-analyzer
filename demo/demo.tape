# VHS Demo Script for Sensor Metrics Analyzer
# Run with: vhs demo.tape

# Output settings
Output demo.gif
Set FontSize 16
Set Width 1400
Set Height 900
Set Theme "Dracula"
Set Padding 20
Set Framerate 30
Set PlaybackSpeed 0.8

# Window chrome for professional look
Set WindowBar Colorful
Set WindowBarSize 40
Set BorderRadius 8

# Start in the project directory
Type "cd sensor-metrics-analyzer-go"
Enter
Sleep 500ms

# Show the help first
Type "# 🔍 Sensor Metrics Analyzer - TUI Demo"
Enter
Sleep 1s
Type "clear"
Enter
Sleep 300ms

# Launch TUI
Type "./bin/metrics-analyzer analyze --format tui --rules ./automated-rules ../metrics.txt"
Sleep 500ms
Enter
Sleep 3s

# Navigate down through results
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms

# View details of selected metric
Enter
Sleep 2.5s

# Navigate to next metric in detail view
Down
Sleep 1.5s

# Go back to list
Escape
Sleep 1s

# Filter by RED status (key 2)
Type "2"
Sleep 2s

# Filter by YELLOW status (key 3)
Type "3"
Sleep 2s

# Filter by GREEN status (key 4)
Type "4"
Sleep 2s

# Back to ALL (key 1)
Type "1"
Sleep 1.5s

# Use search feature
Type "/"
Sleep 500ms
Type "cpu"
Sleep 1s
Enter
Sleep 2s

# Clear search with Escape
Escape
Sleep 1s

# Search for another term
Type "/"
Sleep 300ms
Type "buffer"
Sleep 800ms
Enter
Sleep 2s

# View detail of search result
Enter
Sleep 2.5s

# Go back
Escape
Sleep 1s

# Clear filter
Escape
Sleep 1s

# Show help screen
Type "?"
Sleep 3s

# Exit help
Escape
Sleep 1s

# Final navigation showcase - fast scrolling
Down Down Down Down Down
Sleep 500ms
Up Up Up Up Up
Sleep 500ms

# Go to bottom
Type "G"
Sleep 1s

# Go to top
Type "g"
Sleep 1s

# Quit the application
Type "q"
Sleep 1s

# Show final message
Type "# ✨ That's the Sensor Metrics Analyzer TUI!"
Enter
Sleep 2s

