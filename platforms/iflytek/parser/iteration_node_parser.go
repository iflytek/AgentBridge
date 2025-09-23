package parser

import (
	"agentbridge/internal/models"
	"fmt"
)

// IterationNodeParser parses iteration nodes.
type IterationNodeParser struct {
	*BaseNodeParser
}

func NewIterationNodeParser(variableRefSystem *models.VariableReferenceSystem) *IterationNodeParser {
	return &IterationNodeParser{
		BaseNodeParser: NewBaseNodeParser(variableRefSystem),
	}
}

// GetSupportedType returns the supported node type.
func (p *IterationNodeParser) GetSupportedType() string {
	return "迭代"
}

// ValidateNode validates node data.
func (p *IterationNodeParser) ValidateNode(iflytekNode IFlytekNode) error {
	if iflytekNode.ID == "" {
		return fmt.Errorf("node ID is empty")
	}

	if iflytekNode.Type != p.GetSupportedType() {
		return fmt.Errorf("invalid node type: expected %s, got %s", p.GetSupportedType(), iflytekNode.Type)
	}

	return nil
}

// ParseNode parses a node.
func (p *IterationNodeParser) ParseNode(iflytekNode IFlytekNode) (*models.Node, error) {
	if err := p.ValidateNode(iflytekNode); err != nil {
		return nil, err
	}

	// Parse basic information
	node := p.ParseBasicNodeInfo(iflytekNode, models.NodeTypeIteration)

	// Parse inputs and outputs
	if err := p.parseInputsOutputs(node, iflytekNode.Data); err != nil {
		return nil, err
	}

	// Parse configuration
	config, err := p.parseIterationConfig(iflytekNode.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse iteration config: %w", err)
	}
	node.Config = config

	// Save platform-specific configuration
	p.SavePlatformConfig(node, iflytekNode)

	return node, nil
}

// parseInputsOutputs parses inputs and outputs.
func (p *IterationNodeParser) parseInputsOutputs(node *models.Node, data map[string]interface{}) error {
	// Parse inputs
	if inputs, ok := data["inputs"].([]interface{}); ok {
		nodeInputs, err := p.ParseNodeInputs(inputs)
		if err != nil {
			return fmt.Errorf("failed to parse iteration node inputs: %w", err)
		}
		node.Inputs = nodeInputs
	}

	// Parse outputs
	if outputs, ok := data["outputs"].([]interface{}); ok {
		nodeOutputs, err := p.ParseNodeOutputs(outputs)
		if err != nil {
			return fmt.Errorf("failed to parse iteration node outputs: %w", err)
		}
		node.Outputs = nodeOutputs
	}

	return nil
}

// parseIterationConfig parses iteration configuration.
func (p *IterationNodeParser) parseIterationConfig(data map[string]interface{}) (*models.IterationConfig, error) {
	config := &models.IterationConfig{}

	// Get configuration from nodeParam
	if nodeParam, ok := data["nodeParam"].(map[string]interface{}); ok {
		// Parse iterator configuration
		iteratorConfig, err := p.parseIteratorConfig(nodeParam, data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse iterator config: %w", err)
		}
		config.Iterator = *iteratorConfig

		// Parse execution configuration (use default values)
		config.Execution = models.ExecutionConfig{
			IsParallel:      false,      // Default to serial execution
			ParallelNums:    1,          // Default parallel number is 1
			ErrorHandleMode: "continue", // Default error handling mode is continue
		}

		// Parse sub-workflow configuration
		subWorkflowConfig, err := p.parseSubWorkflowConfig(nodeParam)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sub workflow config: %w", err)
		}
		config.SubWorkflow = *subWorkflowConfig
	}

	return config, nil
}

// parseIteratorConfig parses iterator configuration.
func (p *IterationNodeParser) parseIteratorConfig(nodeParam map[string]interface{}, data map[string]interface{}) (*models.IteratorConfig, error) {
	iteratorConfig := &models.IteratorConfig{
		InputType: "array", // Default to array type
	}

	// Get source node and output information from inputs
	if inputs, ok := data["inputs"].([]interface{}); ok && len(inputs) > 0 {
		if firstInput, ok := inputs[0].(map[string]interface{}); ok {
			// Parse reference information from schema
			if schema, ok := firstInput["schema"].(map[string]interface{}); ok {
				if value, ok := schema["value"].(map[string]interface{}); ok {
					if content, ok := value["content"].(map[string]interface{}); ok {
						if nodeId, ok := content["nodeId"].(string); ok {
							iteratorConfig.SourceNode = nodeId
						}
						if name, ok := content["name"].(string); ok {
							iteratorConfig.SourceOutput = name
						}
					}
				}
			}
		}
	}

	// If no source node found from inputs, provide a placeholder
	// The actual source will be determined by edge connections during workflow processing
	if iteratorConfig.SourceNode == "" {
		iteratorConfig.SourceNode = "placeholder_source"
		iteratorConfig.SourceOutput = "result"
	}

	return iteratorConfig, nil
}

// parseSubWorkflowConfig parses sub-workflow configuration.
func (p *IterationNodeParser) parseSubWorkflowConfig(nodeParam map[string]interface{}) (*models.SubWorkflowConfig, error) {
	subWorkflowConfig := &models.SubWorkflowConfig{
		Nodes: make([]models.Node, 0),
		Edges: make([]models.Edge, 0),
	}

	// Parse iteration start node ID
	if startNodeId, ok := nodeParam["IterationStartNodeId"].(string); ok {
		subWorkflowConfig.StartNodeID = startNodeId
	}

	// Note: Specific nodes and edges of sub-workflow need to be obtained from parent parser
	// Only basic configuration information is set here, specific sub-node parsing needs to be handled at upper level

	return subWorkflowConfig, nil
}
