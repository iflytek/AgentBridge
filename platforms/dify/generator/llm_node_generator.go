package generator

import (
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
	"strings"
)

// LLMNodeGenerator generates LLM nodes
type LLMNodeGenerator struct {
	*BaseNodeGenerator
	nodeMapping map[string]models.Node // Node ID to original node mapping
}

func NewLLMNodeGenerator() *LLMNodeGenerator {
	return &LLMNodeGenerator{
		BaseNodeGenerator: NewBaseNodeGenerator(models.NodeTypeLLM),
		nodeMapping:       make(map[string]models.Node),
	}
}

// SetNodeMapping sets node mapping
func (g *LLMNodeGenerator) SetNodeMapping(nodes []models.Node) {
	if g.nodeMapping == nil {
		g.nodeMapping = make(map[string]models.Node)
	}
	for _, node := range nodes {
		g.nodeMapping[node.ID] = node
	}
}

// GenerateNode generates an LLM node
func (g *LLMNodeGenerator) GenerateNode(node models.Node) (DifyNode, error) {
	if node.Type != models.NodeTypeLLM {
		return DifyNode{}, fmt.Errorf("unsupported node type: %s, expected: %s", node.Type, models.NodeTypeLLM)
	}

	// Generate base node structure
	difyNode := g.generateBaseNode(node)

	// Set LLM node specific data - directly set to data field, don't use config wrapper
	g.setLLMDataFields(&difyNode.Data, node)

	// Restore Dify-specific fields from platform configuration
	if difyConfig := node.PlatformConfig.Dify; difyConfig != nil {
		g.restoreDifyPlatformConfig(difyConfig, &difyNode)
	}

	return difyNode, nil
}

// generateModelConfig dynamically generates model configuration from parsed config
func (g *LLMNodeGenerator) generateModelConfig(node models.Node) map[string]interface{} {
	// Default configuration
	modelConfig := map[string]interface{}{
		"completion_params": map[string]interface{}{
			"temperature": 0.7, // Default value, will be overridden by dynamic config below
		},
		"mode":     "chat",
		"name":     "xdeepseekv32",                                           // Default value, will be overridden by dynamic config below
		"provider": "langgenius/openai_api_compatible/openai_api_compatible", // Default value
	}

	// Dynamically get model configuration from unified DSL config
	if llmConfig, ok := node.Config.(models.LLMConfig); ok {
		// Get model parameters - support iFlytek SparkAgent core parameters: Temperature, MaxTokens, TopK
		if params, ok := modelConfig["completion_params"].(map[string]interface{}); ok {
			// Temperature parameter
			if llmConfig.Parameters.Temperature > 0 {
				params["temperature"] = llmConfig.Parameters.Temperature
			}

			// MaxTokens parameter - iFlytek SparkAgent's maxTokens maps to Dify's max_tokens
			if llmConfig.Parameters.MaxTokens > 0 {
				params["max_tokens"] = llmConfig.Parameters.MaxTokens
			}

			// TopK parameter - iFlytek SparkAgent's topK maps to Dify's top_k
			if llmConfig.Parameters.TopK > 0 {
				params["top_k"] = llmConfig.Parameters.TopK
			}
		}

		// Get model name
		if llmConfig.Model.Name != "" {
			modelConfig["name"] = llmConfig.Model.Name
		}

		// Get model provider (keep Dify format)
		if llmConfig.Model.Provider != "" {
			// iFlytek's provider format might be "iflytek/serviceId", need to convert to Dify format
			if strings.HasPrefix(llmConfig.Model.Provider, "iflytek") {
				// Keep default Dify compatible format
				modelConfig["provider"] = "langgenius/openai_api_compatible/openai_api_compatible"
			} else {
				modelConfig["provider"] = llmConfig.Model.Provider
			}
		}

		// Get mode
		if llmConfig.Model.Mode != "" {
			modelConfig["mode"] = llmConfig.Model.Mode
		}
	}

	return modelConfig
}

// generateVisionConfig generates vision configuration dynamically from parsed settings
func (g *LLMNodeGenerator) generateVisionConfig(node models.Node) map[string]interface{} {
	visionConfig := map[string]interface{}{
		"enabled": false, // Default vision setting
	}

	// Extract vision configuration from unified DSL settings
	if llmConfig, ok := node.Config.(models.LLMConfig); ok {
		if llmConfig.Vision != nil {
			visionConfig["enabled"] = llmConfig.Vision.Enabled
		}
	}

	return visionConfig
}

// setLLMDataFields sets LLM node data fields following official example format
func (g *LLMNodeGenerator) setLLMDataFields(data *DifyNodeData, node models.Node) {
	// Set fields in order following official example format
	// Note: These fields must be directly under data, not wrapped in config

	// Set context field
	contextConfig := g.generateContextConfig(node)

	// Set model field - dynamically retrieve from parsed configuration
	modelConfig := g.generateModelConfig(node)

	// Set prompt_template field
	promptTemplate := g.generatePromptTemplate(node)

	// Set vision field - dynamically retrieve from parsed configuration
	visionConfig := g.generateVisionConfig(node)

	// Set directly to data field, consistent with official example format
	data.Context = contextConfig
	data.Model = modelConfig
	data.PromptTemplate = promptTemplate
	data.Variables = []interface{}{} // Empty interface{} array, consistent with official example
	data.Vision = visionConfig
}

// generateContextConfig generates context configuration
func (g *LLMNodeGenerator) generateContextConfig(node models.Node) map[string]interface{} {
	// LLM nodes in Dify should have disabled context with empty variable_selector
	// Variable references should only exist in prompt template, not in context configuration
	context := map[string]interface{}{
		"enabled":           false,
		"variable_selector": []interface{}{}, // Empty array, consistent with correct Dify examples
	}

	return context
}

// generatePromptTemplate generates prompt template
func (g *LLMNodeGenerator) generatePromptTemplate(node models.Node) []map[string]interface{} {
	// Extract original system template from platform configuration
	systemTemplate := "You are a helpful assistant."
	if node.PlatformConfig.IFlytek != nil {
		if config, ok := node.PlatformConfig.IFlytek["nodeParam"].(map[string]interface{}); ok {
			if template, exists := config["systemTemplate"].(string); exists && template != "" {
				systemTemplate = template
			}
		}
	}

	// Use node description if no platform configuration available
	if systemTemplate == "You are a helpful assistant." && node.Description != "" {
		systemTemplate = node.Description
	}

	// Fix variable reference format in system template
	systemTemplate = g.fixVariableReferences(systemTemplate, node)

	template := []map[string]interface{}{
		{
			"id":   generateRandomUUID(), // Generate random template ID
			"role": "system",
			"text": systemTemplate,
		},
	}

	return template
}

// fixVariableReferences corrects variable reference format from iFlytek to Dify
func (g *LLMNodeGenerator) fixVariableReferences(text string, node models.Node) string {
	replacements := g.buildVariableReplacements(text, node)
	return g.applyReplacements(text, replacements)
}

// buildVariableReplacements builds complete variable replacement mapping
func (g *LLMNodeGenerator) buildVariableReplacements(text string, node models.Node) map[string]string {
	replacements := make(map[string]string)

	// First priority: build from direct input references
	g.addDirectInputReplacements(node, replacements)

	// Second priority: extract from template if no direct references
	if len(replacements) == 0 {
		replacements = g.extractVariablesFromTemplate(text, node)
	}

	// Third priority: handle template variables with source node inference
	g.addTemplateVariableReplacements(text, node, replacements)

	return replacements
}

// addDirectInputReplacements adds replacements from direct input references
func (g *LLMNodeGenerator) addDirectInputReplacements(node models.Node, replacements map[string]string) {
	for _, input := range node.Inputs {
		if g.hasValidReference(input.Reference) {
			starPattern := fmt.Sprintf("{{%s}}", input.Reference.OutputName)
			difyPattern := fmt.Sprintf("{{#%s.%s#}}", input.Reference.NodeID, input.Reference.OutputName)
			replacements[starPattern] = difyPattern
		}
	}
}

// addTemplateVariableReplacements adds replacements for template variables with source node inference
func (g *LLMNodeGenerator) addTemplateVariableReplacements(text string, node models.Node, replacements map[string]string) {
	if len(node.Inputs) == 0 {
		return
	}

	sourceNodeID := g.findSourceNodeID(node)
	if sourceNodeID == "" {
		return
	}

	templateVars := g.extractVariableNamesFromTemplate(text)
	for _, varName := range templateVars {
		starPattern := fmt.Sprintf("{{%s}}", varName)
		if _, exists := replacements[starPattern]; !exists {
			outputFieldName := g.inferOutputFieldName(sourceNodeID, varName)
			difyPattern := fmt.Sprintf("{{#%s.%s#}}", sourceNodeID, outputFieldName)
			replacements[starPattern] = difyPattern
		}
	}
}

// hasValidReference checks if input reference is valid
func (g *LLMNodeGenerator) hasValidReference(ref *models.VariableReference) bool {
	return ref != nil && ref.NodeID != "" && ref.OutputName != ""
}

// findSourceNodeID finds the first available source node ID from inputs
func (g *LLMNodeGenerator) findSourceNodeID(node models.Node) string {
	for _, input := range node.Inputs {
		if input.Reference != nil && input.Reference.NodeID != "" {
			return input.Reference.NodeID
		}
	}
	return ""
}

// applyReplacements applies all replacements to text
func (g *LLMNodeGenerator) applyReplacements(text string, replacements map[string]string) string {
	result := text
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// Post-process to fix common template syntax issues
	result = g.fixCommonTemplateSyntaxIssues(result)

	return result
}

// fixCommonTemplateSyntaxIssues fixes common template syntax problems
func (g *LLMNodeGenerator) fixCommonTemplateSyntaxIssues(text string) string {
	// This function is kept for potential future use
	// Main fixing is now handled in the replacement logic itself
	return text
}

// extractVariablesFromTemplate extracts variables from template and attempts mapping to node references
func (g *LLMNodeGenerator) extractVariablesFromTemplate(text string, node models.Node) map[string]string {
	replacements := make(map[string]string)

	// Find all variable references in template {{variableName}}
	start := 0
	for {
		startIdx := strings.Index(text[start:], "{{")
		if startIdx == -1 {
			break
		}
		startIdx += start

		endIdx := strings.Index(text[startIdx:], "}}")
		if endIdx == -1 {
			// Handle incomplete template variables - look for likely end patterns
			restText := text[startIdx+2:]

			// Find variable name by looking for the next delimiter (space, quote, newline, etc.)
			var varName string

			// Look for common delimiters that might end a variable name
			delimiters := []string{" ", "\"", "\n", "\r", "\t", "|", ":", "，", "。"}
			minIdx := len(restText)

			for _, delim := range delimiters {
				if idx := strings.Index(restText, delim); idx != -1 && idx < minIdx {
					minIdx = idx
				}
			}

			if minIdx < len(restText) && minIdx > 0 {
				varName = strings.TrimSpace(restText[:minIdx])
			} else if len(restText) > 0 {
				varName = strings.TrimSpace(restText)
			}

			if varName != "" {
				sourceNodeID, outputFieldName := g.findSourceNodeAndOutputForVariable(varName, node)
				if sourceNodeID != "" && outputFieldName != "" {
					// Handle both quoted and unquoted patterns
					// Pattern 1: "{{varName"  -> "{{#nodeId.field#}}"
					quotedIncompletePattern := fmt.Sprintf(`"{{%s"`, varName)
					quotedDifyPattern := fmt.Sprintf(`"{{#%s.%s#}}"`, sourceNodeID, outputFieldName)
					replacements[quotedIncompletePattern] = quotedDifyPattern

					// Pattern 2: {{varName  -> {{#nodeId.field#}}
					incompletePattern := fmt.Sprintf("{{%s", varName)
					difyPattern := fmt.Sprintf("{{#%s.%s#}}", sourceNodeID, outputFieldName)
					replacements[incompletePattern] = difyPattern
				}
			}
			break
		}
		endIdx += startIdx

		// Extract variable name
		varName := strings.TrimSpace(text[startIdx+2 : endIdx])
		if varName != "" {
			// Try to find corresponding input reference
			sourceNodeID, outputFieldName := g.findSourceNodeAndOutputForVariable(varName, node)
			if sourceNodeID != "" && outputFieldName != "" {
				starPattern := fmt.Sprintf("{{%s}}", varName)
				difyPattern := fmt.Sprintf("{{#%s.%s#}}", sourceNodeID, outputFieldName)
				replacements[starPattern] = difyPattern
			}
		}

		start = endIdx + 2
	}

	return replacements
}

// findSourceNodeAndOutputForVariable finds source node ID and correct output field name for variable
func (g *LLMNodeGenerator) findSourceNodeAndOutputForVariable(varName string, node models.Node) (string, string) {
	// First check direct input references
	for _, input := range node.Inputs {
		if input.Reference != nil && input.Reference.OutputName == varName {
			return input.Reference.NodeID, input.Reference.OutputName
		}
	}

	// If no direct match, try to find most likely source node from inputs
	// and infer correct output field name based on source node type
	for _, input := range node.Inputs {
		if input.Reference != nil && input.Reference.NodeID != "" {
			// Infer output field name based on source node type
			outputFieldName := g.inferOutputFieldName(input.Reference.NodeID, varName)
			return input.Reference.NodeID, outputFieldName
		}
	}

	return "", ""
}

// inferOutputFieldName infers correct output field name based on source node type
func (g *LLMNodeGenerator) inferOutputFieldName(sourceNodeID, varName string) string {
	// Try to get node type from node mapping
	nodeType := g.getNodeTypeFromConverter(sourceNodeID)
	if nodeType != "" {
		return g.mapOutputFieldByNodeType(nodeType, varName)
	}

	// Fallback to ID-based heuristic method
	if strings.Contains(sourceNodeID, "decision-making") ||
		strings.Contains(sourceNodeID, "classifier") {
		return "class_name"
	}

	// Heuristic judgment based on variable name
	if strings.Contains(varName, "step") || strings.Contains(varName, "class") ||
		strings.Contains(varName, "category") || strings.Contains(varName, "type") {
		return "class_name"
	}

	return varName
}

// getNodeTypeFromConverter gets node type from node mapping
func (g *LLMNodeGenerator) getNodeTypeFromConverter(nodeID string) string {
	if node, exists := g.nodeMapping[nodeID]; exists {
		return string(node.Type)
	}
	return ""
}

// mapOutputFieldByNodeType maps output field names based on node type
func (g *LLMNodeGenerator) mapOutputFieldByNodeType(nodeType, varName string) string {
	switch nodeType {
	case "classifier", "question-classifier":
		return "class_name"
	case "llm":
		return "text"
	case "code":
		// Code node output field name usually matches variable name
		return varName
	case "iteration":
		return "output"
	case "start":
		return varName
	default:
		return varName
	}
}

// findSourceNodeForVariable finds source node ID for variable (backward compatibility)
func (g *LLMNodeGenerator) findSourceNodeForVariable(varName string, node models.Node) string {
	sourceNodeID, _ := g.findSourceNodeAndOutputForVariable(varName, node)
	return sourceNodeID
}

// extractVariableNamesFromTemplate extracts all variable names from template
func (g *LLMNodeGenerator) extractVariableNamesFromTemplate(text string) []string {
	var varNames []string
	seen := make(map[string]bool)
	start := 0

	for {
		varInfo := g.findNextVariablePattern(text, start)
		if varInfo == nil {
			break
		}

		varName := g.processVariableExtraction(text, varInfo)
		if varName != "" && !seen[varName] {
			varNames = append(varNames, varName)
			seen[varName] = true
		}

		start = varInfo.nextStart
	}

	return varNames
}

// variablePatternInfo contains information about found variable pattern
type variablePatternInfo struct {
	startIdx   int
	endIdx     int
	nextStart  int
	hasClosing bool
}

// findNextVariablePattern finds next variable pattern in text
func (g *LLMNodeGenerator) findNextVariablePattern(text string, start int) *variablePatternInfo {
	startIdx := strings.Index(text[start:], "{{")
	if startIdx == -1 {
		return nil
	}
	startIdx += start

	endIdx := strings.Index(text[startIdx:], "}}")
	if endIdx == -1 {
		return &variablePatternInfo{
			startIdx:   startIdx,
			endIdx:     -1,
			nextStart:  startIdx + 2,
			hasClosing: false,
		}
	}

	endIdx += startIdx
	return &variablePatternInfo{
		startIdx:   startIdx,
		endIdx:     endIdx,
		nextStart:  endIdx + 2,
		hasClosing: true,
	}
}

// processVariableExtraction processes variable extraction from pattern info
func (g *LLMNodeGenerator) processVariableExtraction(text string, varInfo *variablePatternInfo) string {
	if varInfo.hasClosing {
		return g.extractValidVariableName(text, varInfo.startIdx, varInfo.endIdx)
	}
	return g.extractIncompleteVariableName(text, varInfo.startIdx)
}

// extractValidVariableName extracts variable name from complete pattern
func (g *LLMNodeGenerator) extractValidVariableName(text string, startIdx, endIdx int) string {
	varName := strings.TrimSpace(text[startIdx+2 : endIdx])
	if g.isSimpleVariableName(varName) {
		return varName
	}
	return ""
}

// extractIncompleteVariableName extracts variable name from incomplete pattern
func (g *LLMNodeGenerator) extractIncompleteVariableName(text string, startIdx int) string {
	remainingText := text[startIdx+2:]
	varEndIdx := g.findVariableEndBoundary(remainingText)
	if varEndIdx <= 0 {
		return ""
	}

	varName := strings.TrimSpace(remainingText[:varEndIdx])
	if g.isValidVariableName(varName) {
		return varName
	}
	return ""
}

// isSimpleVariableName checks if variable name is simple (not already converted)
func (g *LLMNodeGenerator) isSimpleVariableName(varName string) bool {
	return varName != "" && !strings.Contains(varName, "#") && !strings.Contains(varName, ".")
}

// findVariableEndBoundary finds possible ending boundary for variable name
func (g *LLMNodeGenerator) findVariableEndBoundary(text string) int {
	// Common variable name ending markers: Chinese characters, spaces, punctuation
	boundaryChars := []rune("这是为了 。，！？：；（）【】")

	for i, char := range text {
		// Check if Chinese characters encountered
		if char >= 0x4e00 && char <= 0x9fff {
			return i
		}

		// Check if common boundary characters encountered
		for _, boundary := range boundaryChars {
			if char == boundary {
				return i
			}
		}

		// Spaces may also be boundaries
		if char == ' ' || char == '\t' || char == '\n' {
			return i
		}

		// Force truncation if exceeding reasonable variable name length (30 characters)
		if i >= 30 {
			return i
		}
	}

	// Return length of remaining text if no clear boundary found
	return len(text)
}

// isValidVariableName validates whether variable name is legal
func (g *LLMNodeGenerator) isValidVariableName(varName string) bool {
	if !g.isValidVariableLength(varName) {
		return false
	}
	return g.hasOnlyValidChars(varName)
}

// isValidVariableLength checks if variable name length is valid
func (g *LLMNodeGenerator) isValidVariableLength(varName string) bool {
	return len(varName) > 0 && len(varName) <= 50
}

// hasOnlyValidChars checks if variable name contains only valid characters
func (g *LLMNodeGenerator) hasOnlyValidChars(varName string) bool {
	for _, char := range varName {
		if !g.isValidVariableChar(char) {
			return false
		}
	}
	return true
}

// isValidVariableChar checks if a character is valid for variable names
func (g *LLMNodeGenerator) isValidVariableChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '_'
}

// restoreDifyPlatformConfig restores Dify platform-specific configuration
func (g *LLMNodeGenerator) restoreDifyPlatformConfig(config map[string]interface{}, node *DifyNode) {
	// Restore model configuration
	if modelConfig, exists := config["model"].(map[string]interface{}); exists {
		if node.Data.Config == nil {
			node.Data.Config = make(map[string]interface{})
		}
		node.Data.Config["model"] = modelConfig
	}

	// Restore prompt template
	if promptTemplate, exists := config["prompt_template"].([]interface{}); exists {
		if node.Data.Config == nil {
			node.Data.Config = make(map[string]interface{})
		}
		node.Data.Config["prompt_template"] = promptTemplate
	}

	// Restore vision configuration
	if visionConfig, exists := config["vision"].(map[string]interface{}); exists {
		if node.Data.Config == nil {
			node.Data.Config = make(map[string]interface{})
		}
		node.Data.Config["vision"] = visionConfig
	}

	// Restore other node-specific configuration
	if desc, ok := config["desc"].(string); ok {
		node.Data.Desc = desc
	}
	if title, ok := config["title"].(string); ok {
		node.Data.Title = title
	}
}
