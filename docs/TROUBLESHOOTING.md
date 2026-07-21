# Troubleshooting Guide

## GPU Detection Issues

### Symptom: GPU shows "N/A" or "No GPU detected"

The application uses a multi-stage detection chain to identify GPUs. If no GPU is shown, it means all stages failed to return a valid, non-virtual hardware identifier.

#### The Detection Chain
1.  **WMI `Win32_VideoController`**: Primary source for all GPUs (integrated and discrete).
2.  **PowerShell `Get-CimInstance`**: Fallback if WMI queries fail.
3.  **`nvidia-smi`**: Real-time utilization and temperature for NVIDIA GPUs only.
4.  **`LibreHardwareMonitor`**: Real-time sensors for AMD/Intel GPUs (optional external dependency).

#### Common Reasons for Failure
-   **Virtual Drivers**: The app filters out virtual adapters like "Microsoft Basic Render Driver" or Remote Desktop adapters to prevent clutter.
-   **WMI Corruption**: The Windows Management Instrumentation repository may be corrupted.
-   **Restricted Permissions**: WMI or CIM queries might be blocked by system policy.

### Debugging Steps

Run the following commands in PowerShell to see what the OS reports:

```powershell
# 1. Check all video controllers
Get-CimInstance Win32_VideoController | Select-Object Name, AdapterRAM, DriverVersion

# 2. Check for NVIDIA specifically
nvidia-smi
```

If the first command only returns "Microsoft Basic Render Driver", the app is correctly ignoring it as it's not physical hardware.

### Fix Options

1.  **Install Proprietary Drivers**: Ensure you have the latest drivers from NVIDIA, AMD, or Intel. NVIDIA drivers usually include `nvidia-smi`.
2.  **LibreHardwareMonitor**: For AMD or Intel GPU sensors (temperature, fan speed), download and run [LibreHardwareMonitor](https://github.com/LibreHardwareMonitor/LibreHardwareMonitor). The app will automatically pick up its WMI namespace (`root\LibreHardwareMonitor`).
3.  **Repair WMI**: If WMI is failing globally, try repairing the repository:
    ```cmd
    winmgmt /verifyrepository
    winmgmt /salvagerepository
    ```
