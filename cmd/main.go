package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	version     = "v1.0.0"
	appName     = "AI Agents Transformer"
	description = "Cross-Platform AI Agent DSL Converter"
)

var (
	// Global flags
	verbose bool
	quiet   bool
)

var rootCmd = &cobra.Command{
	Use:   "ai-agent-converter",
	Short: description,
	Long: `🚀 AI Agents Transformer - Cross-Platform AI Agent DSL Converter

Cross-platform AI agent workflow DSL converter with iFlytek Spark as the central hub, supporting bidirectional conversion between Spark, Dify, and Coze platforms.

✨ Features:
  • Bidirectional conversion with data integrity guarantee
  • 7 node types support (start, end, llm, code, condition, classifier, iteration)
  • Variable reference resolution and transformation
  • Platform-specific configuration preservation
  • Three-stage validation pipeline (structural, semantic, platform)
  • Unsupported node placeholder conversion mechanism

🎯 Supported Conversion Paths:
  • iFlytek Spark ↔ Dify Platform    ✅ Production Ready
  • iFlytek Spark ↔ Coze Platform    ✅ Production Ready
  • ZIP Format Support for Coze      ✅ Production Ready

🔧 Enterprise Features:
  • Structured logging system
  • Configuration management  
  • Error handling and recovery
  • Performance optimization`,
	Version: version,
	Example: `  # Basic conversion (iFlytek to Dify)
  ai-agent-converter convert --from iflytek --to dify --input agent.yml --output dify.yml

  # Convert iFlytek to Coze
  ai-agent-converter convert --from iflytek --to coze --input agent.yml --output coze.yml

  # Convert Coze ZIP to iFlytek
  ai-agent-converter convert --from coze --to iflytek --input workflow.zip --output agent.yml

  # Auto-detect source format
  ai-agent-converter convert --to coze --input agent.yml --output coze.yml

  # Validate DSL file
  ai-agent-converter validate --input agent.yml

  # Show supported node types
  ai-agent-converter info --nodes`,
}

func init() {
	// Configure global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode, only show errors")
	
	// Add subcommands
	rootCmd.AddCommand(NewConvertCmd())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewInfoCmd())
	rootCmd.AddCommand(NewPlatformsCmd())
	rootCmd.AddCommand(NewBatchCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Error: %v\n", wrapUserFriendlyError(err))
		}
		os.Exit(1)
	}
}

