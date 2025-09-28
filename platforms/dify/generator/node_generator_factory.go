package generator

import (
	"fmt"
	"github.com/iflytek/agentbridge/internal/models"
)

// NodeGeneratorFactory provides node generator factory functionality
type NodeGeneratorFactory struct {
	generators map[models.NodeType]NodeGenerator
}

func NewNodeGeneratorFactory() *NodeGeneratorFactory {
	factory := &NodeGeneratorFactory{
		generators: make(map[models.NodeType]NodeGenerator),
	}

	// Register all node generators
	factory.registerGenerators()

	return factory
}

// registerGenerators registers all node generators
func (f *NodeGeneratorFactory) registerGenerators() {
	f.generators[models.NodeTypeStart] = NewStartNodeGenerator()
	f.generators[models.NodeTypeEnd] = NewEndNodeGenerator()
	f.generators[models.NodeTypeLLM] = NewLLMNodeGenerator()
	f.generators[models.NodeTypeCode] = NewCodeNodeGenerator()
	f.generators[models.NodeTypeCondition] = NewConditionNodeGenerator()
	f.generators[models.NodeTypeClassifier] = NewClassifierNodeGenerator()
	f.generators[models.NodeTypeIteration] = NewIterationNodeGenerator()
}

// GetGenerator returns the node generator for the specified type
func (f *NodeGeneratorFactory) GetGenerator(nodeType models.NodeType) (NodeGenerator, error) {
	generator, exists := f.generators[nodeType]
	if !exists {
		return nil, fmt.Errorf("unsupported node type: %s", nodeType)
	}
	return generator, nil
}

// GetSupportedTypes returns all supported node types
func (f *NodeGeneratorFactory) GetSupportedTypes() []models.NodeType {
	types := make([]models.NodeType, 0, len(f.generators))
	for nodeType := range f.generators {
		types = append(types, nodeType)
	}
	return types
}

// SetNodeMapping sets node mapping for all supported generators
func (f *NodeGeneratorFactory) SetNodeMapping(nodes []models.Node) {
	// Set node mapping for LLM node generator
	if llmGen, ok := f.generators[models.NodeTypeLLM].(*LLMNodeGenerator); ok {
		llmGen.SetNodeMapping(nodes)
	}

	// Set node mapping for Code node generator
	if codeGen, ok := f.generators[models.NodeTypeCode].(*CodeNodeGenerator); ok {
		codeGen.SetNodeMapping(nodes)
	}

	// Set node mapping for Classifier node generator
	if classifierGen, ok := f.generators[models.NodeTypeClassifier].(*ClassifierNodeGenerator); ok {
		classifierGen.SetNodeMapping(nodes)
	}

	// Set node mapping for Condition node generator
	if conditionGen, ok := f.generators[models.NodeTypeCondition].(*ConditionNodeGenerator); ok {
		conditionGen.SetNodeMapping(nodes)
	}

	// Set node mapping for Iteration node generator
	if iterationGen, ok := f.generators[models.NodeTypeIteration].(*IterationNodeGenerator); ok {
		iterationGen.SetNodeMapping(nodes)
	}

	// Future: Add similar settings for other generators that need node mapping
}

// GenerateNode generates a node (convenience method)
func (f *NodeGeneratorFactory) GenerateNode(node models.Node) (DifyNode, error) {
	generator, err := f.GetGenerator(node.Type)
	if err != nil {
		return DifyNode{}, err
	}
	return generator.GenerateNode(node)
}
