# Recording Demos

This project uses [VHS](https://github.com/charmbracelet/vhs) for recording terminal demos.
See the [`demo/`](../../demo/) folder for scripts and instructions.

```bash
# Install VHS
brew install vhs

# Record demos
cd demo
vhs demo.tape       # Full TUI demo
vhs demo-cli.tape   # CLI mode demo

# Publish to charm servers
vhs publish demo.gif
```

