package parser

import (
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
)

// EndNodeParser parses Dify end nodes.
type EndNodeParser struct {
	*BaseNodeParser
	skippedNodeIDs map[string]bool // Track skipped node IDs
}

func NewEndNodeParser(vrs *models.VariableReferenceSystem) NodeParser {
	return &EndNodeParser{
		BaseNodeParser: NewBaseNodeParser("end", vrs),
		skippedNodeIDs: make(map[string]bool),
	}
}

// SetSkippedNodeIDs sets skipped node IDs.
func (p *EndNodeParser) SetSkippedNodeIDs(skippedNodeIDs map[string]bool) {
	p.skippedNodeIDs = skippedNodeIDs
}

// GetSupportedType returns supported node type.
func (p *EndNodeParser) GetSupportedType() string {
	return "end"
}

// ParseNode parses Dify end node.
func (p *EndNodeParser) ParseNode(difyNode DifyNode) (*models.Node, error) {
	// Validate node
	if err := p.ValidateNode(difyNode); err != nil {
		return nil, err
	}

	// Create basic node information
	node := p.parseBasicNodeInfo(difyNode)
	node.Type = models.NodeTypeEnd

	// Parse end node specific configuration
	config := models.EndConfig{
		OutputMode:   "variables", // Dify defaults to variables mode
		StreamOutput: false,       // Default to non-streaming output
	}

	// Parse inputs (end node outputs are actually inputs)
	if difyNode.Data.Outputs != nil {
		outputs := p.parseOutputs(difyNode.Data.Outputs)
		for _, output := range outputs {
			input := models.Input{
				Name:        output.Variable,
				Type:        p.convertDataType(output.ValueType),
				Description: "", // Dify end nodes have no description
			}

			// Parse variable selector as variable reference
			if len(output.ValueSelector) >= 2 {
				sourceNodeID := output.ValueSelector[0]

				// Skip references from filtered nodes (such as "other classification" nodes)
				if p.skippedNodeIDs != nil && p.skippedNodeIDs[sourceNodeID] {
					continue // Skip this input, don't add to end node
				}

				input.Reference = &models.VariableReference{
					Type:       models.ReferenceTypeNodeOutput,
					NodeID:     sourceNodeID,
					OutputName: p.mapOutputName(output.ValueSelector[1]),
					DataType:   p.convertDataType(output.ValueType),
				}
			}

			node.Inputs = append(node.Inputs, input)
		}
	}

	// Check if there is template output mode configuration
	// Dify end nodes may configure templates through other means, use default values for now

	node.Config = config

	return node, nil
}

// ValidateNode validates Dify end node.
func (p *EndNodeParser) ValidateNode(difyNode DifyNode) error {
	// Use base validation
	if err := p.BaseNodeParser.ValidateNode(difyNode); err != nil {
		return err
	}

	// End node specific validation
	if difyNode.Data.Type != "end" {
		return fmt.Errorf("node type must be 'end', got '%s'", difyNode.Data.Type)
	}

	return nil
}

// parseOutputs parses outputs field, supports both array and object formats.
func (p *EndNodeParser) parseOutputs(outputs interface{}) []DifyOutput {
	outputArray := p.convertToArray(outputs)
	if outputArray == nil {
		return []DifyOutput{}
	}

	return p.processOutputArray(outputArray)
}

// convertToArray converts outputs to array format if possible
func (p *EndNodeParser) convertToArray(outputs interface{}) []interface{} {
	outputArray, ok := outputs.([]interface{})
	if !ok {
		return nil
	}
	return outputArray
}

// processOutputArray processes array of output items
func (p *EndNodeParser) processOutputArray(outputArray []interface{}) []DifyOutput {
	var result []DifyOutput

	for _, item := range outputArray {
		if output := p.processOutputItem(item); output != nil {
			result = append(result, *output)
		}
	}

	return result
}

// processOutputItem processes a single output item
func (p *EndNodeParser) processOutputItem(item interface{}) *DifyOutput {
	outputMap := p.convertItemToMap(item)
	if outputMap == nil {
		return nil
	}

	return p.createDifyOutput(outputMap)
}

// convertItemToMap converts item to string map
func (p *EndNodeParser) convertItemToMap(item interface{}) map[string]interface{} {
	// Try direct string map
	if m1, ok := item.(map[string]interface{}); ok {
		return m1
	}

	// Try interface map and convert
	if m2, ok := item.(map[interface{}]interface{}); ok {
		return p.convertInterfaceMap(m2)
	}

	return nil
}

// convertInterfaceMap converts map[interface{}]interface{} to map[string]interface{}
func (p *EndNodeParser) convertInterfaceMap(m2 map[interface{}]interface{}) map[string]interface{} {
	outputMap := make(map[string]interface{})
	for k, v := range m2 {
		if key, ok := k.(string); ok {
			outputMap[key] = v
		}
	}
	return outputMap
}

// createDifyOutput creates DifyOutput from output map
func (p *EndNodeParser) createDifyOutput(outputMap map[string]interface{}) *DifyOutput {
	output := &DifyOutput{}

	p.setOutputVariable(output, outputMap)
	p.setOutputValueType(output, outputMap)
	p.setOutputValueSelector(output, outputMap)

	return output
}

// setOutputVariable sets variable field
func (p *EndNodeParser) setOutputVariable(output *DifyOutput, outputMap map[string]interface{}) {
	if variable, ok := outputMap["variable"].(string); ok {
		output.Variable = variable
	}
}

// setOutputValueType sets value type field
func (p *EndNodeParser) setOutputValueType(output *DifyOutput, outputMap map[string]interface{}) {
	if valueType, ok := outputMap["value_type"].(string); ok {
		output.ValueType = valueType
	}
}

// setOutputValueSelector sets value selector field
func (p *EndNodeParser) setOutputValueSelector(output *DifyOutput, outputMap map[string]interface{}) {
	valueSelector, ok := outputMap["value_selector"].([]interface{})
	if !ok {
		return
	}

	output.ValueSelector = make([]string, len(valueSelector))
	for i, v := range valueSelector {
		if str, ok := v.(string); ok {
			output.ValueSelector[i] = str
		}
	}
}

// mapOutputName maps output names, handling platform differences.
func (p *EndNodeParser) mapOutputName(outputName string) string {
	// Dify -> Unified DSL output name mapping
	switch outputName {
	case "text":
		// Dify LLM node output is usually called "text", but in iFlytek SparkAgent it's called "output"
		return "output"
	default:
		return outputName
	}
}
