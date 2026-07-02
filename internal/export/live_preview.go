package export

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const livePreviewWorkspace = "name:omarchy-themegen-preview"

type hyprWorkspace struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type hyprMonitor struct {
	Name    string `json:"name"`
	Focused bool   `json:"focused"`
}

type hyprClient struct {
	Address   string        `json:"address"`
	Class     string        `json:"class"`
	Title     string        `json:"title"`
	Workspace hyprWorkspace `json:"workspace"`
}

type previewProcess struct {
	Role     string
	Selector string
	Process  *os.Process
}

func GenerateLiveDesktopPreview(outputPath, themeDir string, width, height int) error {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
		return fmt.Errorf("not running inside Hyprland")
	}
	for _, cmd := range []string{"hyprctl", "grim", "magick", "nvim"} {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("%s is required", cmd)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := applyThemeForLivePreview(ctx, themeDir); err != nil {
		return err
	}
	if err := syncLazyVimForLivePreview(ctx); err != nil {
		return err
	}

	originalWorkspace := activeWorkspace(ctx)
	monitor := focusedMonitor(ctx)
	if err := runLive(ctx, "hyprctl", "dispatch", "workspace", livePreviewWorkspace); err != nil {
		return err
	}
	closePreviewWorkspaceClients(ctx)

	processes, err := launchPreviewApps(ctx, themeDir)
	if err != nil {
		return err
	}
	defer cleanupPreviewWorkspace(originalWorkspace, processes)

	time.Sleep(3500 * time.Millisecond)

	tmp, err := os.CreateTemp("", "omarchy-themegen-live-*.png")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	args := []string{tmpPath}
	if monitor != "" {
		args = []string{"-o", monitor, tmpPath}
	}
	if err := runLive(ctx, "grim", args...); err != nil {
		return fmt.Errorf("capture screenshot: %w", err)
	}
	return runLive(ctx, "magick", tmpPath, "-resize", fmt.Sprintf("%dx%d^", width, height), "-gravity", "center", "-extent", fmt.Sprintf("%dx%d", width, height), outputPath)
}

func applyThemeForLivePreview(ctx context.Context, themeDir string) error {
	themeName := filepath.Base(filepath.Clean(themeDir))
	if themeName == "." || themeName == string(filepath.Separator) || themeName == "" {
		return fmt.Errorf("cannot determine theme name from %s", themeDir)
	}
	if p, err := exec.LookPath("omarchy"); err == nil {
		return runLive(ctx, p, "theme", "set", themeName)
	}
	if p, err := exec.LookPath("omarchy-theme-set"); err == nil {
		return runLive(ctx, p, themeName)
	}
	return fmt.Errorf("omarchy is required to apply the theme before live preview")
}

func syncLazyVimForLivePreview(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "nvim", "--headless", "+Lazy! sync", "+qa")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sync LazyVim before live preview: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func launchPreviewApps(ctx context.Context, themeDir string) ([]previewProcess, error) {
	terminal, err := terminalCommand(themeDir)
	if err != nil {
		return nil, err
	}
	var processes []previewProcess
	open := func(role string, args []string) (*previewProcess, error) {
		if len(args) == 0 {
			return nil, nil
		}
		before := previewClientAddressSet(ctx)
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("launch %s: %w", args[0], err)
		}
		proc := previewProcess{Role: role, Selector: waitForNewPreviewClient(ctx, before, 5*time.Second), Process: cmd.Process}
		processes = append(processes, proc)
		time.Sleep(900 * time.Millisecond)
		return &proc, nil
	}

	nvim, err := open("nvim", terminal("nvim", "omarchy-preview-nvim", "cd ~ && nvim . +'lua vim.schedule(function() if _G.Snacks and Snacks.explorer then Snacks.explorer() return end for _,cmd in ipairs({\"Neotree filesystem reveal\",\"NvimTreeOpen\",\"Oil .\"}) do if pcall(vim.cmd, cmd) then return end end end)'"))
	if err != nil {
		killProcesses(processes)
		return nil, err
	}
	btop, err := open("btop", terminal("btop", "omarchy-preview-btop", "while [ $(tput cols) -lt 80 ] || [ $(tput lines) -lt 24 ]; do clear; printf 'Waiting for preview window size...'; sleep 0.2; done; btop"))
	if err != nil {
		killProcesses(processes)
		return nil, err
	}
	if nvim != nil && nvim.Selector != "" {
		_ = runLive(ctx, "hyprctl", "dispatch", "focuswindow", nvim.Selector)
		_ = runLive(ctx, "hyprctl", "dispatch", "togglesplit")
	}
	if _, err := open("ls", terminal("ls", "omarchy-preview-ls", "cd ~/Projects/GO/omarchy-themegen 2>/dev/null || cd ~; clear && ls -la && sleep 20")); err != nil {
		killProcesses(processes)
		return nil, err
	}
	if btop != nil && btop.Selector != "" {
		_ = runLive(ctx, "hyprctl", "dispatch", "focuswindow", btop.Selector)
		_ = runLive(ctx, "hyprctl", "dispatch", "togglesplit")
	}
	if _, err := open("files", fileBrowserCommand()); err != nil {
		killProcesses(processes)
		return nil, err
	}
	return processes, nil
}

func previewClientAddressSet(ctx context.Context) map[string]bool {
	addresses := make(map[string]bool)
	for _, client := range previewWorkspaceClients(ctx) {
		addresses[client.Address] = true
	}
	return addresses
}

func waitForNewPreviewClient(ctx context.Context, before map[string]bool, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, client := range previewWorkspaceClients(ctx) {
			if client.Address != "" && !before[client.Address] {
				return "address:" + client.Address
			}
		}
		time.Sleep(150 * time.Millisecond)
	}
	return ""
}

func previewWorkspaceClients(ctx context.Context) []hyprClient {
	cmd := exec.CommandContext(ctx, "hyprctl", "clients", "-j")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var clients []hyprClient
	if err := json.Unmarshal(out, &clients); err != nil {
		return nil
	}
	var filtered []hyprClient
	for _, client := range clients {
		if isPreviewWorkspace(client.Workspace) && client.Address != "" {
			filtered = append(filtered, client)
		}
	}
	return filtered
}

func closePreviewWorkspaceClients(ctx context.Context) {
	for _, client := range previewWorkspaceClients(ctx) {
		_ = runLive(ctx, "hyprctl", "dispatch", "closewindow", "address:"+client.Address)
	}
	time.Sleep(500 * time.Millisecond)
}

func isPreviewWorkspace(workspace hyprWorkspace) bool {
	return workspace.Name == "omarchy-themegen-preview"
}

func terminalCommand(themeDir string) (func(role, title, shell string) []string, error) {
	if p, err := exec.LookPath("ghostty"); err == nil {
		config := filepath.Join(themeDir, "ghostty.conf")
		if fileExists(config) {
			return func(role, title, shell string) []string {
				return append([]string{p, "--config-file=" + config, "--class=" + previewClass(role), "--title=" + title, "--font-size=8", "--window-padding-x=4", "--window-padding-y=4", "-e"}, previewShellArgs(shell)...)
			}, nil
		}
	}

	if p, err := exec.LookPath("alacritty"); err == nil {
		config := filepath.Join(themeDir, "alacritty.toml")
		if fileExists(config) {
			return func(role, title, shell string) []string {
				return append([]string{p, "--config-file", config, "--class", previewClass(role), "--title", title, "-e"}, previewShellArgs(shell)...)
			}, nil
		}
	}
	if p, err := exec.LookPath("kitty"); err == nil {
		config := filepath.Join(themeDir, "kitty.conf")
		if fileExists(config) {
			return func(role, title, shell string) []string {
				return append([]string{p, "--config", config, "--class", previewClass(role), "--title", title}, previewShellArgs(shell)...)
			}, nil
		}
	}
	if p, err := exec.LookPath("foot"); err == nil {
		config := filepath.Join(themeDir, "foot.ini")
		if fileExists(config) {
			return func(role, title, shell string) []string {
				return append([]string{p, "--config", config, "--app-id", previewClass(role), "--title", title}, previewShellArgs(shell)...)
			}, nil
		}
	}
	return nil, fmt.Errorf("no supported terminal with exported theme config found: tried ghostty, alacritty, kitty, foot")
}

func previewShellArgs(command string) []string {
	shell := os.Getenv("SHELL")
	if shell == "" || !fileExists(shell) {
		for _, candidate := range []string{"zsh", "bash", "sh"} {
			if path, err := exec.LookPath(candidate); err == nil {
				shell = path
				break
			}
		}
	}
	if shell == "" {
		return []string{"sh", "-lc", command}
	}
	return []string{shell, "-lic", command}
}

func previewClass(role string) string {
	return "omarchy-preview-" + role
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func fileBrowserCommand() []string {
	home, _ := os.UserHomeDir()
	candidates := [][]string{
		{"nautilus", "--new-window", home},
		{"thunar", home},
		{"dolphin", home},
		{"nemo", home},
		{"pcmanfm", home},
	}
	for _, candidate := range candidates {
		if p, err := exec.LookPath(candidate[0]); err == nil {
			candidate[0] = p
			return candidate
		}
	}
	return nil
}

func activeWorkspace(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "hyprctl", "activeworkspace", "-j")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	var ws hyprWorkspace
	if err := json.Unmarshal(out, &ws); err != nil {
		return ""
	}
	if ws.ID > 0 {
		return strconv.Itoa(ws.ID)
	}
	if ws.Name == "" {
		return ""
	}
	return "name:" + ws.Name
}

func focusedMonitor(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "hyprctl", "monitors", "-j")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	var monitors []hyprMonitor
	if err := json.Unmarshal(out, &monitors); err != nil {
		return ""
	}
	for _, monitor := range monitors {
		if monitor.Focused {
			return monitor.Name
		}
	}
	if len(monitors) > 0 {
		return monitors[0].Name
	}
	return ""
}

func cleanupPreviewWorkspace(originalWorkspace string, processes []previewProcess) {
	if originalWorkspace != "" {
		_ = runLive(context.Background(), "hyprctl", "dispatch", "workspace", originalWorkspace)
	}
	for i := len(processes) - 1; i >= 0; i-- {
		selector := processes[i].Selector
		if selector == "" {
			continue
		}
		_ = runLive(context.Background(), "hyprctl", "dispatch", "closewindow", selector)
	}
	time.Sleep(400 * time.Millisecond)
	killProcesses(processes)
}

func killProcesses(processes []previewProcess) {
	for i := len(processes) - 1; i >= 0; i-- {
		if processes[i].Process == nil {
			continue
		}
		_ = processes[i].Process.Kill()
		_, _ = processes[i].Process.Wait()
	}
}

func runLive(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
