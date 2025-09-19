package generator

import (
	"ai-agents-transformer/internal/models"
	"ai-agents-transformer/platforms/common"
)

// StartNodeGenerator handles start node generation
type StartNodeGenerator struct {
	*BaseNodeGenerator
}

func NewStartNodeGenerator() *StartNodeGenerator {
	return &StartNodeGenerator{
		BaseNodeGenerator: NewBaseNodeGenerator(models.NodeTypeStart),
	}
}

// GenerateNode generates start node
func (g *StartNodeGenerator) GenerateNode(node models.Node) (IFlytekNode, error) {
	// Validate node
	if err := g.ValidateNode(node); err != nil {
		return IFlytekNode{}, err
	}

	// Generate basic node structure
	iflytekNode := g.createBasicStartNode(node)

	// Generate node outputs
	g.generateNodeOutputs(&iflytekNode, node)

	// Ensure required default output exists
	g.ensureDefaultUserInputOutput(&iflytekNode)

	return iflytekNode, nil
}

// createBasicStartNode creates the basic start node structure
func (g *StartNodeGenerator) createBasicStartNode(node models.Node) IFlytekNode {
	iflytekNode := g.generateBasicNodeInfo(node)

	// Configure start node specific properties
	iflytekNode.Data.AllowInputReference = false
	iflytekNode.Data.AllowOutputReference = true
	iflytekNode.Data.Icon = g.getNodeIcon(models.NodeTypeStart)

	return iflytekNode
}

// generateNodeOutputs generates outputs based on node configuration
func (g *StartNodeGenerator) generateNodeOutputs(iflytekNode *IFlytekNode, node models.Node) {
	if startConfig, ok := common.AsStartConfig(node.Config); ok && startConfig != nil {
		g.generateOutputsFromConfig(iflytekNode, *startConfig)
	} else {
		iflytekNode.Data.Outputs = g.generateOutputs(node.Outputs)
	}
}

// generateOutputsFromConfig generates outputs from start node configuration
func (g *StartNodeGenerator) generateOutputsFromConfig(iflytekNode *IFlytekNode, config models.StartConfig) {
	for _, variable := range config.Variables {
		output := g.createOutputFromVariable(variable)
		iflytekNode.Data.Outputs = append(iflytekNode.Data.Outputs, output)
	}
}

// createOutputFromVariable creates an output from a variable definition
func (g *StartNodeGenerator) createOutputFromVariable(variable models.Variable) IFlytekOutput {
	output := IFlytekOutput{
		ID:         g.generateOutputID(),
		Name:       variable.Name,
		NameErrMsg: "",
		Schema: IFlytekSchema{
			Type:       g.convertDataType(models.UnifiedDataType(variable.Type)),
			Properties: []interface{}{},
			Default:    variable.Default,
		},
		Required: variable.Required,
	}

	// Handle custom parameter type for non-string types
	if models.UnifiedDataType(variable.Type) != models.DataTypeString {
		output.CustomParameterType = "xfyun-file"
	}

	return output
}

// ensureDefaultUserInputOutput ensures AGENT_USER_INPUT output exists
func (g *StartNodeGenerator) ensureDefaultUserInputOutput(iflytekNode *IFlytekNode) {
	if g.hasAgentUserInputOutput(iflytekNode.Data.Outputs) {
		return
	}

	defaultOutput := g.createDefaultUserInputOutput()
	g.prependOutput(iflytekNode, defaultOutput)
}

// hasAgentUserInputOutput checks if AGENT_USER_INPUT output already exists
func (g *StartNodeGenerator) hasAgentUserInputOutput(outputs []IFlytekOutput) bool {
	for _, output := range outputs {
		if output.Name == "AGENT_USER_INPUT" {
			return true
		}
	}
	return false
}

// createDefaultUserInputOutput creates the default user input output
func (g *StartNodeGenerator) createDefaultUserInputOutput() IFlytekOutput {
	return IFlytekOutput{
		ID:         g.generateOutputID(),
		Name:       "AGENT_USER_INPUT",
		NameErrMsg: "",
		Schema: IFlytekSchema{
			Type:    "string",
			Default: "User input content for current conversation round",
		},
		Required:       true,
		DeleteDisabled: true,
	}
}

// prependOutput prepends an output to the beginning of the outputs list
func (g *StartNodeGenerator) prependOutput(iflytekNode *IFlytekNode, output IFlytekOutput) {
	outputs := []IFlytekOutput{output}
	outputs = append(outputs, iflytekNode.Data.Outputs...)
	iflytekNode.Data.Outputs = outputs
}
