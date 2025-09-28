package generator

import (
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
	"strings"
)

// EndNodeGenerator generates Coze end nodes
type EndNodeGenerator struct {
	idGenerator *CozeIDGenerator
}

// NewEndNodeGenerator creates an end node generator
func NewEndNodeGenerator() *EndNodeGenerator {
	return &EndNodeGenerator{
		idGenerator: nil, // Set by the main generator
	}
}

// SetIDGenerator sets the shared ID generator
func (g *EndNodeGenerator) SetIDGenerator(idGenerator *CozeIDGenerator) {
	g.idGenerator = idGenerator
}

// GenerateNode generates a Coze workflow end node
func (g *EndNodeGenerator) GenerateNode(unifiedNode *models.Node) (*CozeNode, error) {
	if err := g.ValidateNode(unifiedNode); err != nil {
		return nil, err
	}

	cozeNodeID := g.idGenerator.MapToCozeNodeID(unifiedNode.ID)

	// Generate inputs from unified node
	inputs := g.generateInputs(unifiedNode)

	return &CozeNode{
		ID:   cozeNodeID,
		Type: "2", // Coze end node type
		Meta: &CozeNodeMeta{
			Position: &CozePosition{
				X: unifiedNode.Position.X, // Read position dynamically
				Y: unifiedNode.Position.Y, // Read position dynamically
			},
		},
		Data: &CozeNodeData{
			Meta: &CozeNodeMetaInfo{
				Title:       g.getNodeTitle(unifiedNode),       // Get title dynamically
				Description: g.getNodeDescription(unifiedNode), // Get description dynamically
				Icon:        g.getNodeIcon(unifiedNode),        // Get icon dynamically
				Subtitle:    "",
				MainColor:   "",
			},
			Outputs: []CozeNodeOutput{}, // End node has no outputs
			Inputs:  inputs,
			Size:    nil,
		},
		Blocks:  []interface{}{},
		Edges:   []interface{}{},
		Version: "",
		Size:    nil,
	}, nil
}

// GenerateSchemaNode generates a Coze schema end node
func (g *EndNodeGenerator) GenerateSchemaNode(unifiedNode *models.Node) (*CozeSchemaNode, error) {
	if err := g.ValidateNode(unifiedNode); err != nil {
		return nil, err
	}

	cozeNodeID := g.idGenerator.MapToCozeNodeID(unifiedNode.ID)

	// Generate inputs for schema node
	inputs := g.generateSchemaInputs(unifiedNode)

	return &CozeSchemaNode{
		Data: &CozeSchemaNodeData{
			NodeMeta: &CozeNodeMetaInfo{
				Description: g.getNodeDescription(unifiedNode), // Use dynamic retrieval
				Icon:        g.getNodeIcon(unifiedNode),        // Use dynamic retrieval
				SubTitle:    "",
				Title:       g.getNodeTitle(unifiedNode), // Use dynamic retrieval
			},
			Inputs:        inputs,
			TerminatePlan: g.selectTerminatePlan(unifiedNode),
		},
		ID: cozeNodeID,
		Meta: &CozeNodeMeta{
			Position: &CozePosition{
				X: unifiedNode.Position.X, // Read position dynamically
				Y: unifiedNode.Position.Y, // Read position dynamically
			},
		},
		Type: "2",
	}, nil
}

// generateInputs generates inputs for the workflow node format
func (g *EndNodeGenerator) generateInputs(unifiedNode *models.Node) interface{} {
	inputParameters := make([]interface{}, 0)

	// Process inputs from unified node
	for _, input := range unifiedNode.Inputs {
		if input.Reference != nil && input.Reference.Type == models.ReferenceTypeNodeOutput {
			// Create individual input parameter
			inputParam := map[string]interface{}{
				"name": input.Name,
				"input": map[string]interface{}{
					"type": g.mapUnifiedTypeToCoze(input.Type),
					"value": map[string]interface{}{
						"type": "ref",
						"content": map[string]interface{}{
							"blockID": g.idGenerator.MapToCozeNodeID(input.Reference.NodeID),
							"name":    g.mapOutputFieldNameForCoze(input.Reference.NodeID, input.Reference.OutputName),
							"source":  "block-output",
						},
						"rawMeta": map[string]interface{}{ // Uses camelCase rawMeta for nodes section
							"type": GetCozeRawMetaType(g.mapUnifiedTypeToCoze(input.Type)),
						},
					},
				},
				"left":      nil,
				"right":     nil,
				"variables": []interface{}{},
			}
			inputParameters = append(inputParameters, inputParam)
		}
	}

	// FIXED: Generate output template from inputs
	outputTemplate := ""
	if len(inputParameters) > 0 {
		// Use the first input parameter name in template
		if param, ok := inputParameters[0].(map[string]interface{}); ok {
			if name, ok := param["name"].(string); ok {
				outputTemplate = "{{" + name + "}}"
			}
		}
	}

	// Return the complete inputs structure
	return map[string]interface{}{
		"inputparameters": inputParameters, // Maintains lowercase format for nodes section
		"settingonerror":  nil,
		"nodebatchinfo":   nil,
		"llmparam":        nil,
		"outputemitter": map[string]interface{}{
			"content": map[string]interface{}{
				"type": "string",
				"value": map[string]interface{}{
					"type":    "literal",
					"content": outputTemplate, // FIXED: Add output template
				},
			},
			"streamingoutput": false,
		},
		"exit": map[string]interface{}{
			"terminateplan": g.selectTerminatePlan(unifiedNode),
		},
		"llm":                nil,
		"loop":               nil,
		"selector":           nil,
		"textprocessor":      nil,
		"subworkflow":        nil,
		"intentdetector":     nil,
		"databasenode":       nil,
		"httprequestnode":    nil,
		"knowledge":          nil,
		"coderunner":         nil,
		"pluginapiparam":     nil,
		"variableaggregator": nil,
		"variableassigner":   nil,
		"qa":                 nil,
		"batch":              nil,
		"comment":            nil,
		"inputreceiver":      nil,
	}
}

// selectTerminatePlan selects end node terminate plan by output mode
func (g *EndNodeGenerator) selectTerminatePlan(unifiedNode *models.Node) string {
	if unifiedNode == nil {
		return "useAnswerContent"
	}
	if cfg, ok := unifiedNode.Config.(models.EndConfig); ok {
		if strings.EqualFold(cfg.OutputMode, "variables") {
			return "returnVariables"
		}
	} else if cfgPtr, ok := unifiedNode.Config.(*models.EndConfig); ok && cfgPtr != nil {
		if strings.EqualFold(cfgPtr.OutputMode, "variables") {
			return "returnVariables"
		}
	}
	return "useAnswerContent"
}

// generateSchemaInputs generates inputs for the schema node format
func (g *EndNodeGenerator) generateSchemaInputs(unifiedNode *models.Node) *CozeNodeInputs {
	inputParameters := make([]CozeInputParameter, 0)

	// Process inputs from unified node
	for _, input := range unifiedNode.Inputs {
		if input.Reference != nil && input.Reference.Type == models.ReferenceTypeNodeOutput {
			inputParam := CozeInputParameter{
				Name: input.Name,
				Input: &CozeInputValue{
					Type: g.mapUnifiedTypeToCoze(input.Type),
					Value: &CozeInputRef{
						Content: &CozeRefContent{
							BlockID: g.idGenerator.MapToCozeNodeID(input.Reference.NodeID),
							Name:    input.Reference.OutputName,
							Source:  "block-output",
						},
						RawMeta: &CozeRawMeta{
							Type: GetCozeRawMetaType(g.mapUnifiedTypeToCoze(input.Type)),
						},
						Type: "ref",
					},
				},
			}
			inputParameters = append(inputParameters, inputParam)
		}
	}

	return &CozeNodeInputs{
		InputParameters: inputParameters,
	}
}

// mapUnifiedTypeToCoze maps unified data types to Coze types
func (g *EndNodeGenerator) mapUnifiedTypeToCoze(unifiedType models.UnifiedDataType) string {
	switch unifiedType {
	case models.DataTypeString:
		return "string"
	case models.DataTypeInteger:
		return "integer"
	case models.DataTypeFloat:
		return "float"
	case models.DataTypeNumber:
		return "float" // Coze primarily uses float for numeric types
	case models.DataTypeBoolean:
		return "boolean"
	case models.DataTypeArrayString, models.DataTypeArrayInteger, models.DataTypeArrayFloat,
		models.DataTypeArrayNumber, models.DataTypeArrayBoolean, models.DataTypeArrayObject:
		return "list" // Coze uses "list" for all array types
	case models.DataTypeObject:
		return "object"
	default:
		return "string"
	}
}

// GetNodeType returns the node type this generator handles
func (g *EndNodeGenerator) GetNodeType() models.NodeType {
	return models.NodeTypeEnd
}

// ValidateNode validates the unified node before generation
func (g *EndNodeGenerator) ValidateNode(unifiedNode *models.Node) error {
	if unifiedNode == nil {
		return fmt.Errorf("unified node is nil")
	}

	if unifiedNode.Type != models.NodeTypeEnd {
		return fmt.Errorf("expected end node, got %s", unifiedNode.Type)
	}

	// Validate node ID format
	if unifiedNode.ID == "" {
		return fmt.Errorf("end node must have a valid ID")
	}

	return nil
}

// getNodeTitle retrieves node title dynamically
func (g *EndNodeGenerator) getNodeTitle(unifiedNode *models.Node) string {
	if unifiedNode.Title != "" {
		return unifiedNode.Title
	}
	return "结束" // Default value
}

// getNodeDescription retrieves node description dynamically
func (g *EndNodeGenerator) getNodeDescription(unifiedNode *models.Node) string {
	if unifiedNode.Description != "" {
		return unifiedNode.Description
	}
	return "工作流的结束节点，用于输出工作流运行后的最终结果。" // Consistent with Coze official description
}

// getNodeIcon retrieves node icon dynamically
func (g *EndNodeGenerator) getNodeIcon(unifiedNode *models.Node) string {
	// Check platform-specific configuration for icon information
	if unifiedNode.PlatformConfig.Coze != nil {
		if icon, exists := unifiedNode.PlatformConfig.Coze["icon"]; exists {
			if iconStr, ok := icon.(string); ok && iconStr != "" {
				return iconStr
			}
		}
	}

	// Check iFlytek platform configuration for icon
	if unifiedNode.PlatformConfig.IFlytek != nil {
		if icon, exists := unifiedNode.PlatformConfig.IFlytek["icon"]; exists {
			if iconStr, ok := icon.(string); ok && iconStr != "" {
				return iconStr
			}
		}
	}

	// Official Coze Exit node icon URL
	return "https://oss-beijing-m8.openstorage.cn/pro-bucket/sparkBot/common/workflow/icon/end-node-icon.png"
}

// mapOutputFieldNameForCoze maps output field names from unified DSL to Coze platform format
func (g *EndNodeGenerator) mapOutputFieldNameForCoze(nodeID, outputName string) string {
	// Map classifier output names from iFlytek format to Coze format
	if outputName == "class_name" {
		// iFlytek classifier outputs "class_name", but Coze uses "classificationId"
		return "classificationId"
	}

	// Default: return original name if no mapping needed
	return outputName
}
