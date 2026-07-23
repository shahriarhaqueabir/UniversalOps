package common

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ── LHM (LibreHardwareMonitor) Manager ──────────────────────────────────────
//
// Manages the bundled LibreHardwareMonitor process lifecycle:
//   - Downloads the portable binary on first use (user consent required)
//   - Starts with admin elevation via UAC when hardware sensors are needed
//   - Stops gracefully on application shutdown
//   - Reports status to the frontend for transparent UX

const (
	lhmVersion     = "0.9.6"
	lhmZipName     = "LibreHardwareMonitor-net472.zip"
	lhmDownloadURL = "https://github.com/LibreHardwareMonitor/LibreHardwareMonitor/releases/download/v" +
		lhmVersion + "/" + lhmZipName
	lhmExeName     = "LibreHardwareMonitor.exe"
	lhmProcessName = "LibreHardwareMonitor"
	lhmLicenseURL  = "https://raw.githubusercontent.com/LibreHardwareMonitor/LibreHardwareMonitor/v" +
		lhmVersion + "/LICENSE"

	// LHMHTTPPort is the default HTTP server port LHM exposes sensor data on.
	LHMHTTPPort = 8085
)

// LHMStatus represents the current state of the bundled LHM instance.
type LHMStatus struct {
	Available  bool   `json:"available"`  // Binary exists on disk
	Running    bool   `json:"running"`    // Process is alive
	NeedsAdmin bool   `json:"needsAdmin"` // Requires elevation to start
	Version    string `json:"version"`    // Bundled version string
	Path       string `json:"path"`       // Path to the binary
	Error      string `json:"error,omitempty"`
}

// LHMManager manages the LibreHardwareMonitor process lifecycle.
type LHMManager struct {
	mu      sync.Mutex
	dirPath string // Directory where LHM is extracted
}

var (
	lhmInstance *LHMManager
	lhmOnce     sync.Once
)

// GetLHMManager returns the singleton LHM manager.
func GetLHMManager() *LHMManager {
	lhmOnce.Do(func() {
		dataDir, _ := ConfigDir()
		lhmDir := filepath.Join(dataDir, "lhm")
		lhmInstance = &LHMManager{
			dirPath: lhmDir,
		}
	})
	return lhmInstance
}

// exePath returns the full path to LibreHardwareMonitor.exe.
func (m *LHMManager) exePath() string {
	return filepath.Join(m.dirPath, lhmExeName)
}

// BinaryPath returns the full path to the LHM executable (exported).
func (m *LHMManager) BinaryPath() string {
	return m.exePath()
}

// DirPath returns the directory where LHM is extracted (exported).
func (m *LHMManager) DirPath() string {
	return m.dirPath
}

// licensePath returns the path to the MPL-2.0 license file.
func (m *LHMManager) licensePath() string {
	return filepath.Join(m.dirPath, "LICENSE")
}

// ── Status ──────────────────────────────────────────────────────────────────

// Status returns the current LHM state for the frontend.
func (m *LHMManager) Status() LHMStatus {
	return LHMStatus{
		Available:  m.IsAvailable(),
		Running:    m.IsRunning(),
		NeedsAdmin: !m.IsRunning() && m.IsAvailable(),
		Version:    lhmVersion,
		Path:       m.exePath(),
	}
}

// IsAvailable returns true if the LHM binary exists on disk.
func (m *LHMManager) IsAvailable() bool {
	info, err := os.Stat(m.exePath())
	return err == nil && info.Size() > 1024*1024 // Sanity: exe should be >1MB
}

// IsRunning returns true if the LHM process is currently alive.
func (m *LHMManager) IsRunning() bool {
	// Use WMI to detect — avoids false positives from partial name matches.
	type Win32Process struct {
		Name string
	}
	var dst []Win32Process
	_ = WMIQueryWithTimeout(
		fmt.Sprintf("SELECT Name FROM Win32_Process WHERE Name='%s.exe'", lhmProcessName),
		&dst,
		2*time.Second,
	)
	return len(dst) > 0
}

// ── Download ────────────────────────────────────────────────────────────────

// Download fetches the LHM portable zip to the data directory.
// Shows progress via callback. Blocks until complete or ctx is cancelled.
func (m *LHMManager) Download(ctx context.Context, progress func(downloaded int64, total int64)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.IsAvailable() {
		return nil // Already have it
	}

	if err := os.MkdirAll(m.dirPath, 0755); err != nil {
		return fmt.Errorf("create lhm dir: %w", err)
	}

	zipPath := filepath.Join(m.dirPath, lhmZipName)

	LogInfo("LHM: Downloading v%s from GitHub...", lhmVersion)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lhmDownloadURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Use a client with a reasonable timeout to prevent hanging
	dlClient := &http.Client{Timeout: 10 * time.Minute}
	resp, err := dlClient.Do(req)
	if err != nil {
		return fmt.Errorf("download LHM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download LHM: HTTP %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength

	out, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("create zip file: %w", err)
	}
	defer func() {
		out.Close()
		// Clean up partial zip on error
		if err != nil {
			_ = os.Remove(zipPath)
		}
	}()

	if progress != nil {
		progress(0, totalSize)
	}

	buf := make([]byte, 32*1024)
	var downloaded int64
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				err = writeErr
				return fmt.Errorf("write zip: %w", err)
			}
			downloaded += int64(n)
			if progress != nil {
				progress(downloaded, totalSize)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			err = readErr
			return fmt.Errorf("read download: %w", err)
		}
	}

	LogInfo("LHM: Download complete (%d bytes), extracting...", downloaded)

	// Extract
	if err = m.extractZip(zipPath); err != nil {
		return fmt.Errorf("extract LHM: %w", err)
	}

	// Download license for MPL-2.0 compliance
	_ = m.downloadLicense(ctx)

	// Remove zip after successful extraction
	_ = os.Remove(zipPath)

	LogInfo("LHM: v%s ready at %s", lhmVersion, m.exePath())
	return nil
}

func (m *LHMManager) extractZip(zipPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Flatten: skip top-level directory prefix if present
		name := f.Name
		if idx := strings.IndexByte(name, '/'); idx >= 0 && idx < len(name)-1 {
			name = name[idx+1:]
		}
		if name == "" {
			continue
		}

		target := filepath.Join(m.dirPath, filepath.FromSlash(name))

		// Guard against zip-slip
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(m.dirPath)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *LHMManager) downloadLicense(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lhmLicenseURL, nil)
	if err != nil {
		return err
	}
	licClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := licClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	out, err := os.Create(m.licensePath())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

// ── Start / Stop ────────────────────────────────────────────────────────────

// Start launches LibreHardwareMonitor with admin elevation (UAC prompt).
// The process runs hidden and is tracked for cleanup on shutdown.
// Returns an error if the user denies elevation or if LHM is not available.
func (m *LHMManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.IsRunning() {
		LogInfo("LHM: Already running")
		return nil
	}

	if !m.IsAvailable() {
		return fmt.Errorf("LHM binary not found at %s — download first", m.exePath())
	}

	// Use PowerShell Start-Process -Verb RunAs to trigger UAC elevation.
	// -WindowStyle Hidden prevents the LHM GUI window from flashing on screen.
	// The UAC dialog will show: program name, publisher, and source.
	psScript := fmt.Sprintf(
		"Start-Process -FilePath '%s' -Verb RunAs -WindowStyle Hidden",
		m.exePath(),
	)

	LogInfo("LHM: Requesting admin elevation via UAC...")
	cmd := HiddenCommand("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Common case: user clicked No on UAC prompt
		LogWarn("LHM: Admin elevation denied or failed: %v (output: %s)", err, strings.TrimSpace(string(output)))
		return fmt.Errorf("admin elevation required — user denied or LHM failed to start: %w", err)
	}

	// Give LHM a moment to initialize its WMI provider
	time.Sleep(3 * time.Second)

	if !m.IsRunning() {
		return fmt.Errorf("LHM process did not start — check if PawnIO driver is installed")
	}

	LogInfo("LHM: Started successfully with admin privileges")
	return nil
}

// Stop terminates the LHM process gracefully.
func (m *LHMManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.IsRunning() {
		return nil
	}

	LogInfo("LHM: Stopping process...")

	// Find PID via WMI
	type Win32Process struct {
		ProcessId int32
		Name      string
	}
	var dst []Win32Process
	_ = WMIQueryWithTimeout(
		fmt.Sprintf("SELECT ProcessId, Name FROM Win32_Process WHERE Name='%s.exe'", lhmProcessName),
		&dst,
		2*time.Second,
	)

	for _, p := range dst {
		// Use taskkill /F for reliable termination (LHM may hold kernel handles)
		killCmd := exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%d", uint32(p.ProcessId)))
		_ = killCmd.Run()
		LogInfo("LHM: Terminated process PID %d", p.ProcessId)
	}

	// Brief pause to let WMI provider unload
	time.Sleep(1 * time.Second)
	return nil
}

// ── Ensure (Download + Start) ───────────────────────────────────────────────

// EnsureLHM guarantees that LHM is available and running.
// If not downloaded, it downloads first (blocking).
// If not running, it attempts to start with admin elevation.
// The userConsent callback is invoked before download — return false to abort.
func (m *LHMManager) Ensure(ctx context.Context, userConsent func() bool) error {
	// Already running? Done.
	if m.IsRunning() {
		return nil
	}

	// Not available yet? Download with user consent.
	if !m.IsAvailable() {
		if userConsent != nil && !userConsent() {
			return fmt.Errorf("user declined LHM download")
		}
		if err := m.Download(ctx, nil); err != nil {
			return fmt.Errorf("download failed: %w", err)
		}
	}

	// Start with admin elevation
	return m.Start()
}

// Remove deletes the bundled LHM installation.
func (m *LHMManager) Remove() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.IsRunning() {
		_ = m.Stop()
	}

	return os.RemoveAll(m.dirPath)
}
