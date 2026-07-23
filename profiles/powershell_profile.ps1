# ──────────────────────────────────────────────────────────────────────────
#  Universal-Ops Operations Platform — PowerShell Diagnostic Profile
# ──────────────────────────────────────────────────────────────────────────
#  Sourced by the Universal-Ops Go backend for the PowerShell Pro tab.
#  Contains utility functions for high-density system diagnostics.
#
#  Naming Convention: Ops-<Noun>
#  Example: Get-OpsOSInfo
#
#  All functions output formatted text suitable for display in the Universal-Ops
#  terminal interface using consistent headers and result pairs.
# ──────────────────────────────────────────────────────────────────────────

# Author: Universal-Ops Operations Platform
# Copyright: (c) Universal-Ops. All rights reserved.

# ── Formatting Helpers ──────────────────────────────────────────────────

function Write-OpsHeader {
    param([string]$Text)
    Write-Output "`n── $Text $(('─' * (70 - $Text.Length)))"
}

function Write-OpsSubHeader {
    param([string]$Text)
    Write-Output "`n• $Text"
}

function Write-OpsResult {
    param([string]$Label, [string]$Value)
    Write-Output "  $($Label.PadRight(20)) > $Value"
}

function Write-OpsError {
    param([string]$Message)
    Write-Output "  ! ERROR: $Message"
}

function Write-OpsTable {
    param($Data, [string[]]$Columns)
    $Data | Select-Object $Columns | Format-Table -AutoSize | Out-String | Write-Output
}

if ($IsLinux -or $IsMacOS) {
    Write-OpsError "This command requires Windows"
    exit
}

# ── Diagnostic Functions ────────────────────────────────────────────────

function Get-OpsOSInfo {
    try {
        $os = Get-CimInstance Win32_OperatingSystem
        $up = (Get-Date) - $os.LastBootUpTime

        Write-OpsSubHeader "Operating System"
        Write-OpsResult "OS Name" $os.Caption
        Write-OpsResult "Version" $os.Version
        Write-OpsResult "Build Number" $os.BuildNumber
        Write-OpsResult "Architecture" $os.OSArchitecture
        Write-OpsResult "Install Date" $os.InstallDate
        Write-OpsResult "Last Boot" $os.LastBootUpTime
        Write-OpsResult "Manufacturer" $os.Manufacturer
        Write-OpsResult "Registered User" $os.RegisteredUser
        Write-OpsResult "Serial Number" $os.SerialNumber
        Write-OpsResult "System Drive" $os.SystemDrive
        Write-OpsResult "Windows Directory" $os.WindowsDirectory
        Write-OpsResult "Uptime" "$($up.Days)d $($up.Hours)h $($up.Minutes)m $($up.Seconds)s"
    } catch {
        Write-OpsError $_.Exception.Message
    }
}

function Get-OpsCPUInfo {
    try {
        $p = Get-CimInstance Win32_Processor
        Write-OpsSubHeader "Processor"
        Write-OpsResult "Name" $p.Name
        Write-OpsResult "Manufacturer" $p.Manufacturer
        Write-OpsResult "Cores" $p.NumberOfCores
        Write-OpsResult "Logical Processors" $p.NumberOfLogicalProcessors
        Write-OpsResult "Max Clock Speed" "$($p.MaxClockSpeed) MHz"
        Write-OpsResult "L2 Cache Size" "$($p.L2CacheSize) KB"
        Write-OpsResult "L3 Cache Size" "$($p.L3CacheSize) KB"
        Write-OpsResult "Architecture" $p.Architecture
        Write-OpsResult "Socket" $p.SocketDesignation
    } catch {
        Write-OpsError $_.Exception.Message
    }
}

function Get-OpsMemoryInfo {
    try {
        $total = (Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory
        $totalGB = [math]::Round($total / 1GB, 2)
        $mems = Get-CimInstance Win32_PhysicalMemory
        $slots = $mems.Count

        Write-OpsSubHeader "Memory"
        Write-OpsResult "Total Memory" "$totalGB GB"
        Write-OpsResult "Installed Modules" $slots

        $os = Get-CimInstance Win32_OperatingSystem
        $used = [math]::Round(($os.TotalVisibleMemorySize - $os.FreePhysicalMemory) / 1MB, 2)
        $avail = [math]::Round($os.FreePhysicalMemory / 1MB, 2)
        $pct = [math]::Round(($used / ($os.TotalVisibleMemorySize / 1MB)) * 100, 1)

        Write-OpsResult "Memory In Use" "$used GB"
        Write-OpsResult "Memory Available" "$avail GB"
        Write-OpsResult "Usage" "$pct%"

        foreach ($m in $mems) {
            $gb = [math]::Round($m.Capacity / 1GB, 0)
            Write-OpsResult "  Slot: $($m.BankLabel)" "$gb GB $($m.Speed)MHz $($m.MemoryType)"
        }
    } catch {
        Write-OpsError $_.Exception.Message
    }
}

function Get-OpsDiskInfo {
    try {
        $disks = Get-CimInstance Win32_DiskDrive
        Write-OpsSubHeader "Physical Disks"
        foreach ($d in $disks) {
            $size = [math]::Round($d.Size / 1GB, 0)
            Write-OpsResult "$($d.Model)" "$size GB  ($($d.MediaType))"
            Write-OpsResult "  Interface" $d.InterfaceType
            Write-OpsResult "  Serial" $d.SerialNumber
            Write-OpsResult "  Partitions" $d.Partitions
        }
    } catch {
        Write-OpsError $_.Exception.Message
    }
}

function Get-OpsVolumeInfo {
    try {
        $vols = Get-CimInstance Win32_Volume | Where-Object { $_.DriveLetter -ne $null }
        Write-OpsSubHeader "Volumes (Disk Usage)"
        foreach ($v in $vols) {
            $cap = [math]::Round($v.Capacity / 1GB, 2)
            $free = [math]::Round($v.FreeSpace / 1GB, 2)
            $used = $cap - $free
            $pct = [math]::Round(($used / $cap) * 100, 1)
            Write-OpsResult "$($v.DriveLetter) ($($v.FileSystem))" "$used GB / $cap GB ($pct% used)"
        }
    } catch {
        Write-OpsError $_.Exception.Message
    }
}

function Get-OpsBiosInfo {
    try {
        $bios = Get-CimInstance Win32_BIOS
        Write-OpsSubHeader "BIOS / Firmware"
        Write-OpsResult "Manufacturer" $bios.Manufacturer
        Write-OpsResult "BIOS Version" $bios.SMBIOSBIOSVersion
        Write-OpsResult "Release Date" $bios.ReleaseDate
        Write-OpsResult "SMBIOS Version" "$($bios.SMBIOSMajorVersion).$($bios.SMBIOSMinorVersion)"
        Write-OpsResult "Serial Number" $bios.SerialNumber
    } catch {
        Write-OpsError $_.Exception.Message
    }
}

function Get-OpsGPUInfo {
    try {
        $gpus = Get-CimInstance Win32_VideoController
        Write-OpsSubHeader "Graphics"
        foreach ($g in $gpus) {
            Write-OpsResult "Adapter" $g.Name
            Write-OpsResult "  Driver Version" $g.DriverVersion
            Write-OpsResult "  Driver Date" $g.DriverDate
            Write-OpsResult "  Resolution" "$($g.CurrentHorizontalResolution)x$($g.CurrentVerticalResolution)"
            Write-OpsResult "  Refresh Rate" "$($g.CurrentRefreshRate) Hz"
            Write-OpsResult "  Adapter RAM" "$([math]::Round($g.AdapterRAM / 1MB, 0)) MB"
            Write-OpsResult "  Status" $g.Status
        }
    } catch {
        Write-OpsError $_.Exception.Message
    }
}

function Get-OpsSystemUptime {
    try {
        $os = Get-CimInstance Win32_OperatingSystem
        $up = (Get-Date) - $os.LastBootUpTime
        Write-OpsResult "System Uptime" "$($up.Days) days, $($up.Hours) hours, $($up.Minutes) minutes"
        Write-OpsResult "Boot Time" $os.LastBootUpTime
    } catch {
        Write-OpsError $_.Exception.Message
    }
}

function Get-OpsEnvironmentInfo {
    try {
        $env = Get-ChildItem Env: | Sort-Object Name
        Write-OpsSubHeader "Environment"
        foreach ($e in $env) {
            $k = $e.Name
            $val = $e.Value
            if ($val.Length -gt 50) { $val = $val.Substring(0, 47) + "..." }
            Write-OpsResult $k $val
        }
    } catch {
        Write-OpsError $_.Exception.Message
    }
}

function Get-OpsSystemInfo {
    Write-OpsHeader "SYSTEM INFORMATION AUDIT"
    Get-OpsOSInfo
    Get-OpsCPUInfo
    Get-OpsMemoryInfo
    Get-OpsDiskInfo
    Get-OpsVolumeInfo
    Get-OpsBiosInfo
    Get-OpsGPUInfo
    Get-OpsSystemUptime
}

function Get-OpsComputerSystem {
    try {
        $cs = Get-CimInstance Win32_ComputerSystem
        Write-OpsSubHeader "Computer System"
        Write-OpsResult "Manufacturer" $cs.Manufacturer
        Write-OpsResult "Model" $cs.Model
        Write-OpsResult "Domain" $cs.Domain
        Write-OpsResult "Domain Role" $cs.DomainRole
        Write-OpsResult "Workgroup" $cs.Workgroup
        Write-OpsResult "Total Physical RAM" "$([math]::Round($cs.TotalPhysicalMemory / 1GB, 2)) GB"
        Write-OpsResult "Current User" $cs.UserName
        Write-OpsResult "Part of Domain" $cs.PartOfDomain
        Write-OpsResult "Hyper-V Present" $cs.HypervisorPresent
    } catch {
        Write-OpsError $_.Exception.Message
    }
}

function Get-OpsNetworkAdaptersDetailed {
    try {
        $adapters = Get-CimInstance Win32_NetworkAdapterConfiguration | Where-Object { $_.IPEnabled -eq $true }
        Write-OpsSubHeader "Network Adapter Details"
        foreach ($a in $adapters) {
            Write-OpsResult $a.Description ""
            Write-OpsResult "  MAC" $a.MACAddress
            Write-OpsResult "  IP" ($a.IPAddress -join ', ')
            Write-OpsResult "  Subnet" ($a.IPSubnet -join ', ')
            Write-OpsResult "  Gateway" ($a.DefaultIPGateway -join ', ')
            Write-OpsResult "  DHCP" $a.DHCPEnabled
            Write-OpsResult "  DNS Servers" ($a.DNSServerSearchOrder -join ', ')
            Write-OpsResult "  DNS Domain" $a.DNSDomain
        }
    } catch {
        Write-OpsError $_.Exception.Message
    }
}

function Get-OpsTopProcesses {
    param([int]$Count = 10)
    try {
        $procs = Get-Process | Sort-Object CPU -Descending | Select-Object -First $Count
        Write-OpsSubHeader "Top $Count Processes by CPU"
        foreach ($p in $procs) {
            Write-OpsResult $p.ProcessName "$([math]::Round($p.CPU, 1))s"
        }

        $procsMem = Get-Process | Sort-Object WorkingSet -Descending | Select-Object -First $Count
        Write-OpsSubHeader "Top $Count Processes by Memory"
        foreach ($p in $procsMem) {
            Write-OpsResult $p.ProcessName "$([math]::Round($p.WorkingSet64 / 1MB, 1)) MB"
        }
    } catch {
        Write-OpsError $_.Exception.Message
    }
}

function Get-OpsProcessDetails {
    param([string]$Name)
    try {
        $proc = Get-Process -Name $Name -ErrorAction Stop
        Write-OpsSubHeader "Process: $Name"
        Write-OpsResult "PID" $proc.Id
        Write-OpsResult "Name" $proc.ProcessName
        Write-OpsResult "CPU (s)" $([math]::Round($proc.CPU, 2))
        Write-OpsResult "Working Set" "$([math]::Round($proc.WorkingSet64 / 1MB, 2)) MB"
        Write-OpsResult "Private Memory" "$([math]::Round($proc.PrivateMemorySize64 / 1MB, 2)) MB"
        Write-OpsResult "Virtual Memory" "$([math]::Round($proc.VirtualMemorySize64 / 1MB, 2)) MB"
        Write-OpsResult "Handles" $proc.HandleCount
        Write-OpsResult "Threads" $proc.Threads.Count
        Write-OpsResult "Priority" $proc.PriorityClass
        Write-OpsResult "Start Time" $proc.StartTime
        Write-OpsResult "Responding" $proc.Responding
        Write-OpsResult "Path" $proc.Path
    } catch {
        Write-OpsError "Process '$Name' not found: $_"
    }
}

function Get-OpsPerformanceCounters {
    try {
        $counters = Get-Counter -Counter "\Processor(_Total)\% Processor Time", "\Memory\Available MBytes"
        Write-OpsSubHeader "Performance Counters (Snapshot)"
        foreach ($c in $counters.CounterSamples) {
            $val = [math]::Round($c.CookedValue, 2)
            Write-OpsResult $c.Path $val
        }

        $cpu = Get-CimInstance Win32_Processor
        Write-OpsResult "CPU Load" "$($cpu.LoadPercentage)%"
    } catch {
        Write-OpsError $_.Exception.Message
    }
}

# ── High-Level Workflows ──────────────────────────────────────────────

function Invoke-OpsDailyOps {
    Write-OpsHeader "DAILY OPERATIONS SNAPSHOT"
    Get-OpsOSInfo
    Get-OpsVolumeInfo
    Get-OpsPerformanceCounters
    Get-OpsTopProcesses -Count 5
}

function Invoke-OpsSystemReview {
    Get-OpsSystemInfo
}

function Invoke-OpsSecurityAudit {
    Write-OpsHeader "SECURITY SURFACE AUDIT"
    Write-OpsSubHeader "Users & Groups"
    net localgroup administrators | Out-String | Write-Output

    Write-OpsSubHeader "Antivirus Status"
    if (Get-Command Get-MpComputerStatus -ErrorAction SilentlyContinue) {
        $status = Get-MpComputerStatus
        Write-OpsResult "Real-Time Protection" $status.RealTimeProtectionEnabled
        Write-OpsResult "Spyware Protection" $status.AntispywareEnabled
        Write-OpsResult "Last Scan" $status.FullScanAge
    } else {
        Write-OpsResult "Defender" "Not accessible"
    }
}

function Invoke-OpsNetworkDiagnostics {
    Write-OpsHeader "NETWORK DIAGNOSTICS"
    Get-OpsNetworkAdaptersDetailed
    Write-OpsSubHeader "Active Connections (TCP)"
    netstat -ano | Select-Object -First 20 | Out-String | Write-Output
}

function Invoke-OpsThreatHunt {
    Write-OpsHeader "THREAT HUNTING PRIMER"
    Write-OpsSubHeader "Suspicious Network Sockets"
    netstat -an | Select-String "LISTENING" | Select-Object -First 10 | Out-String | Write-Output

    Write-OpsSubHeader "Recently Modified Executables (C:\Windows\System32)"
    Get-ChildItem C:\Windows\System32 -Filter *.exe | Sort-Object LastWriteTime -Descending | Select-Object -First 10 -Property Name, LastWriteTime | Format-Table -AutoSize | Out-String | Write-Output
}

function Invoke-OpsChangeAudit {
    Write-OpsHeader "SYSTEM CHANGE AUDIT"
    Write-OpsSubHeader "Recently Installed Programs"
    Get-ItemProperty HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\* | Sort-Object InstallDate -Descending | Select-Object -First 10 -Property DisplayName, InstallDate | Format-Table -AutoSize | Out-String | Write-Output
}

function Invoke-OpsComplianceCheck {
    Write-OpsHeader "COMPLIANCE CHECK"
    Write-OpsResult "Password Policy" "Auditing..."
    net accounts | Out-String | Write-Output
}
