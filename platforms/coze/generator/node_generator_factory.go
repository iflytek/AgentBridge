package generator

import (
	"ai-agents-transformer/internal/models"
	"fmt"
)

// NodeGeneratorFactory creates node generators for different node types
type NodeGeneratorFactory struct {
	generators map[models.NodeType]CozeNodeGenerator
}

// NewNodeGeneratorFactory creates a node generator factory
func NewNodeGeneratorFactory() *NodeGeneratorFactory {
	factory := &NodeGeneratorFactory{
		generators: make(map[models.NodeType]CozeNodeGenerator),
	}

	// Register generators for Phase 1: start and end nodes only
	factory.registerGenerators()
	
	return factory
}

// registerGenerators registers node generators for all supported node types
func (f *NodeGeneratorFactory) registerGenerators() {
	// Phase 1-2: Basic nodes (completed)
	f.generators[models.NodeTypeStart] = NewStartNodeGenerator()
	f.generators[models.NodeTypeEnd] = NewEndNodeGenerator()
	
	// Phase 3: Core nodes (completed)
	f.generators[models.NodeTypeLLM] = NewLLMNodeGenerator()
	f.generators[models.NodeTypeCondition] = NewConditionNodeGenerator()
	
	// Phase 4: Advanced nodes (completed)
	f.generators[models.NodeTypeCode] = NewCodeNodeGenerator()
	f.generators[models.NodeTypeClassifier] = NewClassifierNodeGenerator()
	f.generators[models.NodeTypeIteration] = NewIterationNodeGenerator()
}

// GetNodeGenerator returns the appropriate node generator for the given node type
func (f *NodeGeneratorFactory) GetNodeGenerator(nodeType models.NodeType) (CozeNodeGenerator, error) {
	generator, exists := f.generators[nodeType]
	if !exists {
		return nil, fmt.Errorf("unsupported node type: %s", nodeType)
	}

	return generator, nil
}

// GetSupportedNodeTypes returns list of supported node types
func (f *NodeGeneratorFactory) GetSupportedNodeTypes() []models.NodeType {
	supportedTypes := make([]models.NodeType, 0, len(f.generators))
	for nodeType := range f.generators {
		supportedTypes = append(supportedTypes, nodeType)
	}
	return supportedTypes
}