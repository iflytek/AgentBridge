package parser

import (
	"fmt"

	"github.com/iflytek/agentbridge/internal/models"
)

// IterationNodeParser parses iteration nodes.
type IterationNodeParser struct {
	*BaseNodeParser
}

func NewIterationNodeParser(variableRefSystem *models.VariableReferenceSystem) NodeParser {
	return &IterationNodeParser{
		BaseNodeParser: NewBaseNodeParser("iteration", variableRefSystem),
	}
}

// ParseNode parses iteration node.
func (p *IterationNodeParser) ParseNode(difyNode DifyNode) (*models.Node, error) {
	// Extract basic information
	id := difyNode.ID
	data := difyNode.Data

	title := data.Title
	if title == "" {
		title = "Iteration Node"
	}

	description := data.Desc

	// Parse iteration configuration
	iterationConfig, err := p.parseIterationConfig(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse iteration configuration: %v", err)
	}

	// Parse inputs and outputs
	inputs, err := p.parseIterationInputs(data, iterationConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse iteration inputs: %v", err)
	}

	outputs := p.parseIterationOutputs(data, iterationConfig)

	// Create unified node
	node := &models.Node{
		ID:          id,
		Type:        models.NodeTypeIteration,
		Title:       title,
		Description: description,
		Position:    models.Position{X: difyNode.Position.X, Y: difyNode.Position.Y},
		Size:        models.Size{Width: difyNode.Width, Height: difyNode.Height},
		Config:      iterationConfig,
		Inputs:      inputs,
		Outputs:     outputs,
	}

	return node, nil
}

// parseIterationConfig parses iteration configuration.
func (p *IterationNodeParser) parseIterationConfig(data DifyNodeData) (models.IterationConfig, error) {
	config := models.IterationConfig{}

	// Parse iterator configuration
	if data.IteratorInputType != "" {
		config.Iterator.InputType = data.IteratorInputType
	} else {
		return config, fmt.Errorf("missing iterator_input_type")
	}

	// Parse data source selector
	if len(data.IteratorSelector) >= 2 {
		config.Iterator.SourceNode = data.IteratorSelector[0]
		config.Iterator.SourceOutput = data.IteratorSelector[1]
	}

	// Parse output selector
	if len(data.OutputSelector) >= 2 {
		config.OutputSelector.NodeID = data.OutputSelector[0]
		config.OutputSelector.OutputName = data.OutputSelector[1]
	}

	// Parse execution configuration
	config.Execution.IsParallel = data.IsParallel
	config.Execution.ParallelNums = data.ParallelNums
	config.Execution.ErrorHandleMode = data.ErrorHandleMode

	// Parse start node ID
	config.SubWorkflow.StartNodeID = data.StartNodeID

	// Output type
	config.OutputType = data.OutputType

	return config, nil
}

// parseIterationInputs parses iteration inputs.
func (p *IterationNodeParser) parseIterationInputs(data DifyNodeData, config models.IterationConfig) ([]models.Input, error) {
	inputs := []models.Input{}

	// Iterator input
	if config.Iterator.SourceNode != "" && config.Iterator.SourceOutput != "" {
		input := models.Input{
			Name: "input",
			Type: p.convertDataType(config.Iterator.InputType),
			Reference: &models.VariableReference{
				Type:       models.ReferenceTypeNodeOutput,
				NodeID:     config.Iterator.SourceNode,
				OutputName: config.Iterator.SourceOutput,
				DataType:   p.convertDataType(config.Iterator.InputType),
			},
		}
		inputs = append(inputs, input)
	}

	return inputs, nil
}

// parseIterationOutputs parses iteration outputs.
func (p *IterationNodeParser) parseIterationOutputs(data DifyNodeData, config models.IterationConfig) []models.Output {
	outputs := []models.Output{}

	// Main output
	output := models.Output{
		Name: "output",
		Type: p.convertDataType(config.OutputType),
	}
	if config.OutputType == "" {
		output.Type = models.DataTypeArrayString // Default type
	}

	outputs = append(outputs, output)

	return outputs
}
