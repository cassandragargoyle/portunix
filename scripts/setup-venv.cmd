@echo off
REM Setup Python Virtual Environment for Portunix
REM Usage: scripts\setup-venv.cmd [--with-tests]
REM
REM Creates a virtual environment in .venv\ directory

setlocal EnableDelayedExpansion

set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%SCRIPT_DIR%.."
set "VENV_DIR=%PROJECT_ROOT%\.venv"
set "INSTALL_TESTS=0"

REM Parse arguments
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
echo   --with-tests    Also install test dependencies
echo   --help, -h      Show this help message
exit /b 0

:main
echo ================================
echo Portunix Python Environment Setup
echo ================================
echo.

cd /d "%PROJECT_ROOT%"

REM Detect Python
call :detect_python
if errorlevel 1 exit /b 1

REM Create venv
call :create_venv
if errorlevel 1 exit /b 1

REM Install dependencies
call :install_deps

REM Show instructions
call :show_instructions

exit /b 0

:detect_python
echo [==] Detecting Python...

REM Try py launcher first
py -3 --version >nul 2>&1
if %errorlevel%==0 (
    for /f "tokens=*" %%i in ('py -3 --version 2^>^&1') do set "PY_VERSION=%%i"
    echo [OK] Found Python via py launcher: !PY_VERSION!
    set "PYTHON_CMD=py -3"
    exit /b 0
)

REM Try python command
python --version >nul 2>&1
if %errorlevel%==0 (
    for /f "tokens=*" %%i in ('python --version 2^>^&1') do set "PY_VERSION=%%i"
    echo !PY_VERSION! | findstr /C:"Python 3" >nul
    if !errorlevel!==0 (
        echo [OK] Found Python: !PY_VERSION!
        set "PYTHON_CMD=python"
        exit /b 0
    )
)

echo [X] Python 3 is not installed
echo Install Python 3.8+ from https://python.org or use: portunix install python
exit /b 1

:create_venv
echo [==] Creating virtual environment in %VENV_DIR%...

if exist "%VENV_DIR%" (
    echo [!] Virtual environment already exists
    set /p "RECREATE=Do you want to recreate it? [y/N] "
    if /i "!RECREATE!"=="y" (
        rmdir /s /q "%VENV_DIR%"
    ) else (
        echo [==] Using existing virtual environment
        exit /b 0
    )
)

%PYTHON_CMD% -m venv "%VENV_DIR%"
if errorlevel 1 (
    echo [X] Failed to create virtual environment
    exit /b 1
)
echo [OK] Virtual environment created
exit /b 0

:install_deps
echo [==] Installing dependencies...

set "VENV_PYTHON=%VENV_DIR%\Scripts\python.exe"

REM Upgrade pip
"%VENV_PYTHON%" -m pip install --upgrade pip >nul 2>&1
echo [OK] pip upgraded

REM Install main requirements if exists
if exist "%PROJECT_ROOT%\requirements.txt" (
    REM Check if file has actual dependencies
    findstr /v /r "^#" "%PROJECT_ROOT%\requirements.txt" | findstr /r "." >nul 2>&1
    if !errorlevel!==0 (
        "%VENV_PYTHON%" -m pip install -r "%PROJECT_ROOT%\requirements.txt"
        echo [OK] Main dependencies installed
    ) else (
        echo [OK] No main dependencies to install
    )
)

REM Install test requirements if requested
if "%INSTALL_TESTS%"=="1" (
    if exist "%PROJECT_ROOT%\test\requirements-test.txt" (
        "%VENV_PYTHON%" -m pip install -r "%PROJECT_ROOT%\test\requirements-test.txt"
        echo [OK] Test dependencies installed
    )
)

exit /b 0

:show_instructions
echo.
echo [==] Setup complete!
echo.
echo To activate the virtual environment:
echo.
echo   CMD:           .venv\Scripts\activate.bat
echo   PowerShell:    .venv\Scripts\Activate.ps1
echo   Git Bash:      source .venv/Scripts/activate
echo.
echo Or use the helper scripts:
echo.
echo   CMD:           scripts\activate-venv.cmd
echo   PowerShell:    .\scripts\activate-venv.ps1
echo   Bash:          source scripts/activate-venv.sh
echo.
echo To deactivate: deactivate
echo.
exit /b 0
