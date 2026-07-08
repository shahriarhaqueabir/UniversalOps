package common

import (
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

	// sandboxPollInterval is how often the job assigner polls for new processes.
	sandboxPollInterval = 25 * time.Millisecond

	// sandboxMaxWait is the maximum time the assigner waits for a process to start.
	sandboxMaxWait = 2 * time.Second

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
	cmd *exec.Cmd
	age time.Time
}

// Global state for background job assignment and cleanup.
var (
	pendingJobs   []*activeJob
	pendingJobsMu sync.Mutex
	assignerOnce  sync.Once
)

// startJobAssigner launches the background goroutine that assigns job objects
// to processes after they are started by cmd.Start().
func startJobAssigner() {
	assignerOnce.Do(func() {
		go jobAssignerLoop()
	})
}

// jobAssignerLoop runs in a background goroutine. It polls the pendingJobs
// list, assigns processes to their job objects once they start, and removes
// them from the pending list.
func jobAssignerLoop() {
	for {
		time.Sleep(sandboxPollInterval)

		pendingJobsMu.Lock()
		if len(pendingJobs) == 0 {
			pendingJobsMu.Unlock()
			continue
		}

		var remaining []*activeJob

		for _, aj := range pendingJobs {
			// Skip if process hasn't started yet.
			if aj.cmd.Process == nil {
				// If we've waited too long, give up and clean up.
				if time.Since(aj.age) > sandboxMaxWait {
					windows.CloseHandle(aj.job)
					LogWarn("sandbox: process never started within %v, closing job handle", sandboxMaxWait)
				} else {
					remaining = append(remaining, aj)
				}
				continue
			}

			pid := aj.cmd.Process.Pid
			if pid == 0 {
				windows.CloseHandle(aj.job)
				continue
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
				continue
			}

			// Check if the process has already exited before we can assign the job.
			var exitCode uint32
			if err := windows.GetExitCodeProcess(ph, &exitCode); err != nil || exitCode != stillActive {
				windows.CloseHandle(ph)
				windows.CloseHandle(aj.job)
				continue
			}

			// Assign the process to the job object.
			if err := windows.AssignProcessToJobObject(aj.job, ph); err != nil {
				LogWarn("sandbox: failed to assign process %d to job: %v", pid, err)
				windows.CloseHandle(ph)
				windows.CloseHandle(aj.job)
				continue
			}

			// Configure job limits.
			configureJobLimits(aj.job, aj.cfg)

			// Start cleanup goroutine that closes the job handle when the process exits.
			go cleanupJobHandle(aj.job, ph)
		}

		pendingJobs = remaining
		pendingJobsMu.Unlock()
	}
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

// applyPlatformSandbox applies Windows sandbox restrictions using Job Objects
// and restricted tokens.
//
// It sets up:
//   - Restricted token (when DropPrivileges is true) that disables dangerous
//     privileges while keeping benign ones like SeChangeNotifyPrivilege active
//   - Job Object (when DenyProcessSpawn is true) to prevent child process creation
//     and limit memory usage
//
// Job Object assignment happens asynchronously after cmd.Start() is called
// (via a background goroutine). If sandboxing fails, a warning is logged but
// the command still runs — this is defense-in-depth, not a security boundary.
func applyPlatformSandbox(cmd *exec.Cmd, cfg SandboxConfig) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	sa := cmd.SysProcAttr

	// Basic process isolation: hide the window when network or process
	// isolation is requested. Note: we intentionally do NOT set
	// NoInheritHandles because that prevents stdout/stderr pipes from
	// being passed to the child process on Windows, breaking
	// CombinedOutput() and similar methods. The Job Object provides
	// the real security boundary.
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

	// Create a Job Object for process isolation.
	if cfg.DenyProcessSpawn {
		startJobAssigner()

		job, err := windows.CreateJobObject(nil, nil)
		if err != nil {
			LogWarn("sandbox: failed to create job object: %v (running without process isolation)", err)
			return
		}

		entry := &activeJob{
			job: job,
			cfg: cfg,
			cmd: cmd,
			age: time.Now(),
		}

		pendingJobsMu.Lock()
		pendingJobs = append(pendingJobs, entry)
		pendingJobsMu.Unlock()
	}
}
