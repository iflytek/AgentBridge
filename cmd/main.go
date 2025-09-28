package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	version     = "dev" // Set at build time via -ldflags
	appName     = "AgentBridge"
	description = "Cross-Platform AI Agent DSL Converter"
)

// getVersion returns the version string, attempting to get it from build info first
func getVersion() string {
	// If version was set via ldflags, use it
	if version != "dev" {
		return version
	}

	// Try to get version from build info (works with go install)
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			return info.Main.Version
		}
	}

	// Fallback to default
	return version
}

var (
	// Global flags
	verbose bool
	quiet   bool
)

var rootCmd = &cobra.Command{
	Use:   "agentbridge",
	Short: description,
	Long: `🚀 AgentBridge - Cross-Platform AI Agent DSL Converter

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
	Version: getVersion(),
	Example: `  # Basic conversion (iFlytek to Dify)
  agentbridge convert --from iflytek --to dify --input agent.yml --output dify.yml

  # Convert iFlytek to Coze
  agentbridge convert --from iflytek --to coze --input agent.yml --output coze.yml

  # Convert Coze ZIP to iFlytek
  agentbridge convert --from coze --to iflytek --input workflow.zip --output agent.yml

  # Auto-detect source format
  agentbridge convert --to coze --input agent.yml --output coze.yml

  # Validate DSL file
  agentbridge validate --input agent.yml

  # Show supported node types
  agentbridge info --nodes`,
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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Error: %v\n", wrapUserFriendlyError(err))
		}
		os.Exit(1)
	}
}
