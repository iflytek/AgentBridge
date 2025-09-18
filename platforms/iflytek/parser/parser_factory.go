package parser

import (
	"fmt"
	"ai-agents-transformer/internal/models"
)

// iFlytek node type constants
const (
	IFlytekNodeTypeStart      = "开始节点"
	IFlytekNodeTypeEnd        = "结束节点"
	IFlytekNodeTypeLLM        = "大模型"
	IFlytekNodeTypeCode       = "代码"
	IFlytekNodeTypeCondition  = "分支器"
	IFlytekNodeTypeClassifier = "决策"
	IFlytekNodeTypeIteration  = "迭代"
)

// TypeProvider provides node output type querying interface
type TypeProvider interface {
	GetOutputType(nodeID, outputName string) models.UnifiedDataType
}

// ParserFactory creates node parsers.
type ParserFactory struct {
	parsers map[string]func(*models.VariableReferenceSystem, TypeProvider) NodeParser
}

func NewParserFactory() *ParserFactory {
	factory := &ParserFactory{
		parsers: make(map[string]func(*models.VariableReferenceSystem, TypeProvider) NodeParser),
	}

	// Register basic node parsers
	factory.Register(IFlytekNodeTypeStart, func(vrs *models.VariableReferenceSystem, tp TypeProvider) NodeParser {
		return NewStartNodeParser(vrs)
	})
	factory.Register(IFlytekNodeTypeEnd, func(vrs *models.VariableReferenceSystem, tp TypeProvider) NodeParser {
		return NewEndNodeParser(vrs)
	})
	factory.Register(IFlytekNodeTypeLLM, func(vrs *models.VariableReferenceSystem, tp TypeProvider) NodeParser {
		return NewLLMNodeParser(vrs)
	})
	factory.Register(IFlytekNodeTypeCode, func(vrs *models.VariableReferenceSystem, tp TypeProvider) NodeParser {
		return NewCodeNodeParser(vrs)
	})

	// Register advanced node parsers - condition parser requires TypeProvider
	factory.Register(IFlytekNodeTypeCondition, func(vrs *models.VariableReferenceSystem, tp TypeProvider) NodeParser {
		return NewConditionNodeParserWithTypeProvider(vrs, tp)
	})
	factory.Register(IFlytekNodeTypeClassifier, func(vrs *models.VariableReferenceSystem, tp TypeProvider) NodeParser {
		return NewClassifierNodeParser(vrs)
	})
	factory.Register(IFlytekNodeTypeIteration, func(vrs *models.VariableReferenceSystem, tp TypeProvider) NodeParser {
		return NewIterationNodeParser(vrs)
	})

	return factory
}

// Register registers a parser with TypeProvider support.
func (f *ParserFactory) Register(nodeType string, creator func(*models.VariableReferenceSystem, TypeProvider) NodeParser) {
	f.parsers[nodeType] = creator
}

// CreateParser creates a parser with TypeProvider support.
func (f *ParserFactory) CreateParser(nodeType string, variableRefSystem *models.VariableReferenceSystem, typeProvider TypeProvider) (NodeParser, error) {
	creator, exists := f.parsers[nodeType]
	if !exists {
		return nil, fmt.Errorf("no parser found for node type: %s", nodeType)
	}

	return creator(variableRefSystem, typeProvider), nil
}

// GetSupportedTypes returns supported node types.
func (f *ParserFactory) GetSupportedTypes() []string {
	types := make([]string, 0, len(f.parsers))
	for nodeType := range f.parsers {
		types = append(types, nodeType)
	}
	return types
}

// ParseNodeWithContext uses factory to parse a node with TypeProvider context.
func (f *ParserFactory) ParseNodeWithContext(iflytekNode IFlytekNode, variableRefSystem *models.VariableReferenceSystem, typeProvider TypeProvider) (*models.Node, error) {
	parser, err := f.CreateParser(iflytekNode.Type, variableRefSystem, typeProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser for node type %s: %w", iflytekNode.Type, err)
	}

	return parser.ParseNode(iflytekNode)
}

// CreateParserWithFallback creates a parser with graceful fallback support
func (f *ParserFactory) CreateParserWithFallback(nodeType string, variableRefSystem *models.VariableReferenceSystem, typeProvider TypeProvider) (NodeParser, bool, error) {
	creator, exists := f.parsers[nodeType]
	if !exists {
		return nil, false, nil // Not supported but don't error
	}
	parser := creator(variableRefSystem, typeProvider)
	return parser, true, nil
}

// ParseNodeWithFallback parses a node using fallback mechanism
func (f *ParserFactory) ParseNodeWithFallback(iflytekNode IFlytekNode, variableRefSystem *models.VariableReferenceSystem, typeProvider TypeProvider) (*models.Node, bool, error) {
	parser, supported, err := f.CreateParserWithFallback(iflytekNode.Type, variableRefSystem, typeProvider)
	if err != nil {
		return nil, false, err
	}
	if !supported {
		return nil, false, nil // Node type not supported, return flag
	}
	
	node, err := parser.ParseNode(iflytekNode)
	return node, true, err
}

// ParseNode uses factory to parse a node (backward compatibility).
func (f *ParserFactory) ParseNode(iflytekNode IFlytekNode, variableRefSystem *models.VariableReferenceSystem) (*models.Node, error) {
	// Use nil TypeProvider for backward compatibility
	return f.ParseNodeWithContext(iflytekNode, variableRefSystem, nil)
}
