package common

import (
	"fmt"
	"github.com/iflytek/agentbridge/internal/models"
)

// UnifiedDSLValidator validates unified DSL structures
type UnifiedDSLValidator struct{}

func NewUnifiedDSLValidator() *UnifiedDSLValidator {
	return &UnifiedDSLValidator{}
}

// ValidateWorkflow validates basic workflow structure
func (v *UnifiedDSLValidator) ValidateWorkflow(workflow *models.Workflow) error {
	if workflow == nil {
		return fmt.Errorf("workflow cannot be nil")
	}

	if len(workflow.Nodes) == 0 {
		return fmt.Errorf("workflow must contain at least one node")
	}

	// Validate required start and end nodes
	if err := v.validateRequiredNodes(workflow.Nodes); err != nil {
		return err
	}

	// Validate node reference relationships
	if err := v.validateNodeReferences(workflow); err != nil {
		return err
	}

	return nil
}

// ValidateMetadata validates metadata
func (v *UnifiedDSLValidator) ValidateMetadata(metadata *models.Metadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}

	if metadata.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	return nil
}

// ValidateNode validates a single node
func (v *UnifiedDSLValidator) ValidateNode(node *models.Node) error {
	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}

	if node.ID == "" {
		return fmt.Errorf("node ID is required")
	}

	if node.Type == "" {
		return fmt.Errorf("node type is required")
	}

	// Validate if node type is supported
	if !v.isSupportedNodeType(node.Type) {
		return fmt.Errorf("unsupported node type: %s", node.Type)
	}

	// Validate node configuration
	if err := v.validateNodeConfig(node); err != nil {
		return fmt.Errorf("invalid node config for %s: %w", node.ID, err)
	}

	return nil
}

// ValidateEdge validates edges
func (v *UnifiedDSLValidator) ValidateEdge(edge *models.Edge, nodes []models.Node) error {
	if edge == nil {
		return fmt.Errorf("edge cannot be nil")
	}

	if edge.Source == "" {
		return fmt.Errorf("edge source is required")
	}

	if edge.Target == "" {
		return fmt.Errorf("edge target is required")
	}

	// Validate source and target nodes exist
	if !v.nodeExists(edge.Source, nodes) {
		return fmt.Errorf("source node %s not found", edge.Source)
	}

	if !v.nodeExists(edge.Target, nodes) {
		return fmt.Errorf("target node %s not found", edge.Target)
	}

	return nil
}

// validateRequiredNodes validates required node types
func (v *UnifiedDSLValidator) validateRequiredNodes(nodes []models.Node) error {
	hasStart := false
	hasEnd := false

	for _, node := range nodes {
		switch node.Type {
		case models.NodeTypeStart:
			hasStart = true
		case models.NodeTypeEnd:
			hasEnd = true
		}
	}

	if !hasStart {
		return fmt.Errorf("workflow must contain at least one start node")
	}

	if !hasEnd {
		return fmt.Errorf("workflow must contain at least one end node")
	}

	return nil
}

// validateNodeReferences validates node reference relationships
func (v *UnifiedDSLValidator) validateNodeReferences(workflow *models.Workflow) error {
	nodeMap := make(map[string]models.Node)
	for _, node := range workflow.Nodes {
		nodeMap[node.ID] = node
	}

	// Validate all edge node references
	for _, edge := range workflow.Edges {
		if err := v.ValidateEdge(&edge, workflow.Nodes); err != nil {
			return err
		}
	}

	// Validate node input variable references
	for _, node := range workflow.Nodes {
		for _, input := range node.Inputs {
			if input.Reference != nil && input.Reference.NodeID != "" {
				if _, exists := nodeMap[input.Reference.NodeID]; !exists {
					return fmt.Errorf("node %s references non-existent node %s",
						node.ID, input.Reference.NodeID)
				}
			}
		}
	}

	return nil
}

// validateNodeConfig validates node configuration
func (v *UnifiedDSLValidator) validateNodeConfig(node *models.Node) error {
	if node.Config == nil {
		return nil // Some nodes may not require configuration
	}

	// Validate specific configuration based on node type
	switch node.Type {
	case models.NodeTypeLLM:
		return v.validateLLMConfig(node.Config)
	case models.NodeTypeCode:
		return v.validateCodeConfig(node.Config)
	case models.NodeTypeCondition:
		return v.validateConditionConfig(node.Config)
	case models.NodeTypeClassifier:
		return v.validateClassifierConfig(node.Config)
	case models.NodeTypeIteration:
		return v.validateIterationConfig(node.Config)
	}

	return nil
}

// validateLLMConfig validates LLM node configuration
func (v *UnifiedDSLValidator) validateLLMConfig(config interface{}) error {
	llmConfig, ok := AsLLMConfig(config)
	if !ok || llmConfig == nil {
		return fmt.Errorf("invalid LLM config type")
	}

	if llmConfig.Model.Name == "" {
		return fmt.Errorf("LLM model name is required")
	}

	return nil
}

// validateCodeConfig validates code node configuration
func (v *UnifiedDSLValidator) validateCodeConfig(config interface{}) error {
	codeConfig, ok := AsCodeConfig(config)
	if !ok || codeConfig == nil {
		return fmt.Errorf("invalid code config type")
	}

	if codeConfig.Code == "" {
		return fmt.Errorf("code content is required")
	}

	return nil
}

// validateConditionConfig validates condition node configuration
func (v *UnifiedDSLValidator) validateConditionConfig(config interface{}) error {
	conditionConfig, ok := AsConditionConfig(config)
	if !ok || conditionConfig == nil {
		return fmt.Errorf("invalid condition config type")
	}

	if len(conditionConfig.Cases) == 0 {
		return fmt.Errorf("at least one condition case is required")
	}

	return nil
}

// validateClassifierConfig validates classifier node configuration
func (v *UnifiedDSLValidator) validateClassifierConfig(config interface{}) error {
	classifierConfig, ok := AsClassifierConfig(config)
	if !ok || classifierConfig == nil {
		return fmt.Errorf("invalid classifier config type")
	}

	if len(classifierConfig.Classes) == 0 {
		return fmt.Errorf("at least one class is required")
	}

	return nil
}

// validateIterationConfig validates iteration node configuration
func (v *UnifiedDSLValidator) validateIterationConfig(config interface{}) error {
	iterationConfig, ok := AsIterationConfig(config)
	if !ok || iterationConfig == nil {
		return fmt.Errorf("invalid iteration config type")
	}

	if len(iterationConfig.SubWorkflow.Nodes) == 0 {
		return fmt.Errorf("iteration sub-workflow must have nodes")
	}

	// Create sub-workflow object for validation
	subWorkflow := &models.Workflow{
		Nodes: iterationConfig.SubWorkflow.Nodes,
		Edges: iterationConfig.SubWorkflow.Edges,
	}

	// Recursively validate sub-workflow
	return v.ValidateWorkflow(subWorkflow)
}

// isSupportedNodeType checks if node type is supported
func (v *UnifiedDSLValidator) isSupportedNodeType(nodeType models.NodeType) bool {
	supportedTypes := []models.NodeType{
		models.NodeTypeStart,
		models.NodeTypeEnd,
		models.NodeTypeLLM,
		models.NodeTypeCode,
		models.NodeTypeCondition,
		models.NodeTypeClassifier,
		models.NodeTypeIteration,
	}

	for _, supportedType := range supportedTypes {
		if nodeType == supportedType {
			return true
		}
	}

	return false
}

// nodeExists checks if node exists
func (v *UnifiedDSLValidator) nodeExists(nodeID string, nodes []models.Node) bool {
	for _, node := range nodes {
		if node.ID == nodeID {
			return true
		}
	}
	return false
}
