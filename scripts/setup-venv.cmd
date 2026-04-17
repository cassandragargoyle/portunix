@echo off
REM Setup Python Virtual Environment for Portunix (uv-based)
REM Usage: scripts\setup-venv.cmd [--with-tests]
REM
REM Provisions .\.venv\ via `uv sync` (or `uv sync --group test` with
REM --with-tests). See ADR-039 for the rationale behind adopting uv.

setlocal EnableDelayedExpansion

set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%SCRIPT_DIR%.."
set "INSTALL_TESTS=0"

:parse_args
if "%~1"=="" goto :main
if /i "%~1"=="--with-tests" set "INSTALL_TESTS=1"
if /i "%~1"=="--help" goto :show_help
if /i "%~1"=="-h" goto :show_help
shift
goto :parse_args

:show_help
echo Usage: %~nx0 [options]
echo.
echo Options:
echo   --with-tests    Also install test dependencies (pytest, ...)
echo   --help, -h      Show this help message
echo.
echo This script uses uv ^(https://docs.astral.sh/uv/^) to provision the
echo Python environment. If uv is not installed, run one of:
echo.
echo   portunix install uv
echo   powershell -c "irm https://astral.sh/uv/install.ps1 ^| iex"
exit /b 0

:main
echo ================================
echo Portunix Python Environment Setup
echo ================================
echo.

cd /d "%PROJECT_ROOT%"

call :check_uv
if errorlevel 1 exit /b 1

call :uv_sync
if errorlevel 1 exit /b 1

call :show_instructions
exit /b 0

:check_uv
echo [==] Detecting uv...
uv --version >nul 2>&1
if %errorlevel%==0 (
    for /f "tokens=*" %%i in ('uv --version 2^>^&1') do set "UV_VERSION=%%i"
    echo [OK] Found uv: !UV_VERSION!
    exit /b 0
)
echo [X] uv is not installed or not on PATH
echo.
echo Install uv via one of:
echo   portunix install uv
echo   powershell -c "irm https://astral.sh/uv/install.ps1 ^| iex"
echo.
echo See https://docs.astral.sh/uv/ for details.
exit /b 1

:uv_sync
if "%INSTALL_TESTS%"=="1" (
    echo [==] Provisioning .venv\ with test dependencies ^(uv sync --group test^)...
    uv sync --group test
) else (
    echo [==] Provisioning .venv\ ^(uv sync^)...
    uv sync
)
if errorlevel 1 (
    echo [X] uv sync failed
    exit /b 1
)
echo [OK] Environment ready
exit /b 0

:show_instructions
echo.
echo [==] Setup complete!
echo.
echo Run commands without activation:
echo   uv run pytest test/
echo   uv run scripts\file-server.py --help
echo.
echo Or activate the virtual environment:
echo.
echo   CMD:           .venv\Scripts\activate.bat
echo   PowerShell:    .venv\Scripts\Activate.ps1
echo   Git Bash:      source .venv/Scripts/activate
echo.
echo To deactivate: deactivate
echo.
exit /b 0
