# Issue #250: Replace Windows application icon with the Portunix gear icon

**Status**: ✅ Implemented
**Priority**: Low
**Type**: Enhancement
**Component**: portunix
**Created**: 2026-10-27
**Completed**: 2026-10-27
**GitHub**: #250

## Summary

`portunix.exe` shows a generic house icon (`home.ico`) in Windows Explorer, the taskbar
and installer shortcuts. It should show the Portunix brand icon: a gear with a play
triangle, the same symbol as in the project logo.

## Problem Description

### Current Situation

- `home.ico` in the repository root is a single 64×64 house icon from the initial
  commit
- `versioninfo.json` sets `"IconPath": "home.ico"`; `make gen-resources` runs
  goversioninfo and embeds it into `portunix.syso`, which Go links into the Windows
  build of `portunix.exe`
- `portunix.rc` (`2 ICON "home.ico"`) and `versioninfo.cmd` reference the same file;
  the `.rc` file is not compiled by the Go build, it is read for the version string
- A single 64×64 image is scaled by Windows for the 16/32/48 px and large views,
  so it looks blurry in Explorer

### Target State

- The brand gear icon is the application icon of `portunix.exe`
- The `.ico` contains the standard sizes 16, 32, 48 and 256 px, so Windows does not
  scale it
- The brand icon source (SVG master and a multi-size `.ico`) exists in the
  maintainers' brand assets and is copied into this repository; the SVG is kept next
  to the `.ico` so further sizes can be regenerated

## Proposed Solution

1. Add `assets/brand/portunix.ico` (16/32/48/256 px) and `assets/brand/portunix.svg`
   (vector master)
2. `versioninfo.json`: `"IconPath": "assets/brand/portunix.ico"`
3. Regenerate `portunix.syso` with `make gen-resources`
4. Update the icon reference in `portunix.rc` and `versioninfo.cmd`, or drop it if
   these files no longer need it
5. Remove `home.ico` once nothing references it

## Acceptance Criteria

1. `portunix.exe` built for Windows shows the gear icon in Explorer at every view size
   (small icons, details, large and extra large icons) and in the taskbar
2. The embedded icon contains 16, 32, 48 and 256 px images
   (`portunix.syso` regenerated from `versioninfo.json`)
3. No file in the repository references `home.ico`, and `home.ico` is removed
4. `make gen-resources`, `make build` and the release build (`make-release.py` /
   GoReleaser) succeed; the Windows archive's `portunix.exe` carries the new icon
5. Version information in the file properties (product name, version, copyright)
   is unchanged

## Testing

- Inspect the `.ico`: four images, sizes 16/32/48/256
- `git grep -n home.ico` returns nothing
- Build for Windows, open the folder in Explorer on a Windows VM and switch between
  view sizes; check the taskbar icon while `portunix.exe` runs
- Check file properties → Details: version fields as before
