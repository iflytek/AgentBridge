package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewPlatformsCmd creates the platforms command
func NewPlatformsCmd() *cobra.Command {
	var platformsCmd = &cobra.Command{
		Use:   "platforms",
		Short: "List supported AI agent platforms",
		Long: `Display all currently supported AI agent platforms and their development status.

This command shows the ecosystem of supported platforms and their conversion capabilities.`,
		Example: `  # List all platforms
  agentbridge platforms

  # Show detailed platform information
  agentbridge platforms --detailed`,
		RunE: runPlatforms,
	}

	// Configure platforms command flags
	platformsCmd.Flags().BoolVar(&showDetailed, "detailed", false, "Show detailed platform information")

	return platformsCmd
}

// platformInfo represents platform information structure
type platformInfo struct {
	name        string
	status      string
	description string
	features    []string
}

// runPlatforms executes the platforms command
func runPlatforms(cmd *cobra.Command, args []string) error {
	if !quiet {
		printHeader("Supported AI Agent Platforms")
	}

	platforms := getSupportedPlatforms()
	displayPlatformOverview(platforms)
	displayPlatformSummary(platforms)

	return nil
}

// getSupportedPlatforms returns the list of supported platforms
func getSupportedPlatforms() []platformInfo {
	return []platformInfo{
		{
			name:        "iFlytek Spark Agent",
			status:      "✅ Production Ready (Central Hub)",
			description: "iFlytek Spark AI Agent Development Platform - Central conversion hub",
			features:    []string{"Visual workflow design", "Multi-modal AI capabilities", "Enterprise integration", "Central DSL hub"},
		},
		{
			name:        "Dify Platform",
			status:      "✅ Production Ready",
			description: "Open-source LLM application development platform",
			features:    []string{"LLM orchestration", "RAG capabilities", "API-first design", "Bidirectional with iFlytek"},
		},
		{
			name:        "Coze Platform",
			status:      "✅ Production Ready",
			description: "ByteDance AI agent development platform",
			features:    []string{"Visual workflow builder", "ZIP format support", "Complex iteration nodes", "Bidirectional with iFlytek"},
		},
	}
}

// displayPlatformOverview displays the platform ecosystem overview
func displayPlatformOverview(platforms []platformInfo) {
	fmt.Printf("📊 Platform Ecosystem Overview:\n\n")

	for _, platform := range platforms {
		displayPlatformDetails(platform)
	}
}

// displayPlatformDetails displays details for a single platform
func displayPlatformDetails(platform platformInfo) {
	fmt.Printf("🔹 %s\n", platform.name)
	fmt.Printf("   Status: %s\n", platform.status)
	fmt.Printf("   Description: %s\n", platform.description)

	if showDetailed {
		displayPlatformFeatures(platform.features)
	}
	fmt.Println()
}

// displayPlatformFeatures displays platform features
func displayPlatformFeatures(features []string) {
	fmt.Printf("   Key Features:\n")
	for _, feature := range features {
		fmt.Printf("     • %s\n", feature)
	}
}

// displayPlatformSummary displays the platform summary statistics
func displayPlatformSummary(platforms []platformInfo) {
	fmt.Printf("🎯 Total Platforms: %d (star architecture with iFlytek as hub)\n", len(platforms))
	fmt.Printf("📈 Conversion Coverage: iFlytek↔Dify, iFlytek↔Coze bidirectional support\n")
	fmt.Printf("⚠️  Note: Dify↔Coze direct conversion not supported (use iFlytek as intermediate hub)\n")
}
