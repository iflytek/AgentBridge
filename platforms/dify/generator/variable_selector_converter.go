package generator

import (
	"agentbridge/internal/models"
	"fmt"
	"strings"
)

// VariableSelectorConverter converts variable selectors
type VariableSelectorConverter struct {
	nodeMapping      map[string]models.Node // Node ID to node mapping
	iterationContext string                 // Current iteration node ID (for handling iteration-node-start references)
}

func NewVariableSelectorConverter() *VariableSelectorConverter {
	return &VariableSelectorConverter{
		nodeMapping: make(map[string]models.Node),
	}
}

// SetNodeMapping sets node mapping
func (c *VariableSelectorConverter) SetNodeMapping(nodes []models.Node) {
	c.nodeMapping = make(map[string]models.Node, len(nodes))
	for _, node := range nodes {
		c.nodeMapping[node.ID] = node
	}
}

// SetIterationContext sets current iteration context for handling iteration-node-start references
func (c *VariableSelectorConverter) SetIterationContext(iterationNodeID string) {
	c.iterationContext = iterationNodeID
}

// ConvertVariableReference converts variable reference to Dify variable selector
func (c *VariableSelectorConverter) ConvertVariableReference(ref *models.VariableReference) ([]string, error) {
	if ref == nil {
		return []string{}, nil
	}

	switch ref.Type {
	case models.ReferenceTypeNodeOutput:
		return c.convertNodeOutputReference(ref)
	case models.ReferenceTypeLiteral:
		return c.convertLiteralReference(ref)
	case models.ReferenceTypeTemplate:
		return c.convertTemplateReference(ref)
	default:
		return []string{}, fmt.Errorf("unsupported reference type: %s", ref.Type)
	}
}

// convertNodeOutputReference converts node output reference
func (c *VariableSelectorConverter) convertNodeOutputReference(ref *models.VariableReference) ([]string, error) {
	if ref.NodeID == "" {
		return []string{}, fmt.Errorf("node ID is required for node output reference")
	}

	outputName := ref.OutputName
	if outputName == "" {
		outputName = "output" // Default output name
	}

	// Special handling for iteration-node-start references within iteration context
	// In iFlytek: iteration-node-start::xxx.input → In Dify: iteration.item
	if c.iterationContext != "" &&
		strings.Contains(ref.NodeID, "iteration-node-start") &&
		outputName == "input" {
		return []string{c.iterationContext, "item"}, nil
	}

	// Map to Dify platform fixed field names
	outputName = c.mapToDifyOutputField(ref.NodeID, outputName)

	return []string{ref.NodeID, outputName}, nil
}

// mapToDifyOutputField maps output field names to Dify platform fixed fields
func (c *VariableSelectorConverter) mapToDifyOutputField(nodeID, originalFieldName string) string {
	// Get node information
	node, exists := c.nodeMapping[nodeID]
	if !exists {
		return originalFieldName // Keep original field name if node not found
	}

	// Map to Dify fixed fields based on node type
	switch node.Type {
	case models.NodeTypeLLM:
		if originalFieldName == "output" {
			return "text" // Dify LLM node uses 'text' as output field
		}
	case models.NodeTypeClassifier:
		if originalFieldName == "class_name" {
			return "class_name" // Classifier node keeps original field name
		}
	case models.NodeTypeCode:
		// Code node output field names remain user-defined
		return originalFieldName
	case models.NodeTypeIteration:
		if originalFieldName == "output" {
			return "output" // Iteration node keeps output field name
		}
	}

	return originalFieldName
}

// convertLiteralReference converts literal reference
func (c *VariableSelectorConverter) convertLiteralReference(ref *models.VariableReference) ([]string, error) {
	// Literals usually don't need variable selectors, but may require special handling
	if ref.Value != nil {
		return []string{fmt.Sprintf("%v", ref.Value)}, nil
	}
	return []string{}, nil
}

// convertTemplateReference converts template reference
func (c *VariableSelectorConverter) convertTemplateReference(ref *models.VariableReference) ([]string, error) {
	if ref.Template == "" {
		return []string{}, fmt.Errorf("template is required for template reference")
	}

	// Parse variable references in template
	return c.parseTemplateVariables(ref.Template)
}

// parseTemplateVariables parses variables in template
func (c *VariableSelectorConverter) parseTemplateVariables(template string) ([]string, error) {
	variables := []string{}
	start := 0

	for {
		varRange := c.findNextVariableRange(template, start)
		if varRange.startIdx == -1 {
			break
		}

		varContent := c.extractVariableContent(template, varRange)
		if varContent != "" {
			variableParts := c.parseVariableContent(varContent)
			variables = append(variables, variableParts...)
		}

		start = varRange.endIdx + 2
	}

	return variables, nil
}

// variableRange represents a variable pattern range in template
type variableRange struct {
	startIdx int
	endIdx   int
}

// findNextVariableRange finds next variable range in template
func (c *VariableSelectorConverter) findNextVariableRange(template string, start int) variableRange {
	startIdx := strings.Index(template[start:], "{{")
	if startIdx == -1 {
		return variableRange{startIdx: -1, endIdx: -1}
	}
	startIdx += start

	endIdx := strings.Index(template[startIdx:], "}}")
	if endIdx == -1 {
		return variableRange{startIdx: -1, endIdx: -1}
	}
	endIdx += startIdx

	return variableRange{startIdx: startIdx, endIdx: endIdx}
}

// extractVariableContent extracts variable content from range
func (c *VariableSelectorConverter) extractVariableContent(template string, varRange variableRange) string {
	return strings.TrimSpace(template[varRange.startIdx+2 : varRange.endIdx])
}

// parseVariableContent parses variable content into parts
func (c *VariableSelectorConverter) parseVariableContent(varContent string) []string {
	// Parse variable reference format, e.g., "nodeId.outputName"
	parts := strings.Split(varContent, ".")
	if len(parts) >= 2 {
		return []string{parts[0], parts[1]}
	}
	return []string{varContent}
}

// updateVariableSelectorsByType updates variable selectors based on node type
func (c *VariableSelectorConverter) updateVariableSelectorsByType(node *DifyNode, unifiedNode models.Node) error {
	updaters := c.getNodeTypeUpdaters()

	updater, exists := updaters[unifiedNode.Type]
	if !exists {
		return nil
	}

	return updater(node, unifiedNode)
}

// getNodeTypeUpdaters returns a map of node type to updater functions
func (c *VariableSelectorConverter) getNodeTypeUpdaters() map[models.NodeType]func(*DifyNode, models.Node) error {
	return map[models.NodeType]func(*DifyNode, models.Node) error{
		models.NodeTypeLLM:        c.updateLLMVariableSelectors,
		models.NodeTypeCondition:  c.updateConditionVariableSelectors,
		models.NodeTypeClassifier: c.updateClassifierVariableSelectors,
		models.NodeTypeIteration:  c.updateIterationVariableSelectors,
	}
}

// updateLLMVariableSelectors updates variable selectors for LLM nodes
func (c *VariableSelectorConverter) updateLLMVariableSelectors(node *DifyNode, unifiedNode models.Node) error {
	// Here you can update variable selectors in the prompt template based on input references
	// The specific implementation depends on how variable references are represented in the prompt template
	return nil
}

// updateConditionVariableSelectors updates variable selectors for condition nodes
func (c *VariableSelectorConverter) updateConditionVariableSelectors(node *DifyNode, unifiedNode models.Node) error {
	if cases, ok := node.Data.Config["cases"].([]map[string]interface{}); ok {
		for _, caseItem := range cases {
			if conditions, ok := caseItem["conditions"].([]map[string]interface{}); ok {
				for i := range conditions {
					// Try to get variable selectors from input references
					if len(unifiedNode.Inputs) > 0 {
						firstInput := unifiedNode.Inputs[0]
						if firstInput.Reference != nil {
							selector, err := c.ConvertVariableReference(firstInput.Reference)
							if err == nil && len(selector) > 0 {
								conditions[i]["variable_selector"] = selector
							}
						}
					}
				}
			}
		}
	}
	return nil
}

// updateClassifierVariableSelectors updates variable selectors for classifier nodes
func (c *VariableSelectorConverter) updateClassifierVariableSelectors(node *DifyNode, unifiedNode models.Node) error {
	if len(unifiedNode.Inputs) > 0 {
		firstInput := unifiedNode.Inputs[0]
		if firstInput.Reference != nil {
			selector, err := c.ConvertVariableReference(firstInput.Reference)
			if err == nil && len(selector) > 0 {
				// Set directly to the Data field as classifier nodes do not use Config wrapping
				node.Data.QueryVariableSelector = selector

				// Also update variable references in the instruction field
				if node.Data.Instruction != "" {
					node.Data.Instruction = c.convertInstructionVariableReferences(node.Data.Instruction, selector)
				}
			}
		}
	}
	return nil
}

// updateIterationVariableSelectors updates variable selectors for iteration nodes
func (c *VariableSelectorConverter) updateIterationVariableSelectors(node *DifyNode, unifiedNode models.Node) error {
	if len(unifiedNode.Inputs) > 0 {
		firstInput := unifiedNode.Inputs[0]
		if firstInput.Reference != nil {
			selector, err := c.ConvertVariableReference(firstInput.Reference)
			if err == nil && len(selector) > 0 {
				node.Data.Config["iterator_selector"] = selector
			}
		}
	}
	return nil
}

// convertInstructionVariableReferences converts variable references in instructions
// Converts {{Query}} to {{#nodeId.variableName#}} format
func (c *VariableSelectorConverter) convertInstructionVariableReferences(instruction string, selector []string) string {
	if len(selector) < 2 {
		return instruction
	}

	// Extract node ID and variable name
	nodeID := selector[0]
	variableName := selector[1]

	// Use generic variable reference replacement logic
	// Find and replace variable references in format {{#nodeId.variableName#}} with mapped node IDs
	updatedInstruction := instruction

	// Simple replacement: find reference patterns containing variable name
	if strings.Contains(instruction, fmt.Sprintf(".%s#}}", variableName)) {
		// Find patterns like {{#anyNodeId.variableName#}}
		parts := strings.Split(instruction, "{{#")
		if len(parts) > 1 {
			for i := 1; i < len(parts); i++ {
				if strings.Contains(parts[i], fmt.Sprintf(".%s#}}", variableName)) {
					// Found a matching variable reference, replace the entire reference
					endIndex := strings.Index(parts[i], "#}}")
					if endIndex != -1 {
						parts[i] = fmt.Sprintf("%s.%s#}}", nodeID, variableName) + parts[i][endIndex+3:]
					}
				}
			}
			updatedInstruction = strings.Join(parts, "{{#")
		}
	}

	return updatedInstruction
}
