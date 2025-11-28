# Demo Recordings

This folder contains [VHS](https://github.com/charmbracelet/vhs) scripts for recording terminal demos.

## Scripts

| File | Description | Output |
|------|-------------|--------|
| `demo.tape` | Full TUI demo (~45s) | [View on Charm](https://vhs.charm.sh/vhs-33cWV6RkqrjaeabNAcxVmc.gif) |
| `demo-cli.tape` | CLI mode demo (~40s) | [View on Charm](https://vhs.charm.sh/vhs-5slvsgOGnyRd7JsWNA7vMu.gif) |
| `demo-short.tape` | Short TUI preview (~15s) | - |

## Recording

```bash
# Install VHS
brew install vhs

# Record a demo (from project root)
cd demo
vhs demo.tape

# Publish to Charm servers
vhs publish demo.gif
```

## Notes

- GIF files are gitignored (hosted on vhs.charm.sh instead)
- Edit `.tape` files to modify the demos
- See [VHS documentation](https://github.com/charmbracelet/vhs) for syntax

