@echo off

REM OS validation for local testing
for /f "tokens=4-5 delims=. " %%i in ('ver') do set VERSION=%%i.%%j
if "%OS%" NEQ "Windows_NT" (
    echo ⚠️  WARNING: You are running on non-Windows OS but switching to Windows tester role
    echo    This role is designed for Windows testing and should only accept Windows tests on local host
    echo    For container/VM testing, accept tests based on container/VM OS, not host OS
    echo.
)

if exist ".claude\roles\current.md" (
    del ".claude\roles\current.md"
)
mklink ".claude\roles\current.md" "tester-windows.md"
echo 🧪  Switched to role: TESTER (Windows)
echo.
echo 📋 Role Guidelines:
echo    • Local host testing: Only accept Windows tests
echo    • Container/VM testing: Accept tests based on container/VM OS
echo    • Always document tested OS in acceptance protocol