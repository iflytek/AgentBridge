package generator

import (
	"fmt"
	"github.com/iflytek/agentbridge/internal/models"
)

// StartNodeGenerator generates Coze start nodes
type StartNodeGenerator struct {
	idGenerator *CozeIDGenerator
}

// NewStartNodeGenerator creates a start node generator
func NewStartNodeGenerator() *StartNodeGenerator {
	return &StartNodeGenerator{
		idGenerator: nil, // Set by the main generator
	}
}

// SetIDGenerator sets the shared ID generator
func (g *StartNodeGenerator) SetIDGenerator(idGenerator *CozeIDGenerator) {
	g.idGenerator = idGenerator
}

// GenerateNode generates a Coze workflow start node
func (g *StartNodeGenerator) GenerateNode(unifiedNode *models.Node) (*CozeNode, error) {
	if err := g.ValidateNode(unifiedNode); err != nil {
		return nil, err
	}

	cozeNodeID := g.idGenerator.MapToCozeNodeID(unifiedNode.ID)

	// Generate outputs from unified node
	outputs := g.generateOutputs(unifiedNode)

	return &CozeNode{
		ID:   cozeNodeID,
		Type: "1", // Coze start node type
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
			Outputs: outputs,
			Inputs:  nil,
			Size:    nil,
		},
		Blocks:  []interface{}{},
		Edges:   []interface{}{},
		Version: "",
		Size:    nil,
	}, nil
}

// GenerateSchemaNode generates a Coze schema start node
func (g *StartNodeGenerator) GenerateSchemaNode(unifiedNode *models.Node) (*CozeSchemaNode, error) {
	if err := g.ValidateNode(unifiedNode); err != nil {
		return nil, err
	}

	cozeNodeID := g.idGenerator.MapToCozeNodeID(unifiedNode.ID)

	// Generate outputs and trigger parameters
	outputs := g.generateOutputs(unifiedNode)

	return &CozeSchemaNode{
		Data: &CozeSchemaNodeData{
			NodeMeta: &CozeNodeMetaInfo{
				Description: g.getNodeDescription(unifiedNode), // Get description dynamically
				Icon:        g.getNodeIcon(unifiedNode),        // Get icon dynamically
				SubTitle:    "",
				Title:       g.getNodeTitle(unifiedNode), // Get title dynamically
			},
			Outputs:           outputs,
			TriggerParameters: outputs, // Same as outputs for start node
		},
		ID: cozeNodeID,
		Meta: &CozeNodeMeta{
			Position: &CozePosition{
				X: unifiedNode.Position.X, // Read position dynamically
				Y: unifiedNode.Position.Y, // Read position dynamically
			},
		},
		Type: "1",
	}, nil
}

// generateOutputs generates outputs from unified node
func (g *StartNodeGenerator) generateOutputs(unifiedNode *models.Node) []CozeNodeOutput {
	outputs := make([]CozeNodeOutput, 0)

	// Process outputs from unified node
	for _, output := range unifiedNode.Outputs {
		cozeOutput := CozeNodeOutput{
			Name:     output.Name,
			Required: output.Required, // Read required field dynamically
			Type:     g.mapUnifiedTypeToCoze(output.Type),
		}
		outputs = append(outputs, cozeOutput)
	}

	// If no outputs defined, add default ones
	if len(outputs) == 0 {
		// Add default AGENT_USER_INPUT for compatibility
		outputs = append(outputs, CozeNodeOutput{
			Name:     "AGENT_USER_INPUT",
			Required: true,
			Type:     "string",
		})
	}

	return outputs
}

// mapUnifiedTypeToCoze maps unified data types to Coze types
func (g *StartNodeGenerator) mapUnifiedTypeToCoze(unifiedType models.UnifiedDataType) string {
	switch unifiedType {
	case models.DataTypeString:
		return "string"
	case models.DataTypeInteger: // Integer preserves precise type
		return "integer"
	case models.DataTypeFloat: // Float preserves precise type
		return "float"
	case models.DataTypeNumber: // Maintain backward compatibility
		return "float" // Coze primarily uses float for numeric types
	case models.DataTypeBoolean:
		return "boolean"
	default:
		return "string"
	}
}

// GetNodeType returns the node type this generator handles
func (g *StartNodeGenerator) GetNodeType() models.NodeType {
	return models.NodeTypeStart
}

// ValidateNode validates the unified node before generation
func (g *StartNodeGenerator) ValidateNode(unifiedNode *models.Node) error {
	if unifiedNode == nil {
		return fmt.Errorf("unified node is nil")
	}

	if unifiedNode.Type != models.NodeTypeStart {
		return fmt.Errorf("expected start node, got %s", unifiedNode.Type)
	}

	// Validate node ID format
	if unifiedNode.ID == "" {
		return fmt.Errorf("start node must have a valid ID")
	}

	// Validate output definitions (Entry nodes should have output definitions)
	if len(unifiedNode.Outputs) == 0 {
		// Allow empty output definitions, will use default AGENT_USER_INPUT
	}

	return nil
}

// getNodeTitle retrieves node title dynamically
func (g *StartNodeGenerator) getNodeTitle(unifiedNode *models.Node) string {
	if unifiedNode.Title != "" {
		return unifiedNode.Title
	}
	return "开始" // Default value
}

// getNodeDescription retrieves node description dynamically
func (g *StartNodeGenerator) getNodeDescription(unifiedNode *models.Node) string {
	if unifiedNode.Description != "" {
		return unifiedNode.Description
	}
	return "工作流的开启节点，用于定义流程调用所需的业务变量信息。" // Consistent with Coze official description
}

// getNodeIcon retrieves node icon dynamically
func (g *StartNodeGenerator) getNodeIcon(unifiedNode *models.Node) string {
	// Check platform-specific configuration for icon information
	if unifiedNode.PlatformConfig.Coze != nil {
		if icon, exists := unifiedNode.PlatformConfig.Coze["icon"]; exists {
			if iconStr, ok := icon.(string); ok && iconStr != "" {
				return iconStr
			}
		}
	}

	// Check iFlytek platform configuration for icon (may need conversion)
	if unifiedNode.PlatformConfig.IFlytek != nil {
		if icon, exists := unifiedNode.PlatformConfig.IFlytek["icon"]; exists {
			if iconStr, ok := icon.(string); ok && iconStr != "" {
				return iconStr
			}
		}
	}

	// Official Coze Entry node icon URL
	return "https://oss-beijing-m8.openstorage.cn/pro-bucket/sparkBot/common/workflow/icon/start-node-icon.png"
}
