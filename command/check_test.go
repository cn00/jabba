package command

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name     string
		shell    string
		osType   string
		psModule string
		expected string
	}{
		{"bash", "/bin/bash", "", "", "bash"},
		{"zsh", "/usr/bin/zsh", "", "", "zsh"},
		{"fish", "/usr/local/bin/fish", "", "", "fish"},
		{"nushell", "/usr/bin/nu", "", "", "nushell"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("SHELL", tt.shell)
			if tt.osType != "" {
				os.Setenv("OSTYPE", tt.osType)
			}
			if tt.psModule != "" {
				os.Setenv("PSModulePath", tt.psModule)
			}

			result := detectShell()
			if result != tt.expected {
				t.Errorf("detectShell() = %v, want %v", result, tt.expected)
			}

			os.Unsetenv("SHELL")
			os.Unsetenv("OSTYPE")
			os.Unsetenv("PSModulePath")
		})
	}
}

func TestGetIntegrationFile(t *testing.T) {
	tests := []struct {
		shell    string
		expected string
	}{
		{"bash", "jabba.sh"},
		{"zsh", "jabba.sh"},
		{"fish", "jabba.fish"},
		{"powershell", "jabba.ps1"},
		{"nushell", "jabba.nu"},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			result := getIntegrationFile(tt.shell)
			if filepath.Base(result) != tt.expected {
				t.Errorf("getIntegrationFile(%s) = %v, want filename %v", tt.shell, result, tt.expected)
			}
		})
	}
}

func TestGetRCFile(t *testing.T) {
	tests := []struct {
		shell    string
		expected string
	}{
		{"bash", ".bashrc"},
		{"zsh", ".zshrc"},
		{"fish", "config.fish"},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			result := getRCFile(tt.shell)
			if !filepath.IsAbs(result) {
				t.Errorf("getRCFile(%s) should return absolute path", tt.shell)
			}
			if filepath.Base(result) != tt.expected {
				t.Errorf("getRCFile(%s) = %v, want filename %v", tt.shell, result, tt.expected)
			}
		})
	}
}

func TestCheckRCFileHasSource(t *testing.T) {
	tmpDir := t.TempDir()
	rcFile := filepath.Join(tmpDir, ".bashrc")
	integrationFile := filepath.Join(tmpDir, "jabba.sh")

	// Test with no file
	if checkRCFileHasSource(rcFile, integrationFile) {
		t.Error("checkRCFileHasSource should return false for non-existent file")
	}

	// Test with file without source line
	os.WriteFile(rcFile, []byte("export PATH=/usr/bin:$PATH\n"), 0644)
	if checkRCFileHasSource(rcFile, integrationFile) {
		t.Error("checkRCFileHasSource should return false when source line is missing")
	}

	// Test with file with source line
	os.WriteFile(rcFile, []byte("export PATH=/usr/bin:$PATH\nsource ~/.jabba/jabba.sh\n"), 0644)
	if !checkRCFileHasSource(rcFile, integrationFile) {
		t.Error("checkRCFileHasSource should return true when source line is present")
	}
}

func TestCheck(t *testing.T) {
	// Set up test environment
	originalHome := os.Getenv("JABBA_HOME")
	tmpDir := t.TempDir()
	os.Setenv("JABBA_HOME", tmpDir)
	defer os.Setenv("JABBA_HOME", originalHome)

	result, err := Check()
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.JabbaHome == "" {
		t.Error("Check() should return JABBA_HOME")
	}

	if result.CurrentShell == "" {
		t.Error("Check() should detect current shell")
	}

	if result.IntegrationFile == "" {
		t.Error("Check() should return integration file path")
	}
}

func TestCheckWithIntegrationFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	tmpDir := t.TempDir()
	originalHome := os.Getenv("JABBA_HOME")
	os.Setenv("JABBA_HOME", tmpDir)
	os.Setenv("SHELL", "/bin/bash")
	defer func() {
		os.Setenv("JABBA_HOME", originalHome)
		os.Unsetenv("SHELL")
	}()

	// Create integration file
	integrationFile := filepath.Join(tmpDir, "jabba.sh")
	os.WriteFile(integrationFile, []byte("# test"), 0644)

	result, err := Check()
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if !result.ShellIntegrationOK {
		t.Error("Check() should detect existing integration file")
	}
}
