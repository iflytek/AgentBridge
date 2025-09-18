package generator

import (
	"ai-agents-transformer/internal/models"
	"crypto/rand"
	"fmt"
)

// BaseNodeGenerator provides base node generation functionality for iFlytek SparkAgent
type BaseNodeGenerator struct {
	nodeType models.NodeType
}

func NewBaseNodeGenerator(nodeType models.NodeType) *BaseNodeGenerator {
	return &BaseNodeGenerator{
		nodeType: nodeType,
	}
}

// GetSupportedType returns the supported node type
func (g *BaseNodeGenerator) GetSupportedType() models.NodeType {
	return g.nodeType
}

// ValidateNode performs basic node validation
func (g *BaseNodeGenerator) ValidateNode(node models.Node) error {
	if node.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	if node.Type == "" {
		return fmt.Errorf("node type cannot be empty")
	}

	if node.Title == "" {
		return fmt.Errorf("node title cannot be empty")
	}

	return nil
}

// generateBasicNodeInfo generates basic node information
func (g *BaseNodeGenerator) generateBasicNodeInfo(node models.Node) IFlytekNode {
	iflytekNode := IFlytekNode{
		ID:       g.generateIFlytekNodeID(node.Type),
		Dragging: false,
		Selected: false,
		Width:    node.Size.Width,
		Height:   node.Size.Height,
		Position: IFlytekPosition{
			X: node.Position.X,
			Y: node.Position.Y,
		},
		PositionAbsolute: IFlytekPosition{
			X: node.Position.X,
			Y: node.Position.Y,
		},
		Type: g.convertNodeType(node.Type),
		Data: IFlytekNodeData{
			AllowInputReference:  true,
			AllowOutputReference: true,
			Label:                node.Title,
			Status:               "",
			NodeMeta: IFlytekNodeMeta{
				AliasName: g.getNodeTypeDisplayName(node.Type),
				NodeType:  g.getNodeCategory(node.Type),
			},
			Inputs:      []IFlytekInput{},
			Outputs:     []IFlytekOutput{},
			References:  []IFlytekReference{},
			NodeParam:   make(map[string]interface{}),
			Description: node.Description,
			Updatable:   false,
		},
	}

	// restore platform-specific configuration
	if node.PlatformConfig.IFlytek != nil {
		config := node.PlatformConfig.IFlytek
		if allowInputRef, ok := config["allowInputReference"].(bool); ok {
			iflytekNode.Data.AllowInputReference = allowInputRef
		}
		if allowOutputRef, ok := config["allowOutputReference"].(bool); ok {
			iflytekNode.Data.AllowOutputReference = allowOutputRef
		}
		if status, ok := config["status"].(string); ok {
			iflytekNode.Data.Status = status
		}
		if parentID, ok := config["parentId"].(string); ok && parentID != "" {
			iflytekNode.ParentID = &parentID
			iflytekNode.Data.ParentID = &parentID
		}
	}

	return iflytekNode
}

// determineLabelByID determines display label based on iFlytek node ID and title mapping
// Priority order: title mapping if exists, then common prefix type mapping, then default node type
func (g *BaseNodeGenerator) determineLabelByID(nodeID string, nodeTitleMapping map[string]string) string {
	if nodeID == "" {
		return ""
	}

	// priority 1: title mapping
	if label := g.getLabelFromTitleMapping(nodeID, nodeTitleMapping); label != "" {
		return label
	}

	// priority 2: common prefix mapping
	if label := g.getLabelFromNodeType(nodeID); label != "" {
		return label
	}

	// priority 3: default
	return "节点"
}

// getLabelFromTitleMapping gets label from title mapping
func (g *BaseNodeGenerator) getLabelFromTitleMapping(nodeID string, nodeTitleMapping map[string]string) string {
	if nodeTitleMapping == nil {
		return ""
	}

	if title, ok := nodeTitleMapping[nodeID]; ok && title != "" {
		return title
	}

	return ""
}

// getLabelFromNodeType gets label based on node type prefix
func (g *BaseNodeGenerator) getLabelFromNodeType(nodeID string) string {
	prefixMappings := g.getNodeTypePrefixMappings()

	for prefix, label := range prefixMappings {
		if g.hasPrefix(nodeID, prefix) {
			return label
		}
	}

	return ""
}

// getNodeTypePrefixMappings returns node type prefix to label mappings
func (g *BaseNodeGenerator) getNodeTypePrefixMappings() map[string]string {
	return map[string]string{
		"node-start::":      "开始",
		"node-end::":        "结束",
		"spark-llm::":       "大模型",
		"ifly-code::":       "代码",
		"if-else::":         "分支器",
		"decision-making::": "决策",
		"iteration::":       "迭代",
	}
}

// hasPrefix checks if nodeID has the specified prefix
func (g *BaseNodeGenerator) hasPrefix(nodeID, prefix string) bool {
	return len(nodeID) >= len(prefix) && nodeID[:len(prefix)] == prefix
}

// generateOutputs generates node outputs
func (g *BaseNodeGenerator) generateOutputs(outputs []models.Output) []IFlytekOutput {
	iflytekOutputs := make([]IFlytekOutput, 0, len(outputs))

	for _, output := range outputs {
		iflytekOutput := IFlytekOutput{
			ID:         g.generateOutputID(),
			Name:       output.Name,
			NameErrMsg: "",
			Schema: IFlytekSchema{
				Type:    g.convertDataType(output.Type),
				Default: output.Description,
			},
		}

		iflytekOutputs = append(iflytekOutputs, iflytekOutput)
	}

	return iflytekOutputs
}

// convertNodeType converts node type
func (g *BaseNodeGenerator) convertNodeType(nodeType models.NodeType) string {
	switch nodeType {
	case models.NodeTypeStart:
		return "开始节点"
	case models.NodeTypeEnd:
		return "结束节点"
	case models.NodeTypeLLM:
		return "大模型"
	case models.NodeTypeCode:
		return "代码"
	case models.NodeTypeCondition:
		return "分支器"
	case models.NodeTypeClassifier:
		return "决策"
	case models.NodeTypeIteration:
		return "迭代"
	default:
		return string(nodeType)
	}
}

// getNodeTypeDisplayName returns node type display name
func (g *BaseNodeGenerator) getNodeTypeDisplayName(nodeType models.NodeType) string {
	switch nodeType {
	case models.NodeTypeStart:
		return "开始节点"
	case models.NodeTypeEnd:
		return "结束节点"
	case models.NodeTypeLLM:
		return "大模型"
	case models.NodeTypeCode:
		return "代码"
	case models.NodeTypeCondition:
		return "分支器"
	case models.NodeTypeClassifier:
		return "决策"
	case models.NodeTypeIteration:
		return "迭代"
	default:
		return string(nodeType)
	}
}

// getNodeCategory returns node category
func (g *BaseNodeGenerator) getNodeCategory(nodeType models.NodeType) string {
	switch nodeType {
	case models.NodeTypeStart, models.NodeTypeEnd, models.NodeTypeLLM, models.NodeTypeIteration:
		return "基础节点"
	case models.NodeTypeCondition, models.NodeTypeClassifier:
		return "分支器"
	case models.NodeTypeCode:
		return "工具"
	default:
		return "基础节点"
	}
}

// convertDataType converts data type
func (g *BaseNodeGenerator) convertDataType(dataType models.UnifiedDataType) string {
	switch dataType {
	case models.DataTypeString:
		return "string"
	case models.DataTypeInteger:
		return "integer"
	case models.DataTypeFloat:
		return "number"
	case models.DataTypeNumber:
		return "number"
	case models.DataTypeBoolean:
		return "boolean"
	case models.DataTypeArrayString:
		return "array-string"
	case models.DataTypeArrayObject:
		return "array-object"
	case models.DataTypeObject:
		return "object"
	default:
		return "string"
	}
}

// generateInputID generates input ID
func (g *BaseNodeGenerator) generateInputID() string {
	return generateUUID()
}

// generateOutputID generates output ID
func (g *BaseNodeGenerator) generateOutputID() string {
	return generateUUID()
}

// generateRefID generates reference ID
func (g *BaseNodeGenerator) generateRefID() string {
	return generateUUID()
}

// generateIFlytekNodeID generates iFlytek SparkAgent compliant node ID
func (g *BaseNodeGenerator) generateIFlytekNodeID(nodeType models.NodeType) string {
	uuid := generateRealUUID()

	switch nodeType {
	case models.NodeTypeStart:
		return "node-start::" + uuid
	case models.NodeTypeEnd:
		return "node-end::" + uuid
	case models.NodeTypeLLM:
		return "spark-llm::" + uuid
	case models.NodeTypeCode:
		return "ifly-code::" + uuid
	case models.NodeTypeCondition:
		return "if-else::" + uuid
	case models.NodeTypeClassifier:
		return "decision-making::" + uuid
	case models.NodeTypeIteration:
		return "iteration::" + uuid
	default:
		return "node-unknown::" + uuid
	}
}

// generateSpecialNodeID generates special node ID for iteration child nodes
func (g *BaseNodeGenerator) generateSpecialNodeID(nodePrefix string) string {
	uuid := generateRealUUID()
	return nodePrefix + "::" + uuid
}

// generateRealUUID generates cryptographically secure UUID
func generateRealUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// fallback to deterministic generation if crypto/rand fails
		return fmt.Sprintf("%x-%x-%x-%x-%x",
			b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	}

	// set UUID version 4 and variant bits
	b[6] = (b[6] & 0x0f) | 0x40 // UUID version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// generateUUID generates UUID for backward compatibility
func generateUUID() string {
	return generateRealUUID()
}

// getNodeIcon returns the standard icon URL for each node type
func (g *BaseNodeGenerator) getNodeIcon(nodeType models.NodeType) string {
	switch nodeType {
	case models.NodeTypeStart:
		return "https://oss-beijing-m8.openstorage.cn/pro-bucket/sparkBot/common/workflow/icon/start-node-icon.png"
	case models.NodeTypeEnd:
		return "https://oss-beijing-m8.openstorage.cn/pro-bucket/sparkBot/common/workflow/icon/end-node-icon.png"
	case models.NodeTypeLLM:
		return "https://oss-beijing-m8.openstorage.cn/pro-bucket/sparkBot/common/workflow/icon/largeModelIcon.png"
	case models.NodeTypeCode:
		return "https://oss-beijing-m8.openstorage.cn/pro-bucket/sparkBot/common/workflow/icon/codeIcon.png"
	case models.NodeTypeCondition:
		return "https://oss-beijing-m8.openstorage.cn/pro-bucket/sparkBot/common/workflow/icon/if-else-node-icon.png"
	case models.NodeTypeClassifier:
		return "https://oss-beijing-m8.openstorage.cn/pro-bucket/sparkBot/common/workflow/icon/designMakeIcon.png"
	case models.NodeTypeIteration:
		return "https://oss-beijing-m8.openstorage.cn/pro-bucket/sparkBot/common/workflow/icon/iteration-icon.png"
	default:
		return "https://oss-beijing-m8.openstorage.cn/pro-bucket/sparkBot/common/workflow/icon/start-node-icon.png"
	}
}
