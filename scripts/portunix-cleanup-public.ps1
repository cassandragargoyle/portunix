# cleanup-public.ps1
# ---------------------------------
# This script cleans the branch intended for GitHub from private files.
# Run it in the worktree folder (e.g., ../portunix-public).

# List of files/folders you don't want on GitHub.
# Add or remove as needed.
# Execute:
# powershell -ExecutionPolicy Bypass -File .\cleanup-public.ps1
$remove = @(
  'CLAUDE.md',
  'GEMINI.md',
  'NOTES.md',
  'bin/',
  '*.exe',
  'docs/private/**',
  'config/dev/**',
  'package.portunix.linux.bat'
  'package.portunix.windows.bat',
  'build.portunix.linux.arm.bat',
  'build.portunix.linux.bat',
  'build.portunix.linux.sh',
  'app/service_lnx.go',
  'cmd/login.go',
  'build.portunix.windows.bat',
  'scripts\package-win.ps1'
)

Write-Host "Removing private files from public branch..."

foreach ($item in $remove) {
    # --cached = the file is removed from Git, but stays on disk
    # If you want to remove it physically as well, remove --cached
    # git rm -r -f --ignore-unmatch --cached -- $item
    git rm -r -f --ignore-unmatch -- $item
}

# Commit only if there is something to commit
$changes = git status --porcelain
 if ($changes) {
    git commit -m "public: remove private files before publishing"
    Write-Host "Cleanup commit created."
} else {
    Write-Host "No files to remove. Nothing committed."
}
