package parser

import (
	"github.com/iflytek/agentbridge/core/interfaces"
	"github.com/iflytek/agentbridge/internal/models"
	"github.com/iflytek/agentbridge/platforms/common"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Compile-time interface check
var _ interfaces.DSLParser = (*DifyParser)(nil)

// DifyParser parses Dify DSL to unified format.
type DifyParser struct {
	*common.BaseParser
	factory           *ParserFactory
	variableRefSystem *models.VariableReferenceSystem
	skippedNodeIDs    map[string]bool // Track skipped node IDs
}

func NewDifyParser() *DifyParser {
	variableRefSystem := models.NewVariableReferenceSystem()

	return &DifyParser{
		BaseParser:        common.NewBaseParser(models.PlatformDify),
		factory:           NewParserFactory(),
		variableRefSystem: variableRefSystem,
	}
}

// Parse parses Dify DSL to unified format.
func (p *DifyParser) Parse(data []byte) (*models.UnifiedDSL, error) {
	// Validate input data
	if err := p.Validate(data); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Parse YAML
	var difyDSL DifyDSL
	if err := yaml.Unmarshal(data, &difyDSL); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Build unified DSL
	unifiedDSL := &models.UnifiedDSL{
		Version: "1.0",
		Metadata: models.Metadata{
			Name:        difyDSL.App.Name,
			Description: difyDSL.App.Description,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		PlatformMetadata: models.PlatformMetadata{
			Dify: &models.DifyMetadata{
				Icon:                difyDSL.App.Icon,
				IconBackground:      difyDSL.App.IconBackground,
				Mode:                difyDSL.App.Mode,
				UseIconAsAnswerIcon: difyDSL.App.UseIconAsAnswerIcon,
				Kind:                difyDSL.Kind,
				AppVersion:          difyDSL.Version,
			},
		},
		Workflow: models.Workflow{
			Nodes:     []models.Node{},
			Edges:     []models.Edge{},
			Variables: []models.Variable{},
		},
	}

	// Parse UI configuration
	if err := p.parseUIConfig(&difyDSL, unifiedDSL); err != nil {
		return nil, fmt.Errorf("failed to parse UI config: %w", err)
	}

	// Parse nodes
	if err := p.parseNodes(difyDSL.Workflow.Graph.Nodes, unifiedDSL); err != nil {
		return nil, fmt.Errorf("failed to parse nodes: %w", err)
	}

	// Parse connection relationships
	if err := p.parseEdges(difyDSL.Workflow.Graph.Edges, unifiedDSL); err != nil {
		return nil, fmt.Errorf("failed to parse edges: %w", err)
	}

	// Process iteration child node relationships
	if err := p.processIterationRelationships(unifiedDSL); err != nil {
		return nil, fmt.Errorf("failed to process iteration relationships: %w", err)
	}

	// Print conversion summary after parsing is complete
	p.printConversionSummary(unifiedDSL)

	return unifiedDSL, nil
}

// Validate validates Dify DSL format.
func (p *DifyParser) Validate(data []byte) error {
	var difyDSL DifyDSL
	if err := yaml.Unmarshal(data, &difyDSL); err != nil {
		return fmt.Errorf("invalid YAML format: %w", err)
	}

	// Validate required fields
	if difyDSL.App.Name == "" {
		return fmt.Errorf("app name is required")
	}

	if difyDSL.Workflow.Graph.Nodes == nil {
		return fmt.Errorf("workflow nodes are required")
	}

	return nil
}

// parseUIConfig parses UI configuration.
func (p *DifyParser) parseUIConfig(difyDSL *DifyDSL, unifiedDSL *models.UnifiedDSL) error {
	features := difyDSL.Workflow.Features

	uiConfig := &models.UIConfig{}

	// Parse opening statement
	if features.OpeningStatement != "" {
		uiConfig.OpeningStatement = features.OpeningStatement
	}

	// Parse suggested questions
	if len(features.SuggestedQuestions) > 0 {
		uiConfig.SuggestedQuestions = features.SuggestedQuestions
	}

	// Parse icons
	uiConfig.Icon = difyDSL.App.Icon
	uiConfig.IconBackground = difyDSL.App.IconBackground

	// Only set UIConfig when there is valid configuration
	if uiConfig.OpeningStatement != "" || len(uiConfig.SuggestedQuestions) > 0 ||
		uiConfig.Icon != "" || uiConfig.IconBackground != "" {
		unifiedDSL.Metadata.UIConfig = uiConfig
	}

	return nil
}

// parseNodes parses nodes.
func (p *DifyParser) parseNodes(difyNodes []DifyNode, unifiedDSL *models.UnifiedDSL) error {
	// Track skipped node IDs for edge filtering
	skippedNodeIDs := make(map[string]bool)
	p.skippedNodeIDs = skippedNodeIDs // Ensure the parser instance has access to skipped node IDs

	for _, difyNode := range difyNodes {
		// Skip nodes with title "other classification" as they are handled through default intent mechanism in iFlytek
		if difyNode.Data.Title == "其他分类" {
			skippedNodeIDs[difyNode.ID] = true
			continue
		}

		// Use fallback parsing to handle unsupported node types
		node, supported, err := p.factory.ParseNodeWithFallback(difyNode, p.variableRefSystem)
		if err != nil {
			return fmt.Errorf("failed to parse node %s: %w", difyNode.ID, err)
		}

		if !supported {
			// Convert unsupported nodes to code node placeholders
			fmt.Printf("⚠️  Converting unsupported node type '%s' (ID: %s) to code node placeholder\n",
				difyNode.Data.Type, difyNode.ID)

			node, err = p.convertUnsupportedNodeToCodeNode(difyNode)
			if err != nil {
				return fmt.Errorf("failed to convert unsupported node %s: %w", difyNode.ID, err)
			}
		}

		// If it's an end node, we need to pass skipped node IDs to handle references correctly
		if node.Type == models.NodeTypeEnd {
			// Re-parse with end node parser context
			endParser := NewEndNodeParser(p.variableRefSystem).(*EndNodeParser)
			endParser.SetSkippedNodeIDs(skippedNodeIDs)
			node, err = endParser.ParseNode(difyNode)
			if err != nil {
				return fmt.Errorf("failed to parse end node %s: %w", difyNode.ID, err)
			}
		}

		// Check if the node itself has iteration information and mark it
		p.markNodeIterationFromNodeData(node, difyNode.Data)
		p.markNodeIterationFromParentID(node, difyNode.ParentID)

		unifiedDSL.Workflow.Nodes = append(unifiedDSL.Workflow.Nodes, *node)
	}

	// Save skipped node IDs for edge parsing
	p.skippedNodeIDs = skippedNodeIDs

	return nil
}

// parseEdges parses connection relationships.
func (p *DifyParser) parseEdges(difyEdges []DifyEdge, unifiedDSL *models.UnifiedDSL) error {
	for _, difyEdge := range difyEdges {
		// Skip edges involving filtered nodes
		if p.skippedNodeIDs != nil {
			if p.skippedNodeIDs[difyEdge.Source] || p.skippedNodeIDs[difyEdge.Target] {
				continue
			}
		}

		// Skip edges involving non-existent nodes (filtered out as unsupported)
		if !p.nodeExistsInDSL(difyEdge.Source, unifiedDSL) || !p.nodeExistsInDSL(difyEdge.Target, unifiedDSL) {
			continue
		}

		edge := models.Edge{
			ID:           difyEdge.ID,
			Source:       difyEdge.Source,
			Target:       difyEdge.Target,
			SourceHandle: difyEdge.SourceHandle,
			TargetHandle: difyEdge.TargetHandle,
			Type:         p.convertEdgeType(difyEdge.Type),
		}

		// Parse platform-specific configuration
		if difyEdge.Data != nil {
			edge.PlatformConfig = models.PlatformConfig{
				Dify: map[string]interface{}{
					"zIndex":        difyEdge.ZIndex,
					"isInLoop":      difyEdge.Data.IsInLoop,
					"isInIteration": difyEdge.Data.IsInIteration,
					"iterationId":   difyEdge.Data.IterationID,
					"sourceType":    difyEdge.Data.SourceType,
					"targetType":    difyEdge.Data.TargetType,
				},
			}
		}

		unifiedDSL.Workflow.Edges = append(unifiedDSL.Workflow.Edges, edge)
	}

	return nil
}

// convertEdgeType converts connection type.
func (p *DifyParser) convertEdgeType(difyType string) models.EdgeType {
	switch difyType {
	case "custom":
		return models.EdgeTypeDefault
	default:
		return models.EdgeTypeDefault
	}
}

// processIterationRelationships processes iteration child node relationships.
func (p *DifyParser) processIterationRelationships(unifiedDSL *models.UnifiedDSL) error {
	// Create iteration node mapping
	iterationMap := make(map[string]bool)
	for _, node := range unifiedDSL.Workflow.Nodes {
		if node.Type == models.NodeTypeIteration {
			iterationMap[node.ID] = true
		}
	}

	// Process iteration information in edges, update node iteration relationships
	for _, edge := range unifiedDSL.Workflow.Edges {
		if edge.PlatformConfig.Dify != nil {
			if isInIteration, ok := edge.PlatformConfig.Dify["isInIteration"].(bool); ok && isInIteration {
				if iterationId, ok := edge.PlatformConfig.Dify["iterationId"].(string); ok && iterationId != "" {
					// Find source and target nodes, mark them as iteration child nodes
					p.markNodeAsIterationChild(unifiedDSL, edge.Source, iterationId)
					p.markNodeAsIterationChild(unifiedDSL, edge.Target, iterationId)
				}
			}
		}
	}

	return nil
}

// markNodeAsIterationChild marks a node as an iteration child node.
func (p *DifyParser) markNodeAsIterationChild(unifiedDSL *models.UnifiedDSL, nodeID, iterationID string) {
	for i := range unifiedDSL.Workflow.Nodes {
		if unifiedDSL.Workflow.Nodes[i].ID == nodeID {
			p.setNodeIterationConfig(&unifiedDSL.Workflow.Nodes[i], iterationID)
			break
		}
	}
}

// markNodeIterationFromNodeData marks a node as iteration child based on node data
func (p *DifyParser) markNodeIterationFromNodeData(node *models.Node, data DifyNodeData) {
	if data.IsInIteration && data.IterationID != "" {
		p.setNodeIterationConfig(node, data.IterationID)
	}
}

// markNodeIterationFromParentID marks a node as iteration child based on parent ID
func (p *DifyParser) markNodeIterationFromParentID(node *models.Node, parentID string) {
	if parentID != "" {
		p.setNodeIterationConfig(node, parentID)
	}
}

// setNodeIterationConfig sets iteration configuration for different node types
func (p *DifyParser) setNodeIterationConfig(node *models.Node, iterationID string) {
	switch node.Type {
	case models.NodeTypeStart:
		p.setStartNodeIteration(node, iterationID)
	case models.NodeTypeCode:
		p.setCodeNodeIteration(node, iterationID)
	case models.NodeTypeLLM:
		p.setLLMNodeIteration(node, iterationID)
	case models.NodeTypeCondition:
		p.setConditionNodeIteration(node, iterationID)
	case models.NodeTypeClassifier:
		p.setClassifierNodeIteration(node, iterationID)
	}
}

// setStartNodeIteration sets iteration config for start node
func (p *DifyParser) setStartNodeIteration(node *models.Node, iterationID string) {
	if startConfig, ok := node.Config.(models.StartConfig); ok {
		startConfig.IsInIteration = true
		startConfig.ParentID = iterationID
		node.Config = startConfig
	} else {
		node.Config = models.StartConfig{
			IsInIteration: true,
			ParentID:      iterationID,
		}
	}
}

// setCodeNodeIteration sets iteration config for code node
func (p *DifyParser) setCodeNodeIteration(node *models.Node, iterationID string) {
	if codeConfig, ok := node.Config.(models.CodeConfig); ok {
		codeConfig.IsInIteration = true
		codeConfig.IterationID = iterationID
		node.Config = codeConfig
	} else {
		node.Config = models.CodeConfig{
			IsInIteration: true,
			IterationID:   iterationID,
		}
	}
}

// setLLMNodeIteration sets iteration config for LLM node
func (p *DifyParser) setLLMNodeIteration(node *models.Node, iterationID string) {
	if llmConfig, ok := node.Config.(models.LLMConfig); ok {
		llmConfig.IsInIteration = true
		llmConfig.IterationID = iterationID
		node.Config = llmConfig
	} else {
		node.Config = models.LLMConfig{
			IsInIteration: true,
			IterationID:   iterationID,
		}
	}
}

// setConditionNodeIteration sets iteration config for condition node
func (p *DifyParser) setConditionNodeIteration(node *models.Node, iterationID string) {
	if conditionConfig, ok := node.Config.(models.ConditionConfig); ok {
		conditionConfig.IsInIteration = true
		conditionConfig.IterationID = iterationID
		node.Config = conditionConfig
	} else {
		node.Config = models.ConditionConfig{
			IsInIteration: true,
			IterationID:   iterationID,
		}
	}
}

// setClassifierNodeIteration sets iteration config for classifier node
func (p *DifyParser) setClassifierNodeIteration(node *models.Node, iterationID string) {
	if classifierConfig, ok := node.Config.(models.ClassifierConfig); ok {
		classifierConfig.IsInIteration = true
		classifierConfig.IterationID = iterationID
		node.Config = classifierConfig
	} else {
		node.Config = models.ClassifierConfig{
			IsInIteration: true,
			IterationID:   iterationID,
		}
	}
}

// convertUnsupportedNodeToCodeNode converts unsupported nodes to code node placeholders
func (p *DifyParser) convertUnsupportedNodeToCodeNode(difyNode DifyNode) (*models.Node, error) {
	// Get node title for type description
	nodeTitle := p.extractNodeTitle(difyNode)

	// Use code node parser to create code node
	codeParser := NewCodeNodeParser(p.variableRefSystem)

	// Create modified node with code node type
	modifiedNode := difyNode
	modifiedNode.Data.Type = "code"

	modifiedNode.Data.Title = fmt.Sprintf("暂不兼容的节点-%s（请根据需求手动实现）", nodeTitle)

	// Set default code configuration
	modifiedNode.Data.Code = fmt.Sprintf(`# 抱歉！当前兼容性工具不支持转换此类节点: %s

# 请根据业务需求手动补充实现逻辑
`, difyNode.Data.Type)
	modifiedNode.Data.CodeLanguage = "python3"

	// Create default output if none exist to maintain connections
	if modifiedNode.Data.Outputs == nil {
		modifiedNode.Data.Outputs = map[string]interface{}{
			"result": map[string]interface{}{
				"type":     "string",
				"children": nil,
			},
		}
	}

	// Parse using code node parser
	return codeParser.ParseNode(modifiedNode)
}

// extractNodeTitle extracts node title
func (p *DifyParser) extractNodeTitle(difyNode DifyNode) string {
	if difyNode.Data.Title != "" {
		return difyNode.Data.Title
	}
	return difyNode.Data.Type
}

// nodeExistsInDSL checks if a node exists in the unified DSL
func (p *DifyParser) nodeExistsInDSL(nodeID string, unifiedDSL *models.UnifiedDSL) bool {
	for _, node := range unifiedDSL.Workflow.Nodes {
		if node.ID == nodeID {
			return true
		}
	}
	return false
}

// printConversionSummary prints detailed conversion statistics
func (p *DifyParser) printConversionSummary(unifiedDSL *models.UnifiedDSL) {
	totalNodes := len(unifiedDSL.Workflow.Nodes)
	fmt.Printf("✅ Conversion Summary: All %d nodes processed successfully\n", totalNodes)

	// Count nodes converted to code placeholders by checking titles
	convertedCount := 0
	for _, node := range unifiedDSL.Workflow.Nodes {
		if strings.Contains(node.Title, "暂不兼容的节点-") {
			convertedCount++
		}
	}

	if convertedCount > 0 {
		fmt.Printf("ℹ️  %d unsupported nodes were converted to code node placeholders\n", convertedCount)
		fmt.Printf("ℹ️  Please manually adjust these placeholder nodes as needed\n")
	}
}
