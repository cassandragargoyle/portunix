@echo off
if exist ".claude\roles\current.md" (
    del ".claude\roles\current.md"
)
mklink ".claude\roles\current.md" "architect.md"
echo 🏛️  Switched to role: ARCHITECT (Windows)