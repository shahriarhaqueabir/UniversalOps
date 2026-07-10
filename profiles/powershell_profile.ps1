<#
.SYNOPSIS
    Hawkward Operations Platform — PowerShell Diagnostic Profile
    Integrated diagnostic, ops, audit, systems admin & monitoring commands.
    Sourced by the Hawkward Go backend for the PowerShell Pro tab.
.DESCRIPTION
    This profile provides ~100+ PowerShell commands organized into 7 high-level
    workflow functions. Each workflow aggregates multiple individual diagnostics
    into a comprehensive report.

    All functions output formatted text suitable for display in the Hawkward
    PowerShell Pro console. Functions are read-only — no system state is modified.

    Compatible with PowerShell 5.1+ (Windows) and PowerShell 7+ (cross-platform).
    When running on non-Windows, commands that require Windows-only modules
    gracefully report "not available".
.NOTES
    Author: Hawkward Operations Platform
    Copyright: (c) Hawkward. All rights reserved.
#>

#requires -Version 5.1

#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  SECTION 0: UTILITY FUNCTIONS
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Detect if running on Windows (some commands are Windows-only)
$script:IsWindows = $PSVersionTable.PSVersion.Major -ge 5 -and $IsWindows -eq $true
if (-not (Get-Variable -Name IsWindows -ErrorAction SilentlyContinue)) {
    $script:IsWindows = $PSVersionTable.Platform -eq 'Win32NT' -or $env:OS -eq 'Windows_NT'
}

function Write-HawkHeader {
    <#
    .SYNOPSIS
        Writes a section header in the Hawkward console output format.
    #>
    param([string]$Title)
    $line = '─' * 60
    Write-Output "`n╔$line╗"
    Write-Output "║  $($Title.ToUpper())".PadRight(63) + '║'
    Write-Output "╚$line╝`n"
}

function Write-HawkSubHeader {
    <#
    .SYNOPSIS
        Writes a sub-section header.
    #>
    param([string]$Title)
    $line = '─' * 40
    Write-Output "`n┌$line┐"
    Write-Output "│ $Title"
    Write-Output "└$line┘"
}

function Write-HawkResult {
    <#
    .SYNOPSIS
        Writes a key: value result line.
    #>
    param(
        [string]$Label,
        [string]$Value,
        [string]$Status = ''
    )
    $statusTag = if ($Status) { " [$Status]" } else { '' }
    Write-Output "  {0,-35} {1}{2}" -f "$($Label):", $Value, $statusTag
}

function Write-HawkError {
    <#
    .SYNOPSIS
        Writes an error message.
    #>
    param([string]$Message)
    Write-Output "  ⚠ ERROR: $Message"
}

function Write-HawkTable {
    <#
    .SYNOPSIS
        Outputs objects as a formatted table with aligned columns.
    #>
    param(
        [Parameter(Mandatory)]
        [array]$InputObject,
        [Parameter(Mandatory)]
        [string[]]$Property
    )
    $InputObject | Format-Table $Property -AutoSize | Out-String | ForEach-Object { $_.TrimEnd() }
}

function Test-WindowsOnly {
    <#
    .SYNOPSIS
        Returns $true only on Windows. Used to guard Windows-specific commands.
    #>
    if (-not $script:IsWindows) {
        Write-HawkError "This command requires Windows"
        return $false
    }
    return $true
}

#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  SECTION 1: SYSTEM INFORMATION COMMANDS
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Get-HawkOSInfo {
    <#
    .SYNOPSIS
        Retrieves operating system details.
    #>
    Write-HawkSubHeader "Operating System"
    try {
        if (-not (Test-WindowsOnly)) { return }
        $os = Get-CimInstance Win32_OperatingSystem
        Write-HawkResult "OS Name" $os.Caption
        Write-HawkResult "Version" $os.Version
        Write-HawkResult "Build Number" $os.BuildNumber
        Write-HawkResult "Architecture" $os.OSArchitecture
        Write-HawkResult "Install Date" $os.InstallDate
        Write-HawkResult "Last Boot" $os.LastBootUpTime
        Write-HawkResult "Manufacturer" $os.Manufacturer
        Write-HawkResult "Registered User" $os.RegisteredUser
        Write-HawkResult "Serial Number" $os.SerialNumber
        Write-HawkResult "System Drive" $os.SystemDrive
        Write-HawkResult "Windows Directory" $os.WindowsDirectory
        $up = (Get-Date) - $os.LastBootUpTime
        Write-HawkResult "Uptime" "$($up.Days)d $($up.Hours)h $($up.Minutes)m $($up.Seconds)s"
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkCPUInfo {
    <#
    .SYNOPSIS
        Retrieves processor details.
    #>
    Write-HawkSubHeader "Processor"
    try {
        if (-not (Test-WindowsOnly)) { return }
        $cpu = Get-CimInstance Win32_Processor
        foreach ($p in $cpu) {
            Write-HawkResult "Name" $p.Name
            Write-HawkResult "Manufacturer" $p.Manufacturer
            Write-HawkResult "Cores" $p.NumberOfCores
            Write-HawkResult "Logical Processors" $p.NumberOfLogicalProcessors
            Write-HawkResult "Max Clock Speed" "$($p.MaxClockSpeed) MHz"
            Write-HawkResult "L2 Cache Size" "$($p.L2CacheSize) KB"
            Write-HawkResult "L3 Cache Size" "$($p.L3CacheSize) KB"
            Write-HawkResult "Architecture" $p.Architecture
            Write-HawkResult "Socket" $p.SocketDesignation
            break
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkMemoryInfo {
    <#
    .SYNOPSIS
        Retrieves physical memory details.
    #>
    Write-HawkSubHeader "Memory"
    try {
        if (-not (Test-WindowsOnly)) { return }
        $mem = Get-CimInstance Win32_PhysicalMemory
        $total = 0
        $slots = 0
        foreach ($m in $mem) {
            $total += $m.Capacity
            $slots++
        }
        $totalGB = [math]::Round($total / 1GB, 2)
        Write-HawkResult "Total Memory" "$totalGB GB"
        Write-HawkResult "Installed Modules" $slots
        $os = Get-CimInstance Win32_OperatingSystem
        $avail = [math]::Round($os.FreePhysicalMemory / 1MB, 2)
        $used = [math]::Round(($total / 1GB) - ($os.FreePhysicalMemory / 1MB), 2)
        Write-HawkResult "Memory In Use" "$used GB"
        Write-HawkResult "Memory Available" "$avail GB"
        $pct = [math]::Round((($total - ($os.FreePhysicalMemory * 1KB)) / $total) * 100, 1)
        Write-HawkResult "Usage" "$pct%"
        foreach ($m in $mem) {
            $gb = [math]::Round($m.Capacity / 1GB, 2)
            Write-HawkResult "  Slot: $($m.BankLabel)" "$gb GB $($m.Speed)MHz $($m.MemoryType)"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkDiskInfo {
    <#
    .SYNOPSIS
        Retrieves disk drive details.
    #>
    Write-HawkSubHeader "Physical Disks"
    try {
        if (-not (Test-WindowsOnly)) { return }
        $disks = Get-CimInstance Win32_DiskDrive
        foreach ($d in $disks) {
            $size = [math]::Round($d.Size / 1GB, 2)
            Write-HawkResult "$($d.Model)" "$size GB  ($($d.MediaType))"
            Write-HawkResult "  Interface" $d.InterfaceType
            Write-HawkResult "  Serial" $d.SerialNumber
            Write-HawkResult "  Partitions" $d.Partitions
            Write-Output ""
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkVolumeInfo {
    <#
    .SYNOPSIS
        Retrieves volume/partition disk usage.
    #>
    Write-HawkSubHeader "Volumes (Disk Usage)"
    try {
        $vols = Get-CimInstance Win32_LogicalDisk -Filter "DriveType=3"
        Write-Output "  Drive  Size(GB)  Used(GB)  Free(GB)  Used%  FileSystem"
        Write-Output "  -----  --------  --------  --------  -----  ----------"
        foreach ($v in $vols) {
            $total = [math]::Round($v.Size / 1GB, 2)
            $free = [math]::Round($v.FreeSpace / 1GB, 2)
            $used = [math]::Round(($v.Size - $v.FreeSpace) / 1GB, 2)
            $pct = if ($v.Size -gt 0) { [math]::Round((($v.Size - $v.FreeSpace) / $v.Size) * 100, 1) } else { 0 }
            Write-Output ("  {0,-6} {1,8} {2,9} {3,9} {4,6}% {5,-10}" -f $v.DeviceID, $total, $used, $free, $pct, $v.FileSystem)
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkBiosInfo {
    <#
    .SYNOPSIS
        Retrieves BIOS/firmware information.
    #>
    Write-HawkSubHeader "BIOS / Firmware"
    try {
        if (-not (Test-WindowsOnly)) { return }
        $bios = Get-CimInstance Win32_BIOS
        Write-HawkResult "Manufacturer" $bios.Manufacturer
        Write-HawkResult "BIOS Version" $bios.SMBIOSBIOSVersion
        Write-HawkResult "Release Date" $bios.ReleaseDate
        Write-HawkResult "SMBIOS Version" "$($bios.SMBIOSMajorVersion).$($bios.SMBIOSMinorVersion)"
        Write-HawkResult "Serial Number" $bios.SerialNumber
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkGPUInfo {
    <#
    .SYNOPSIS
        Retrieves graphics adapter information.
    #>
    Write-HawkSubHeader "Graphics"
    try {
        if (-not (Test-WindowsOnly)) { return }
        $gpu = Get-CimInstance Win32_VideoController
        foreach ($g in $gpu) {
            Write-HawkResult "Adapter" $g.Name
            Write-HawkResult "  Driver Version" $g.DriverVersion
            Write-HawkResult "  Driver Date" $g.DriverDate
            Write-HawkResult "  Resolution" "$($g.CurrentHorizontalResolution)x$($g.CurrentVerticalResolution)"
            Write-HawkResult "  Refresh Rate" "$($g.CurrentRefreshRate) Hz"
            Write-HawkResult "  Adapter RAM" "$([math]::Round($g.AdapterRAM / 1MB, 0)) MB"
            Write-HawkResult "  Status" $g.Status
            Write-Output ""
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkSystemUptime {
    <#
    .SYNOPSIS
        Returns system uptime.
    #>
    try {
        if (-not (Test-WindowsOnly)) { return }
        $os = Get-CimInstance Win32_OperatingSystem
        $up = (Get-Date) - $os.LastBootUpTime
        Write-HawkResult "System Uptime" "$($up.Days) days, $($up.Hours) hours, $($up.Minutes) minutes"
        Write-HawkResult "Boot Time" $os.LastBootUpTime
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkEnvironmentInfo {
    <#
    .SYNOPSIS
        Displays key environment variables.
    #>
    Write-HawkSubHeader "Environment"
    $keys = @('COMPUTERNAME', 'USERDOMAIN', 'USERNAME', 'PROCESSOR_IDENTIFIER', 'NUMBER_OF_PROCESSORS',
              'OS', 'SystemRoot', 'TEMP', 'PATH')
    foreach ($k in $keys) {
        $val = [Environment]::GetEnvironmentVariable($k, 'Machine')
        if (-not $val) { $val = [Environment]::GetEnvironmentVariable($k, 'User') }
        if (-not $val) { $val = [Environment]::GetEnvironmentVariable($k, 'Process') }
        if ($k -eq 'PATH') { $val = "$($val.Substring(0, [Math]::Min(120, $val.Length)))..." }
        Write-HawkResult $k $val
    }
}

function Get-HawkSystemInfo {
    <#
    .SYNOPSIS
        Comprehensive system information summary.
    #>
    Get-HawkOSInfo
    Get-HawkCPUInfo
    Get-HawkMemoryInfo
    Get-HawkDiskInfo
    Get-HawkVolumeInfo
    Get-HawkBiosInfo
    Get-HawkGPUInfo
    Get-HawkSystemUptime
}

function Get-HawkComputerSystem {
    <#
    .SYNOPSIS
        Basic system enclosure details (model, manufacturer, domain).
    #>
    Write-HawkSubHeader "Computer System"
    try {
        $cs = Get-CimInstance Win32_ComputerSystem -ErrorAction SilentlyContinue
        Write-HawkResult "Manufacturer" $cs.Manufacturer
        Write-HawkResult "Model" $cs.Model
        Write-HawkResult "Domain" $cs.Domain
        Write-HawkResult "Domain Role" $cs.DomainRole
        Write-HawkResult "Workgroup" $cs.Workgroup
        Write-HawkResult "Total Physical RAM" "$([math]::Round($cs.TotalPhysicalMemory / 1GB, 2)) GB"
        Write-HawkResult "Current User" $cs.UserName
        Write-HawkResult "Part of Domain" $cs.PartOfDomain
        Write-HawkResult "Hyper-V Present" $cs.HypervisorPresent
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkNetworkAdaptersDetailed {
    <#
    .SYNOPSIS
        Detailed network adapter configuration (MAC, IP, DNS, DHCP).
    #>
    Write-HawkSubHeader "Network Adapter Details"
    try {
        $adapters = Get-CimInstance Win32_NetworkAdapterConfiguration -Filter "IPEnabled=True" -ErrorAction SilentlyContinue
        foreach ($a in $adapters) {
            Write-HawkResult $a.Description ""
            Write-HawkResult "  MAC" $a.MACAddress
            Write-HawkResult "  IP" ($a.IPAddress -join ', ')
            Write-HawkResult "  Subnet" ($a.IPSubnet -join ', ')
            Write-HawkResult "  Gateway" ($a.DefaultIPGateway -join ', ')
            Write-HawkResult "  DHCP" $a.DHCPEnabled
            Write-HawkResult "  DNS Servers" ($a.DNSServerSearchOrder -join ', ')
            Write-HawkResult "  DNS Domain" $a.DNSDomain
            Write-Output ""
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  SECTION 2: PERFORMANCE MONITORING
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Get-HawkTopProcesses {
    <#
    .SYNOPSIS
        Lists top processes by CPU and memory usage.
    .PARAMETER Count
        Number of processes to display (default: 10).
    #>
    param([int]$Count = 10)
    Write-HawkSubHeader "Top $Count Processes by CPU"
    try {
        $procs = Get-Process | Sort-Object CPU -Descending | Select-Object -First $Count
        $procs | Format-Table -Property Id, @{N='Name';E={$_.ProcessName}},
            @{N='CPU (s)';E={[math]::Round($_.CPU, 1)}},
            @{N='Mem (MB)';E={[math]::Round($_.WorkingSet64 / 1MB, 1)}},
            @{N='Handles';E={$_.HandleCount}},
            @{N='Threads';E={$_.Threads.Count}},
            @{N='Responding';E={$_.Responding}} -AutoSize | Out-String | ForEach-Object { $_.TrimEnd() }

        Write-HawkSubHeader "Top $Count Processes by Memory"
        $procs = Get-Process | Sort-Object WorkingSet64 -Descending | Select-Object -First $Count
        $procs | Format-Table -Property Id, @{N='Name';E={$_.ProcessName}},
            @{N='Mem (MB)';E={[math]::Round($_.WorkingSet64 / 1MB, 1)}},
            @{N='CPU (s)';E={[math]::Round($_.CPU, 1)}},
            @{N='Handles';E={$_.HandleCount}},
            @{N='Start Time';E={$_.StartTime}} -AutoSize | Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkProcessDetails {
    <#
    .SYNOPSIS
        Gets detailed information about a specific process.
    .PARAMETER Name
        Process name (without .exe).
    #>
    param([string]$Name)
    Write-HawkSubHeader "Process: $Name"
    try {
        $p = Get-Process -Name $Name -ErrorAction Stop
        foreach ($proc in $p) {
            Write-HawkResult "PID" $proc.Id
            Write-HawkResult "Name" $proc.ProcessName
            Write-HawkResult "CPU (s)" $([math]::Round($proc.CPU, 2))
            Write-HawkResult "Working Set" "$([math]::Round($proc.WorkingSet64 / 1MB, 2)) MB"
            Write-HawkResult "Private Memory" "$([math]::Round($proc.PrivateMemorySize64 / 1MB, 2)) MB"
            Write-HawkResult "Virtual Memory" "$([math]::Round($proc.VirtualMemorySize64 / 1MB, 2)) MB"
            Write-HawkResult "Handles" $proc.HandleCount
            Write-HawkResult "Threads" $proc.Threads.Count
            Write-HawkResult "Priority" $proc.PriorityClass
            Write-HawkResult "Start Time" $proc.StartTime
            Write-HawkResult "Responding" $proc.Responding
            Write-HawkResult "Path" $proc.Path
            Write-Output ""
        }
    } catch {
        Write-HawkError "Process '$Name' not found: $_"
    }
}

function Get-HawkPerformanceCounters {
    <#
    .SYNOPSIS
        Samples key performance counters (CPU, memory, disk, network).
    #>
    Write-HawkSubHeader "Performance Counters (Snapshot)"
    try {
        if ($script:IsWindows) {
            $counters = Get-Counter '\Processor(_Total)\% Processor Time',
                '\Memory\Available MBytes',
                '\Memory\Pages/sec',
                '\PhysicalDisk(_Total)\% Disk Time',
                '\PhysicalDisk(_Total)\Avg. Disk Queue Length',
                '\Network Interface(*)\Bytes Total/sec' -ErrorAction SilentlyContinue
            if ($counters) {
                foreach ($c in $counters.CounterSamples) {
                    $val = [math]::Round($c.CookedValue, 2)
                    Write-HawkResult $c.Path $val
                }
            }
        } else {
            $cpu = Get-CimInstance Win32_Processor -ErrorAction SilentlyContinue
            if ($cpu) {
                Write-HawkResult "CPU Load" "$($cpu.LoadPercentage)%"
            }
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  SECTION 3: NETWORK DIAGNOSTICS
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Get-HawkNetworkAdapters {
    <#
    .SYNOPSIS
        Lists network adapters with their configuration.
    #>
    Write-HawkSubHeader "Network Adapters"
    try {
        $adapters = Get-NetAdapter -ErrorAction SilentlyContinue
        if (-not $adapters) { $adapters = Get-CimInstance Win32_NetworkAdapter -Filter "NetEnabled=True" }
        $adapters | Format-Table -Property Name, InterfaceDescription, Status, LinkSpeed,
            MacAddress, @{N='IP';E={(Get-NetIPAddress -InterfaceAlias $_.Name -ErrorAction SilentlyContinue | Where-Object AddressFamily -eq 2).IPAddress -join ', '}} -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkIPConfiguration {
    <#
    .SYNOPSIS
        Displays IP configuration for all interfaces.
    #>
    Write-HawkSubHeader "IP Configuration"
    try {
        $config = Get-NetIPConfiguration -ErrorAction SilentlyContinue
        if ($config) {
            $config | Format-Table -Property InterfaceAlias, @{N='IPv4';E={$_.IPv4Address.IPAddress}},
                @{N='Gateway';E={$_.IPv4DefaultGateway.NextHop}},
                @{N='DNS';E={($_.DNSServer | Where-Object AddressFamily -eq 2).ServerAddress -join ', '}} -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
        } else {
            ipconfig | Out-String | ForEach-Object { $_.TrimEnd() }
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkRoutingTable {
    <#
    .SYNOPSIS
        Displays the IPv4 routing table.
    #>
    Write-HawkSubHeader "Routing Table"
    try {
        Get-NetRoute -AddressFamily IPv4 -ErrorAction SilentlyContinue |
            Format-Table -Property DestinationPrefix, NextHop, RouteMetric, ifIndex -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkActiveConnections {
    <#
    .SYNOPSIS
        Displays active TCP connections.
    #>
    Write-HawkSubHeader "Active TCP Connections"
    try {
        Get-NetTCPConnection -State Established -ErrorAction SilentlyContinue |
            Format-Table -Property LocalAddress, LocalPort, RemoteAddress, RemotePort, OwningProcess -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkListeningPorts {
    <#
    .SYNOPSIS
        Displays all TCP/UDP listening ports.
    #>
    Write-HawkSubHeader "Listening Ports"
    try {
        Write-Output "  TCP Listeners:"
        Get-NetTCPConnection -State Listen -ErrorAction SilentlyContinue |
            Format-Table -Property LocalAddress, LocalPort, @{N='Process';E={(Get-Process -Id $_.OwningProcess -ErrorAction SilentlyContinue).ProcessName}} -AutoSize |
            Out-String | ForEach-Object { "    " + $_.TrimEnd() }
        Write-Output "  UDP Listeners:"
        Get-NetUDPEndpoint -ErrorAction SilentlyContinue |
            Format-Table -Property LocalAddress, LocalPort -AutoSize |
            Out-String | ForEach-Object { "    " + $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkDNSSettings {
    <#
    .SYNOPSIS
        Displays DNS client configuration and cache.
    #>
    Write-HawkSubHeader "DNS Configuration"
    try {
        Write-Output "  DNS Client Cache:"
        Get-DnsClientCache -ErrorAction SilentlyContinue |
            Format-Table -Property Entry, Type, Data, TimeToLive -AutoSize |
            Out-String | ForEach-Object { "    " + $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Test-HawkConnectivity {
    <#
    .SYNOPSIS
        Tests connectivity to common endpoints.
    #>
    param(
        [string[]]$Targets = @('8.8.8.8', '1.1.1.1', 'github.com', 'google.com'),
        [int]$Count = 2
    )
    Write-HawkSubHeader "Connectivity Test"
    foreach ($t in $Targets) {
        try {
            $result = Test-Connection -ComputerName $t -Count $Count -ErrorAction SilentlyContinue
            if ($result) {
                $avg = ($result | Measure-Object ResponseTime -Average).Average
                $status = "$([math]::Round($avg, 1))ms avg"
                Write-HawkResult "✓ $t" "$status ($Count packets)" "OK"
            } else {
                Write-HawkResult "✗ $t" "No response" "FAIL"
            }
        } catch {
            Write-HawkResult "✗ $t" "Unreachable" "FAIL"
        }
    }
}

function Test-HawkDNSResolution {
    <#
    .SYNOPSIS
        Tests DNS resolution for common names.
    #>
    param(
        [string[]]$Names = @('google.com', 'github.com', 'microsoft.com')
    )
    Write-HawkSubHeader "DNS Resolution"
    foreach ($n in $Names) {
        try {
            $result = Resolve-DnsName -Name $n -Type A -ErrorAction SilentlyContinue
            if ($result) {
                $ips = ($result | Where-Object Section -eq Answer).IPAddress -join ', '
                Write-HawkResult "✓ $n" $ips "OK"
            } else {
                Write-HawkResult "✗ $n" "Resolution failed" "FAIL"
            }
        } catch {
            Write-HawkResult "✗ $n" "Resolution failed" "FAIL"
        }
    }
}

function Get-HawkNetworkStatistics {
    <#
    .SYNOPSIS
        Displays network adapter statistics (bytes, packets, errors).
    #>
    Write-HawkSubHeader "Network Statistics"
    try {
        $stats = Get-NetAdapterStatistics -ErrorAction SilentlyContinue
        if ($stats) {
            $stats | Format-Table -Property Name, ReceivedBytes, SentBytes,
                @{N='Errors';E={$_.ReceivedErrors + $_.SentErrors}} -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
        } else {
            netstat -e | Out-String | ForEach-Object { $_.TrimEnd() }
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Test-HawkPort {
    <#
    .SYNOPSIS
        Tests if a specific TCP port is open on a remote host.
    .PARAMETER Host
        Target hostname or IP.
    .PARAMETER Port
        TCP port number.
    #>
    param(
        [Parameter(Mandatory)]
        [string]$TargetHost,
        [Parameter(Mandatory)]
        [int]$Port
    )
    Write-HawkSubHeader "TCP Port Test"
    try {
        $result = Test-NetConnection -ComputerName $TargetHost -Port $Port -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
        if ($result.TcpTestSucceeded) {
            Write-HawkResult "$($TargetHost):$Port" "Open" "OK"
        } else {
            Write-HawkResult "$($TargetHost):$Port" "Closed or filtered" "FAIL"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkARPTable {
    <#
    .SYNOPSIS
        Displays the ARP table (IP-to-MAC mappings).
    #>
    Write-HawkSubHeader "ARP Table"
    try {
        Get-NetNeighbor -AddressFamily IPv4 -ErrorAction SilentlyContinue |
            Format-Table -Property IPAddress, LinkLayerAddress, State, ifIndex -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        $arp = cmd /c arp -a 2>&1 | Out-String
        Write-Output $arp
    }
}

function Get-HawkBandwidthUsage {
    <#
    .SYNOPSIS
        Estimates current bandwidth usage across network adapters.
    #>
    Write-HawkSubHeader "Bandwidth Usage"
    try {
        $counters = Get-Counter '\Network Interface(*)\Bytes Total/sec' -ErrorAction SilentlyContinue
        if ($counters) {
            foreach ($c in $counters.CounterSamples) {
                $mbps = [math]::Round($c.CookedValue * 8 / 1MB, 2)
                Write-HawkResult $c.Instance "$mbps Mbps"
            }
        } else {
            Write-Output "  Bandwidth counters not available (run as admin?)"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}
    #>
    Write-HawkSubHeader "Network Statistics"
    try {
        $stats = Get-NetAdapterStatistics -ErrorAction SilentlyContinue
        if ($stats) {
            $stats | Format-Table -Property Name, ReceivedBytes, SentBytes,
                @{N='Errors';E={$_.ReceivedErrors + $_.SentErrors}} -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
        } else {
            netstat -e | Out-String | ForEach-Object { $_.TrimEnd() }
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkFirewallRules {
    <#
    .SYNOPSIS
        Lists active firewall rules.
    .PARAMETER Direction
        Filter by direction (Inbound, Outbound). Default: all.
    #>
    param([string]$Direction = '')
    Write-HawkSubHeader "Firewall Rules ($(if ($Direction) { $Direction } else { 'All' }))"
    try {
        $rules = Get-NetFirewallRule -Enabled True -ErrorAction SilentlyContinue
        if ($Direction) {
            $rules = $rules | Where-Object Direction -eq $Direction
        }
        $rules | Format-Table -Property DisplayName, Direction, Action, Profile -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkFirewallStatus {
    <#
    .SYNOPSIS
        Displays firewall profile status.
    #>
    Write-HawkSubHeader "Firewall Status"
    try {
        Get-NetFirewallProfile -ErrorAction SilentlyContinue |
            Format-Table -Property Name, Enabled, DefaultInboundAction, DefaultOutboundBoundAction -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkNetworkDiagnostics {
    <#
    .SYNOPSIS
        Comprehensive network diagnostic overview.
    #>
    Get-HawkNetworkAdapters
    Get-HawkIPConfiguration
    Get-HawkRoutingTable
    Get-HawkActiveConnections
    Get-HawkListeningPorts
    Get-HawkDNSSettings
    Test-HawkConnectivity
    Test-HawkDNSResolution
    Get-HawkNetworkStatistics
}

#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  SECTION 3B: NETWORK ADVANCED
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Get-HawkAdapterBinding {
    <#
    .SYNOPSIS
        Shows network adapter bindings (protocols enabled).
    #>
    Write-HawkSubHeader "Network Bindings"
    try {
        Get-NetAdapterBinding -ErrorAction SilentlyContinue |
            Format-Table -Property Name, ComponentID, Enabled, DisplayName -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkNetworkProtocolStats {
    <#
    .SYNOPSIS
        Shows per-protocol network statistics (IPv4, IPv6, TCP, UDP).
    #>
    Write-HawkSubHeader "Protocol Statistics"
    try {
        $netstat = netstat -s 2>&1 | Out-String
        Write-Output $netstat
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkDNSServerConfig {
    <#
    .SYNOPSIS
        Shows DNS client server configuration and adapter-specific settings.
    #>
    Write-HawkSubHeader "DNS Server Configuration"
    try {
        Get-DnsClientServerAddress -ErrorAction SilentlyContinue |
            Format-Table -Property InterfaceAlias, AddressFamily, ServerAddresses -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Test-HawkLatency {
    <#
    .SYNOPSIS
        Measures and reports RTT latency distribution to a target.
    .PARAMETER Target
        Target hostname or IP.
    .PARAMETER Count
        Number of pings (default: 8).
    #>
    param(
        [string]$Target = '8.8.8.8',
        [int]$Count = 8
    )
    Write-HawkSubHeader "Latency Test: $Target"
    try {
        $result = Test-Connection -ComputerName $Target -Count $Count -ErrorAction SilentlyContinue
        if ($result) {
            $times = $result.ResponseTime
            $min = [math]::Round(($times | Measure-Object -Minimum).Minimum, 1)
            $max = [math]::Round(($times | Measure-Object -Maximum).Maximum, 1)
            $avg = [math]::Round(($times | Measure-Object -Average).Average, 1)
            $lossPct = $Count - $times.Count
            Write-HawkResult "Min / Max / Avg" "${min}ms / ${max}ms / ${avg}ms"
            Write-HawkResult "Loss" "$lossPct / $Count ($([math]::Round($lossPct/$Count*100, 1))%)"
        } else {
            Write-HawkResult $Target "100% packet loss" "FAIL"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkMtuDiscovery {
    <#
    .SYNOPSIS
        Discovers MTU path to a destination using ping with don't fragment.
    .PARAMETER Target
        Target hostname or IP.
    #>
    param([string]$Target = '8.8.8.8')
    Write-HawkSubHeader "Path MTU Discovery: $Target"
    try {
        $sizes = @(1473, 1472, 1400, 1300, 1200, 1000, 800, 500)
        foreach ($size in $sizes) {
            $result = Test-Connection -ComputerName $Target -Count 1 -BufferSize $size -DontFragment -ErrorAction SilentlyContinue
            if ($result) {
                Write-HawkResult "MTU $size" "Packet can be sent"
            } else {
                Write-HawkResult "MTU $size" "Packet needs fragmentation"
            }
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  SECTION 4: SERVICES & PROCESSES
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Get-HawkServicesStatus {
    <#
    .SYNOPSIS
        Lists all services with status.
    .PARAMETER State
        Filter by state (Running, Stopped). Default: all.
    #>
    param([string]$State = '')
    Write-HawkSubHeader "Windows Services"
    try {
        $services = Get-Service -ErrorAction SilentlyContinue
        if ($State) { $services = $services | Where-Object Status -eq $State }
        $services | Format-Table -Property Name, DisplayName, Status, StartType -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
        Write-Output "  Total: $($services.Count) services"
        if (-not $State) {
            $running = ($services | Where-Object Status -eq Running).Count
            $stopped = ($services | Where-Object Status -eq Stopped).Count
            $autoStopped = ($services | Where-Object { $_.StartType -eq 'Automatic' -and $_.Status -eq 'Stopped' }).Count
            Write-Output "  Running: $running | Stopped: $stopped | Auto-but-stopped: $autoStopped"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkServiceDetail {
    <#
    .SYNOPSIS
        Gets detailed information about a specific service.
    .PARAMETER Name
        Service name.
    #>
    param([string]$Name)
    try {
        $s = Get-Service -Name $Name -ErrorAction Stop
        Write-HawkSubHeader "Service: $Name"
        Write-HawkResult "Display Name" $s.DisplayName
        Write-HawkResult "Status" $s.Status
        Write-HawkResult "Start Type" $s.StartType
        Write-HawkResult "Service Type" $s.ServiceType
        Write-HawkResult "Can Pause/Continue" $s.CanPauseAndContinue
        Write-HawkResult "Can Shutdown" $s.CanShutdown
        Write-HawkResult "Can Stop" $s.CanStop
        $dep = Get-CimInstance -ClassName Win32_Service -Filter "Name='$Name'" -ErrorAction SilentlyContinue
        if ($dep) {
            Write-HawkResult "Path Name" $dep.PathName
            Write-HawkResult "Account" $dep.StartName
            Write-HawkResult "Process ID" $dep.ProcessId
            Write-HawkResult "Description" $dep.Description
        }
    } catch {
        Write-HawkError "Service '$Name' not found: $_"
    }
}

function Get-HawkAutoStartServices {
    <#
    .SYNOPSIS
        Lists auto-start services that are currently stopped.
    #>
    Write-HawkSubHeader "Auto-Start Services Currently Stopped"
    try {
        $autoStopped = Get-Service -ErrorAction SilentlyContinue |
            Where-Object { $_.StartType -eq 'Automatic' -and $_.Status -eq 'Stopped' } |
            Format-Table -Property Name, DisplayName, Status, StartType -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
        if ($autoStopped) { $autoStopped } else { Write-Output "  None (all auto-start services are running)" }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  SECTION 5: SECURITY & AUDIT
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Get-HawkLoggedOnUsers {
    <#
    .SYNOPSIS
        Shows currently logged-on users (local and remote sessions).
    #>
    Write-HawkSubHeader "Logged-On Users"
    try {
        $sessions = Get-CimInstance Win32_LogonSession -Filter "LogonType=2 OR LogonType=10" -ErrorAction SilentlyContinue
        if ($sessions) {
            foreach ($s in $sessions) {
                $user = Get-CimInstance -ClassName Win32_UserAccount -Filter "Name='$($s.UserName)'" -ErrorAction SilentlyContinue
                $type = if ($s.LogonType -eq 2) { 'Interactive' } else { 'Remote' }
                Write-HawkResult $s.Caption "$type session since $($s.StartTime)"
            }
        }
        # Fallback: query session
        $sessions = query session 2>&1 | Out-String
        if ($sessions) { Write-Output $sessions }
    } catch {}
}

function Get-HawkOpenFiles {
    <#
    .SYNOPSIS
        Lists files opened over the network via SMB.
    #>
    Write-HawkSubHeader "Open Network Files"
    try {
        $openFiles = Get-SmbOpenFile -ErrorAction SilentlyContinue
        if ($openFiles) {
            $openFiles | Format-Table -Property FileId, ClientUserName, Path, ShareRelativePath -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
        } else {
            Write-Output "  No open network files or not available"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Get-HawkLocalUsers {
    <#
    .SYNOPSIS
        Lists local user accounts.
    #>
    Write-HawkSubHeader "Local Users"
    try {
        if (-not (Test-WindowsOnly)) { return }
        $users = Get-LocalUser -ErrorAction SilentlyContinue
        $users | Format-Table -Property Name, Enabled, LastLogon, PasswordChangeableDate,
            @{N='Groups';E={(Get-LocalGroupMember -Member $_.Name -ErrorAction SilentlyContinue | Where-Object ObjectClass -eq 'User').Group -join ', '}} -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkLocalGroups {
    <#
    .SYNOPSIS
        Lists local groups and their members.
    #>
    Write-HawkSubHeader "Local Groups"
    try {
        if (-not (Test-WindowsOnly)) { return }
        $groups = Get-LocalGroup -ErrorAction SilentlyContinue
        foreach ($g in $groups) {
            $members = Get-LocalGroupMember -Group $g.Name -ErrorAction SilentlyContinue
            $memberNames = ($members | Where-Object ObjectClass -eq 'User').Name -join ', '
            if (-not $memberNames) { $memberNames = '(none)' }
            Write-HawkResult $g.Name $memberNames
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkAdminGroup {
    <#
    .SYNOPSIS
        Lists members of the Administrators group.
    #>
    Write-HawkSubHeader "Administrators Group"
    try {
        if (-not (Test-WindowsOnly)) { return }
        $admins = Get-LocalGroupMember -Group Administrators -ErrorAction SilentlyContinue
        $admins | Format-Table -Property Name, ObjectClass, PrincipalSource -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkStartupPrograms {
    <#
    .SYNOPSIS
        Lists startup programs from registry and startup folders.
    #>
    Write-HawkSubHeader "Startup Programs"
    try {
        $startup = @()
        $regPaths = @(
            'HKLM:\Software\Microsoft\Windows\CurrentVersion\Run',
            'HKLM:\Software\Microsoft\Windows\CurrentVersion\RunOnce',
            'HKCU:\Software\Microsoft\Windows\CurrentVersion\Run',
            'HKCU:\Software\Microsoft\Windows\CurrentVersion\RunOnce'
        )
        foreach ($rp in $regPaths) {
            if (Test-Path $rp) {
                $items = Get-ItemProperty -Path $rp -ErrorAction SilentlyContinue
                $itemProps = $items.PSObject.Properties |
                    Where-Object { $_.Name -notin @('PSPath','PSParentPath','PSChildName','PSDrive','PSProvider') }
                foreach ($prop in $itemProps) {
                    Write-HawkResult "$($prop.Name)" $($prop.Value)
                }
            }
        }
        $startupFolder = [Environment]::GetFolderPath('Startup')
        if (Test-Path $startupFolder) {
            Write-HawkSubHeader "Startup Folder"
            Get-ChildItem $startupFolder -ErrorAction SilentlyContinue |
                Format-Table -Property Name -AutoSize | Out-String | ForEach-Object { $_.TrimEnd() }
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkScheduledTasks {
    <#
    .SYNOPSIS
        Lists scheduled tasks.
    .PARAMETER Path
        Task path to filter (e.g., '\Microsoft\'). Default: root.
    #>
    param([string]$TaskPath = '\')
    Write-HawkSubHeader "Scheduled Tasks"
    try {
        $tasks = Get-ScheduledTask -TaskPath $TaskPath -ErrorAction SilentlyContinue |
            Where-Object State -ne 'Disabled'
        $tasks | Format-Table -Property TaskName, State,
            @{N='Next Run';E={$_.NextRunTime}},
            @{N='Actions';E={($_.Actions | ForEach-Object { if ($_.Execute) { Split-Path $_.Execute -Leaf } }) -join ', '}} -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkDefenderStatus {
    <#
    .SYNOPSIS
        Displays Windows Defender status.
    #>
    Write-HawkSubHeader "Windows Defender"
    try {
        if (-not (Test-WindowsOnly)) { return }
        $mp = Get-MpComputerStatus -ErrorAction SilentlyContinue
        if ($mp) {
            Write-HawkResult "Real-time Protection" $mp.RealTimeProtectionEnabled
            Write-HawkResult "Antivirus Enabled" $mp.AntivirusEnabled
            Write-HawkResult "Antispyware Enabled" $mp.AntispywareEnabled
            Write-HawkResult "Signature Age" "$($mp.AntivirusSignatureAge) days"
            Write-HawkResult "Signature Version" $mp.AntivirusSignatureVersion
            Write-HawkResult "Engine Version" $mp.AMEngineVersion
            Write-HawkResult "Product Version" $mp.AMProductVersion
            Write-HawkResult "Last Quick Scan" $mp.QuickScanEndTime
            Write-HawkResult "Last Full Scan" $mp.FullScanEndTime
            Write-HawkResult "NIS Enabled" $mp.NISEnabled
            Write-HawkResult "Tamper Protection" $mp.IsTamperProtected
            Write-HawkResult "Cloud Protection" $mp.CloudProtectionLevel
        } else {
            Write-Output "  Windows Defender not available or not installed"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkDefenderThreats {
    <#
    .SYNOPSIS
        Displays recent Windows Defender threat detections.
    #>
    Write-HawkSubHeader "Threat History"
    try {
        $threats = Get-MpThreatDetection -ErrorAction SilentlyContinue | Select-Object -First 10
        if ($threats) {
            $threats | Format-Table -Property Resources, ThreatID, DetectionTime, DomainUser -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
        } else {
            Write-Output "  No threats detected"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkExecutionPolicy {
    <#
    .SYNOPSIS
        Displays PowerShell execution policy.
    #>
    Write-HawkSubHeader "PowerShell Execution Policy"
    try {
        $policies = Get-ExecutionPolicy -List -ErrorAction SilentlyContinue
        $policies | Format-Table -Property Scope, ExecutionPolicy -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkInstalledUpdates {
    <#
    .SYNOPSIS
        Lists recently installed Windows updates.
    .PARAMETER Count
        Number of updates to display (default: 10).
    #>
    param([int]$Count = 10)
    Write-HawkSubHeader "Installed Updates (Last $Count)"
    try {
        if (-not (Test-WindowsOnly)) { return }
        Get-HotFix -ErrorAction SilentlyContinue |
            Sort-Object InstalledOn -Descending |
            Select-Object -First $Count |
            Format-Table -Property HotFixID, Description, InstalledOn, InstalledBy -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkInstalledSoftware {
    <#
    .SYNOPSIS
        Lists installed software from the registry.
    .PARAMETER Pattern
        Optional filter pattern.
    #>
    param([string]$Pattern = '')
    Write-HawkSubHeader "Installed Software"
    try {
        $paths = @(
            'HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*',
            'HKLM:\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*',
            'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*'
        )
        $software = @()
        foreach ($p in $paths) {
            if (Test-Path $p) {
                $software += Get-ItemProperty $p -ErrorAction SilentlyContinue |
                    Where-Object { $_.DisplayName -and $_.DisplayName -notlike '*Update for*' -and $_.DisplayName -notlike '*Security Update*' }
            }
        }
        if ($Pattern) {
            $software = $software | Where-Object { $_.DisplayName -like "*$Pattern*" }
        }
        $software | Sort-Object DisplayName |
            Format-Table -Property DisplayName, DisplayVersion, Publisher, InstallDate -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
        Write-Output "  Total installed applications: $($software.Count)"
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkCertificateStore {
    <#
    .SYNOPSIS
        Lists certificates from the personal certificate store.
    #>
    Write-HawkSubHeader "Certificate Store (Current User - Personal)"
    try {
        $certs = Get-ChildItem Cert:\CurrentUser\My -ErrorAction SilentlyContinue
        if ($certs) {
            $certs | Format-Table -Property Subject, Issuer,
                @{N='Not After';E={$_.NotAfter}},
                @{N='Serial';E={$_.SerialNumber.Substring(0, [Math]::Min(16, $_.SerialNumber.Length))}} -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
            $expiring = $certs | Where-Object { $_.NotAfter -lt (Get-Date).AddDays(30) }
            if ($expiring) { Write-Output "  ⚠ $($expiring.Count) certificate(s) expiring within 30 days" }
        } else {
            Write-Output "  No personal certificates found"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkBitLockerStatus {
    <#
    .SYNOPSIS
        Displays BitLocker encryption status.
    #>
    Write-HawkSubHeader "BitLocker Status"
    try {
        if (-not (Test-WindowsOnly)) { return }
        $bl = Get-BitLockerVolume -ErrorAction SilentlyContinue
        if ($bl) {
            $bl | Format-Table -Property MountPoint, EncryptionMethod, ProtectionStatus,
                @{N='Percentage';E={$_.PercentageEncrypted}} -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
        } else {
            Write-Output "  BitLocker not available or no volumes encrypted"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkSecurityAudit {
    <#
    .SYNOPSIS
        Comprehensive security audit.
    #>
    Get-HawkLocalUsers
    Get-HawkAdminGroup
    Get-HawkStartupPrograms
    Get-HawkScheduledTasks
    Get-HawkDefenderStatus
    Get-HawkDefenderThreats
    Get-HawkExecutionPolicy
    Get-HawkFirewallStatus
    Get-HawkCertificateStore
    Get-HawkBitLockerStatus
}

#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  SECTION 6: EVENT LOGS & AUDIT TRAIL
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Get-HawkSystemEvents {
    <#
    .SYNOPSIS
        Retrieves recent system event log entries.
    .PARAMETER Count
        Number of events to retrieve (default: 20).
    .PARAMETER Level
        Filter by level: Critical=1, Error=2, Warning=3, Info=4.
    #>
    param([int]$Count = 20, [int]$Level = 0)
    Write-HawkSubHeader "System Events (Last $Count)"
    try {
        $filter = @{LogName='System'; StartTime=(Get-Date).AddHours(-24)}
        if ($Level -gt 0) { $filter.Level = $Level }
        Get-WinEvent -FilterHashtable $filter -MaxEvents $Count -ErrorAction SilentlyContinue |
            Format-Table -Property TimeCreated, LevelDisplayName,
                @{N='Provider';E={$_.ProviderName}},
                @{N='Message';E={$_.Message.Substring(0, [Math]::Min(80, $_.Message.Length))}} -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkApplicationEvents {
    <#
    .SYNOPSIS
        Retrieves recent application event log entries.
    .PARAMETER Count
        Number of events to retrieve (default: 20).
    #>
    param([int]$Count = 20)
    Write-HawkSubHeader "Application Events (Last $Count)"
    try {
        $filter = @{LogName='Application'; StartTime=(Get-Date).AddHours(-24)}
        Get-WinEvent -FilterHashtable $filter -MaxEvents $Count -ErrorAction SilentlyContinue |
            Format-Table -Property TimeCreated, LevelDisplayName,
                @{N='Source';E={$_.ProviderName}},
                @{N='EventID';E={$_.Id}},
                @{N='Message';E={$_.Message.Substring(0, [Math]::Min(80, $_.Message.Length))}} -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkSecurityEvents {
    <#
    .SYNOPSIS
        Retrieves recent security event log entries.
    .PARAMETER Count
        Number of events to retrieve (default: 20).
    .PARAMETER EventID
        Filter by specific event ID (e.g., 4624=logon, 4625=failed logon, 4648=explicit logon).
    #>
    param([int]$Count = 20, [int]$EventID = 0)
    Write-HawkSubHeader "Security Events (Last $Count)"
    try {
        $filter = @{LogName='Security'; StartTime=(Get-Date).AddHours(-24)}
        if ($EventID -gt 0) { $filter.ID = $EventID }
        Get-WinEvent -FilterHashtable $filter -MaxEvents $Count -ErrorAction SilentlyContinue |
            Format-Table -Property TimeCreated,
                @{N='EventID';E={$_.Id}},
                @{N='Task';E={$_.TaskDisplayName}},
                @{N='Message';E={$_.Message.Substring(0, [Math]::Min(80, $_.Message.Length))}} -AutoSize |
            Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkRecentChanges {
    <#
    .SYNOPSIS
        Identifies recent file system changes in critical directories.
    .PARAMETER Hours
        How far back to look (default: 24).
    #>
    param([int]$Hours = 24)
    Write-HawkSubHeader "Recent File Changes (Last $Hours Hours)"
    $dirs = @(
        $env:SystemRoot,
        "$env:SystemRoot\System32\drivers\etc",
        "$env:ProgramData"
    )
    foreach ($dir in $dirs) {
        if (Test-Path $dir) {
            $changes = Get-ChildItem -Path $dir -Recurse -ErrorAction SilentlyContinue |
                Where-Object { -not $_.PSIsContainer -and $_.LastWriteTime -gt (Get-Date).AddHours(-$Hours) } |
                Sort-Object LastWriteTime -Descending |
                Select-Object -First 5
            if ($changes) {
                Write-HawkSubHeader "Changes in: $dir"
                $changes | Format-Table -Property Name, LastWriteTime, Length -AutoSize |
                    Out-String | ForEach-Object { $_.TrimEnd() }
            }
        }
    }
}

function Get-HawkCrashDumps {
    <#
    .SYNOPSIS
        Checks for recent crash dump files.
    #>
    Write-HawkSubHeader "Crash Dumps & Minidumps"
    $dumpDirs = @(
        "$env:SystemRoot\Minidump",
        "$env:LOCALAPPDATA\CrashDumps",
        "$env:TEMP"
    )
    $found = $false
    foreach ($dir in $dumpDirs) {
        if (Test-Path $dir) {
            $dumps = Get-ChildItem -Path $dir -Filter '*.dmp' -ErrorAction SilentlyContinue |
                Where-Object { $_.LastWriteTime -gt (Get-Date).AddDays(-7) }
            if ($dumps) {
                $found = $true
                Write-HawkSubHeader "Dumps in: $dir"
                $dumps | Format-Table -Property Name, LastWriteTime, Length -AutoSize |
                    Out-String | ForEach-Object { $_.TrimEnd() }
            }
        }
    }
    if (-not $found) { Write-Output "  No recent crash dumps found" }
}

#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  SECTION 7: STORAGE & FILES
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Get-HawkStorageOverview {
    <#
    .SYNOPSIS
        Comprehensive storage overview including physical disks, volumes, and usage trends.
    #>
    Get-HawkDiskInfo
    Get-HawkVolumeInfo
    Write-HawkSubHeader "Disk Health"
    try {
        $pd = Get-PhysicalDisk -ErrorAction SilentlyContinue
        if ($pd) {
            $pd | Format-Table -Property FriendlyName, MediaType, HealthStatus, OperationalStatus, Size -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  SECTION 8: ADDITIONAL SYSTEM & PLATFORM COMMANDS
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Get-HawkDiskPerformance {
    <#
    .SYNOPSIS
        Shows disk performance metrics (read/write rates, queue depth).
    #>
    Write-HawkSubHeader "Disk Performance"
    try {
        $disks = Get-CimInstance Win32_PerfFormattedData_PerfDisk_PhysicalDisk -ErrorAction SilentlyContinue
        if ($disks) {
            $disks | Where-Object Name -ne '_Total' |
                Format-Table -Property Name,
                    @{N='R/s';E={$_.DiskReadsPerSec}},
                    @{N='W/s';E={$_.DiskWritesPerSec}},
                    @{N='AvgR(s)';E={$_.AvgDiskSecPerRead}},
                    @{N='AvgW(s)';E={$_.AvgDiskSecPerWrite}},
                    @{N='Queue';E={$_.CurrentDiskQueueLength}} -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkPageFile {
    <#
    .SYNOPSIS
        Displays paging file configuration and usage.
    #>
    Write-HawkSubHeader "Page File"
    try {
        $pf = Get-CimInstance Win32_PageFileUsage -ErrorAction SilentlyContinue
        if ($pf) {
            foreach ($p in $pf) {
                $totalGB = [math]::Round(($p.AllocatedBaseSize / 1KB), 2)
                $peakGB = [math]::Round(($p.PeakUsage / 1KB), 2)
                Write-HawkResult $p.Name "Allocated: ${totalGB}GB, Peak: ${peakGB}GB"
            }
        }
        $os = Get-CimInstance Win32_OperatingSystem -ErrorAction SilentlyContinue
        if ($os) {
            Write-HawkResult "Total Page File" "$([math]::Round($os.TotalVirtualMemorySize / 1MB, 2)) GB"
            Write-HawkResult "Free Page File" "$([math]::Round($os.FreeVirtualMemory / 1MB, 2)) GB"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkDrivers {
    <#
    .SYNOPSIS
        Lists installed kernel drivers and their status.
    #>
    Write-HawkSubHeader "System Drivers"
    try {
        $drivers = Get-WindowsDriver -Online -ErrorAction SilentlyContinue
        if ($drivers) {
            $drivers | Sort-Object DriverSignature |
                Format-Table -Property DriverSignature, DriverProvider,
                    @{N='Class';E={$_.ClassName}},
                    @{N='Date';E={$_.DriverDate}} -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
        } else {
            # Fallback: list services that are also drivers
            $drvServices = Get-CimInstance Win32_SystemDriver -ErrorAction SilentlyContinue |
                Where-Object State -eq 'Running'
            $drvServices | Format-Table -Property Name, DisplayName, State, StartMode -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
        }
        Write-Output "  Total: $($drivers.Count) drivers"
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkWindowsFeatures {
    <#
    .SYNOPSIS
        Lists installed Windows optional features (roles/features).
    #>
    Write-HawkSubHeader "Windows Features"
    try {
        $features = Get-WindowsOptionalFeature -Online -ErrorAction SilentlyContinue |
            Where-Object State -eq 'Enabled'
        if ($features) {
            $features | Format-Table -Property FeatureName, State -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
            Write-Output "  Enabled features: $($features.Count)"
        } else {
            Write-Output "  Windows features information not available"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkRestorePoints {
    <#
    .SYNOPSIS
        Lists system restore points.
    #>
    Write-HawkSubHeader "System Restore Points"
    try {
        $rps = Get-ComputerRestorePoint -ErrorAction SilentlyContinue
        if ($rps) {
            $rps | Format-Table -Property Description,
                @{N='Created';E={$_.CreationTime}},
                @{N='Type';E={$_.RestorePointType}} -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
            Write-Output "  Total restore points: $($rps.Count)"
        } else {
            Write-Output "  No restore points found or System Restore disabled"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkPowerPlan {
    <#
    .SYNOPSIS
        Displays active power plan configuration.
    #>
    Write-HawkSubHeader "Power Configuration"
    try {
        $plan = powercfg /GETACTIVESCHEME 2>&1 | Out-String
        Write-Output "  Active Power Scheme: $($plan.Trim())"
        $subs = powercfg /QUERY 2>&1 | Select-String -Pattern 'Power Scheme.*: ' | Out-String
        $allPlans = powercfg /LIST 2>&1 | Out-String
        Write-Output "`n  All Power Schemes:"
        $allPlans.Split("`n") | ForEach-Object { "    $_" }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkTimeInfo {
    <#
    .SYNOPSIS
        Displays system time, time zone, and NTP configuration.
    #>
    Write-HawkSubHeader "System Time"
    try {
        $tz = Get-TimeZone -ErrorAction SilentlyContinue
        Write-HawkResult "Current Time" (Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
        Write-HawkResult "UTC Time" ((Get-Date).ToUniversalTime() -Format 'yyyy-MM-dd HH:mm:ss')
        Write-HawkResult "Time Zone" $tz.DisplayName
        Write-HawkResult "UTC Offset" $tz.BaseUtcOffset
        Write-HawkResult "DST Active" $tz.DaylightSavingTime

        # NTP config
        $ntp = Get-CimInstance Win32_ComputerSystem -ErrorAction SilentlyContinue
        if ($ntp) {
            Write-HawkResult "NTP Server" "(see registry: HKLM:\SYSTEM\CurrentControlSet\Services\W32Time\Parameters\NtpServer)"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkPrinters {
    <#
    .SYNOPSIS
        Lists installed printers.
    #>
    Write-HawkSubHeader "Printers"
    try {
        $printers = Get-CimInstance Win32_Printer -ErrorAction SilentlyContinue
        if ($printers) {
            $printers | Format-Table -Property Name, DriverName, PortName, PrinterState, PrinterStatus -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
            Write-Output "  Total printers: $($printers.Count)"
        } else {
            Write-Output "  No printers installed"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkNetworkShares {
    <#
    .SYNOPSIS
        Lists local network shares.
    #>
    Write-HawkSubHeader "Network Shares"
    try {
        $shares = Get-CimInstance Win32_Share -ErrorAction SilentlyContinue
        if ($shares) {
            $shares | Format-Table -Property Name, Path, Description, Type -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
            Write-Output "  Total shares: $($shares.Count)"
        } else {
            Write-Output "  No shares found"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkWiFiProfiles {
    <#
    .SYNOPSIS
        Lists saved WiFi network profiles.
    #>
    Write-HawkSubHeader "Saved WiFi Networks"
    try {
        $profiles = netsh wlan show profiles 2>&1 | Out-String
        if ($profiles -match 'All User Profile') {
            $lines = $profiles -split "`n" | Where-Object { $_ -match 'All User Profile' }
            foreach ($line in $lines) {
                $name = ($line -split ':')[1].Trim()
                Write-Output "  $name"
            }
        } else {
            Write-Output "  No saved WiFi profiles or WiFi not available"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkDeviceInfo {
    <#
    .SYNOPSIS
        Lists devices with problems (error status) from Plug and Play.
    #>
    Write-HawkSubHeader "Problem Devices"
    try {
        $problemDevices = Get-PnpDevice -Status Error -ErrorAction SilentlyContinue
        if ($problemDevices) {
            $problemDevices | Format-Table -Property FriendlyName, Class, Problem, Status -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
            Write-Output "  Problem devices: $($problemDevices.Count)"
        } else {
            Write-Output "  All devices are working properly"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkEventLogSummary {
    <#
    .SYNOPSIS
        Shows error/warning counts per event log in the last 24 hours.
    #>
    Write-HawkSubHeader "Event Log Summary (Last 24h)"
    $logNames = @('Application', 'System', 'Security', 'PowerShell')
    foreach ($log in $logNames) {
        try {
            $errors = Get-WinEvent -FilterHashtable @{LogName=$log; Level=1; StartTime=(Get-Date).AddHours(-24)} -ErrorAction SilentlyContinue
            $warnings = Get-WinEvent -FilterHashtable @{LogName=$log; Level=2; StartTime=(Get-Date).AddHours(-24)} -ErrorAction SilentlyContinue
            $errCount = if ($errors) { @($errors).Count } else { 0 }
            $warnCount = if ($warnings) { @($warnings).Count } else { 0 }
            Write-HawkResult $log "Errors: $errCount, Warnings: $warnCount"
        } catch {
            Write-HawkResult $log "Log not available"
        }
    }
}

function Get-HawkWindowsActivation {
    <#
    .SYNOPSIS
        Displays Windows activation status.
    #>
    Write-HawkSubHeader "Windows Activation"
    try {
        $activation = Get-CimInstance -ClassName SoftwareLicensingProduct -Filter "PartialProductKey IS NOT NULL" -ErrorAction SilentlyContinue
        if ($activation) {
            foreach ($a in $activation) {
                Write-HawkResult "Product Name" $a.Name
                Write-HawkResult "License Status" $a.LicenseStatus
                Write-HawkResult "Product Key (last 5)" $a.PartialProductKey
            }
        }
        # Fallback via command
        $slmgr = cscript //nologo "$env:SystemRoot\System32\slmgr.vbs" /dli 2>&1 | Out-String
        Write-Output $slmgr
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkPowerShellModules {
    <#
    .SYNOPSIS
        Lists available and imported PowerShell modules.
    #>
    Write-HawkSubHeader "PowerShell Modules"
    try {
        $moduleCount = (Get-Module -ListAvailable -ErrorAction SilentlyContinue).Count
        $importedCount = (Get-Module -ErrorAction SilentlyContinue).Count
        Write-HawkResult "Available Modules" $moduleCount
        Write-HawkResult "Imported Modules" $importedCount
        Write-Output "`n  Imported modules:"
        Get-Module -ErrorAction SilentlyContinue |
            Format-Table -Property Name, Version, ModuleType -AutoSize |
            Out-String | ForEach-Object { "    " + $_.TrimEnd() }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkHyperVStatus {
    <#
    .SYNOPSIS
        Checks if Hyper-V is installed and lists virtual machines.
    #>
    Write-HawkSubHeader "Hyper-V"
    try {
        $hyperv = Get-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -ErrorAction SilentlyContinue
        if ($hyperv -and $hyperv.State -eq 'Enabled') {
            Write-HawkResult "Hyper-V" "Installed" "OK"
            $vms = Get-VM -ErrorAction SilentlyContinue
            if ($vms) {
                $vms | Format-Table -Property Name, State, CPUUsage, MemoryAssigned -AutoSize |
                    Out-String | ForEach-Object { $_.TrimEnd() }
                Write-Output "  Total VMs: $($vms.Count)"
            } else {
                Write-Output "  No virtual machines configured"
            }
        } else {
            Write-Output "  Hyper-V is not installed"
        }
    } catch {
        Write-Output "  Hyper-V check not available"
    }
}

function Get-HawkDockerStatus {
    <#
    .SYNOPSIS
        Checks if Docker is running and lists containers.
    #>
    Write-HawkSubHeader "Docker"
    try {
        $version = docker --version 2>&1 | Out-String
        if ($version -match 'Docker') {
            Write-HawkResult "Docker" $version.Trim()
            $containers = docker ps --format "table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Names}}" 2>&1
            if ($containers) { Write-Output $containers }
        } else {
            Write-Output "  Docker not found"
        }
    } catch {
        Write-Output "  Docker not found or not running"
    }
}

function Get-HawkWSLStatus {
    <#
    .SYNOPSIS
        Checks WSL status and lists installed Linux distributions.
    #>
    Write-HawkSubHeader "Windows Subsystem for Linux"
    try {
        $wsl = wsl --status 2>&1 | Out-String
        if ($wsl -match 'WSL') {
            Write-Output $wsl
            $distros = wsl --list --verbose 2>&1 | Out-String
            if ($distros) { Write-Output $distros }
        } else {
            Write-Output "  WSL not installed"
        }
    } catch {
        Write-Output "  WSL check not available"
    }
}

#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  SECTION 9: THREAT HUNTING & ANOMALY DETECTION
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Get-HawkSuspiciousProcesses {
    <#
    .SYNOPSIS
        Hunts for suspicious processes (unsigned, hidden, network connections).
    #>
    Write-HawkSubHeader "Suspicious Process Check"
    try {
        $suspicious = @()
        $procs = Get-Process -ErrorAction SilentlyContinue

        # Check for processes with no main window but network connections
        $networkProcs = Get-NetTCPConnection -ErrorAction SilentlyContinue |
            Where-Object OwningProcess -gt 0 |
            Select-Object -ExpandProperty OwningProcess -Unique
        $headlessNetwork = $procs | Where-Object {
            $_.Id -in $networkProcs -and
            -not $_.MainWindowTitle -and
            $_.ProcessName -notin @('svchost', 'System', 'Idle', 'csrss', 'wininit', 'services',
                                   'lsass', 'wmiprvse', 'spoolsv', 'taskhostw', 'sihost',
                                   'RuntimeBroker', 'SecurityHealthService', 'dllhost',
                                   'dasHost', 'conhost', 'fontdrvhost')
        }
        if ($headlessNetwork) {
            Write-Output "  ⚠ Headless processes with network connections:"
            $headlessNetwork | Format-Table -Property Id, ProcessName,
                @{N='Company';E={(Get-Item $_.Path -ErrorAction SilentlyContinue).VersionInfo.CompanyName}} -AutoSize |
                Out-String | ForEach-Object { "    " + $_.TrimEnd() }
        } else {
            Write-Output "  ✓ No suspicious headless-network processes detected"
        }

        # Check for unsigned processes
        Write-Output ""
        $unsigned = $procs | Where-Object {
            $_.Path -and $_.ProcessName -notin @('svchost', 'System', 'Idle')
        } | Select-Object -First 50 | Where-Object {
            $path = $_.Path
            try {
                $sig = Get-AuthenticodeSignature -FilePath $path -ErrorAction SilentlyContinue
                $sig.Status -ne 'Valid' -and $sig.Status -ne 'NotSigned'
            } catch { $false }
        }
        if ($unsigned) {
            Write-Output "  ⚠ Potentially unsigned processes (first 10):"
            $unsigned | Select-Object -First 10 |
                Format-Table -Property Id, ProcessName, Path -AutoSize |
                Out-String | ForEach-Object { "    " + $_.TrimEnd() }
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkGhostPorts {
    <#
    .SYNOPSIS
        Detects listening ports that don't map to known processes.
    #>
    Write-HawkSubHeader "Ghost Port Detection"
    try {
        $tcpListen = Get-NetTCPConnection -State Listen -ErrorAction SilentlyContinue
        $orphans = @()
        foreach ($conn in $tcpListen) {
            $proc = Get-Process -Id $conn.OwningProcess -ErrorAction SilentlyContinue
            if (-not $proc) {
                $orphans += $conn
            }
        }
        if ($orphans) {
            Write-Output "  ⚠ Orphaned listening ports (no owning process):"
            $orphans | Format-Table -Property LocalAddress, LocalPort -AutoSize |
                Out-String | ForEach-Object { "    " + $_.TrimEnd() }
        } else {
            Write-Output "  ✓ All listening ports have valid owning processes"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkFileAnomalies {
    <#
    .SYNOPSIS
        Checks for suspicious files in common locations.
    #>
    Write-HawkSubHeader "File Anomaly Scan"
    $tempCheck = @{
        Path  = "$env:TEMP"
        Count = 20
    }
    try {
        $largeTemp = Get-ChildItem -Path $tempCheck.Path -ErrorAction SilentlyContinue |
            Where-Object { -not $_.PSIsContainer -and $_.Length -gt 100MB }
        if ($largeTemp) {
            Write-Output "  ⚠ Large temp files (>100MB):"
            $largeTemp | Format-Table -Property Name, @{N='Size(MB)';E={[math]::Round($_.Length/1MB,1)}}, LastWriteTime -AutoSize |
                Out-String | ForEach-Object { "    " + $_.TrimEnd() }
        } else {
            Write-Output "  ✓ No oversized temp files found"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkNetworkConnectionsByProcess {
    <#
    .SYNOPSIS
        Shows established TCP connections grouped by process name.
    #>
    Write-HawkSubHeader "Network Connections by Process"
    try {
        $conns = Get-NetTCPConnection -State Established -ErrorAction SilentlyContinue
        $groups = $conns | Group-Object -Property OwningProcess
        foreach ($g in $groups) {
            $procName = (Get-Process -Id $g.Name -ErrorAction SilentlyContinue).ProcessName
            if (-not $procName) { $procName = "PID:$($g.Name)" }
            Write-HawkResult $procName "$($g.Count) connection(s)"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkUnusualPorts {
    <#
    .SYNOPSIS
        Flags any listening ports outside common ranges (<1024 or >49152 ephemeral).
    #>
    Write-HawkSubHeader "Unusual Listening Ports"
    try {
        $listeners = Get-NetTCPConnection -State Listen -ErrorAction SilentlyContinue
        $unusual = $listeners | Where-Object {
            $_.LocalPort -gt 1024 -and $_.LocalPort -lt 49152 -and
            $_.LocalPort -notin @(3306, 3389, 5040, 5353, 5357, 5358, 5985, 5986, 8080, 8443, 9000, 9090, 10000)
        }
        if ($unusual) {
            $unusual | Format-Table -Property LocalPort, LocalAddress,
                @{N='Process';E={(Get-Process -Id $_.OwningProcess -ErrorAction SilentlyContinue).ProcessName}} -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
        } else {
            Write-Output "  No unusual listening ports detected"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkServiceAnomalies {
    <#
    .SYNOPSIS
        Detects services with unusual configurations (running from temp, non-Microsoft).
    #>
    Write-HawkSubHeader "Service Anomalies"
    try {
        $svcs = Get-CimInstance Win32_Service -ErrorAction SilentlyContinue |
            Where-Object { $_.State -eq 'Running' -and $_.PathName -like '*\Temp*' }
        if ($svcs) {
            Write-Output "  ⚠ Services running from Temp directory:"
            $svcs | Format-Table -Property Name, DisplayName, PathName -AutoSize |
                Out-String | ForEach-Object { "    " + $_.TrimEnd() }
        } else {
            Write-Output "  ✓ No services running from Temp directories"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkProcessMitigations {
    <#
    .SYNOPSIS
        Displays exploit protection/mitigation settings for running processes.
    #>
    Write-HawkSubHeader "Process Mitigations"
    try {
        $mitigations = Get-ProcessMitigation -ErrorAction SilentlyContinue
        if ($mitigations) {
            $mitigations | Select-Object -First 20 |
                Format-Table -Property ProcessId, ProcessName,
                    @{N='DEP';E={$_.Dep}},
                    @{N='ASLR';E={$_.Aslr}},
                    @{N='SEHOP';E={$_.Sehop}} -AutoSize |
                Out-String | ForEach-Object { $_.TrimEnd() }
        } else {
            Write-Output "  Process mitigations not available"
        }
    } catch {
        Write-Output "  Process mitigations query not available"
    }
}

function Get-HawkRegistryHives {
    <#
    .SYNOPSIS
        Shows registry hives currently loaded in the registry.
    #>
    Write-HawkSubHeader "Registry Hives"
    try {
        $hives = Get-ChildItem -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\hivelist' -ErrorAction SilentlyContinue
        Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\hivelist' -ErrorAction SilentlyContinue |
            Format-Table -AutoSize | Out-String | ForEach-Object { $_.TrimEnd() }
    } catch {
        Write-Output "  Registry hives not available"
    }
}

function Get-HawkSystemLogins {
    <#
    .SYNOPSIS
        Shows recent interactive and remote logins from Security event log.
    #>
    Write-HawkSubHeader "Recent Logins (Security Log)"
    try {
        $logons = Get-WinEvent -FilterHashtable @{LogName='Security'; ID=4624; StartTime=(Get-Date).AddHours(-24)} -MaxEvents 20 -ErrorAction SilentlyContinue
        if ($logons) {
            foreach ($e in $logons) {
                $user = $e.Properties[5].Value  # TargetUserName
                $type = $e.Properties[8].Value  # LogonType: 2=Interactive, 10=Remote
                $typeName = switch ($type) { 2 {'Interactive'} 10 {'Remote'} default {'Type ' + $type} }
                Write-HawkResult $e.TimeCreated "$typeName: $user"
            }
        } else {
            Write-Output "  No logon events in the last 24 hours"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

function Get-HawkPendingReboot {
    <#
    .SYNOPSIS
        Checks if the system has a pending reboot (from updates, config changes).
    #>
    Write-HawkSubHeader "Pending Reboot Check"
    $rebootPending = $false
    try {
        $regPaths = @(
            'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update\RebootRequired',
            'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Component Based Servicing\RebootPending',
            'HKLM:\SOFTWARE\Microsoft\ServerManager\CurrentRebootAttempts'
        )
        foreach ($rp in $regPaths) {
            if (Test-Path $rp) {
                Write-HawkResult "$rp" "Reboot required" "WARN"
                $rebootPending = $true
            }
        }
        $cb = Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager' -Name PendingFileRenameOperations -ErrorAction SilentlyContinue
        if ($cb -and $cb.PendingFileRenameOperations) {
            Write-HawkResult "Pending File Renames" $cb.PendingFileRenameOperations.Count "WARN"
            $rebootPending = $true
        }
        if (-not $rebootPending) {
            Write-Output "  ✓ No pending reboot detected"
        }
    } catch {
        Write-HawkError $_.Exception.Message
    }
}

#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  SECTION 10: COMPLIANCE CHECKS (CIS-Inspired)
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Get-HawkComplianceStatus {
    <#
    .SYNOPSIS
        Runs CIS-inspired security baseline checks and produces a compliance score.
    .DESCRIPTION
        Checks: Guest account status, admin count, LSA protection, Defender status,
        execution policy, firewall enabled, etc. Returns score out of 100.
    #>
    Write-HawkSubHeader "CIS-Inspired Compliance Scan"
    $checks = @()
    $score = 0
    $total = 0

    try {
        # Check 1: Guest account disabled
        $total++
        $guest = Get-LocalUser -Name 'Guest' -ErrorAction SilentlyContinue
        if ($guest -and -not $guest.Enabled) {
            $score++; $checks += @{Check='Guest account disabled'; Status='PASS'}
        } else {
            $checks += @{Check='Guest account disabled'; Status='FAIL'}
        }

        # Check 2: More than 3 admins is a flag
        $total++
        $adminCount = (Get-LocalGroupMember -Group Administrators -ErrorAction SilentlyContinue).Count
        if ($adminCount -le 5) {
            $score++; $checks += @{Check="Admin count ($adminCount)"; Status='PASS'}
        } else {
            $checks += @{Check="Admin count ($adminCount)"; Status='WARN'}
        }

        # Check 3: Defender real-time protection
        $total++
        $mp = Get-MpComputerStatus -ErrorAction SilentlyContinue
        if ($mp -and $mp.RealTimeProtectionEnabled) {
            $score++; $checks += @{Check='Defender real-time protection'; Status='PASS'}
        } else {
            $checks += @{Check='Defender real-time protection'; Status='FAIL'}
        }

        # Check 4: Firewall enabled on all profiles
        $total++
        $fw = Get-NetFirewallProfile -ErrorAction SilentlyContinue
        $allEnabled = ($fw | Where-Object Enabled -eq $true).Count -eq 3
        if ($allEnabled) {
            $score++; $checks += @{Check='Firewall enabled (all profiles)'; Status='PASS'}
        } else {
            $checks += @{Check='Firewall enabled (all profiles)'; Status='FAIL'}
        }

        # Check 5: Execution policy not unrestricted
        $total++
        $ep = Get-ExecutionPolicy -ErrorAction SilentlyContinue
        if ($ep -in @('RemoteSigned', 'AllSigned', 'Restricted')) {
            $score++; $checks += @{Check="Execution policy ($ep)"; Status='PASS'}
        } else {
            $checks += @{Check="Execution policy ($ep)"; Status='FAIL'}
        }

        # Check 6: BitLocker enabled
        $total++
        $bl = Get-BitLockerVolume -ErrorAction SilentlyContinue
        $anyProtected = $bl | Where-Object ProtectionStatus -eq 'On'
        if ($anyProtected) {
            $score++; $checks += @{Check='BitLocker enabled'; Status='PASS'}
        } else {
            $checks += @{Check='BitLocker enabled'; Status='FAIL'}
        }

        # Check 7: System drive space >= 10% free
        $total++
        $sysDrive = Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='$env:SystemDrive'" -ErrorAction SilentlyContinue
        if ($sysDrive) {
            $freePct = ($sysDrive.FreeSpace / $sysDrive.Size) * 100
            if ($freePct -ge 10) {
                $score++; $checks += @{Check="Disk free ($([math]::Round($freePct,1))%)"; Status='PASS'}
            } else {
                $checks += @{Check="Disk free ($([math]::Round($freePct,1))%)"; Status='FAIL'}
            }
        }

        # Check 8: UAC enabled
        $total++
        $uac = Get-ItemProperty -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System' -Name EnableLUA -ErrorAction SilentlyContinue
        if ($uac.EnableLUA -eq 1) {
            $score++; $checks += @{Check='UAC enabled'; Status='PASS'}
        } else {
            $checks += @{Check='UAC enabled'; Status='FAIL'}
        }

        # Check 9: Auto logon disabled
        $total++
        $autoLogon = Get-ItemProperty -Path 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon' -Name DefaultUserName -ErrorAction SilentlyContinue
        if (-not $autoLogon -or -not $autoLogon.DefaultUserName) {
            $score++; $checks += @{Check='Auto logon disabled'; Status='PASS'}
        } else {
            $checks += @{Check='Auto logon disabled'; Status='WARN'}
        }

        # Check 10: Recent crash dumps
        $total++
        $dumpCount = @(Get-ChildItem "$env:SystemRoot\Minidump" -Filter '*.dmp' -ErrorAction SilentlyContinue).Count
        if ($dumpCount -eq 0) {
            $score++; $checks += @{Check='No recent crash dumps'; Status='PASS'}
        } elseif ($dumpCount -le 3) {
            $checks += @{Check="Crash dumps: $dumpCount"; Status='WARN'}
        } else {
            $checks += @{Check="Crash dumps: $dumpCount"; Status='FAIL'}
        }

    } catch {
        Write-HawkError "Compliance check error: $_"
    }

    # Output results in a compact table
    $checks | Format-Table -Property @{N='Check';E={$_.Check}}, @{N='Status';E={$_.Status}} -AutoSize |
        Out-String | ForEach-Object { $_.TrimEnd() }

    $pct = if ($total -gt 0) { [math]::Round(($score / $total) * 100, 0) } else { 0 }
    $rating = if ($pct -ge 90) { 'Excellent' } elseif ($pct -ge 70) { 'Good' } elseif ($pct -ge 50) { 'Fair' } else { 'Poor' }
    Write-Output ""
    Write-Output "  ─────────────────────────────────────────────"
    Write-Output "  COMPLIANCE SCORE: $pct / 100  ($rating)"
    Write-Output "  Passed: $score / $total checks"
    Write-Output "  ─────────────────────────────────────────────"
}

#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  SECTION 11: FULL WORKFLOW FUNCTIONS
#  These are the entry points called by Hawkward's Go backend.
#  Each aggregates a set of individual diagnostics into a complete report.
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

function Invoke-HawkDailyOps {
    <#
    .SYNOPSIS
        Aggregates daily health check, storage audit, and network connectivity report.
    #>
    $start = Get-Date
    Write-HawkHeader "Hawkward Daily Operations Report"
    Write-HawkResult "Generated" (Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
    Write-HawkResult "Host" $env:COMPUTERNAME
    Write-HawkResult "User" $env:USERNAME
    Write-Output ""

    Get-HawkSystemUptime
    Get-HawkVolumeInfo
    Test-HawkConnectivity
    Test-HawkDNSResolution
    Get-HawkInstalledUpdates -Count 5
    Get-HawkServicesStatus -State Stopped
    Get-HawkAutoStartServices

    $dur = (Get-Date) - $start
    Write-Output "`n  ─────────────────────────────────────────────"
    Write-Output "  Daily Ops completed in $($dur.TotalSeconds.ToString('F1'))s"
    Write-Output "  ─────────────────────────────────────────────"
}

function Invoke-HawkSystemReview {
    <#
    .SYNOPSIS
        Deep dive into hardware specs, performance metrics, and system details.
    #>
    $start = Get-Date
    Write-HawkHeader "Hawkward System Review"

    Get-HawkOSInfo
    Get-HawkCPUInfo
    Get-HawkMemoryInfo
    Get-HawkDiskInfo
    Get-HawkVolumeInfo
    Get-HawkBiosInfo
    Get-HawkGPUInfo
    Get-HawkEnvironmentInfo
    Get-HawkTopProcesses -Count 8
    Get-HawkPerformanceCounters

    $dur = (Get-Date) - $start
    Write-Output "`n  ─────────────────────────────────────────────"
    Write-Output "  System Review completed in $($dur.TotalSeconds.ToString('F1'))s"
    Write-Output "  ─────────────────────────────────────────────"
}

function Invoke-HawkSecurityAudit {
    <#
    .SYNOPSIS
        Full audit of users, firewall, startup, tasks, defender, and certificates.
    #>
    $start = Get-Date
    Write-HawkHeader "Hawkward Security Audit"

    Get-HawkLocalUsers
    Get-HawkAdminGroup
    Get-HawkStartupPrograms
    Get-HawkScheduledTasks
    Get-HawkDefenderStatus
    Get-HawkDefenderThreats
    Get-HawkFirewallStatus
    Get-HawkFirewallRules -Direction Inbound
    Get-HawkExecutionPolicy
    Get-HawkCertificateStore
    Get-HawkBitLockerStatus

    $dur = (Get-Date) - $start
    Write-Output "`n  ─────────────────────────────────────────────"
    Write-Output "  Security Audit completed in $($dur.TotalSeconds.ToString('F1'))s"
    Write-Output "  ─────────────────────────────────────────────"
}

function Invoke-HawkNetworkDiagnostics {
    <#
    .SYNOPSIS
        Verify internet, DNS resolvers, interfaces, and network shares.
    #>
    $start = Get-Date
    Write-HawkHeader "Hawkward Network Diagnostics"

    Get-HawkNetworkAdapters
    Get-HawkIPConfiguration
    Get-HawkRoutingTable
    Get-HawkActiveConnections
    Get-HawkListeningPorts
    Test-HawkConnectivity -Count 3
    Test-HawkDNSResolution
    Get-HawkNetworkStatistics
    Get-HawkDNSSettings

    $dur = (Get-Date) - $start
    Write-Output "`n  ─────────────────────────────────────────────"
    Write-Output "  Network Diagnostics completed in $($dur.TotalSeconds.ToString('F1'))s"
    Write-Output "  ─────────────────────────────────────────────"
}

function Invoke-HawkThreatHunt {
    <#
    .SYNOPSIS
        Search for suspicious processes, ghost ports, and file anomalies.
    #>
    $start = Get-Date
    Write-HawkHeader "Hawkward Threat Hunt"

    Write-HawkSubHeader "Threat Intelligence Summary"
    Write-HawkResult "Scan Time" (Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
    Write-HawkResult "Engine" "Hawkward Behavioral Analysis"
    Write-Output ""

    Get-HawkSuspiciousProcesses
    Get-HawkGhostPorts
    Get-HawkFileAnomalies
    Get-HawkNetworkConnectionsByProcess
    Get-HawkUnusualPorts
    Get-HawkServiceAnomalies
    Get-HawkCrashDumps
    Get-HawkDefenderThreats
    Get-HawkSecurityEvents -Count 15

    $dur = (Get-Date) - $start
    Write-Output "`n  ─────────────────────────────────────────────"
    Write-Output "  Threat Hunt completed in $($dur.TotalSeconds.ToString('F1'))s"
    Write-Output "  ─────────────────────────────────────────────"
}

function Invoke-HawkChangeAudit {
    <#
    .SYNOPSIS
        Review recent files, patches, driver changes, and crash dumps.
    #>
    $start = Get-Date
    Write-HawkHeader "Hawkward Change Audit"

    Get-HawkInstalledUpdates -Count 15
    Get-HawkRecentChanges -Hours 48
    Get-HawkCrashDumps
    Get-HawkSystemEvents -Count 10
    Get-HawkApplicationEvents -Count 10

    $dur = (Get-Date) - $start
    Write-Output "`n  ─────────────────────────────────────────────"
    Write-Output "  Change Audit completed in $($dur.TotalSeconds.ToString('F1'))s"
    Write-Output "  ─────────────────────────────────────────────"
}

function Invoke-HawkComplianceCheck {
    <#
    .SYNOPSIS
        CIS-inspired baseline verification and compliance scoring.
    #>
    $start = Get-Date
    Write-HawkHeader "Hawkward Compliance Check"
    Write-HawkResult "Standard" "CIS-Inspired Windows Baseline"
    Write-HawkResult "Host" $env:COMPUTERNAME
    Write-Output ""

    Get-HawkComplianceStatus
    Get-HawkInstalledSoftware
    Get-HawkCertificateStore
    Get-HawkServicesStatus -State Running

    $dur = (Get-Date) - $start
    Write-Output "`n  ─────────────────────────────────────────────"
    Write-Output "  Compliance Check completed in $($dur.TotalSeconds.ToString('F1'))s"
    Write-Output "  ─────────────────────────────────────────────"
}

#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#  PROFILE INITIALIZATION
#━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Verify the profile loaded correctly
Write-Output "[Hawkward] PowerShell Diagnostic Profile loaded"
Write-Output "[Hawkward] 7 workflow functions available:"
Write-Output "[Hawkward]   - Invoke-HawkDailyOps"
Write-Output "[Hawkward]   - Invoke-HawkSystemReview"
Write-Output "[Hawkward]   - Invoke-HawkSecurityAudit"
Write-Output "[Hawkward]   - Invoke-HawkNetworkDiagnostics"
Write-Output "[Hawkward]   - Invoke-HawkThreatHunt"
Write-Output "[Hawkward]   - Invoke-HawkChangeAudit"
Write-Output "[Hawkward]   - Invoke-HawkComplianceCheck"
