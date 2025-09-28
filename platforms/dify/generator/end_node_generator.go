package generator

import (
	"fmt"
	"github.com/iflytek/agentbridge/internal/models"
	"strings"
)

// EndNodeGenerator generates end nodes
type EndNodeGenerator struct {
	*BaseNodeGenerator
}

func NewEndNodeGenerator() *EndNodeGenerator {
	return &EndNodeGenerator{
		BaseNodeGenerator: NewBaseNodeGenerator(models.NodeTypeEnd),
	}
}

// GenerateNode generates an end node
func (g *EndNodeGenerator) GenerateNode(node models.Node) (DifyNode, error) {
	if node.Type != models.NodeTypeEnd {
		return DifyNode{}, fmt.Errorf("unsupported node type: %s, expected: %s", node.Type, models.NodeTypeEnd)
	}

	// Generate base node structure
	difyNode := g.generateBaseNode(node)

	// Set end node specific data
	// For end nodes, generate outputs from inputs (conceptual mapping)
	difyNode.Data.Outputs = g.generateOutputsFromInputs(node.Inputs)

	// End node needs empty config field (outputs directly at data level)
	if difyNode.Data.Config == nil {
		difyNode.Data.Config = make(map[string]interface{})
	}

	// Restore Dify-specific fields from platform configuration
	if difyConfig := node.PlatformConfig.Dify; difyConfig != nil {
		g.restoreDifyPlatformConfig(difyConfig, &difyNode)
	}

	return difyNode, nil
}

// GenerateNodeWithWorkflowContext generates an end node (with workflow context)
func (g *EndNodeGenerator) GenerateNodeWithWorkflowContext(node models.Node, workflow *models.Workflow) (DifyNode, error) {
	if node.Type != models.NodeTypeEnd {
		return DifyNode{}, fmt.Errorf("unsupported node type: %s, expected: %s", node.Type, models.NodeTypeEnd)
	}

	// Generate base node structure
	difyNode := g.generateBaseNode(node)

	// Build end node outputs with workflow context
	difyNode.Data.Outputs = g.generateSmartOutputsFromWorkflow(node, workflow)

	// End node needs empty config field (outputs directly at data level)
	if difyNode.Data.Config == nil {
		difyNode.Data.Config = make(map[string]interface{})
	}

	// Restore Dify-specific fields from platform configuration
	if difyConfig := node.PlatformConfig.Dify; difyConfig != nil {
		g.restoreDifyPlatformConfig(difyConfig, &difyNode)
	}

	return difyNode, nil
}

// generateOutputsFromInputs generates output definitions from inputs (for end nodes)
func (g *EndNodeGenerator) generateOutputsFromInputs(inputs []models.Input) []DifyOutput {
	difyOutputs := make([]DifyOutput, 0, len(inputs))

	for _, input := range inputs {
		// Only generate fields actually needed by Dify: variable, value_selector, value_type
		difyOutput := DifyOutput{
			Variable:      input.Name,
			ValueSelector: []string{}, // Will be filled by actual reference relationships
		}

		// Set value_type based on input type (Dify format)
		difyOutput.ValueType = g.mapUnifiedTypeToString(input.Type)

		// If there's reference information, set value_selector
		if input.Reference != nil {
			outputName := input.Reference.OutputName

			// Infer node type from source node ID, then map output name
			mappedOutputName := g.mapOutputNameByNodeID(input.Reference.NodeID, outputName)

			difyOutput.ValueSelector = []string{
				input.Reference.NodeID,
				mappedOutputName,
			}
		}

		difyOutputs = append(difyOutputs, difyOutput)
	}

	return difyOutputs
}

// generateSmartOutputsFromWorkflow builds end node outputs with workflow context
func (g *EndNodeGenerator) generateSmartOutputsFromWorkflow(endNode models.Node, workflow *models.Workflow) []DifyOutput {
	// Handle explicit inputs and use workflow to infer types and mapped output names
	if len(endNode.Inputs) > 0 {
		outputs := make([]DifyOutput, 0, len(endNode.Inputs))
		for _, input := range endNode.Inputs {
			out := DifyOutput{
				Variable:      input.Name,
				ValueSelector: []string{},
			}

			// Default value_type based on declared input type
			out.ValueType = g.mapUnifiedTypeToString(input.Type)

			if input.Reference != nil && input.Reference.NodeID != "" {
				// Find source node in workflow for accurate type inference
				sourceNode := g.findNodeByID(input.Reference.NodeID, workflow.Nodes)
				// Map output name according to platform-specific rules
				mappedOutputName := g.mapOutputNameByNodeID(input.Reference.NodeID, input.Reference.OutputName)
				out.ValueSelector = []string{input.Reference.NodeID, mappedOutputName}

				// If source node is found, prefer its output type for value_type
				if sourceNode != nil {
					inferredType := g.getNodeOutputType(sourceNode)
					if inferredType != "" {
						out.ValueType = inferredType
						out.Type = inferredType
					}
				} else {
					// Fallback for iteration pattern by NodeID prefix
					if strings.HasPrefix(input.Reference.NodeID, "iteration::") {
						out.ValueType = "array[string]"
						out.Type = "array[string]"
					}
				}
			}

			outputs = append(outputs, out)
		}
		return outputs
	}

	// If no input configuration, infer from connection relationships
	incomingEdges := g.findIncomingEdges(endNode.ID, workflow.Edges)
	difyOutputs := make([]DifyOutput, 0)

	// Create an output for each source node connected to end node
	for _, edge := range incomingEdges {
		sourceNode := g.findNodeByID(edge.Source, workflow.Nodes)
		if sourceNode == nil {
			continue
		}

		// Determine output name and type based on source node type
		outputName := g.getNodeOutputName(sourceNode)
		outputType := g.getNodeOutputType(sourceNode)

		// Use source node title or ID as variable name, not hardcoded result format
		variableName := g.generateVariableNameFromNode(sourceNode)

		// Map output name using node type information
		mappedOutputName := g.mapOutputNameToDifyStandardWithNodeType(sourceNode.Type, outputName)

		difyOutput := DifyOutput{
			Variable:      variableName,
			ValueSelector: []string{sourceNode.ID, mappedOutputName},
			ValueType:     outputType,
			Type:          outputType,
		}

		difyOutputs = append(difyOutputs, difyOutput)
	}

	return difyOutputs
}

// generateVariableNameFromNode generates meaningful variable names from node information
func (g *EndNodeGenerator) generateVariableNameFromNode(node *models.Node) string {
	// Try to get name from outputs
	if name := g.getNameFromOutputs(node); name != "" {
		return name
	}

	// Try to get name from node title
	if name := g.getNameFromTitle(node); name != "" {
		return name
	}

	// Generate generic name by node type
	return g.getGenericNameByType(node.Type)
}

// getNameFromOutputs gets variable name from node outputs
func (g *EndNodeGenerator) getNameFromOutputs(node *models.Node) string {
	if len(node.Outputs) == 0 {
		return ""
	}

	// Single output case
	if len(node.Outputs) == 1 {
		return g.getValidOutputName(node.Outputs[0])
	}

	// Multiple outputs case
	return g.getFirstValidOutputName(node.Outputs)
}

// getValidOutputName returns valid output name or empty string
func (g *EndNodeGenerator) getValidOutputName(output models.Output) string {
	if g.isValidOutputName(output.Name) {
		return g.sanitizeVariableName(output.Name)
	}
	return ""
}

// getFirstValidOutputName gets first valid output name from multiple outputs
func (g *EndNodeGenerator) getFirstValidOutputName(outputs []models.Output) string {
	for _, output := range outputs {
		if name := g.getValidOutputName(output); name != "" {
			return name
		}
	}
	return ""
}

// isValidOutputName checks if output name is valid (not empty or system built-in)
func (g *EndNodeGenerator) isValidOutputName(name string) bool {
	return name != "" && name != "AGENT_USER_INPUT"
}

// getNameFromTitle gets variable name from node title
func (g *EndNodeGenerator) getNameFromTitle(node *models.Node) string {
	if node.Title != "" {
		return g.sanitizeVariableName(node.Title)
	}
	return ""
}

// getGenericNameByType generates generic variable name by node type
func (g *EndNodeGenerator) getGenericNameByType(nodeType models.NodeType) string {
	typeNames := map[models.NodeType]string{
		models.NodeTypeLLM:        "llm_result",
		models.NodeTypeCode:       "code_output",
		models.NodeTypeClassifier: "classifier_result",
		models.NodeTypeIteration:  "iteration_result",
	}

	if name, exists := typeNames[nodeType]; exists {
		return name
	}
	return "node_output"
}

// sanitizeVariableName cleans variable name to ensure compliance with naming conventions
func (g *EndNodeGenerator) sanitizeVariableName(name string) string {
	result := g.processVariableChars(name)
	result = g.normalizeVariableName(result)
	return g.ensureNonEmptyVariableName(result)
}

// processVariableChars processes variable characters
func (g *EndNodeGenerator) processVariableChars(name string) string {
	result := ""
	for _, char := range name {
		if g.isValidVariableChar(char) {
			result += string(char)
		} else if g.shouldReplaceWithUnderscore(char) {
			result += "_"
		}
	}
	return result
}

// isValidVariableChar checks if character is valid for variable names
func (g *EndNodeGenerator) isValidVariableChar(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')
}

// shouldReplaceWithUnderscore checks if character should be replaced with underscore
func (g *EndNodeGenerator) shouldReplaceWithUnderscore(char rune) bool {
	return char == ' ' || char == '-'
}

// normalizeVariableName normalizes variable name to lowercase
func (g *EndNodeGenerator) normalizeVariableName(result string) string {
	return strings.ToLower(result)
}

// ensureNonEmptyVariableName ensures variable name is not empty
func (g *EndNodeGenerator) ensureNonEmptyVariableName(result string) string {
	if result == "" {
		return "output"
	}
	return result
}

// findIncomingEdges finds all edges connecting to the specified node
func (g *EndNodeGenerator) findIncomingEdges(nodeID string, edges []models.Edge) []models.Edge {
	incomingEdges := make([]models.Edge, 0)
	for _, edge := range edges {
		if edge.Target == nodeID {
			incomingEdges = append(incomingEdges, edge)
		}
	}
	return incomingEdges
}

// findNodeByID finds node by ID
func (g *EndNodeGenerator) findNodeByID(nodeID string, nodes []models.Node) *models.Node {
	for _, node := range nodes {
		if node.ID == nodeID {
			return &node
		}
	}
	return nil
}

// getNodeOutputName gets standard output name based on node type
func (g *EndNodeGenerator) getNodeOutputName(node *models.Node) string {
	switch node.Type {
	case models.NodeTypeLLM:
		return "text" // LLM node standard output name in Dify
	case models.NodeTypeCode:
		return "result" // Code node standard output name
	case models.NodeTypeClassifier:
		return "class_name" // Classifier node standard output name
	case models.NodeTypeIteration:
		return "output" // Iteration node standard output name
	default:
		return "output" // Default output name
	}
}

// getNodeOutputType gets output data type based on node type
func (g *EndNodeGenerator) getNodeOutputType(node *models.Node) string {
	switch node.Type {
	case models.NodeTypeLLM:
		return "string" // LLM node output string
	case models.NodeTypeCode:
		// Get actual type from node's output definition
		if len(node.Outputs) > 0 {
			return g.mapUnifiedTypeToString(node.Outputs[0].Type)
		}
		return "string" // Default string
	case models.NodeTypeClassifier:
		return "string" // Classifier output string
	case models.NodeTypeIteration:
		return "array[string]" // Iteration node usually outputs an array
	default:
		return "string" // Default string
	}
}

// mapUnifiedTypeToString maps unified DSL types to strings
func (g *EndNodeGenerator) mapUnifiedTypeToString(dataType models.UnifiedDataType) string {
	// Use unified mapping system
	mapping := models.GetDefaultDataTypeMapping()
	return mapping.ToDifyType(dataType)
}

// mapOutputNameToDifyStandardWithNodeType maps output names to Dify standard format based on node type
func (g *EndNodeGenerator) mapOutputNameToDifyStandardWithNodeType(nodeType models.NodeType, outputName string) string {
	switch nodeType {
	case models.NodeTypeLLM:
		// LLM node's output in Dify is standard name "text"
		if outputName == "output" {
			return "text"
		}
	case models.NodeTypeCode:
		// Code node output name remains unchanged
		return outputName
	case models.NodeTypeStart:
		// Start node output name remains unchanged
		return outputName
	case models.NodeTypeIteration:
		// Iteration node output name remains unchanged
		return outputName
	case models.NodeTypeClassifier:
		// Classifier node output name remains unchanged
		return outputName
	}

	// Default to original name
	return outputName
}

// mapOutputNameByNodeID infers node type from node ID and maps output name
func (g *EndNodeGenerator) mapOutputNameByNodeID(nodeID, outputName string) string {
	// Infer node type from node ID prefix and map output name with platform-specific rules
	if strings.HasPrefix(nodeID, "spark-llm::") {
		// iFlytek SparkAgent LLM node output name is "output", needs to be mapped to Dify's "text"
		if outputName == "output" {
			return "text"
		}
		return outputName
	} else if strings.HasPrefix(nodeID, "ifly-code::") {
		// Code node output name is user-defined on both platforms, remains unchanged
		// iFlytek SparkAgent: "result", Dify: user-defined (usually "result")
		return outputName
	} else if strings.HasPrefix(nodeID, "node-start::") {
		// Start node output name is user-defined on both platforms, remains unchanged
		return outputName
	} else if strings.HasPrefix(nodeID, "iteration::") {
		// Iteration node output name needs special handling: always "output" in Dify
		// But ensure iteration node's output_selector correctly points to the final processing node
		return "output" // Dify iteration node standard output name
	} else if strings.HasPrefix(nodeID, "decision-making::") {
		// Classifier node output name is fixed "class_name" on both platforms, remains unchanged
		return outputName
	}

	// Default to original name
	return outputName
}

// restoreDifyPlatformConfig restores Dify platform-specific configuration
func (g *EndNodeGenerator) restoreDifyPlatformConfig(config map[string]interface{}, node *DifyNode) {
	g.restoreOutputsConfig(config, node)
	g.restoreNodeMetadata(config, node)
}

// restoreOutputsConfig restores output configuration
func (g *EndNodeGenerator) restoreOutputsConfig(config map[string]interface{}, node *DifyNode) {
	outputsConfig, exists := config["outputs"].([]interface{})
	if !exists {
		return
	}

	outputs := make([]DifyOutput, 0, len(outputsConfig))
	for _, outputConfig := range outputsConfig {
		if output := g.parseOutputConfig(outputConfig); output != nil {
			outputs = append(outputs, *output)
		}
	}
	node.Data.Outputs = outputs
}

// parseOutputConfig parses single output configuration
func (g *EndNodeGenerator) parseOutputConfig(outputConfig interface{}) *DifyOutput {
	outputMap, ok := outputConfig.(map[string]interface{})
	if !ok {
		return nil
	}

	output := DifyOutput{ValueSelector: []string{}}

	if variable, ok := outputMap["variable"].(string); ok {
		output.Variable = variable
	}
	if valueType, ok := outputMap["value_type"].(string); ok {
		output.Type = valueType
	}

	g.parseValueSelector(outputMap, &output)
	return &output
}

// parseValueSelector parses value selector from output map
func (g *EndNodeGenerator) parseValueSelector(outputMap map[string]interface{}, output *DifyOutput) {
	valueSelector, ok := outputMap["value_selector"].([]interface{})
	if !ok {
		return
	}

	for _, selector := range valueSelector {
		if selectorStr, ok := selector.(string); ok {
			output.ValueSelector = append(output.ValueSelector, selectorStr)
		}
	}
}

// restoreNodeMetadata restores node metadata configuration
func (g *EndNodeGenerator) restoreNodeMetadata(config map[string]interface{}, node *DifyNode) {
	if desc, ok := config["desc"].(string); ok {
		node.Data.Desc = desc
	}
	if title, ok := config["title"].(string); ok {
		node.Data.Title = title
	}
}
