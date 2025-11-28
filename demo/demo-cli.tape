# VHS Demo Script - CLI Mode (Console + Markdown)
# Run with: vhs demo-cli.tape

Output demo-cli.gif
Set FontSize 15
Set Width 1400
Set Height 900
Set Theme "Dracula"
Set Padding 20
Set Framerate 24
Set PlaybackSpeed 0.7
Set WindowBar Colorful
Set BorderRadius 8

# Title
Type "# 📊 Sensor Metrics Analyzer - CLI Demo"
Enter
Sleep 1s
Type "clear"
Enter
Sleep 300ms

# Show help
Type "# First, let's see the available options"
Enter
Sleep 800ms
Type "./bin/metrics-analyzer analyze --help"
Enter
Sleep 4s

Type "clear"
Enter
Sleep 300ms

# Console output (default)
Type "# 🖥️  Console Output (default mode)"
Enter
Sleep 800ms
Type "./bin/metrics-analyzer analyze --rules ./automated-rules ../metrics.txt"
Enter
Sleep 5s

Type "clear"
Enter
Sleep 300ms

# Generate markdown report
Type "# 📝 Generate Markdown Report"
Enter
Sleep 800ms
Type "./bin/metrics-analyzer analyze --format markdown --rules ./automated-rules --output report.md ../metrics.txt"
Enter
Sleep 3s

# Show the generated file
Type "ls -la report.md"
Enter
Sleep 1.5s

# Display with bat
Type "# 👀 View the report with syntax highlighting"
Enter
Sleep 800ms
Type "bat report.md"
Enter
Sleep 6s

# Show just the summary section
Type "clear"
Enter
Sleep 300ms
Type "# 🔍 Quick peek at critical issues"
Enter
Sleep 800ms
Type "bat report.md --line-range 1:50"
Enter
Sleep 5s

# Validate rules
Type "clear"
Enter
Sleep 300ms
Type "# ✅ Validate rule files"
Enter
Sleep 800ms
Type "./bin/metrics-analyzer validate ./automated-rules"
Enter
Sleep 2s

# List rules
Type "# 📋 List all available rules"
Enter
Sleep 800ms
Type "./bin/metrics-analyzer list-rules ./automated-rules"
Enter
Sleep 4s

# Cleanup
Type "rm report.md"
Enter
Sleep 500ms

Type "# ✨ That's the CLI mode!"
Enter
Sleep 2s

