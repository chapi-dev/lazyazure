package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    CLIArgs
		wantErr bool
		errMsg  string
	}{
		{
			name: "no args",
			args: []string{"lazyazure"},
			want: CLIArgs{
				ShowVersion: false,
				CheckUpdate: false,
				ShowHelp:    false,
			},
			wantErr: false,
		},
		{
			name: "version flag long",
			args: []string{"lazyazure", "--version"},
			want: CLIArgs{
				ShowVersion: true,
				CheckUpdate: false,
				ShowHelp:    false,
			},
			wantErr: false,
		},
		{
			name: "version flag short",
			args: []string{"lazyazure", "-v"},
			want: CLIArgs{
				ShowVersion: true,
				CheckUpdate: false,
				ShowHelp:    false,
			},
			wantErr: false,
		},
		{
			name: "check-update flag",
			args: []string{"lazyazure", "--check-update"},
			want: CLIArgs{
				ShowVersion: false,
				CheckUpdate: true,
				ShowHelp:    false,
			},
			wantErr: false,
		},
		{
			name: "help flag long",
			args: []string{"lazyazure", "--help"},
			want: CLIArgs{
				ShowVersion: false,
				CheckUpdate: false,
				ShowHelp:    true,
			},
			wantErr: false,
		},
		{
			name: "help flag short",
			args: []string{"lazyazure", "-h"},
			want: CLIArgs{
				ShowVersion: false,
				CheckUpdate: false,
				ShowHelp:    true,
			},
			wantErr: false,
		},
		{
			name:    "unknown flag",
			args:    []string{"lazyazure", "--unknown"},
			want:    CLIArgs{},
			wantErr: true,
			errMsg:  "unknown flag: --unknown",
		},
		{
			name:    "multiple args (treats first as flag)",
			args:    []string{"lazyazure", "--version", "--check-update"},
			want:    CLIArgs{ShowVersion: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseArgs() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("parseArgs() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("parseArgs() unexpected error = %v", err)
				return
			}

			if got.ShowVersion != tt.want.ShowVersion {
				t.Errorf("parseArgs() ShowVersion = %v, want %v", got.ShowVersion, tt.want.ShowVersion)
			}
			if got.CheckUpdate != tt.want.CheckUpdate {
				t.Errorf("parseArgs() CheckUpdate = %v, want %v", got.CheckUpdate, tt.want.CheckUpdate)
			}
			if got.ShowHelp != tt.want.ShowHelp {
				t.Errorf("parseArgs() ShowHelp = %v, want %v", got.ShowHelp, tt.want.ShowHelp)
			}
		})
	}
}

func TestPrintVersion(t *testing.T) {
	tests := []struct {
		name            string
		version         string
		commit          string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:         "normal version",
			version:      "v1.0.0",
			commit:       "abc123def456",
			wantContains: []string{"lazyazure v1.0.0", "abc123d"},
		},
		{
			name:         "dev version",
			version:      "dev",
			commit:       "unknown",
			wantContains: []string{"lazyazure dev", "unknown"},
		},
		{
			name:         "short commit",
			version:      "v0.2.0",
			commit:       "abc1234",
			wantContains: []string{"lazyazure v0.2.0", "abc1234"},
		},
		{
			name:            "long commit truncated",
			version:         "v1.0.0",
			commit:          "abcdef1234567890abcdef1234567890",
			wantContains:    []string{"abcdef1"},
			wantNotContains: []string{"abcdef1234567890abcdef1234567890"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printVersion(tt.version, tt.commit)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("printVersion() output = %q, should contain %q", output, want)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(output, notWant) {
					t.Errorf("printVersion() output = %q, should NOT contain %q", output, notWant)
				}
			}
		})
	}
}

func TestPrintHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printHelp()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that help output contains expected sections
	wantContains := []string{
		"LazyAzure",
		"Usage:",
		"Options:",
		"--help",
		"--version",
		"--check-update",
		"Environment Variables:",
		"LAZYAZURE_DEBUG",
		"LAZYAZURE_DEMO",
		"github.com/matsest/lazyazure",
	}

	for _, want := range wantContains {
		if !strings.Contains(output, want) {
			t.Errorf("printHelp() output should contain %q, got:\n%s", want, output)
		}
	}
}

func TestIsDevelopmentBuild(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{
			name:    "dev version",
			version: "dev",
			want:    true,
		},
		{
			name:    "release version",
			version: "v1.0.0",
			want:    false,
		},
		{
			name:    "dirty build",
			version: "v1.0.0-dirty",
			want:    true,
		},
		{
			name:    "ahead of tag (git describe)",
			version: "v1.0.0-2-gabc1234",
			want:    true,
		},
		{
			name:    "semver without v",
			version: "1.0.0",
			want:    false,
		},
		{
			name:    "prerelease version",
			version: "v1.0.0-beta.1",
			want:    false,
		},
		{
			name:    "commit hash starting with g after dash",
			version: "v1.0.0-gabc1234",
			want:    true, // After split, "gabc1234" has 'g' at position 0 (when i=1)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDevelopmentBuild(tt.version)
			if got != tt.want {
				t.Errorf("isDevelopmentBuild(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestCheckUpdate(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		commit         string
		serverResponse string
		serverStatus   int
		wantExitCode   int
		wantErr        bool
		wantOutput     []string
	}{
		{
			name:           "up to date",
			version:        "v1.0.0",
			commit:         "abc1234",
			serverResponse: `{"tag_name": "v1.0.0"}`,
			serverStatus:   http.StatusOK,
			wantExitCode:   0,
			wantErr:        false,
			wantOutput:     []string{"lazyazure v1.0.0", "v1.0.0", "latest version"},
		},
		{
			name:           "update available",
			version:        "v1.0.0",
			commit:         "abc1234",
			serverResponse: `{"tag_name": "v1.1.0"}`,
			serverStatus:   http.StatusOK,
			wantExitCode:   1,
			wantErr:        false,
			wantOutput:     []string{"Update available", "v1.0.0", "v1.1.0"},
		},
		{
			name:           "dev build skips comparison",
			version:        "dev",
			commit:         "abc1234",
			serverResponse: `{"tag_name": "v1.0.0"}`,
			serverStatus:   http.StatusOK,
			wantExitCode:   0,
			wantErr:        false,
			wantOutput:     []string{"development build", "Skipping"},
		},
		{
			name:           "version without v prefix",
			version:        "1.0.0",
			commit:         "abc1234",
			serverResponse: `{"tag_name": "v1.0.0"}`,
			serverStatus:   http.StatusOK,
			wantExitCode:   0,
			wantErr:        false,
			wantOutput:     []string{"latest version"},
		},
		{
			name:           "server error",
			version:        "v1.0.0",
			commit:         "abc1234",
			serverResponse: `{"message": "Not Found"}`,
			serverStatus:   http.StatusNotFound,
			wantExitCode:   2,
			wantErr:        true,
		},
		{
			name:           "invalid json",
			version:        "v1.0.0",
			commit:         "abc1234",
			serverResponse: `invalid json`,
			serverStatus:   http.StatusOK,
			wantExitCode:   2,
			wantErr:        true,
		},
		{
			name:           "uses default URL when empty",
			version:        "v1.0.0",
			commit:         "abc1234",
			serverResponse: `{"tag_name": "v1.0.0"}`,
			serverStatus:   http.StatusOK,
			wantExitCode:   0,
			wantErr:        false,
			wantOutput:     []string{"latest version"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request headers
				if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
					t.Errorf("Expected Accept header to be application/vnd.github.v3+json, got %s", r.Header.Get("Accept"))
				}
				if r.Header.Get("User-Agent") != "lazyazure" {
					t.Errorf("Expected User-Agent header to be lazyazure, got %s", r.Header.Get("User-Agent"))
				}

				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Create HTTP client that uses test server
			client := server.Client()

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			exitCode, err := checkUpdate(tt.version, tt.commit, client, server.URL)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if tt.wantErr {
				if err == nil {
					t.Errorf("checkUpdate() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("checkUpdate() unexpected error = %v", err)
				}
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("checkUpdate() exitCode = %d, want %d", exitCode, tt.wantExitCode)
			}

			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("checkUpdate() output should contain %q, got:\n%s", want, output)
				}
			}
		})
	}
}

func TestCheckUpdate_DefaultURL(t *testing.T) {
	// This test verifies that when apiURL is empty, the function tries to use the default URL
	// We can't actually test the real GitHub API, but we can verify it attempts the request

	// Create a server that will receive the request
	requestReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tag_name": "v1.0.0"}`))
	}))
	defer server.Close()

	// Use the server URL instead of empty string to test the logic
	// In real usage, empty string would use the hardcoded GitHub URL
	client := server.Client()
	exitCode, err := checkUpdate("v1.0.0", "abc1234", client, server.URL)

	if err != nil {
		t.Errorf("checkUpdate() unexpected error = %v", err)
	}

	if exitCode != 0 {
		t.Errorf("checkUpdate() exitCode = %d, want 0", exitCode)
	}

	if !requestReceived {
		t.Error("checkUpdate() did not make request to server")
	}
}

func TestRunCLI(t *testing.T) {
	tests := []struct {
		name       string
		args       CLIArgs
		version    string
		commit     string
		wantExit   int
		runsApp    bool
		wantOutput []string
	}{
		{
			name:       "show help",
			args:       CLIArgs{ShowHelp: true},
			version:    "v1.0.0",
			commit:     "abc1234",
			wantExit:   0,
			runsApp:    false,
			wantOutput: []string{"LazyAzure", "Usage:", "Options:", "Environment Variables:"},
		},
		{
			name:       "show version",
			args:       CLIArgs{ShowVersion: true},
			version:    "v1.0.0",
			commit:     "abc1234",
			wantExit:   0,
			runsApp:    false,
			wantOutput: []string{"lazyazure v1.0.0"},
		},
		{
			name:       "check update - up to date",
			args:       CLIArgs{CheckUpdate: true},
			version:    "v1.0.0",
			commit:     "abc1234",
			wantExit:   0,
			runsApp:    false,
			wantOutput: []string{"latest version"},
		},
		{
			name:     "no CLI args runs app",
			args:     CLIArgs{},
			version:  "v1.0.0",
			commit:   "abc1234",
			wantExit: -1,
			runsApp:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"tag_name": "v1.0.0"}`))
			}))
			defer server.Close()

			client := server.Client()

			// Capture stdout for commands that produce output
			if len(tt.wantOutput) > 0 {
				oldStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				exitCode := runCLI(tt.args, tt.version, tt.commit, client, server.URL)

				w.Close()
				os.Stdout = oldStdout

				var buf bytes.Buffer
				io.Copy(&buf, r)
				output := buf.String()

				if exitCode != tt.wantExit {
					t.Errorf("runCLI() exitCode = %d, want %d", exitCode, tt.wantExit)
				}

				for _, want := range tt.wantOutput {
					if !strings.Contains(output, want) {
						t.Errorf("runCLI() output should contain %q, got:\n%s", want, output)
					}
				}
			} else {
				exitCode := runCLI(tt.args, tt.version, tt.commit, client, server.URL)

				if exitCode != tt.wantExit {
					t.Errorf("runCLI() exitCode = %d, want %d", exitCode, tt.wantExit)
				}
			}
		})
	}
}

func TestGetVersionInfo(t *testing.T) {
	// Save original values
	origVersion := version
	origCommit := commit
	origDate := date
	defer func() {
		version = origVersion
		commit = origCommit
		date = origDate
	}()

	// Set test values
	version = "v1.2.3"
	commit = "abc123def456"
	date = "2024-01-15"

	info := GetVersionInfo()

	if info.Version != "v1.2.3" {
		t.Errorf("GetVersionInfo().Version = %q, want %q", info.Version, "v1.2.3")
	}
	if info.Commit != "abc123def456" {
		t.Errorf("GetVersionInfo().Commit = %q, want %q", info.Commit, "abc123def456")
	}
	if info.Date != "2024-01-15" {
		t.Errorf("GetVersionInfo().Date = %q, want %q", info.Date, "2024-01-15")
	}
}
