# Issue #124: Download Progress Indicator

## Summary

When downloading large files (e.g., Docker Desktop Installer at 571MB), there is no visual feedback showing download progress. Users see only the initial message and then nothing until download completes, making it appear the application is frozen.

## Problem Statement

Current behavior:
```
📥 Downloading: Docker Desktop Installer.exe
📦 Size: 571.52 MB
```

Then silence for several minutes while the file downloads. Users have no way to know:
- If the download is still in progress
- How much has been downloaded
- Estimated time remaining

## Proposed Solution

Add a progress indicator during file downloads:

```
📥 Downloading: Docker Desktop Installer.exe
📦 Size: 571.52 MB
⏳ Progress: [████████████░░░░░░░░] 62% (354.5 MB / 571.5 MB) - 2m 15s remaining
```

Or simpler spinning indicator with percentage:
```
📥 Downloading: Docker Desktop Installer.exe (62% - 354.5 MB / 571.5 MB)
```

## Technical Considerations

1. **Terminal compatibility**: Progress bar must work in various terminals (Windows CMD, PowerShell, Linux terminals)
2. **Carriage return**: Use `\r` to update line in place
3. **Fallback**: If terminal doesn't support in-place updates, show periodic updates (every 10%)
4. **Speed calculation**: Show download speed (MB/s) and ETA

## Affected Components

- `src/helpers/ptx-installer/engine/download.go` - DownloadFile functions
- Any other download functions in the codebase

## Implementation Notes

Go example for progress display:
```go
func downloadWithProgress(url, dest string) error {
    resp, _ := http.Get(url)
    defer resp.Body.Close()

    total := resp.ContentLength
    var downloaded int64

    file, _ := os.Create(dest)
    defer file.Close()

    buf := make([]byte, 32*1024)
    for {
        n, err := resp.Body.Read(buf)
        if n > 0 {
            file.Write(buf[:n])
            downloaded += int64(n)
            percent := float64(downloaded) / float64(total) * 100
            fmt.Printf("\r⏳ Progress: %.1f%% (%s / %s)",
                percent, formatBytes(downloaded), formatBytes(total))
        }
        if err == io.EOF {
            break
        }
    }
    fmt.Println() // New line after progress
    return nil
}
```

## Priority

Medium

## Labels

enhancement, user-experience, ptx-installer, download

---
**Created**: 2026-01-03
**Status**: Implemented
**Implemented**: 2026-01-03
