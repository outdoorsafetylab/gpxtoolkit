package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/log"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CommandRequest represents the structure of a command execution request
type CommandRequest struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Flags   map[string]string `json:"flags,omitempty"`
}

// CommandResponse represents the structure of a command execution response
type CommandResponse struct {
	Success     bool              `json:"success"`
	Command     string            `json:"command"`
	Stdout      string            `json:"stdout,omitempty"`
	Stderr      string            `json:"stderr,omitempty"`
	ExitCode    int               `json:"exit_code,omitempty"`
	Duration    string            `json:"duration,omitempty"`
	Error       string            `json:"error,omitempty"`
	OutputFiles []string          `json:"output_files,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CommandController handles command execution requests
type CommandController struct {
	Service elevation.Service
}

// Handler processes HTTP requests for command execution
func (c *CommandController) Handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get command from form
	commandName := r.FormValue("command")
	if commandName == "" {
		http.Error(w, "Command is required", http.StatusBadRequest)
		return
	}

	// Get command arguments
	args := r.Form["args"]

	// Get command flags
	flags := make(map[string]string)
	for key, values := range r.Form {
		if key != "command" && key != "args" && key != "file" {
			if len(values) > 0 {
				flags[key] = values[0]
			}
		}
	}

	// Create temporary directory for command execution
	tempDir, err := os.MkdirTemp("", "gpxtoolkit-*")
	if err != nil {
		log.Errorf("Failed to create temp directory: %v", err)
		http.Error(w, "Failed to create temp directory: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	// Process uploaded files directly to temp directory
	inputFiles, err := c.processUploadedFiles(r, tempDir)
	if err != nil {
		log.Errorf("Failed to process uploaded files: %v", err)
		http.Error(w, "Failed to process files: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Execute command
	response, err := c.executeCommand(commandName, args, flags, inputFiles, tempDir)
	if err != nil {
		log.Errorf("Failed to execute command: %v", err)
		response = &CommandResponse{
			Success:  false,
			Command:  commandName,
			Error:    err.Error(),
			Duration: time.Since(startTime).String(),
		}
	} else {
		response.Duration = time.Since(startTime).String()
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Errorf("Failed to encode response: %v", err)
	}
}

// processUploadedFiles handles file uploads and returns the list of file paths
func (c *CommandController) processUploadedFiles(r *http.Request, tempDir string) ([]string, error) {
	var inputFiles []string

	// Process multipart files
	for _, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			file, err := fileHeader.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open uploaded file %s: %w", fileHeader.Filename, err)
			}
			defer file.Close()

			// Create temp file
			tempFile, err := os.CreateTemp(tempDir, filepath.Base(fileHeader.Filename)+"-*")
			if err != nil {
				return nil, fmt.Errorf("failed to create temp file for %s: %w", fileHeader.Filename, err)
			}
			defer tempFile.Close()

			// Copy file content
			if _, err := io.Copy(tempFile, file); err != nil {
				return nil, fmt.Errorf("failed to copy file %s: %w", fileHeader.Filename, err)
			}

			inputFiles = append(inputFiles, tempFile.Name())
		}
	}

	return inputFiles, nil
}

// executeCommand runs the specified command with the given arguments and files
func (c *CommandController) executeCommand(commandName string, args []string, flags map[string]string, inputFiles []string, tempDir string) (*CommandResponse, error) {
	// Build command line arguments
	cmdArgs := []string{commandName}

	// Add flags
	for flag, value := range flags {
		if value == "true" {
			cmdArgs = append(cmdArgs, "--"+flag)
		} else if value != "false" {
			cmdArgs = append(cmdArgs, "--"+flag, value)
		}
	}

	// Add file arguments
	for _, file := range inputFiles {
		cmdArgs = append(cmdArgs, "--file", file)
	}

	// Add additional args
	cmdArgs = append(cmdArgs, args...)

	// Create command context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Execute the command using the current executable
	cmd := exec.CommandContext(ctx, os.Args[0], cmdArgs...)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set working directory to temp directory
	cmd.Dir = tempDir

	// Set environment variables for elevation service if available
	if c.Service != nil {
		// This would need to be implemented based on your elevation service configuration
		// For now, we'll use the default environment
	}

	// Execute command
	err := cmd.Run()

	// Build response
	response := &CommandResponse{
		Success:  err == nil,
		Command:  commandName,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: cmd.ProcessState.ExitCode(),
	}

	// Handle errors
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			response.Error = "Command execution timed out"
		} else {
			response.Error = err.Error()
		}
	}

	// Look for output files in temp directory
	outputFiles, err := c.findOutputFiles(tempDir)
	if err != nil {
		log.Warnf("Failed to find output files: %v", err)
	} else {
		response.OutputFiles = outputFiles
	}

	// Add metadata
	response.Metadata = map[string]string{
		"temp_dir": tempDir,
		"args":     strings.Join(cmdArgs, " "),
	}

	return response, nil
}

// findOutputFiles looks for generated output files in the temp directory
func (c *CommandController) findOutputFiles(tempDir string) ([]string, error) {
	var outputFiles []string

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			// Skip input files (they usually have the original extension)
			ext := filepath.Ext(entry.Name())
			if ext == ".gpx" || ext == ".csv" || ext == ".svg" {
				outputFiles = append(outputFiles, entry.Name())
			}
		}
	}

	return outputFiles, nil
}
