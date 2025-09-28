package generator

import (
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
)

// NodeGeneratorFactory creates iFlytek SparkAgent node generators
type NodeGeneratorFactory struct {
	generators map[models.NodeType]NodeGenerator
	idMapping  map[string]string
}

func NewNodeGeneratorFactory() *NodeGeneratorFactory {
	factory := &NodeGeneratorFactory{
		generators: make(map[models.NodeType]NodeGenerator),
		idMapping:  make(map[string]string),
	}

	// register all node generators
	factory.registerGenerators()

	return factory
}

// SetIDMapping sets ID mapping
func (f *NodeGeneratorFactory) SetIDMapping(idMapping map[string]string) {
	f.idMapping = idMapping

	// set ID mapping for supported generators
	if endGen, ok := f.generators[models.NodeTypeEnd].(*EndNodeGenerator); ok {
		endGen.SetIDMapping(idMapping)
	}
	if llmGen, ok := f.generators[models.NodeTypeLLM].(*LLMNodeGenerator); ok {
		llmGen.SetIDMapping(idMapping)
	}
	if condGen, ok := f.generators[models.NodeTypeCondition].(*ConditionNodeGenerator); ok {
		condGen.SetIDMapping(idMapping)
	}
	if codeGen, ok := f.generators[models.NodeTypeCode].(*CodeNodeGenerator); ok {
		codeGen.SetIDMapping(idMapping)
	}
	if classifierGen, ok := f.generators[models.NodeTypeClassifier].(*ClassifierNodeGenerator); ok {
		classifierGen.SetIDMapping(idMapping)
	}
	if iterationGen, ok := f.generators[models.NodeTypeIteration].(*IterationNodeGenerator); ok {
		iterationGen.SetIDMapping(idMapping)
	}
}

// SetNodeTitleMapping sets node title mapping
func (f *NodeGeneratorFactory) SetNodeTitleMapping(nodeTitleMapping map[string]string) {
	// set node title mapping for supported generators
	if endGen, ok := f.generators[models.NodeTypeEnd].(*EndNodeGenerator); ok {
		endGen.SetNodeTitleMapping(nodeTitleMapping)
	}
	if llmGen, ok := f.generators[models.NodeTypeLLM].(*LLMNodeGenerator); ok {
		llmGen.SetNodeTitleMapping(nodeTitleMapping)
	}
	if codeGen, ok := f.generators[models.NodeTypeCode].(*CodeNodeGenerator); ok {
		codeGen.SetNodeTitleMapping(nodeTitleMapping)
	}
	if condGen, ok := f.generators[models.NodeTypeCondition].(*ConditionNodeGenerator); ok {
		condGen.SetNodeTitleMapping(nodeTitleMapping)
	}
	if classifierGen, ok := f.generators[models.NodeTypeClassifier].(*ClassifierNodeGenerator); ok {
		classifierGen.SetNodeTitleMapping(nodeTitleMapping)
	}
	if iterationGen, ok := f.generators[models.NodeTypeIteration].(*IterationNodeGenerator); ok {
		iterationGen.SetNodeTitleMapping(nodeTitleMapping)
	}
}

// registerGenerators registers all node generators
func (f *NodeGeneratorFactory) registerGenerators() {
	f.generators[models.NodeTypeStart] = NewStartNodeGenerator()
	f.generators[models.NodeTypeEnd] = NewEndNodeGenerator()
	f.generators[models.NodeTypeLLM] = NewLLMNodeGenerator()
	f.generators[models.NodeTypeCondition] = NewConditionNodeGenerator()
	f.generators[models.NodeTypeCode] = NewCodeNodeGenerator()
	f.generators[models.NodeTypeClassifier] = NewClassifierNodeGenerator()
	f.generators[models.NodeTypeIteration] = NewIterationNodeGenerator()
}

// GetGenerator returns generator for specified node type
func (f *NodeGeneratorFactory) GetGenerator(nodeType models.NodeType) (NodeGenerator, error) {
	// For classifier nodes, create an instance each time to avoid mapping conflicts
	if nodeType == models.NodeTypeClassifier {
		classifierGen := NewClassifierNodeGenerator()
		classifierGen.SetIDMapping(f.idMapping)
		return classifierGen, nil
	}

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

// GenerateNode generates node (convenience method)
func (f *NodeGeneratorFactory) GenerateNode(node models.Node) (IFlytekNode, error) {
	generator, err := f.GetGenerator(node.Type)
	if err != nil {
		return IFlytekNode{}, err
	}
	return generator.GenerateNode(node)
}
