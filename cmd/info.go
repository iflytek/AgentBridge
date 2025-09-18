package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// NewInfoCmd creates the info command
func NewInfoCmd() *cobra.Command {
	var infoCmd = &cobra.Command{
		Use:   "info",
		Short: "Display tool information",
		Long: `Display detailed information about the tool, including supported node types, data type mappings, and more.

This command provides comprehensive information about the converter's capabilities.`,
		Example: `  # Show supported node types
  ai-agent-converter info --nodes

  # Show data type mappings
  ai-agent-converter info --types

  # Show all information
  ai-agent-converter info --all`,
		RunE: runInfo,
	}

	// Configure info command flags
	infoCmd.Flags().BoolVar(&showNodes, "nodes", false, "Show supported node types")
	infoCmd.Flags().BoolVar(&showTypes, "types", false, "Show data type mappings")
	infoCmd.Flags().BoolVar(&showAll, "all", false, "Show all information")

	return infoCmd
}

// runInfo executes the info command
func runInfo(cmd *cobra.Command, args []string) error {
	if !quiet {
		printHeader("Tool Information")
	}

	if showAll || showNodes {
		printSupportedNodes()
	}

	if showAll || showTypes {
		printDataTypeMapping()
	}

	if !showNodes && !showTypes && !showAll {
		printGeneralInfo()
	}

	return nil
}

// printSupportedNodes prints supported node types information
func printSupportedNodes() {
	fmt.Println("\n📋 Supported Node Types:")

	nodeTypes := []struct {
		unified string
		iflytek string
		dify    string
		desc    string
	}{
		{"start", "start-node", "start", "Workflow entry point"},
		{"end", "end-node", "end", "Workflow exit point"},
		{"llm", "llm-node", "llm", "Large Language Model node"},
		{"code", "code-node", "code", "Code execution node"},
		{"condition", "condition-node", "if-else", "Conditional branching node"},
		{"classifier", "classifier-node", "question-classifier", "Question classification node"},
		{"iteration", "iteration-node", "iteration", "Loop iteration node"},
	}

	fmt.Printf("%-12s %-15s %-20s %s\n", "Unified", "iFlytek Spark", "Dify", "Description")
	fmt.Println(strings.Repeat("-", 75))

	for _, nt := range nodeTypes {
		fmt.Printf("%-12s %-15s %-20s %s\n", nt.unified, nt.iflytek, nt.dify, nt.desc)
	}
}

// printDataTypeMapping prints data type mapping information
func printDataTypeMapping() {
	fmt.Println("\n🔄 Data Type Mappings:")

	dataTypes := []struct {
		unified string
		iflytek string
		dify    string
		desc    string
	}{
		{"string", "string", "string", "String type"},
		{"number", "integer/number", "number", "Numeric type"},
		{"boolean", "boolean", "boolean", "Boolean type"},
		{"array[string]", "array-string", "array[string]", "String array"},
		{"array[object]", "array-object", "array[object]", "Object array"},
		{"object", "object", "object", "Object type"},
	}

	fmt.Printf("%-15s %-15s %-15s %s\n", "Unified", "iFlytek Spark", "Dify", "Description")
	fmt.Println(strings.Repeat("-", 70))

	for _, dt := range dataTypes {
		fmt.Printf("%-15s %-15s %-15s %s\n", dt.unified, dt.iflytek, dt.dify, dt.desc)
	}
}

// printGeneralInfo prints general tool information
func printGeneralInfo() {
	fmt.Printf("📦 Tool Version: %s\n", version)
	fmt.Printf("🎯 Supported Platforms: iFlytek Spark Agent (Hub), Dify, Coze\n")
	fmt.Printf("🔄 Conversion Paths: iFlytek↔Dify, iFlytek↔Coze (Star Architecture)\n")
	fmt.Printf("📋 Supported Nodes: 7 fundamental node types\n")
	fmt.Printf("🔗 Supported Connections: Default connections, conditional connections\n")
	fmt.Printf("📊 Data Types: 6 unified data types\n")
	fmt.Printf("⚡ Features: Bidirectional conversion, unsupported node placeholder conversion, ZIP format support\n")

	fmt.Println("\n💡 Usage Tips:")
	fmt.Println("   • Use --verbose to see detailed conversion process")
	fmt.Println("   • Use validate command to verify file format")
	fmt.Println("   • Supports automatic format detection")
	fmt.Println("   • Coze ZIP format is automatically detected and supported")
	fmt.Println("   • For Dify↔Coze conversion, use iFlytek as intermediate hub")
	fmt.Println("   • Converted files can be directly imported to target platforms")
}