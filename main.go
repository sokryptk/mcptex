package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"log/slog"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

var (
	logger    = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx       = context.Background()
	outputDir = flag.String("output-dir", "output", "Name of the output directory")
	port      = flag.String("port", "8080", "Port to run the server on")
)

type CustomLogger struct {
	logger *slog.Logger
}

func (c *CustomLogger) Infof(msg string, args ...any) {
	c.logger.Info(msg, args...)
}

func (c *CustomLogger) Errorf(msg string, args ...any) {
	c.logger.Error(msg, args...)
}

func main() {
	var transport = "http"
	flag.StringVar(&transport, "transport", transport, "Transport protocol to use (http or stdio)")
	flag.StringVar(&transport, "t", transport, "Transport protocol to use (http or stdio)")
	flag.Parse()

	s := server.NewMCPServer(
		"MCPTex",
		"0.1.0",
		server.WithToolCapabilities(false),
		server.WithLogging(),
		server.WithRecovery(),
	)

	convertLatexToPdf := mcp.NewTool("convert_latex_to_pdf",
		mcp.WithDescription("Converts a given LaTeX ready document to PDF"),
		mcp.WithString("document",
			mcp.Required(),
			mcp.Description("LaTeX document to convert"),
		),
		mcp.WithArray("options",
			mcp.Description("Options for the conversion, e.g., compiler options -mltex, -etex"),
		),
	)

	s.AddTool(convertLatexToPdf, convertPdfToLatex)

	if transport == "http" {
		httpServer := server.NewStreamableHTTPServer(s, server.WithLogger(&CustomLogger{logger: logger}))
		logger.Info("HTTP server listening on :8080/mcp")

		if err := httpServer.Start(fmt.Sprintf(":%s", *port)); err != nil {
			logger.Error("Server error: %v", err)
			os.Exit(1)
		}
	} else {
		err := server.ServeStdio(s)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to start server", "error", err)
			os.Exit(1)
		}
	}
}

func helloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Hello, %s!", name)), nil
}
