<p align="center">
  <img src="https://raw.githubusercontent.com/wasmdesk/brand/main/png/color/256/wasmdesk.png" alt="wasmdesk" width="88" height="88">
</p>

<h1 align="center">wasmaqua</h1>
<p align="center"><strong>A macOS-Aqua-themed window manager, in your browser, written in Ruby.</strong></p>

<p align="center">
  A sibling of <a href="https://github.com/wasmdesk/wasmbox">wasmbox</a> with the
  same external-client protocol but a different look: three traffic-light
  buttons on the LEFT (red close · yellow minimize · green maximize), a 28&nbsp;px
  gradient titlebar, centered title text — the canonical macOS window chrome.
</p>

<p align="center">
  <a href="https://github.com/wasmdesk"><img alt="part of wasmdesk" src="https://img.shields.io/badge/wasmdesk-the%20WASM%20desktop-1a7f37?style=flat-square"></a>
  <a href="https://github.com/go-embedded-ruby/ruby"><img alt="built on go-embedded-ruby" src="https://img.shields.io/badge/runs%20on-go--embedded--ruby-9B1C2E?style=flat-square"></a>
  <img alt="WebAssembly" src="https://img.shields.io/badge/WebAssembly-CGO%3D0-654FF0?style=flat-square&logo=webassembly&logoColor=white">
  <a href="LICENSE"><img alt="License: BSD-3-Clause" src="https://img.shields.io/badge/license-BSD--3--Clause-blue?style=flat-square"></a>
</p>

---

## ⚠️ Superseded by `wasmbox --frame=aqua`

**As of 2026-06-30**, the wasmaqua look-and-feel is bundled into
wasmbox itself as the `AquaFrame` window-decoration preset. One
binary, one codebase, two looks (and counting):

```text
# Openbox look (wasmbox default)
http://localhost:8080/

# Aqua look (what wasmaqua shipped)
http://localhost:8080/?frame=aqua

# Aqua chrome + WhiteSur palette ≈ macOS Big Sur
http://localhost:8080/?frame=aqua-whitesur-light
```

See [wasmbox/compositor/02_frame.rb](https://github.com/wasmdesk/wasmbox/blob/main/compositor/02_frame.rb)
for the Frame strategy + the 16-entry FrameRegistry (2 plain layouts
+ 14 layout×palette combos covering Adwaita / Juno / WhiteSur /
Solarized).

This repo is **frozen** but kept buildable for legacy users; new
decoration work goes into wasmbox. Migration is a one-flag change
on the URL — no client-side changes needed because the
external-client wire protocol was always identical between wasmbox
and wasmaqua.

---

## Why a second WM? (historical)

`wasmbox` ships an Openbox/Fluxbox-style decoration (single close-X on the
right, dark red titlebar). `wasmaqua` is the same compositor codebase
re-themed for users who prefer the macOS Aqua look:

| feature        | wasmbox                             | wasmaqua                                    |
|----------------|-------------------------------------|---------------------------------------------|
| titlebar       | 22&nbsp;px, flat `#9b1c2e` active   | 28&nbsp;px, gradient `#ECECEC` → `#D6D6D6`  |
| buttons        | close-X + minimize, on the **right**| 3 circles (red / yellow / green) on the **left** |
| title text     | left-aligned, white on dark         | centered, dark on light                     |
| frame border   | red active / grey inactive          | hairline grey + 1&nbsp;px faked drop-shadow |
| close glyph    | × stroke                            | filled red circle                           |
| minimize glyph | `_` bar                             | filled yellow circle                        |
| maximize       | n/a                                 | filled green circle (zoom-toggle)           |

The external-client wire protocol is **identical** — `hello`/`welcome`/
`commit`/`input`/`set_title`/`request_close`/`closed` — so the same
`wasmdesk/wasmdock`, `go-quake1/cmd/quake-wasmbox`, terminal, files and
hello-world clients run unmodified.

## Build + run

```sh
task build       # builds wasmaqua.wasm + cmd/serve into ./bin
task serve       # serves http://localhost:8081/ with COOP/COEP
task test        # native Ruby-half assertions (rbtest)
```

Then point a SAB-capable browser (Chrome ≥ 92, Firefox ≥ 79) at
<http://localhost:8081/>. The compositor boots in a dedicated Web Worker,
takes over the `<canvas>`, paints the Aqua desktop, and waits for spawn
requests via `wasmaquaSpawnExternal(url)` / `wasmaquaSpawnFromOCI(ref)`
(also exposed under the legacy `wasmboxSpawnExternal` names).

## Layout

```
wasmaqua/
├── main.go                  # //go:embed compositor.rb → wasm
├── compositor.rb            # Ruby WM (macOS Aqua Theme + 3 traffic lights)
├── index.html               # main-thread bootstrap + DOM-event relay
├── bridge.js                # main ↔ worker message protocol constants
├── compositor.worker.js     # the wasm + Ruby VM live here
├── ociapps-loader.js        # OCI registry → blob-URL loader (shared with wasmbox)
├── coi-serviceworker.js     # COOP/COEP shim for static hosts
├── cmd/
│   ├── serve/               # dev HTTP server (COOP/COEP + .wasm MIME)
│   └── rbtest/              # native pure-Ruby assertions for the WM half
└── Taskfile.yml
```

## License

BSD-3-Clause — see [LICENSE](LICENSE).
