package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// IsInstalled checks if git is installed and available in PATH.
func IsInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// Version returns the installed git version.
func Version() (string, error) {
	cmd := exec.Command("git", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// InstallCommand returns the command to install git for the current OS.
func InstallCommand() (string, string) {
	switch runtime.GOOS {
	case "darwin":
		// Check if Homebrew is installed
		if _, err := exec.LookPath("brew"); err == nil {
			return "brew", "brew install git"
		}
		return "xcode-select", "xcode-select --install"
	case "linux":
		// Try to detect package manager
		if _, err := exec.LookPath("apt-get"); err == nil {
			return "apt-get", "sudo apt-get update && sudo apt-get install -y git"
		}
		if _, err := exec.LookPath("yum"); err == nil {
			return "yum", "sudo yum install -y git"
		}
		if _, err := exec.LookPath("dnf"); err == nil {
			return "dnf", "sudo dnf install -y git"
		}
		if _, err := exec.LookPath("pacman"); err == nil {
			return "pacman", "sudo pacman -S --noconfirm git"
		}
		if _, err := exec.LookPath("apk"); err == nil {
			return "apk", "sudo apk add git"
		}
		return "", ""
	case "windows":
		// Check if winget, choco, or scoop is available
		if _, err := exec.LookPath("winget"); err == nil {
			return "winget", "winget install --id Git.Git -e --source winget"
		}
		if _, err := exec.LookPath("choco"); err == nil {
			return "choco", "choco install git -y"
		}
		if _, err := exec.LookPath("scoop"); err == nil {
			return "scoop", "scoop install git"
		}
		return "", ""
	default:
		return "", ""
	}
}

// PromptInstall asks the user for confirmation and installs git.
func PromptInstall() error {
	pkgMgr, installCmd := InstallCommand()
	if installCmd == "" {
		return fmt.Errorf("cannot determine how to install git on %s. Please install git manually", runtime.GOOS)
	}

	fmt.Printf("git is required but not installed.\n")
	fmt.Printf("Install using %s? [Y/n]: ", pkgMgr)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input != "" && input != "y" && input != "yes" {
		return fmt.Errorf("git installation cancelled")
	}

	fmt.Printf("Installing git...\n")
	return Install()
}

// Install installs git using the appropriate package manager.
func Install() error {
	_, installCmd := InstallCommand()
	if installCmd == "" {
		return fmt.Errorf("cannot determine how to install git on %s", runtime.GOOS)
	}

	// Use shell to execute the install command
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/C", installCmd)
	default:
		cmd = exec.Command("sh", "-c", installCmd)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install git: %w", err)
	}

	// Verify installation
	if !IsInstalled() {
		return fmt.Errorf("git installation completed but git is not in PATH. Please restart your terminal")
	}

	version, _ := Version()
	fmt.Printf("git installed successfully: %s\n", version)
	return nil
}

// EnsureInstalled checks if git is installed, prompts for installation if not.
func EnsureInstalled() error {
	if IsInstalled() {
		return nil
	}
	return PromptInstall()
}

// Clone clones a repository to the specified path.
func Clone(url, destPath string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", url, destPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CloneQuiet clones a repository quietly.
func CloneQuiet(url, destPath string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", "--quiet", url, destPath)
	return cmd.Run()
}

// Pull pulls the latest changes in a repository.
func Pull(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "pull", "--ff-only")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// PullQuiet pulls quietly.
func PullQuiet(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "pull", "--ff-only", "--quiet")
	return cmd.Run()
}

// Fetch fetches the latest changes without merging.
func Fetch(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "fetch", "--quiet")
	return cmd.Run()
}

// GetCurrentCommit returns the current commit SHA.
func GetCurrentCommit(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetRemoteCommit returns the latest remote commit SHA.
func GetRemoteCommit(repoPath, branch string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "origin/"+branch)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetDefaultBranch returns the default branch name.
func GetDefaultBranch(repoPath string) (string, error) {
	// Try to get from remote HEAD
	cmd := exec.Command("git", "-C", repoPath, "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()
	if err == nil {
		ref := strings.TrimSpace(string(output))
		// refs/remotes/origin/main -> main
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1], nil
		}
	}

	// Fallback: try common branch names
	for _, branch := range []string{"main", "master"} {
		cmd := exec.Command("git", "-C", repoPath, "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch)
		if cmd.Run() == nil {
			return branch, nil
		}
	}

	return "", fmt.Errorf("cannot determine default branch")
}

// HasChanges checks if there are new commits on remote.
func HasChanges(repoPath, branch string) (bool, error) {
	if err := Fetch(repoPath); err != nil {
		return false, err
	}

	local, err := GetCurrentCommit(repoPath)
	if err != nil {
		return false, err
	}

	remote, err := GetRemoteCommit(repoPath, branch)
	if err != nil {
		return false, err
	}

	return local != remote, nil
}

// ListChangedFiles returns files changed between two commits.
func ListChangedFiles(repoPath, fromCommit, toCommit string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "diff", "--name-only", fromCommit, toCommit)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}
