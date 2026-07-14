# vscode/media/ — static assets for the VS Code extension

> 🌐 English | [繁體中文](README.zh-TW.md)

Static files referenced by [`../package.json`](../package.json)'s
`contributes` section. One file today:

- `auspex.svg` — the icon for the Auspex activity-bar container and
  its Status view (`viewsContainers.activitybar[].icon` and
  `views.auspex[].icon` both point at `media/auspex.svg`). Drawn
  stroke-only with `currentColor` so it follows the editor theme; the
  motif is a bird-augur's sighting arc over a horizon (see the SVG's
  own comment).

Paths in `package.json` are relative to the extension root, so
renaming or moving files here requires updating that manifest in the
same change. Extension behavior and development workflow are
documented in [`../README.md`](../README.md).
