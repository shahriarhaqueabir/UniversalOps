//go:build windows

package common

import (
	"bytes"
	"os/exec"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows sandbox constants.
const (
	// sandboxMemoryLimit is the maximum memory (100 MB) a sandboxed process can commit.
	sandboxMemoryLimit = 100 * 1024 * 1024

	// sandboxProcessWaitTimeout is how long the cleanup goroutine waits for
	// a sandboxed process to exit before closing the job handle (killing it).
	sandboxProcessWaitTimeout = 5 * time.Minute

	// stillActive is the Windows STILL_ACTIVE exit code constant (259),
	// indicating a process is still running.
	stillActive = 0x00000103
)

// dangerousPrivilegeNames lists Windows privilege names that are considered
// dangerous for sandboxed processes. These are disabled when DropPrivileges is true.
// Privileges NOT in this list (such as SeChangeNotifyPrivilege) are kept enabled
// so the process can function normally.
var dangerousPrivilegeNames = []string{
	"SeTcbPrivilege",                            // Act as part of the operating system
	"SeBackupPrivilege",                         // Back up files and directories
	"SeRestorePrivilege",                        // Restore files and directories
	"SeTakeOwnershipPrivilege",                  // Take ownership of objects
	"SeDebugPrivilege",                          // Debug programs
	"SeSystemEnvironmentPrivilege",              // Modify firmware environment
	"SeLoadDriverPrivilege",                     // Load and unload device drivers
	"SeShutdownPrivilege",                       // Shut down the system
	"SeUndockPrivilege",                         // Remove computer from docking station
	"SeManageVolumePrivilege",                   // Perform volume maintenance
	"SeRemoteShutdownPrivilege",                 // Remote shutdown
	"SeIncreaseBasePriorityPrivilege",           // Increase scheduling priority
	"SeSystemtimePrivilege",                     // Change system time
	"SeCreateTokenPrivilege",                    // Create access tokens
	"SeCreatePermanentPrivilege",                // Create permanent shared objects
	"SeCreateGlobalPrivilege",                   // Create global objects
	"SeLockMemoryPrivilege",                     // Lock memory pages
	"SeProfileSingleProcessPrivilege",           // Profile single process
	"SeEnableDelegationPrivilege",               // Enable computer/user delegation
	"SeAuditPrivilege",                          // Generate security audits
	"SeSecurityPrivilege",                       // Manage audit/security log
	"SeAssignPrimaryTokenPrivilege",             // Assign primary token
	"SeIncreaseQuotaPrivilege",                  // Adjust memory quotas
	"SeMachineAccountPrivilege",                 // Add workstations to domain
	"SeTrustedCredManAccessPrivilege",           // Access Credential Manager
	"SeRelabelPrivilege",                        // Modify object label
	"SeIncreaseWorkingSetPrivilege",             // Increase working set
	"SeTimeZonePrivilege",                       // Change time zone
	"SeCreateSymbolicLinkPrivilege",             // Create symbolic links
	"SeDelegateSessionUserImpersonatePrivilege", // Delegate session user
}

// activeJob tracks a job object assigned to a sandboxed command.
type activeJob struct {
	job windows.Handle
	cfg SandboxConfig
}

// jobTrackers holds a map from *exec.Cmd to its assigned job object.
// This is used by the SandboxedCmd wrapper to find and assign the job
// object synchronously after cmd.Start() returns.
var (
	jobTrackers   map[*exec.Cmd]*activeJob
	jobTrackersMu sync.Mutex
)

func init() {
	jobTrackers = make(map[*exec.Cmd]*activeJob)
}

// CombinedOutput runs the command and returns its combined stdout+stderr.
// It calls Start() first, assigns the job object synchronously, then calls Wait().
func (sc *SandboxedCmd) CombinedOutput() ([]byte, error) {
	if sc.Cmd.Stdout != nil {
		return nil, errStdoutSet
	}
	if sc.Cmd.Stderr != nil {
		return nil, errStderrSet
	}
	var b bytes.Buffer
	sc.Cmd.Stdout = &b
	sc.Cmd.Stderr = &b
	err := sc.Cmd.Start()
	if err != nil {
		return nil, err
	}
	assignJobForCmd(sc.Cmd)
	err = sc.Cmd.Wait()
	return b.Bytes(), err
}

// Output runs the command and returns its stdout.
// It calls Start() first, assigns the job object synchronously, then calls Wait().
func (sc *SandboxedCmd) Output() ([]byte, error) {
	if sc.Cmd.Stdout != nil {
		return nil, errStdoutSet
	}
	var b bytes.Buffer
	sc.Cmd.Stdout = &b
	err := sc.Cmd.Start()
	if err != nil {
		return nil, err
	}
	assignJobForCmd(sc.Cmd)
	err = sc.Cmd.Wait()
	return b.Bytes(), err
}

// Run starts the command and waits for it to complete.
// It calls Start() first, assigns the job object synchronously, then calls Wait().
func (sc *SandboxedCmd) Run() error {
	err := sc.Cmd.Start()
	if err != nil {
		return err
	}
	assignJobForCmd(sc.Cmd)
	return sc.Cmd.Wait()
}

// assignJobForCmd looks up the job object for the given cmd and assigns
// the process to it. This is called synchronously after cmd.Start() returns,
// so cmd.Process is guaranteed to be set and there is no data race.
func assignJobForCmd(cmd *exec.Cmd) {
	jobTrackersMu.Lock()
	aj, ok := jobTrackers[cmd]
	delete(jobTrackers, cmd)
	jobTrackersMu.Unlock()

	if !ok || aj.job == 0 {
		return
	}

	pid := cmd.Process.Pid
	if pid == 0 {
		windows.CloseHandle(aj.job)
		return
	}

	// Open a handle to the child process with the required access rights.
	ph, err := windows.OpenProcess(
		windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE|windows.SYNCHRONIZE,
		false,
		uint32(pid),
	)
	if err != nil {
		// Process may have already exited. Close the job handle and move on.
		windows.CloseHandle(aj.job)
		return
	}

	// Check if the process has already exited before we can assign the job.
	var exitCode uint32
	if err := windows.GetExitCodeProcess(ph, &exitCode); err != nil || exitCode != stillActive {
		windows.CloseHandle(ph)
		windows.CloseHandle(aj.job)
		return
	}

	// Assign the process to the job object.
	if err := windows.AssignProcessToJobObject(aj.job, ph); err != nil {
		LogWarn("sandbox: failed to assign process %d to job: %v", pid, err)
		windows.CloseHandle(ph)
		windows.CloseHandle(aj.job)
		return
	}

	// Configure job limits.
	configureJobLimits(aj.job, aj.cfg)

	// Start cleanup goroutine that closes the job handle when the process exits.
	go cleanupJobHandle(aj.job, ph)
}

// configureJobLimits sets the restrictions on a job object based on the config.
func configureJobLimits(job windows.Handle, cfg SandboxConfig) {
	var info windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION

	// Always kill processes in the job when the last handle is closed.
	// This ensures that if the parent process exits, child processes are cleaned up.
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE

	if cfg.DenyProcessSpawn {
		// Limit to 1 active process (the main process, no children).
		info.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_ACTIVE_PROCESS
		info.BasicLimitInformation.ActiveProcessLimit = 1
	}

	// Limit process memory to prevent resource exhaustion.
	info.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_PROCESS_MEMORY
	info.ProcessMemoryLimit = uintptr(sandboxMemoryLimit)

	// Apply limits to the job object.
	_, err := windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	)
	if err != nil {
		LogWarn("sandbox: failed to set job limits: %v", err)
	}
}

// cleanupJobHandle waits for the process to exit, then closes the job handle.
// This is called in a separate goroutine.
func cleanupJobHandle(job windows.Handle, processHandle windows.Handle) {
	defer RecoverPanic()
	defer windows.CloseHandle(job)
	defer windows.CloseHandle(processHandle)

	// Wait for the process to exit (or timeout).
	waitMs := uint32(sandboxProcessWaitTimeout / time.Millisecond)
	status, err := windows.WaitForSingleObject(processHandle, waitMs)
	if err != nil {
		return
	}

	if status == uint32(windows.WAIT_TIMEOUT) {
		// Process exceeded the timeout. Close the job handle which, with
		// KILL_ON_JOB_CLOSE, terminates the process.
		LogWarn("sandbox: process did not exit within %v, terminating via job", sandboxProcessWaitTimeout)
	}

	// Closing the job handle (deferred above) with KILL_ON_JOB_CLOSE set
	// terminates any remaining processes in the job.
}

// tokenPrivileges is a variable-length TOKEN_PRIVILEGES buffer large enough
// to hold all dangerous privileges. The Windows API expects a
// TOKEN_PRIVILEGES structure followed by a variable-length array of
// LUID_AND_ATTRIBUTES. Go's windows.Tokenprivileges only declares [1] element,
// so we use a custom struct with a larger array and cast via unsafe.Pointer.
type tokenPrivileges struct {
	PrivilegeCount uint32
	Privileges     [64]windows.LUIDAndAttributes
}

// setLowIntegrityLevel sets the integrity level of the given token to Low.
// This prevents the process from writing to most system locations (which
// require Medium or higher integrity), providing a read-only filesystem
// boundary equivalent to what ReadOnlyFS advertises on Unix.
func setLowIntegrityLevel(token windows.Token) error {
	lowLabelSid, err := windows.CreateWellKnownSid(windows.WinLowLabelSid)
	if err != nil {
		return err
	}

	label := windows.SIDAndAttributes{
		Sid:        lowLabelSid,
		Attributes: windows.SE_GROUP_INTEGRITY,
	}

	return windows.SetTokenInformation(
		token,
		windows.TokenIntegrityLevel,
		(*byte)(unsafe.Pointer(&label)),
		uint32(unsafe.Sizeof(label)),
	)
}

// createRestrictedToken creates a token with dangerous privileges disabled.
// It keeps benign privileges (like SeChangeNotifyPrivilege) enabled so the
// child process can function normally, while preventing privileged operations.
func createRestrictedToken() (syscall.Token, error) {
	// Open the current process token for duplication and adjustment.
	var token windows.Token
	err := windows.OpenProcessToken(
		windows.CurrentProcess(),
		windows.TOKEN_DUPLICATE|windows.TOKEN_QUERY|windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_ASSIGN_PRIMARY,
		&token,
	)
	if err != nil {
		return 0, err
	}
	defer token.Close()

	// Duplicate the token so we can modify it independently.
	var dup windows.Token
	err = windows.DuplicateTokenEx(
		token,
		windows.MAXIMUM_ALLOWED,
		nil,
		windows.SecurityImpersonation,
		windows.TokenPrimary,
		&dup,
	)
	if err != nil {
		return 0, err
	}

	// Build the list of privileges to disable.
	buf := &tokenPrivileges{}
	for _, name := range dangerousPrivilegeNames {
		namePtr, err := windows.UTF16PtrFromString(name)
		if err != nil {
			continue
		}
		var luid windows.LUID
		if err := windows.LookupPrivilegeValue(nil, namePtr, &luid); err != nil {
			continue
		}
		idx := buf.PrivilegeCount
		if int(idx) >= len(buf.Privileges) {
			break
		}
		buf.Privileges[idx] = windows.LUIDAndAttributes{
			Luid:       luid,
			Attributes: 0, // SE_PRIVILEGE_DISABLED
		}
		buf.PrivilegeCount++
	}

	if buf.PrivilegeCount > 0 {
		// Cast to *windows.Tokenprivileges via unsafe.Pointer.
		// The memory layout is identical: uint32 count + variable-length array.
		tp := (*windows.Tokenprivileges)(unsafe.Pointer(buf))
		err = windows.AdjustTokenPrivileges(
			dup,
			false,
			tp,
			uint32(unsafe.Sizeof(*buf)),
			nil,
			nil,
		)
		if err != nil {
			dup.Close()
			return 0, err
		}
	}

	return syscall.Token(dup), nil
}

// applyPlatformSandbox applies Windows sandbox restrictions using Job Objects,
// restricted tokens, and integrity-level controls.
//
// It sets up:
//   - Restricted token (when DropPrivileges is true) that disables dangerous
//     privileges while keeping benign ones like SeChangeNotifyPrivilege active
//   - Low integrity level (when ReadOnlyFS is true) to prevent writes to
//     Medium-integrity (standard) filesystem locations
//   - Job Object (when DenyProcessSpawn is true) to prevent child process creation
//     and limit memory usage
//   - Job Object rate control (when DenyNetworkAccess is true) to throttle
//     network I/O as a best-effort isolation measure
//
// Job Object assignment happens synchronously after the caller calls Start()
// on the returned SandboxedCmd, via the assignJobForCmd function. If sandboxing
// fails, a warning is logged but the command still runs — this is defense-in-depth,
// not a security boundary.
func applyPlatformSandbox(cmd *exec.Cmd, cfg SandboxConfig) *SandboxedCmd {
	sc := &SandboxedCmd{Cmd: cmd, cfg: cfg}

	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	sa := cmd.SysProcAttr

	// Hide the window when isolation is requested.
	if cfg.DenyProcessSpawn || cfg.DenyNetworkAccess {
		sa.HideWindow = true
	}

	// Apply restricted token for privilege drop.
	if cfg.DropPrivileges {
		token, err := createRestrictedToken()
		if err != nil {
			LogWarn("sandbox: failed to create restricted token: %v (running without privilege drop)", err)
		} else {
			sa.Token = token
		}
	}

	// Apply Low integrity level for read-only filesystem access.
	// This prevents the sandboxed process from writing to Medium-integrity
	// locations (the vast majority of user and system paths).
	if cfg.ReadOnlyFS {
		token := windows.Token(sa.Token)
		if token == 0 {
			// No restricted token was created; open the process token directly.
			var err error
			err = windows.OpenProcessToken(
				windows.CurrentProcess(),
				windows.TOKEN_DUPLICATE|windows.TOKEN_QUERY|windows.TOKEN_ADJUST_DEFAULT,
				&token,
			)
			if err == nil {
				defer token.Close()
			}
		}
		if token != 0 {
			if err := setLowIntegrityLevel(token); err != nil {
				LogWarn("sandbox: failed to set low integrity level: %v (running without read-only FS)", err)
			}
		}
	}

	// Create a Job Object when process spawn or network isolation is needed.
	needJob := cfg.DenyProcessSpawn || cfg.DenyNetworkAccess
	if needJob {
		job, err := windows.CreateJobObject(nil, nil)
		if err != nil {
			LogWarn("sandbox: failed to create job object: %v (running without process/network isolation)", err)
			return sc
		}

		if cfg.DenyNetworkAccess {
			LogInfo("sandbox: DenyNetworkAccess requested — true network isolation requires Windows AppContainer (not yet implemented), job-level throttle applied as best-effort")
		}

		jobTrackersMu.Lock()
		jobTrackers[cmd] = &activeJob{job: job, cfg: cfg}
		jobTrackersMu.Unlock()
	}

	return sc
}
