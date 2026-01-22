<#
.SYNOPSIS
    Portunix Docusaurus QuickStart - Create documentation site in minutes

.DESCRIPTION
    This script sets up a complete Docusaurus documentation environment using Portunix.
    It handles all prerequisites including Portunix installation, container runtime,
    and project initialization.

.QUICK START
    # Option 1: Direct execution (recommended)
    irm https://github.com/cassandragargoyle/Portunix/releases/latest/download/quickstart-docusaurus.ps1 | iex

    # Option 2: Download and run
    Invoke-WebRequest -Uri "https://github.com/cassandragargoyle/Portunix/releases/latest/download/quickstart-docusaurus.ps1" -OutFile "quickstart-docusaurus.ps1"
    .\quickstart-docusaurus.ps1

    # Option 3: With parameters (non-interactive)
    .\quickstart-docusaurus.ps1 -ProjectPath "C:\Projects\my-docs" -ProjectName "my-docs" -NonInteractive

.PARAMETER ProjectPath
    Path where the documentation project will be created. Default: ./my-docs

.PARAMETER ProjectName
    Name of the documentation project. Default: my-docs

.PARAMETER NonInteractive
    Run without interactive prompts (for CI/CD usage)

.PARAMETER SkipPortunixCheck
    Skip Portunix installation/update check

.PARAMETER SkipContainerCheck
    Skip container runtime check

.NOTES
    Requires: Windows 10/11 with PowerShell 5.1+
    Container: Docker or Podman (will be installed automatically if needed)
    Author: CassandraGargoyle Team
    Version: 1.0.0
#>

[CmdletBinding()]
param(
    [string]$ProjectPath = "",
    [string]$ProjectName = "",
    [switch]$NonInteractive,
    [switch]$SkipPortunixCheck,
    [switch]$SkipContainerCheck
)

# ============================================================================
# Configuration
# ============================================================================

$script:Version = "1.0.0"
$script:PortunixMinVersion = "1.9.0"
$script:DefaultProjectPath = ".\my-docs"
$script:DefaultProjectName = "my-docs"
$script:PortunixPath = $null  # Will be set by Find-Portunix

# ============================================================================
# Helper Functions
# ============================================================================

function Write-Banner {
    $banner = @"

    ____             __                   _
   / __ \____  _____/ /___  ______  (_)  __
  / /_/ / __ \/ ___/ __/ / / / __ \/ / |/_/
 / ____/ /_/ / /  / /_/ /_/ / / / / />  <
/_/    \____/_/   \__/\__,_/_/ /_/_/_/|_|

    Docusaurus QuickStart v$script:Version

"@
    Write-Host $banner -ForegroundColor Cyan
}

function Write-Step {
    param([string]$Step, [string]$Message)
    Write-Host ""
    Write-Host "[$Step] " -ForegroundColor Yellow -NoNewline
    Write-Host $Message -ForegroundColor White
}

function Write-Status {
    param([string]$Status, [string]$Message)
    switch ($Status) {
        "OK"      { Write-Host "  [OK] " -ForegroundColor Green -NoNewline; Write-Host $Message }
        "FAIL"    { Write-Host "  [FAIL] " -ForegroundColor Red -NoNewline; Write-Host $Message }
        "WARN"    { Write-Host "  [WARN] " -ForegroundColor Yellow -NoNewline; Write-Host $Message }
        "INFO"    { Write-Host "  [INFO] " -ForegroundColor Cyan -NoNewline; Write-Host $Message }
        "SKIP"    { Write-Host "  [SKIP] " -ForegroundColor Gray -NoNewline; Write-Host $Message }
        default   { Write-Host "  $Message" }
    }
}

function Write-Error-Box {
    param([string]$Title, [string[]]$Messages)
    Write-Host ""
    Write-Host "  +-----------------------------------------+" -ForegroundColor Red
    Write-Host "  | ERROR: $Title" -ForegroundColor Red
    Write-Host "  +-----------------------------------------+" -ForegroundColor Red
    foreach ($msg in $Messages) {
        Write-Host "  | $msg" -ForegroundColor Red
    }
    Write-Host "  +-----------------------------------------+" -ForegroundColor Red
    Write-Host ""
}

function Write-Success-Box {
    param([string]$Title, [string[]]$Messages)
    Write-Host ""
    Write-Host "  +-----------------------------------------+" -ForegroundColor Green
    Write-Host "  | $Title" -ForegroundColor Green
    Write-Host "  +-----------------------------------------+" -ForegroundColor Green
    foreach ($msg in $Messages) {
        Write-Host "  | $msg" -ForegroundColor White
    }
    Write-Host "  +-----------------------------------------+" -ForegroundColor Green
    Write-Host ""
}

function Get-UserInput {
    param(
        [string]$Prompt,
        [string]$Default = ""
    )

    if ($NonInteractive) {
        return $Default
    }

    $displayDefault = if ($Default) { " [$Default]" } else { "" }
    Write-Host "  $Prompt$displayDefault`: " -ForegroundColor White -NoNewline
    $input = Read-Host

    if ([string]::IsNullOrWhiteSpace($input)) {
        return $Default
    }
    return $input
}

function Get-UserConfirmation {
    param(
        [string]$Prompt,
        [bool]$Default = $true
    )

    if ($NonInteractive) {
        return $Default
    }

    $defaultStr = if ($Default) { "Y/n" } else { "y/N" }
    Write-Host "  $Prompt [$defaultStr]: " -ForegroundColor White -NoNewline
    $input = Read-Host

    if ([string]::IsNullOrWhiteSpace($input)) {
        return $Default
    }

    return $input -match "^[Yy]"
}

function Test-CommandExists {
    param([string]$Command)
    $null -ne (Get-Command $Command -ErrorAction SilentlyContinue)
}

function Find-Portunix {
    # Check if portunix is in PATH
    $inPath = Get-Command "portunix" -ErrorAction SilentlyContinue
    if ($inPath) {
        return $inPath.Source
    }

    # Check common installation locations
    $commonPaths = @(
        "C:\Portunix\portunix.exe",
        "$env:LOCALAPPDATA\Programs\Portunix\portunix.exe",
        "$env:ProgramFiles\Portunix\portunix.exe"
    )

    foreach ($path in $commonPaths) {
        if (Test-Path $path) {
            return $path
        }
    }

    return $null
}

# ============================================================================
# Portunix Installation Functions
# ============================================================================

function Get-PortunixVersion {
    $script:PortunixPath = Find-Portunix
    if (-not $script:PortunixPath) {
        return $null
    }

    try {
        $versionOutput = & $script:PortunixPath --version 2>&1
        # Return full version string (e.g., "1.9.0", "1.9.0-dev", "1.9.0-SNAPSHOT")
        if ($versionOutput -match "v?(\d+\.\d+\.\d+(?:-[a-zA-Z0-9]+)?)") {
            return $matches[1]
        }
        # Handle "dev" version without numeric part
        if ($versionOutput -match "\bdev\b") {
            return "dev"
        }
    } catch {
        return $null
    }
    return $null
}

function Test-IsDevVersion {
    param([string]$Version)

    if (-not $Version) { return $false }

    # Dev versions: "dev" alone, or suffix like -dev, -SNAPSHOT, -alpha, -beta, -rc
    return ($Version -eq "dev") -or ($Version -match "-(?:dev|SNAPSHOT|alpha|beta|rc)")
}

function Get-VersionNumeric {
    param([string]$Version)

    if (-not $Version) { return $null }

    # Extract numeric part (e.g., "1.9.0" from "1.9.0-dev")
    if ($Version -match "^(\d+\.\d+\.\d+)") {
        return $matches[1]
    }
    return $null
}

function Test-PortunixVersionOk {
    param([string]$CurrentVersion)

    if (-not $CurrentVersion) { return $false }

    $numericVersion = Get-VersionNumeric $CurrentVersion
    if (-not $numericVersion) { return $false }

    $current = [version]$numericVersion
    $minimum = [version]$script:PortunixMinVersion

    return $current -ge $minimum
}

function Install-Portunix {
    Write-Status "INFO" "Installing Portunix..."

    try {
        # Determine architecture
        $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

        # Get latest version from GitHub API
        Write-Status "INFO" "Checking latest Portunix version..."
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        $apiUrl = "https://api.github.com/repos/cassandragargoyle/Portunix/releases/latest"
        $release = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing
        $version = $release.tag_name  # e.g., "v1.10.0"
        $versionNum = $version.TrimStart('v')  # e.g., "1.10.0"

        # Construct direct download URL
        $fileName = "portunix_${versionNum}_windows_${arch}.zip"
        $releaseUrl = "https://github.com/cassandragargoyle/Portunix/releases/download/${version}/${fileName}"
        Write-Status "INFO" "Downloading: $fileName"

        # Create temp directory
        $tempDir = Join-Path $env:TEMP "portunix-install"
        if (Test-Path $tempDir) {
            Remove-Item -Recurse -Force $tempDir
        }
        New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

        # Download release archive
        Write-Status "INFO" "Downloading Portunix from GitHub..."
        $zipPath = Join-Path $tempDir "portunix.zip"

        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        Invoke-WebRequest -Uri $releaseUrl -OutFile $zipPath -UseBasicParsing

        # Extract archive
        Write-Status "INFO" "Extracting archive..."
        Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force

        # Find portunix.exe in extracted files
        $portunixExe = Get-ChildItem -Path $tempDir -Recurse -Filter "portunix.exe" | Select-Object -First 1
        if (-not $portunixExe) {
            throw "portunix.exe not found in downloaded archive"
        }

        # Install to C:\Portunix
        $installDir = "C:\Portunix"
        if (-not (Test-Path $installDir)) {
            New-Item -ItemType Directory -Path $installDir -Force | Out-Null
        }

        # Copy all files from the extracted directory
        $sourceDir = $portunixExe.DirectoryName
        Copy-Item -Path "$sourceDir\*" -Destination $installDir -Recurse -Force
        Write-Status "OK" "Portunix installed to $installDir"

        # Add to PATH if not already there
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($userPath -notlike "*$installDir*") {
            [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
            Write-Status "OK" "Added $installDir to PATH"
        }

        # Refresh PATH for current session
        $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")

        # Cleanup
        Remove-Item -Recurse -Force $tempDir -ErrorAction SilentlyContinue

        # Set portunix path
        $script:PortunixPath = Join-Path $installDir "portunix.exe"

        if (Test-Path $script:PortunixPath) {
            Write-Status "OK" "Portunix installed successfully"
            return $true
        } else {
            Write-Status "FAIL" "Installation completed but portunix.exe not found"
            return $false
        }
    } catch {
        Write-Status "FAIL" "Installation failed: $($_.Exception.Message)"
        return $false
    }
}

function Update-Portunix {
    Write-Status "INFO" "Updating Portunix..."

    try {
        & $script:PortunixPath update
        if ($LASTEXITCODE -eq 0) {
            Write-Status "OK" "Portunix updated successfully"
            return $true
        } else {
            Write-Status "WARN" "Update check completed (may already be latest)"
            return $true
        }
    } catch {
        Write-Status "WARN" "Update failed: $($_.Exception.Message)"
        return $true  # Continue anyway with current version
    }
}

# ============================================================================
# Container Runtime Functions
# ============================================================================

function Get-ContainerRuntimeInfo {
    # Use portunix system info to detect container runtime
    try {
        $systemInfo = & $script:PortunixPath system info --json 2>&1
        if ($LASTEXITCODE -eq 0 -and $systemInfo) {
            $info = $systemInfo | ConvertFrom-Json

            if ($info.capabilities) {
                $caps = $info.capabilities

                # Primary check: container_available
                if ($caps.container_available -eq $true) {
                    # Determine which runtime
                    $runtime = if ($caps.docker -eq $true) { "docker" } elseif ($caps.podman -eq $true) { "podman" } else { "unknown" }

                    # Check if daemon is running
                    $isRunning = -not ($caps.compose.warning -match "not running")
                    $warning = $caps.compose.warning

                    return @{
                        Runtime = $runtime
                        Installed = $true
                        Running = $isRunning
                        Warning = $warning
                    }
                }
            }
        }
    } catch {
        Write-Status "WARN" "Could not parse system info: $($_.Exception.Message)"
    }

    return @{ Runtime = $null; Installed = $false; Running = $false; Warning = $null }
}

# ============================================================================
# Main Workflow
# ============================================================================

function Start-QuickStart {
    $ErrorActionPreference = "Stop"

    Write-Banner

    # -------------------------------------------------------------------------
    # Step 1: Check Portunix
    # -------------------------------------------------------------------------
    Write-Step "1/6" "Checking Portunix installation..."

    if ($SkipPortunixCheck) {
        Write-Status "SKIP" "Portunix check skipped"
    } else {
        $portunixVersion = Get-PortunixVersion

        if (-not $portunixVersion) {
            Write-Status "WARN" "Portunix is not installed"

            if (Get-UserConfirmation -Prompt "Install Portunix now?") {
                if (-not (Install-Portunix)) {
                    Write-Error-Box "Installation Failed" @(
                        "Could not install Portunix automatically.",
                        "",
                        "Manual installation:",
                        "  irm https://portunix.ai/install.ps1 | iex",
                        "",
                        "Then run this script again."
                    )
                    exit 1
                }
                $portunixVersion = Get-PortunixVersion
            } else {
                Write-Error-Box "Portunix Required" @(
                    "This script requires Portunix to function.",
                    "",
                    "Install manually:",
                    "  irm https://portunix.ai/install.ps1 | iex"
                )
                exit 1
            }
        }

        Write-Status "OK" "Portunix $portunixVersion found"

        # Skip update check for dev/snapshot versions
        if (Test-IsDevVersion $portunixVersion) {
            Write-Status "INFO" "Development version detected - skipping update check"
        } elseif (-not (Test-PortunixVersionOk $portunixVersion)) {
            Write-Status "WARN" "Portunix version $portunixVersion is older than recommended ($script:PortunixMinVersion)"

            if (Get-UserConfirmation -Prompt "Update Portunix to latest version?") {
                Update-Portunix
            }
        } else {
            # Check for updates anyway (optional)
            if (-not $NonInteractive) {
                if (Get-UserConfirmation -Prompt "Check for Portunix updates?" -Default $false) {
                    Update-Portunix
                }
            }
        }
    }

    # -------------------------------------------------------------------------
    # Step 2: Check Container Runtime
    # -------------------------------------------------------------------------
    Write-Step "2/6" "Checking container runtime..."

    if ($SkipContainerCheck) {
        Write-Status "SKIP" "Container check skipped"
    } else {
        $containerInfo = Get-ContainerRuntimeInfo

        if ($containerInfo.Installed) {
            $runtimeName = $containerInfo.Runtime
            if ($containerInfo.Running) {
                Write-Status "OK" "Container runtime: $runtimeName (running)"
            } else {
                Write-Status "OK" "Container runtime: $runtimeName (installed)"
                if ($containerInfo.Warning) {
                    Write-Status "WARN" $containerInfo.Warning
                }
            }
        } else {
            Write-Status "INFO" "No container runtime currently installed"
            Write-Status "INFO" "Portunix will automatically install Docker/Podman when needed"
            Write-Host ""
            Write-Host "  Note: When you run a playbook that requires containers," -ForegroundColor Gray
            Write-Host "  Portunix will prompt you to install the container runtime." -ForegroundColor Gray
            Write-Host ""

            if (-not $NonInteractive) {
                if (-not (Get-UserConfirmation -Prompt "Continue without container runtime?")) {
                    Write-Host ""
                    Write-Host "  You can install container runtime manually:" -ForegroundColor Yellow
                    Write-Host "    portunix install docker" -ForegroundColor Cyan
                    Write-Host "    portunix install podman" -ForegroundColor Cyan
                    Write-Host ""
                    exit 0
                }
            }
        }
    }

    # -------------------------------------------------------------------------
    # Step 3: Configure Project
    # -------------------------------------------------------------------------
    Write-Step "3/6" "Configuring project..."

    # Project path
    if ([string]::IsNullOrWhiteSpace($ProjectPath)) {
        $ProjectPath = Get-UserInput -Prompt "Project path" -Default $script:DefaultProjectPath
    }

    # Expand to absolute path
    $ProjectPath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($ProjectPath)

    # Project name (derive from path if not specified)
    if ([string]::IsNullOrWhiteSpace($ProjectName)) {
        $defaultName = Split-Path $ProjectPath -Leaf
        if ([string]::IsNullOrWhiteSpace($defaultName)) {
            $defaultName = $script:DefaultProjectName
        }
        $ProjectName = Get-UserInput -Prompt "Project name" -Default $defaultName
    }

    Write-Status "INFO" "Project path: $ProjectPath"
    Write-Status "INFO" "Project name: $ProjectName"

    # Check if path exists
    if (Test-Path $ProjectPath) {
        Write-Status "WARN" "Directory already exists: $ProjectPath"
        if (-not (Get-UserConfirmation -Prompt "Continue and potentially overwrite?")) {
            Write-Status "INFO" "Aborted by user"
            exit 0
        }
    }

    # -------------------------------------------------------------------------
    # Step 4: Create Project Directory
    # -------------------------------------------------------------------------
    Write-Step "4/6" "Creating project directory..."

    try {
        if (-not (Test-Path $ProjectPath)) {
            New-Item -ItemType Directory -Path $ProjectPath -Force | Out-Null
            Write-Status "OK" "Created: $ProjectPath"
        } else {
            Write-Status "OK" "Using existing: $ProjectPath"
        }

        Set-Location $ProjectPath
        Write-Status "INFO" "Working directory: $(Get-Location)"
    } catch {
        Write-Error-Box "Directory Error" @(
            "Could not create project directory.",
            "",
            "Error: $($_.Exception.Message)",
            "",
            "Check permissions and try again."
        )
        exit 1
    }

    # -------------------------------------------------------------------------
    # Step 5: Generate Playbook and Create Project
    # -------------------------------------------------------------------------
    Write-Step "5/6" "Setting up Docusaurus project..."

    $playbookFile = "$ProjectName.ptxbook"

    try {
        # Generate playbook from template
        Write-Status "INFO" "Generating playbook from template..."
        & $script:PortunixPath playbook init $ProjectName --template static-docs --engine docusaurus --target container

        if (-not (Test-Path $playbookFile)) {
            throw "Playbook file was not created"
        }
        Write-Status "OK" "Playbook created: $playbookFile"

        # Validate playbook
        Write-Status "INFO" "Validating playbook..."
        & $script:PortunixPath playbook validate $playbookFile
        if ($LASTEXITCODE -ne 0) {
            throw "Playbook validation failed"
        }
        Write-Status "OK" "Playbook is valid"

        # Run create script to initialize Docusaurus
        Write-Status "INFO" "Creating Docusaurus project (this may take a few minutes)..."
        Write-Host "  Command: $script:PortunixPath playbook run $playbookFile --script create" -ForegroundColor Gray
        & $script:PortunixPath playbook run $playbookFile --script create

        if ($LASTEXITCODE -ne 0) {
            throw "Docusaurus project creation failed"
        }
        Write-Status "OK" "Docusaurus project created successfully"

    } catch {
        Write-Error-Box "Setup Failed" @(
            "Could not create Docusaurus project.",
            "",
            "Error: $($_.Exception.Message)",
            "",
            "Try running manually:",
            "  cd $ProjectPath",
            "  portunix playbook init $ProjectName --template static-docs --engine docusaurus --target container",
            "  portunix playbook run $playbookFile --script create"
        )
        exit 1
    }

    # -------------------------------------------------------------------------
    # Step 6: Success!
    # -------------------------------------------------------------------------
    Write-Step "6/6" "Setup complete!"

    Write-Success-Box "Docusaurus Project Ready!" @(
        "",
        "Location: $ProjectPath",
        "Playbook: $playbookFile",
        ""
    )

    Write-Host "  Available Commands:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "    Start development server:" -ForegroundColor White
    Write-Host "      portunix playbook run $playbookFile --script dev" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "    Build for production:" -ForegroundColor White
    Write-Host "      portunix playbook run $playbookFile --script build" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "    List all available scripts:" -ForegroundColor White
    Write-Host "      portunix playbook run $playbookFile --list-scripts" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  Documentation:" -ForegroundColor Yellow
    Write-Host "    Portunix: https://github.com/cassandragargoyle/Portunix" -ForegroundColor Gray
    Write-Host "    Docusaurus: https://docusaurus.io/docs" -ForegroundColor Gray
    Write-Host ""

    # Offer to start dev server
    if (-not $NonInteractive) {
        if (Get-UserConfirmation -Prompt "Start development server now?") {
            Write-Host ""
            Write-Host "  Starting development server in new window..." -ForegroundColor Cyan
            Write-Host "  Server will be available at: http://localhost:3000" -ForegroundColor Green
            Write-Host "  Close the new window to stop the server" -ForegroundColor Gray
            Write-Host ""

            # Start dev server in a new terminal window so the quickstart script can exit
            $devCommand = "$script:PortunixPath playbook run $playbookFile --script dev"
            Start-Process powershell -ArgumentList "-NoExit", "-Command", $devCommand

            Write-Host "  Development server started in new window" -ForegroundColor Green
            Write-Host ""
        }
    }
}

# ============================================================================
# Entry Point
# ============================================================================

try {
    Start-QuickStart
} catch {
    Write-Error-Box "Unexpected Error" @(
        $_.Exception.Message,
        "",
        "If this persists, please report at:",
        "https://github.com/cassandragargoyle/Portunix/issues"
    )
    exit 1
}
