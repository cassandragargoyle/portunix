#requires -version 5.1
<#
  Prepare portable Win32-OpenSSH on Windows (Variant B):
  - Create/normalize 'sshd' service first (so NT SERVICE\sshd exists)
  - Use SID-based identity resolution (works on localized Windows)
  - Apply ACL fixes (prefer portable scripts, fallback to in-box)
  - Generate host keys, open firewall, start service

  Run as Administrator.
#>

[CmdletBinding()]
param(
  # If the script sits in the same folder as sshd.exe, default is fine:
  [Parameter(Mandatory=$false)]
  [ValidateScript({ Test-Path $_ -PathType Container })]
  [string]$OpenSSHPath = $PSScriptRoot,

  [ValidateRange(1,65535)]
  [int]$Port = 22,

  [switch]$NoFirewall,
  [switch]$NoStart
)

function Test-Admin {
  $id = [Security.Principal.WindowsIdentity]::GetCurrent()
  $p  = New-Object Security.Principal.WindowsPrincipal($id)
  if (-not $p.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw "Run this script in an elevated PowerShell (Run as Administrator)."
  }
}

function Resolve-SIDName([string]$Sid) {
  try {
    return ([Security.Principal.SecurityIdentifier]$Sid).Translate([Security.Principal.NTAccount]).Value
  } catch {
    Write-Verbose "Could not translate SID ${Sid}: $($_.Exception.Message)"
    return $null
  }
}

function Ensure-Service {
  param(
    [Parameter(Mandatory=$true)][string]$Name,
    [Parameter(Mandatory=$true)][string]$BinPath,
    [string]$StartType = 'auto'
  )
  $svc = Get-Service -Name $Name -ErrorAction SilentlyContinue
  if (-not $svc) {
    Write-Verbose "Creating service '$Name' with binPath: $BinPath"
    $null = sc.exe create $Name binPath= "`"$BinPath`"" start= $StartType
    if ($LASTEXITCODE -ne 0) { throw "Failed to create service '$Name' ($LASTEXITCODE)." }
  } else {
    Write-Verbose "Service '$Name' already exists. Ensuring binPath and start type."
    $null = sc.exe config $Name binPath= "`"$BinPath`""
    $null = sc.exe config $Name start= $StartType
  }
}

function Ensure-FirewallRule {
  param(
    [Parameter(Mandatory=$true)][string]$Name,
    [Parameter(Mandatory=$true)][int]$Port
  )
  $rule = netsh advfirewall firewall show rule name="$Name" | Out-String
  if ($rule -match "No rules match") {
    Write-Verbose "Creating firewall rule '$Name' on TCP $Port"
    $null = netsh advfirewall firewall add rule name="$Name" dir=in action=allow protocol=TCP localport=$Port
  } else {
    Write-Verbose "Firewall rule '$Name' already exists."
  }
}

function Ensure-HostKeys {
  $progDataSsh = Join-Path $env:PROGRAMDATA 'ssh'
  if (-not (Test-Path $progDataSsh)) {
    $null = New-Item -ItemType Directory -Path $progDataSsh -Force
  }
  $keygen = Join-Path $OpenSSHPath 'ssh-keygen.exe'
  if (-not (Test-Path $keygen)) { throw "ssh-keygen.exe not found at $keygen" }
  & $keygen -A | Out-Null
  if ($LASTEXITCODE -ne 0) { throw "Host key generation failed." }
  Write-Verbose "Host keys ensured in $progDataSsh"
}

function Set-WDAGUserPassword {
  param(
    [string]$Password = "<YOUR_PASSWORD>"
  )
  
  try {
    Write-Verbose "Setting password for WDAGUtilityAccount..."
    $securePassword = ConvertTo-SecureString $Password -AsPlainText -Force
    $user = Get-LocalUser -Name "WDAGUtilityAccount" -ErrorAction SilentlyContinue
    
    if ($user) {
      Set-LocalUser -Name "WDAGUtilityAccount" -Password $securePassword
      Write-Verbose "Password set for WDAGUtilityAccount"
      
      # Enable the account if disabled
      if (-not $user.Enabled) {
        Enable-LocalUser -Name "WDAGUtilityAccount"
        Write-Verbose "WDAGUtilityAccount enabled"
      }
      
      return $Password
    } else {
      Write-Warning "WDAGUtilityAccount not found"
      return $null
    }
  } catch {
    Write-Warning "Failed to set password for WDAGUtilityAccount: $($_.Exception.Message)"
    return $null
  }
}

function Get-SandboxIP {
  try {
    # Get all network adapters and find the one with a valid IP
    $adapters = Get-NetAdapter | Where-Object { $_.Status -eq 'Up' }
    foreach ($adapter in $adapters) {
      $ip = Get-NetIPAddress -InterfaceIndex $adapter.InterfaceIndex -AddressFamily IPv4 -ErrorAction SilentlyContinue | 
            Where-Object { $_.IPAddress -ne '127.0.0.1' -and $_.IPAddress -ne '169.254.*' }
      if ($ip) {
        Write-Verbose "Found sandbox IP: $($ip.IPAddress)"
        return $ip.IPAddress
      }
    }
    
    # Fallback to Get-WmiObject
    $ip = Get-WmiObject -Class Win32_NetworkAdapterConfiguration | 
          Where-Object { $_.DefaultIPGateway -ne $null } | 
          Select-Object -First 1 -ExpandProperty IPAddress | 
          Where-Object { $_ -match '^192\.168\.|^10\.|^172\.' }
    
    if ($ip) {
      Write-Verbose "Found sandbox IP (fallback): $ip"
      return $ip
    }
    
    Write-Warning "Could not determine sandbox IP address"
    return "localhost"
  } catch {
    Write-Warning "Error getting sandbox IP: $($_.Exception.Message)"
    return "localhost"
  }
}

function Try-RunAclFixes {
  $portableFixHost = Join-Path $OpenSSHPath 'FixHostFilePermissions.ps1'
  $portableFixUser = Join-Path $OpenSSHPath 'FixUserFilePermissions.ps1'
  $portableUtils   = Join-Path $OpenSSHPath 'OpenSSHUtils.psm1'

  $inboxDir     = Join-Path $env:WINDIR 'System32\OpenSSH'
  $inboxFixHost = Join-Path $inboxDir 'FixHostFilePermissions.ps1'
  $inboxFixUser = Join-Path $inboxDir 'FixUserFilePermissions.ps1'
  $inboxUtils   = Join-Path $inboxDir 'OpenSSHUtils.psm1'

  $hasPortable = (Test-Path $portableFixHost) -and (Test-Path $portableFixUser)
  $hasInbox    = (Test-Path $inboxFixHost) -and (Test-Path $inboxFixUser)
  $usedPortable = $false

  if ($hasPortable) {
    if (Test-Path $portableUtils) { Import-Module $portableUtils -Force -ErrorAction Stop }
    Write-Verbose "Running portable ACL fixes..."
    & $portableFixHost -Confirm:$false
    & $portableFixUser -Confirm:$false
    $usedPortable = $true
  }
  elseif ($hasInbox) {
    if (Test-Path $inboxUtils) { Import-Module $inboxUtils -Force -ErrorAction Stop }
    Write-Verbose "Running in-box ACL fixes..."
    & $inboxFixHost -Confirm:$false
    & $inboxFixUser -Confirm:$false
  }
  else {
    Write-Warning "No FixHostFilePermissions.ps1/FixUserFilePermissions.ps1 found (portable or in-box). Skipping ACL fix step."
    return
  }

  if ($LASTEXITCODE -ne 0) {
    Write-Warning "ACL fix scripts returned a non-zero exit code."
  } else {
    $mode = 'in-box'
    if ($usedPortable) { $mode = 'portable' }
    Write-Verbose "ACL fixes completed ($mode)."
  }
}

try {
  Test-Admin

  $OpenSSHPath = (Resolve-Path $OpenSSHPath).Path
  $sshdPath    = Join-Path $OpenSSHPath 'sshd.exe'
  if (-not (Test-Path $sshdPath)) { throw "sshd.exe not found at '$sshdPath'. Check -OpenSSHPath." }

  # Resolve localized names (for logging only)
  $BuiltinAdmins = Resolve-SIDName 'S-1-5-32-544'
  $BuiltinUsers  = Resolve-SIDName 'S-1-5-32-545'
  $SystemAccount = Resolve-SIDName 'S-1-5-18'
  $AuthUsers     = Resolve-SIDName 'S-1-5-11'
  Write-Verbose "Admins=$BuiltinAdmins; Users=$BuiltinUsers; System=$SystemAccount; AU=$AuthUsers"

  # 1) Set password for WDAGUtilityAccount
  $password = Set-WDAGUserPassword
  if ($password) {
    Write-Host "WDAGUtilityAccount password set successfully" -ForegroundColor Green
  }

  # 2) Get sandbox IP address
  $sandboxIP = Get-SandboxIP
  Write-Host "Sandbox IP: $sandboxIP" -ForegroundColor Cyan
  
  # Write IP and password to shared folder for host access
  $tmpDir = 'C:\Portunix\.tmp'
  if (-not (Test-Path $tmpDir)) {
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null
  }
  $infoFile = Join-Path $tmpDir 'ssh_info.txt'
  @"
Sandbox SSH Information
======================
IP Address: $sandboxIP
Username: WDAGUtilityAccount
Password: $password
SSH Command: ssh WDAGUtilityAccount@$sandboxIP
"@ | Out-File -FilePath $infoFile -Encoding utf8
  Write-Verbose "SSH info written to $infoFile"

  # 3) Create/normalize the service first so NT SERVICE\sshd exists
  Ensure-Service -Name 'sshd' -BinPath $sshdPath -StartType 'auto'

  # 4) Ensure sshd_config (copy default if shipped)
  $etcConfigSource = Join-Path $OpenSSHPath 'sshd_config_default'
  $etcConfigDest   = Join-Path $env:PROGRAMDATA 'ssh\sshd_config'
  if (-not (Test-Path $etcConfigDest)) {
    if (Test-Path $etcConfigSource) {
      $null = New-Item -ItemType Directory -Path (Split-Path $etcConfigDest) -Force
      Copy-Item $etcConfigSource $etcConfigDest -Force
      Write-Verbose "Copied default sshd_config to $etcConfigDest"
    }
  }

  # 5) Host keys
  Ensure-HostKeys

  # 6) ACL fixes
  Try-RunAclFixes

  # 7) Firewall
  if (-not $NoFirewall) {
    Ensure-FirewallRule -Name "OpenSSH-Server-In-TCP-$Port" -Port $Port
  } else {
    Write-Verbose "Skipping firewall step per -NoFirewall."
  }

  # 8) Start service
  if (-not $NoStart) {
    if ($Port -ne 22 -and (Test-Path $etcConfigDest)) {
      (Get-Content $etcConfigDest) |
        ForEach-Object {
          if ($_ -match '^\s*#?\s*Port\s+\d+') { "Port $Port" } else { $_ }
        } |
        Set-Content $etcConfigDest -Encoding ascii
      Write-Verbose "Set Port $Port in $etcConfigDest"
    }
    Start-Service -Name sshd -ErrorAction Stop
    Set-Service -Name sshd -StartupType Automatic
    Write-Host "sshd service is running."
  } else {
    Write-Verbose "Skipping service start per -NoStart."
  }

  # 9) Status
  sc.exe qc sshd | Out-Host
  Get-Service sshd | Format-Table -Auto | Out-Host

  Write-Host "`nPortable OpenSSH prepared successfully." -ForegroundColor Green
  Write-Host "Path: $OpenSSHPath"
  Write-Host "Port: $Port"
  Write-Host "Sandbox IP: $sandboxIP" -ForegroundColor Cyan
  Write-Host "SSH User: WDAGUtilityAccount" -ForegroundColor Yellow
  Write-Host "SSH Password: $password" -ForegroundColor Yellow
  Write-Host "SSH Command: ssh WDAGUtilityAccount@$sandboxIP" -ForegroundColor Green
  if (-not $NoFirewall) { Write-Host "Firewall rule: OpenSSH-Server-In-TCP-$Port" }
}
catch {
  Write-Error $_.Exception.Message
  if ($_.ScriptStackTrace) { Write-Verbose $_.ScriptStackTrace }
  exit 1
}
