package parser

import (
	"fmt"
	"ai-agents-transformer/internal/models"
)

// ClassifierNodeParser parses Dify classifier nodes.
type ClassifierNodeParser struct {
	*BaseNodeParser
}

func NewClassifierNodeParser(variableRefSystem *models.VariableReferenceSystem) *ClassifierNodeParser {
	return &ClassifierNodeParser{
		BaseNodeParser: NewBaseNodeParser("question-classifier", variableRefSystem),
	}
}

// ParseNode parses classifier node.
func (p *ClassifierNodeParser) ParseNode(difyNode DifyNode) (*models.Node, error) {
	// Use base parser for basic parsing
	node, err := p.BaseNodeParser.ParseNode(difyNode)
	if err != nil {
		return nil, fmt.Errorf("base parsing failed: %w", err)
	}

	// Set node type
	node.Type = models.NodeTypeClassifier

	// Parse classifier configuration
	config, err := p.parseClassifierConfig(difyNode.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse classifier configuration: %w", err)
	}
	node.Config = config

	// Parse input parameters
	inputs, err := p.parseInputs(difyNode.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input parameters: %w", err)
	}
	node.Inputs = inputs

	// Set output parameters - iFlytek SparkAgent classifier only allows 1 output
	node.Outputs = []models.Output{
		{
			Name:        "class_name",
			Type:        models.DataTypeString,
			Description: "Classification result name",
		},
	}

	return node, nil
}

// parseClassifierConfig parses classifier configuration.
func (p *ClassifierNodeParser) parseClassifierConfig(data DifyNodeData) (models.ClassifierConfig, error) {
	config := models.ClassifierConfig{
		Classes: []models.ClassifierClass{},
	}

	// Parse model configuration
	if data.Model != nil {
		modelConfig, err := p.parseModelConfig(*data.Model)
		if err != nil {
			return config, fmt.Errorf("failed to parse model configuration: %w", err)
		}
		config.Model = modelConfig
	}

	// Parse classification classes, filter out "other classification"
	for _, difyClass := range data.Classes {
		// Skip "other classification" as it will be handled as default intent in iFlytek
		if difyClass.Name == "其他分类" {
			continue
		}
		
		class := models.ClassifierClass{
			ID:   difyClass.ID,
			Name: difyClass.Name,
			// Dify Class structure doesn't have description field, use name as description
			Description: difyClass.Name,
		}
		config.Classes = append(config.Classes, class)
	}

	// Parse query variable
	if len(data.QueryVariableSelector) >= 2 {
		nodeID := data.QueryVariableSelector[0]
		outputName := data.QueryVariableSelector[1]
		config.QueryVariable = fmt.Sprintf("%s.%s", nodeID, outputName)
	}

	// Parse instructions (optional, prefer instruction over instructions)
	if data.Instruction != "" {
		config.Instructions = data.Instruction
	} else if data.Instructions != "" {
		config.Instructions = data.Instructions
	}

	return config, nil
}

// parseModelConfig parses model configuration.
func (p *ClassifierNodeParser) parseModelConfig(modelData DifyModel) (models.ModelConfig, error) {
	config := models.ModelConfig{
		Provider: modelData.Provider,
		Name:     modelData.Name,
		Mode:     modelData.Mode,
	}

	return config, nil
}

// parseInputs parses input parameters.
func (p *ClassifierNodeParser) parseInputs(data DifyNodeData) ([]models.Input, error) {
	var inputs []models.Input

	// Classifier node mainly specifies input source through query_variable_selector
	if len(data.QueryVariableSelector) >= 2 {
		nodeID := data.QueryVariableSelector[0]
		outputName := data.QueryVariableSelector[1]

		// Create variable reference
		reference := &models.VariableReference{
			Type:       models.ReferenceTypeNodeOutput,
			NodeID:     nodeID,
			OutputName: outputName,
			DataType:   models.DataTypeString,
		}

		input := models.Input{
			Name:      "Query",
			Type:      models.DataTypeString,
			Required:  true,
			Reference: reference,
		}

		inputs = append(inputs, input)
	}

	return inputs, nil
}


