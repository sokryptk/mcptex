package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"os"
	"os/exec"
	"strings"
)

func convertPdfToLatex(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	document, err := request.RequireString("document")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	options := request.GetStringSlice("options", []string{})

	err = os.Mkdir("output", 0755)
	if err != nil && !os.IsExist(err) {
		logger.ErrorContext(ctx, "Failed to create output directory", "error", err)
		return mcp.NewToolResultError("Failed to convert LaTeX to pdf"), nil
	}

	name, err := uuid.NewV7()
	if err != nil {
		logger.ErrorContext(ctx, "Failed to generate UUID", "error", err)
		return mcp.NewToolResultError("Failed to convert LaTeX to pdf"), nil
	}

	tempFolder, err := os.MkdirTemp(*outputDir, fmt.Sprintf("%s", name.String()))
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create temporary directory", "error", err)
		return mcp.NewToolResultError("Failed to convert LaTeX to PDF"), nil
	}

	defer os.RemoveAll(tempFolder)
	_, err = exec.LookPath("xelatex")
	if err != nil {
		logger.ErrorContext(ctx, "xelatex not found in PATH", "error", err)
		return mcp.NewToolResultError("xelatex not found in PATH"), err
	}

	defaultOpts := []string{
		fmt.Sprintf("-output-directory=%s", tempFolder),
		"-halt-on-error",
	}

	randomPdfName := fmt.Sprintf("%s/texput.pdf", tempFolder)
	cmd := exec.Command("xelatex", append(defaultOpts, options...)...)
	cmd.Stdin = strings.NewReader(document)
	var errBuffer bytes.Buffer
	cmd.Stdout = &errBuffer

	err = cmd.Run()
	if err != nil {
		logger.ErrorContext(ctx, "Failed to run xelatex command", "error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Failed to convert LaTeX to PDF: %s", errBuffer.String()),
		), nil
	}

	if _, err := os.Stat(randomPdfName); os.IsNotExist(err) {
		logger.ErrorContext(ctx, "Generated PDF file does not exist", "file", randomPdfName)
		return mcp.NewToolResultError("Failed to convert LaTeX to PDF"), err
	}

	logger.InfoContext(ctx, "Successfully converted LaTeX to PDF", "file", randomPdfName)

	file, err := os.Open(randomPdfName)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to open generated PDF file", "error", err)
		return mcp.NewToolResultError("Failed to convert LaTeX to PDF"), err
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to read generated PDF file", "error", err)
		return mcp.NewToolResultError("Failed to convert LaTeX to PDF"), fmt.Errorf("failed to read generated PDF file")
	}

	blob := mcp.BlobResourceContents{
		URI:      randomPdfName,
		Blob:     base64.StdEncoding.EncodeToString(fileContent),
		MIMEType: "application/pdf",
	}

	return mcp.NewToolResultResource(randomPdfName, blob), nil
}
