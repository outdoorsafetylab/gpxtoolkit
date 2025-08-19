/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gpxtoolkit/log"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var (
	cgiURL     = "http://localhost:8080"
	cgiTimeout = 5 * time.Minute
)

// cgiCmd represents the cgi command
var cgiCmd = &cobra.Command{
	Use:   "cgi",
	Short: "Execute commands via HTTP service (local command simulator)",
	Long: `Execute commands via HTTP service (local command simulator).

This command acts as a bridge between local CLI usage and the HTTP service.
It simulates local command execution by making HTTP requests to the service.

Examples:
  # Get statistics via HTTP service
  gpxtoolkit cgi stats --file track.gpx --alpha 0.3

  # Convert CSV to GPX via HTTP service
  gpxtoolkit cgi csv2gpx --file points.csv

  # Elevation correction via HTTP service
  gpxtoolkit cgi elev --file track.gpx --waypoints

  # Set custom service URL
  gpxtoolkit cgi --url http://remote-server:8080 stats --file track.gpx
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("command name is required")
		}

		commandName := args[0]
		commandArgs := args[1:]

		// Execute command via HTTP service
		return executeViaHTTP(commandName, commandArgs)
	},
}

// executeViaHTTP executes a command via the HTTP service
func executeViaHTTP(commandName string, args []string) error {
	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add command
	if err := writer.WriteField("command", commandName); err != nil {
		return fmt.Errorf("failed to write command field: %w", err)
	}

	// Add command arguments
	for _, arg := range args {
		if err := writer.WriteField("args", arg); err != nil {
			return fmt.Errorf("failed to write arg field: %w", err)
		}
	}

	// Add all flags from the current command context
	if err := addAllFlagsToForm(writer, commandName); err != nil {
		return fmt.Errorf("failed to add flags: %w", err)
	}

	// Add files from root command
	if err := addFilesToForm(writer); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	// Close multipart writer
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", cgiURL+"/cgi/execute", body)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: cgiTimeout,
	}

	// Execute request
	log.Infof("Executing command '%s' via HTTP service at %s", commandName, cgiURL)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Handle command execution result
	return handleCommandResult(result)
}

// addAllFlagsToForm adds all flags from the current command context to the multipart form
func addAllFlagsToForm(writer *multipart.Writer, commandName string) error {
	// Add elevation service flags
	if elevationURL != "" {
		if err := writer.WriteField("elevation-url", elevationURL); err != nil {
			return err
		}
	}
	if elevationToken != "" {
		if err := writer.WriteField("elevation-token", elevationToken); err != nil {
			return err
		}
	}
	if googleElevationAPIKey != "" {
		if err := writer.WriteField("elevation-api-key", googleElevationAPIKey); err != nil {
			return err
		}
	}
	if keepCreator {
		if err := writer.WriteField("keep-creator", "true"); err != nil {
			return err
		}
	}

	// For now, we'll rely on the command line parsing to handle specific flags
	// The milestone command will need to be updated to handle its own flags
	// This is a limitation of the current approach - we need to make the flags
	// accessible or use a different method

	return nil
}

// addFilesToForm adds input files to the multipart form
func addFilesToForm(writer *multipart.Writer) error {
	if len(files) == 0 {
		// Try to read from stdin if no files specified
		if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
			// Stdin is a pipe or file, create a temporary file
			tempFile, err := os.CreateTemp("", "gpxtoolkit-stdin-*")
			if err != nil {
				return fmt.Errorf("failed to create temp file for stdin: %w", err)
			}
			defer os.Remove(tempFile.Name())
			defer tempFile.Close()

			if _, err := io.Copy(tempFile, os.Stdin); err != nil {
				return fmt.Errorf("failed to copy stdin to temp file: %w", err)
			}

			// Add the temp file to the form
			file, err := os.Open(tempFile.Name())
			if err != nil {
				return fmt.Errorf("failed to open temp file: %w", err)
			}
			defer file.Close()

			part, err := writer.CreateFormFile("file", "stdin.gpx")
			if err != nil {
				return fmt.Errorf("failed to create form file: %w", err)
			}

			if _, err := io.Copy(part, file); err != nil {
				return fmt.Errorf("failed to copy file to form: %w", err)
			}
		}
		return nil
	}

	// Add specified files
	for _, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", filePath, err)
		}
		defer file.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			return fmt.Errorf("failed to create form file for %s: %w", filePath, err)
		}

		if _, err := io.Copy(part, file); err != nil {
			return fmt.Errorf("failed to copy file %s to form: %w", filePath, err)
		}
	}

	return nil
}

// handleCommandResult processes the command execution result
func handleCommandResult(result map[string]interface{}) error {
	// Check if command was successful
	success, ok := result["success"].(bool)
	if !ok {
		return fmt.Errorf("invalid response format: missing success field")
	}

	if !success {
		// Command failed, print error
		if errorMsg, ok := result["error"].(string); ok {
			return fmt.Errorf("command execution failed: %s", errorMsg)
		}
		return fmt.Errorf("command execution failed with unknown error")
	}

	// Print stdout if available
	if stdout, ok := result["stdout"].(string); ok && stdout != "" {
		fmt.Print(stdout)
	}

	// Print stderr if available
	if stderr, ok := result["stderr"].(string); ok && stderr != "" {
		fmt.Fprint(os.Stderr, stderr)
	}

	// Handle output files
	if outputFiles, ok := result["output_files"].([]interface{}); ok && len(outputFiles) > 0 {
		fmt.Fprintf(os.Stderr, "\nGenerated output files:\n")
		for _, file := range outputFiles {
			if fileName, ok := file.(string); ok {
				fmt.Fprintf(os.Stderr, "  - %s\n", fileName)
			}
		}
	}

	// Print execution info
	if duration, ok := result["duration"].(string); ok {
		fmt.Fprintf(os.Stderr, "\nExecution time: %s\n", duration)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(cgiCmd)

	// Local flags for cgi command
	cgiCmd.Flags().StringVar(&cgiURL, "url", cgiURL, "HTTP service URL")
	cgiCmd.Flags().DurationVar(&cgiTimeout, "timeout", cgiTimeout, "HTTP request timeout")

	// Inherit root command flags
	cgiCmd.PersistentFlags().StringArrayVarP(&files, "file", "f", files, "GPX file name; will read from stdin if this is not specified")
	cgiCmd.PersistentFlags().StringVar(&elevationURL, "elevation-url", "", "URL for elevation service")
	cgiCmd.PersistentFlags().StringVar(&elevationToken, "elevation-token", "", "auth token of elevation service")
	cgiCmd.PersistentFlags().StringVar(&googleElevationAPIKey, "elevation-api-key", "", "API key of Google Elevation API")
	cgiCmd.PersistentFlags().BoolVar(&keepCreator, "keep-creator", keepCreator, "Keep the creator of the original GPX")
}
