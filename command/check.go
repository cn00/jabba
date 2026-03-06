package command

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Jabba-Team/jabba/cfg"
)

type CheckResult struct {
	JabbaHome           string
	JavaHome            string
	ShellIntegrationOK  bool
	IntegrationFile     string
	RCFile              string
	RCFileHasSource     bool
	CurrentShell        string
	IsFunction          bool
}

func Check() (*CheckResult, error) {
	result := &CheckResult{
		JabbaHome:    cfg.Dir(),
		JavaHome:     os.Getenv("JAVA_HOME"),
		CurrentShell: detectShell(),
	}

	// Determine integration file based on shell
	result.IntegrationFile = getIntegrationFile(result.CurrentShell)
	result.RCFile = getRCFile(result.CurrentShell)

	// Check if integration file exists
	if _, err := os.Stat(result.IntegrationFile); err == nil {
		result.ShellIntegrationOK = true
	}

	// Check if RC file has source line
	if result.RCFile != "" {
		result.RCFileHasSource = checkRCFileHasSource(result.RCFile, result.IntegrationFile)
	}

	// Check if jabba is loaded as function (via JABBA_SHELL_INTEGRATION env var)
	result.IsFunction = os.Getenv("JABBA_SHELL_INTEGRATION") == "ON"

	return result, nil
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		if runtime.GOOS == "windows" {
			// Check if running in PowerShell
			if os.Getenv("PSModulePath") != "" {
				return "powershell"
			}
			// Check if running in Git Bash/MSYS/Cygwin
			osType := os.Getenv("OSTYPE")
			if strings.Contains(osType, "msys") || strings.Contains(osType, "cygwin") {
				return "bash"
			}
			return "powershell"
		}
		return "bash"
	}

	// Extract shell name from path
	shellName := filepath.Base(shell)
	if strings.Contains(shellName, "zsh") {
		return "zsh"
	} else if strings.Contains(shellName, "fish") {
		return "fish"
	} else if strings.Contains(shellName, "bash") {
		return "bash"
	} else if strings.Contains(shellName, "nu") {
		return "nushell"
	}

	return "bash" // default
}

func getIntegrationFile(shell string) string {
	jabbaHome := cfg.Dir()
	switch shell {
	case "powershell":
		return filepath.Join(jabbaHome, "jabba.ps1")
	case "fish":
		return filepath.Join(jabbaHome, "jabba.fish")
	case "nushell":
		return filepath.Join(jabbaHome, "jabba.nu")
	default:
		return filepath.Join(jabbaHome, "jabba.sh")
	}
}

func getRCFile(shell string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch shell {
	case "powershell":
		if runtime.GOOS == "windows" {
			return filepath.Join(os.Getenv("USERPROFILE"), "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
		}
		return ""
	case "zsh":
		return filepath.Join(home, ".zshrc")
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish")
	case "nushell":
		if runtime.GOOS == "windows" {
			return filepath.Join(os.Getenv("APPDATA"), "nushell", "config.nu")
		}
		return filepath.Join(home, ".config", "nushell", "config.nu")
	case "bash":
		return filepath.Join(home, ".bashrc")
	default:
		return filepath.Join(home, ".bashrc")
	}
}

func checkRCFileHasSource(rcFile, integrationFile string) bool {
	content, err := os.ReadFile(rcFile)
	if err != nil {
		return false
	}

	// Look for reference to jabba integration file
	return strings.Contains(string(content), "jabba.sh") ||
		strings.Contains(string(content), "jabba.ps1") ||
		strings.Contains(string(content), "jabba.fish") ||
		strings.Contains(string(content), "jabba.nu")
}

func PrintCheckResult(result *CheckResult) {
	fmt.Println("Jabba Installation Check")
	fmt.Println("========================")
	fmt.Printf("JABBA_HOME:           %s\n", result.JabbaHome)
	fmt.Printf("JAVA_HOME:            %s\n", result.JavaHome)
	fmt.Printf("Current Shell:        %s\n", result.CurrentShell)
	fmt.Printf("Integration File:     %s\n", result.IntegrationFile)

	if result.ShellIntegrationOK {
		fmt.Printf("Integration Exists:   ✓ Yes\n")
	} else {
		fmt.Printf("Integration Exists:   ✗ No (run 'jabba init' to create)\n")
	}

	fmt.Printf("RC File:              %s\n", result.RCFile)
	if result.RCFileHasSource {
		fmt.Printf("RC File Configured:   ✓ Yes\n")
	} else {
		fmt.Printf("RC File Configured:   ✗ No (run 'jabba init' to configure)\n")
	}

	if result.IsFunction {
		fmt.Printf("Shell Integration:    ✓ Active (jabba is a function)\n")
	} else {
		fmt.Printf("Shell Integration:    ✗ Not active (jabba is binary only)\n")
		fmt.Println("\nTo activate shell integration, restart your shell or run:")
		if result.CurrentShell == "powershell" {
			fmt.Printf("  . \"%s\"\n", result.IntegrationFile)
		} else {
			fmt.Printf("  source \"%s\"\n", result.IntegrationFile)
		}
	}
}
