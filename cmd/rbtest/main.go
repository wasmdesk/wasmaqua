// SPDX-License-Identifier: BSD-3-Clause
//
// Command rbtest runs the pure-Ruby half of wasmaqua/compositor.rb (Theme,
// Window, ExternalWindow, WindowManager) on the native go-embedded-ruby
// interpreter and asserts the macOS-Aqua decoration model.
//
// The JS-touching Compositor class plus the boot block at the bottom of
// compositor.rb are deliberately skipped: the test loads only the bytes BEFORE
// `class Compositor`, then appends a Ruby assertion script. This way the same
// file that ships inside wasmaqua.wasm is the file under test — no shadow copy
// to drift from.
//
// Exit code is 0 on success, 1 on any failed assertion (Ruby `raise`s and the
// Go wrapper surfaces the error). `task test` invokes it.
//
//go:build !js
// +build !js

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	ruby "github.com/go-embedded-ruby/ruby"
)

// splitMarker is the first line of the Compositor class definition. Everything
// before it is the pure WM half (safe off-wasm); everything from it onward
// touches the JS bridge.
const splitMarker = "class Compositor"

func main() {
	path := flag.String("rb", "compositor.rb", "path to compositor.rb (relative to cwd)")
	raw := flag.Bool("raw", false, "run -rb file as-is without splitting on `class Compositor` or appending the WM test script (debug aid)")
	flag.Parse()
	src, err := os.ReadFile(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rbtest: read %s: %v\n", *path, err)
		os.Exit(2)
	}
	compositorRB := string(src)
	var script string
	if *raw {
		script = compositorRB
	} else {
		idx := strings.Index(compositorRB, splitMarker)
		if idx < 0 {
			fmt.Fprintln(os.Stderr, "rbtest: cannot locate `class Compositor` in compositor.rb")
			os.Exit(2)
		}
		pure := compositorRB[:idx]
		script = pure + "\n" + testScript
	}
	var out bytes.Buffer
	if err := ruby.Run(script, &out); err != nil {
		fmt.Fprintln(os.Stderr, out.String())
		fmt.Fprintf(os.Stderr, "rbtest FAIL: %v\n", err)
		os.Exit(1)
	}
	os.Stdout.Write(out.Bytes())
	fmt.Println("rbtest: PASS")
}

// testScript exercises the macOS-Aqua decoration model. Each `assert` raises
// on failure so ruby.Run returns a non-nil error and rbtest exits non-zero.
const testScript = `
def assert(cond, msg)
  raise "ASSERT FAILED: #{msg}" unless cond
end

def assert_eq(actual, expected, msg)
  unless actual == expected
    raise "ASSERT_EQ FAILED (#{msg}): expected #{expected.inspect}, got #{actual.inspect}"
  end
end

# ---- Theme palette + sizing -----------------------------------------------
assert_eq(Theme::TITLE_H,       28,        "titlebar height = 28")
assert_eq(Theme::BTN_R,         6,         "traffic-light radius = 6")
assert_eq(Theme::BTN_DIAM,      12,        "traffic-light diameter = 12")
assert_eq(Theme::BTN_GAP,       8,         "button gap = 8")
assert_eq(Theme::CLOSE_RED,     "#FF5F57", "red close colour")
assert_eq(Theme::MIN_YELLOW,    "#FEBC2E", "yellow minimize colour")
assert_eq(Theme::MAX_GREEN,     "#28C840", "green maximize colour")
assert_eq(Theme::CLOSE_RED_OUT, "#E0443E", "red outline")
assert_eq(Theme::TITLE_ACTIVE,  "#ECECEC", "active titlebar gradient top")
assert_eq(Theme::TITLE_ACTIVE_2,"#D6D6D6", "active titlebar gradient bottom")
assert_eq(Theme::TITLE_BORDER,  "#BFBFBF", "titlebar bottom-edge hairline")
assert_eq(Theme::TITLE_TEXT_ON, "#3C3C43", "active title text")
assert_eq(Theme::TITLE_TEXT_OFF,"#9B9B9F", "inactive title text")

# ---- Three button rects, anchored to the LEFT side ------------------------
w = Window.new(7, "demo", 100, 200, 400, 300, "#ffffff")
cr = w.close_rect
mr = w.minimize_rect
xr = w.maximize_rect

# All three are BTN_DIAM x BTN_DIAM, vertically centered in the titlebar.
[cr, mr, xr].each_with_index do |r, i|
  assert_eq(r[2], Theme::BTN_DIAM, "button[#{i}] width = BTN_DIAM")
  assert_eq(r[3], Theme::BTN_DIAM, "button[#{i}] height = BTN_DIAM")
end
# All three share the same y (same row).
assert_eq(cr[1], mr[1], "close y == minimize y (same row)")
assert_eq(mr[1], xr[1], "minimize y == maximize y (same row)")

# Vertical centering: y = frame_top + (TITLE_H - BTN_DIAM) / 2.
expected_y = w.frame_top + (Theme::TITLE_H - Theme::BTN_DIAM) / 2
assert_eq(cr[1], expected_y, "buttons vertically centered in titlebar")

# Left-anchored: close starts at x + BTN_GAP, then min, then max — each
# BTN_GAP apart. That puts close FAR LEFT (the macOS layout), not on the
# right like wasmbox/Fluxbox.
assert_eq(cr[0], w.x + Theme::BTN_GAP,
          "close at x + BTN_GAP from the left edge")
assert_eq(mr[0], cr[0] + Theme::BTN_DIAM + Theme::BTN_GAP,
          "minimize sits BTN_GAP right of close")
assert_eq(xr[0], mr[0] + Theme::BTN_DIAM + Theme::BTN_GAP,
          "maximize sits BTN_GAP right of minimize")

# And they're nowhere near the right edge (sanity: macOS layout, not Fluxbox).
assert(cr[0] < w.right - Theme::BTN_DIAM * 3,
       "close button on the LEFT half of the titlebar")

# ---- Hit-tests on the three buttons ---------------------------------------
def center(rect)
  [rect[0] + rect[2]/2, rect[1] + rect[3]/2]
end
ccx, ccy = center(cr)
mcx, mcy = center(mr)
xcx, xcy = center(xr)
assert(w.on_close?(ccx, ccy),    "on_close? hits center of red button")
assert(w.on_minimize?(mcx, mcy), "on_minimize? hits center of yellow button")
assert(w.on_maximize?(xcx, xcy), "on_maximize? hits center of green button")
# Each button rejects the other two centers.
assert(!w.on_close?(mcx, mcy),    "on_close? misses yellow center")
assert(!w.on_close?(xcx, xcy),    "on_close? misses green center")
assert(!w.on_minimize?(ccx, ccy), "on_minimize? misses red center")
assert(!w.on_minimize?(xcx, xcy), "on_minimize? misses green center")
assert(!w.on_maximize?(ccx, ccy), "on_maximize? misses red center")
assert(!w.on_maximize?(mcx, mcy), "on_maximize? misses yellow center")
# A click on the FAR RIGHT side of the titlebar (where wasmbox put close)
# must NOT trigger close in wasmaqua: the right side is empty in this theme.
right_x = w.right - 8
mid_y   = w.frame_top + Theme::TITLE_H / 2
assert(!w.on_close?(right_x, mid_y),    "right-side click does NOT close (close is on the LEFT)")
assert(!w.on_minimize?(right_x, mid_y), "right-side click does NOT minimize")
assert(!w.on_maximize?(right_x, mid_y), "right-side click does NOT maximize")

# ---- Panels + popups have no buttons --------------------------------------
panel = Window.new(8, "dock", 0, 700, 800, 28, "#000000", "panel")
assert_eq(panel.close_rect,    [panel.x, panel.y, 0, 0], "panel close_rect collapses")
assert_eq(panel.minimize_rect, [panel.x, panel.y, 0, 0], "panel minimize_rect collapses")
assert_eq(panel.maximize_rect, [panel.x, panel.y, 0, 0], "panel maximize_rect collapses")
assert(!panel.on_close?(panel.x + 10, panel.y + 10),    "panel never reports on_close?")
assert(!panel.on_minimize?(panel.x + 10, panel.y + 10), "panel never reports on_minimize?")
assert(!panel.on_maximize?(panel.x + 10, panel.y + 10), "panel never reports on_maximize?")

popup = Window.new(9, "menu", 50, 50, 120, 80, "#ffffff", "popup")
assert_eq(popup.maximize_rect, [popup.x, popup.y, 0, 0], "popup maximize_rect collapses")
assert(!popup.on_maximize?(popup.x + 2, popup.y + 2), "popup never reports on_maximize?")

# ---- close click really removes the window from the WM --------------------
wm = WindowManager.new
target = wm.spawn("close-me", 200, 150)
other  = wm.spawn("survivor", 200, 150)
assert_eq(wm.windows.length, 2, "two windows before close")
# Simulate what Compositor#on_mousedown does on a close hit-test:
wm.close(target)
ids = wm.windows.map(&:id)
assert(!ids.include?(target.id), "closed window removed from WM")
assert(ids.include?(other.id),   "other window survives")
assert_eq(wm.windows.length, 1,  "one window after close")

# ---- zoom_toggle (green maximize) round-trips -----------------------------
zw = Window.new(20, "zoom", 100, 200, 400, 300, "#ffffff")
orig = [zw.x, zw.y, zw.w, zw.h]
zw.zoom_toggle(1024, 768)
assert_eq(zw.x, 0,                       "zoomed x = 0")
assert_eq(zw.y, Theme::TITLE_H,          "zoomed y = TITLE_H (titlebar still visible)")
assert_eq(zw.w, 1024,                    "zoomed w = canvas_w")
assert_eq(zw.h, 768 - Theme::TITLE_H,    "zoomed h = canvas_h - TITLE_H")
zw.zoom_toggle(1024, 768)
assert_eq([zw.x, zw.y, zw.w, zw.h], orig, "second zoom restores original rect")
# A panel zoom is a no-op (no decoration).
pn = Window.new(21, "panel", 0, 700, 800, 28, "#000000", "panel")
before = [pn.x, pn.y, pn.w, pn.h]
pn.zoom_toggle(1024, 768)
assert_eq([pn.x, pn.y, pn.w, pn.h], before, "panel zoom_toggle is a no-op")

# (LAYOUT_KEY namespace check is done in the file diff — the constant lives
#  inside class Compositor, which the rbtest split skips.)

# ---- AquaChrome.decoration_spec (widgets paint spec, pure) ----------------
# The spec re-expresses the SAME colours + geometry the hand-drawn
# draw_window path uses, in FRAME-LOCAL coordinates. It feeds Widgets.decoration
# (the JS-touching render lives in class Compositor, skipped here) so only the
# pure Hash builder is asserted.
dw = Window.new(30, "xterm", 60, 60, 240, 180, "#ffffff")
# frame_local: surface rect -> frame-space (origin at win.x, frame_top).
assert_eq(AquaChrome.frame_local(dw, [dw.x + 8, dw.frame_top + 8, 12, 12]), [8, 8, 12, 12],
          "frame_local translates to frame-space origin")
# Active spec.
sa = AquaChrome.decoration_spec(dw, true)
assert_eq(sa["title"],        "xterm",               "spec carries the title")
assert(sa["title_center"],                           "Aqua title is centred")
assert_eq(sa["title_color"],  Theme::TITLE_ACTIVE,   "active titlebar = TITLE_ACTIVE")
assert_eq(sa["title_ink"],    Theme::TITLE_TEXT_ON,  "active title ink = TITLE_TEXT_ON")
assert_eq(sa["hairline"],     Theme::TITLE_BORDER,   "spec has the titlebar hairline")
assert_eq(sa["titlebar"],     [0, 0, 240, Theme::TITLE_H], "titlebar rect is frame-local, full width")
assert_eq(sa["buttons"].length, 3,                   "three traffic-light buttons")
assert_eq(sa["buttons"][0]["shape"], "circle",       "close button is a circle")
assert_eq(sa["buttons"][0]["face"],  Theme::CLOSE_RED,  "active close = CLOSE_RED")
assert_eq(sa["buttons"][0]["rect"],  [8, 8, 12, 12], "close rect is frame-local")
assert_eq(sa["buttons"][1]["face"],  Theme::MIN_YELLOW, "active minimize = MIN_YELLOW")
assert_eq(sa["buttons"][2]["face"],  Theme::MAX_GREEN,  "active maximize = MAX_GREEN")
assert_eq(sa["border_color"], Theme::BORDER_ACTIVE,  "active border = BORDER_ACTIVE")
assert_eq(sa["shadow"],       Theme::SHADOW,         "unshaded spec has the drop shadow")
assert(sa["show_grip"],                              "unshaded spec shows the resize grip")
assert_eq(sa["grip_color"],   Theme::RESIZE_GRIP,    "grip colour = RESIZE_GRIP")
# Inactive spec: traffic-lights grey out, title + border dim.
si = AquaChrome.decoration_spec(dw, false)
assert_eq(si["title_color"],  Theme::TITLE_INACTIVE, "inactive titlebar = TITLE_INACTIVE")
assert_eq(si["title_ink"],    Theme::TITLE_TEXT_OFF, "inactive title ink = TITLE_TEXT_OFF")
assert_eq(si["border_color"], Theme::BORDER_INACTIVE,"inactive border = BORDER_INACTIVE")
assert_eq(si["buttons"][0]["face"], AquaChrome::TL_DIM_FILL, "inactive close greys out")
assert(si["buttons"][0]["face"] != sa["buttons"][0]["face"], "active/inactive close differ")
# Shaded window: only titlebar + buttons, no border / grip / shadow.
dw.shaded = true
ss = AquaChrome.decoration_spec(dw, true)
assert(!ss.key?("border"),    "shaded spec has no border")
assert(!ss.key?("grip"),      "shaded spec has no resize grip")
assert(!ss.key?("shadow"),    "shaded spec has no drop shadow")
assert_eq(ss["buttons"].length, 3, "shaded spec still shows the three buttons")

# ---- Menu#widget_items (widgets menu builder, pure) ----------------------
mroot = RootMenu.build(WindowManager.new)
items = mroot.widget_items
assert_eq(items.length, mroot.entries.length, "one widget item per menu entry")
# A submenu parent (Applications) carries a ">" chevron shortcut + an action.
apps = items.find { |it| it["label"] == "Applications" }
assert(!apps.nil?,                     "Applications entry present in widget items")
assert_eq(apps["shortcut"], ">",       "submenu parent gets a chevron shortcut")
assert_eq(apps["action"],   "x",       "submenu parent carries an action marker (paints enabled)")
# A separator becomes { "separator" => true } with no label.
sep = items.find { |it| it["separator"] }
assert(!sep.nil?,                      "separator preserved as a widget item")
assert(!sep.key?("label"),             "separator item has no label")
# A leaf entry (Exit) carries an action but no chevron.
exit_it = items.find { |it| it["label"] == "Exit" }
assert(!exit_it.nil?,                  "Exit leaf present")
assert_eq(exit_it["action"], "x",      "leaf carries an action marker")
assert(!exit_it.key?("shortcut"),      "leaf has no chevron shortcut")
`
