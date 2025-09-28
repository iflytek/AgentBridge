package generator

import (
	"github.com/iflytek/agentbridge/internal/models"
	"github.com/iflytek/agentbridge/platforms/common"
	"fmt"
	"regexp"
	"strings"
)

// ClassifierNodeGenerator handles iFlytek SparkAgent classifier node generation
type ClassifierNodeGenerator struct {
	*BaseNodeGenerator
	idMapping         map[string]string
	classIDToIntentID map[string]string // Classification ID to intent ID mapping
	nodeTitleMapping  map[string]string // iFlytek SparkAgent ID -> node title mapping
}

func NewClassifierNodeGenerator() *ClassifierNodeGenerator {
	return &ClassifierNodeGenerator{
		BaseNodeGenerator: NewBaseNodeGenerator(models.NodeTypeClassifier),
		idMapping:         make(map[string]string),
		classIDToIntentID: make(map[string]string),
		nodeTitleMapping:  make(map[string]string),
	}
}

// GenerateNode generates classifier node
func (g *ClassifierNodeGenerator) GenerateNode(node models.Node) (IFlytekNode, error) {
	// validate node type
	if node.Type != models.NodeTypeClassifier {
		return IFlytekNode{}, fmt.Errorf("expected classifier node, got %s", node.Type)
	}

	// parse classifier configuration
	classifierConfig, ok := common.AsClassifierConfig(node.Config)
	if !ok || classifierConfig == nil {
		return IFlytekNode{}, fmt.Errorf("invalid classifier config type")
	}

	// generate basic node structure
	iflytekNode := g.generateBasicNodeInfo(node)

	// set classifier specific properties
	iflytekNode.Type = "决策"
	iflytekNode.Data.NodeMeta = IFlytekNodeMeta{
		AliasName: "决策",
		NodeType:  "基础节点",
	}

	// set node parameters
	nodeParam, err := g.generateNodeParam(*classifierConfig, node.Inputs)
	if err != nil {
		return IFlytekNode{}, fmt.Errorf("failed to generate node parameters: %w", err)
	}
	iflytekNode.Data.NodeParam = nodeParam

	// set inputs/outputs (using mapped IDs)
	iflytekNode.Data.Inputs = g.generateInputsWithMapping(node.Inputs)
	iflytekNode.Data.Outputs = g.BaseNodeGenerator.generateOutputs(node.Outputs)

	// set reference information
	iflytekNode.Data.References = g.generateReferences(node.Inputs)

	// set icon and description
	iflytekNode.Data.Icon = g.getNodeIcon(models.NodeTypeClassifier)
	iflytekNode.Data.Description = "结合输入的参数与填写的意图，决定后续的逻辑走向"

	return iflytekNode, nil
}

// generateNodeParam generates node parameters
func (g *ClassifierNodeGenerator) generateNodeParam(config models.ClassifierConfig, inputs []models.Input) (map[string]interface{}, error) {
	nodeParam := map[string]interface{}{
		"topK":    4,
		"modelId": 141,
		"chatHistory": map[string]interface{}{
			"isEnabled": false,
			"rounds":    1,
		},
		"reasonMode":      1,
		"auditing":        "default",
		"llmId":           141,
		"promptPrefix":    g.generatePromptPrefix(config, inputs),
		"url":             "wss://maas-api.cn-huabei-1.xf-yun.com/v1.1/chat",
		"multiMode":       false,
		"uid":             "20718349453", // Default value obtained from examples
		"patchId":         "0",
		"isThink":         false,
		"searchDisable":   true,
		"domain":          "xdeepseekv3",
		"appId":           "12a0a7e2", // Default value obtained from examples
		"maxTokens":       8192,
		"temperature":     0.5,
		"model":           "spark",
		"useFunctionCall": true,
		"serviceId":       "xdeepseekv3",
		"llmIdErrMsg":     "",
	}

	// Generate intent chains
	intentChains := make([]map[string]interface{}, 0)
	var defaultIntentID string

	// create normal intents for all actual classifications
	for i, class := range config.Classes {
		intentID := fmt.Sprintf("intent-one-of::%s", generateUUID())

		// normal classification intent
		intentChain := map[string]interface{}{
			"intentType":        2, // Normal classification intent type
			"name":              g.cleanVariableReferences(class.Name),
			"description":       g.cleanVariableReferences(class.Description),
			"id":                intentID,
			"nameErrMsg":        "",
			"descriptionErrMsg": "",
		}
		intentChains = append(intentChains, intentChain)

		// save classification ID to intent ID mapping
		g.classIDToIntentID[class.ID] = intentID

		// Also map Dify number format (1, 2, 3...) to intent ID for edge conversion
		difyNumberHandle := fmt.Sprintf("%d", i+1)
		g.classIDToIntentID[difyNumberHandle] = intentID
	}

	// create additional default intent
	defaultIntentID = fmt.Sprintf("intent-one-of::%s", generateUUID())
	defaultIntentChain := map[string]interface{}{
		"intentType":        1, // Default intent type
		"name":              "default",
		"description":       "Default intent",
		"id":                defaultIntentID,
		"nameErrMsg":        "",
		"descriptionErrMsg": "",
	}
	intentChains = append(intentChains, defaultIntentChain)

	// save special mapping if default intent exists
	if defaultIntentID != "" {
		g.classIDToIntentID["__default__"] = defaultIntentID
	}

	nodeParam["intentChains"] = intentChains

	return nodeParam, nil
}

// generatePromptPrefix generates prompt prefix from Dify instructions
func (g *ClassifierNodeGenerator) generatePromptPrefix(config models.ClassifierConfig, inputs []models.Input) string {
	if config.Instructions == "" {
		return ""
	}

	return g.unifyVariableReferencesToQuery(config.Instructions)
}

// processInstructionReferences processes variable references in instructions
func (g *ClassifierNodeGenerator) processInstructionReferences(instructions string, inputs []models.Input) string {
	processed := instructions

	// Convert Dify variable references to iFlytek format
	// Pattern: {{#nodeID.variableName#}} -> {{variableName}}
	for _, input := range inputs {
		if input.Reference != nil {
			difyPattern := fmt.Sprintf("{{#%s.%s#}}", input.Reference.NodeID, input.Reference.OutputName)
			iflytekPattern := fmt.Sprintf("{{%s}}", input.Name)
			processed = replaceVariableReference(processed, difyPattern, iflytekPattern)
		}
	}

	return processed
}

// replaceVariableReference safely replaces variable references
func replaceVariableReference(text, pattern, replacement string) string {
	if text == "" || pattern == "" {
		return text
	}
	return strings.ReplaceAll(text, pattern, replacement)
}

// cleanVariableReferences removes all variable references from description text
func (g *ClassifierNodeGenerator) cleanVariableReferences(text string) string {
	if text == "" {
		return ""
	}

	// Use regex to remove all Dify variable references: {{#...#}}
	re := regexp.MustCompile(`\{\{#[^}]*#\}\}`)
	return re.ReplaceAllString(text, "")
}

// unifyVariableReferencesToQuery converts all variable references to {{Query}}
func (g *ClassifierNodeGenerator) unifyVariableReferencesToQuery(instructions string) string {
	if instructions == "" {
		return ""
	}

	// Replace all Dify variable references with {{Query}}
	re := regexp.MustCompile(`\{\{#[^}]*#\}\}`)
	return re.ReplaceAllString(instructions, "{{Query}}")
}

// generateReferences generates reference information
func (g *ClassifierNodeGenerator) generateReferences(inputs []models.Input) []IFlytekReference {
	references := make([]IFlytekReference, 0)

	// generate reference information for each input with references
	for _, input := range inputs {
		if input.Reference != nil && input.Reference.NodeID != "" {
			// get mapped node ID
			mappedNodeID := input.Reference.NodeID
			if g.idMapping != nil {
				if mapped, exists := g.idMapping[input.Reference.NodeID]; exists {
					mappedNodeID = mapped
				}
			}

			ref := IFlytekReference{
				Children: []IFlytekReference{
					{
						References: []IFlytekRefDetail{
							{
								OriginID: mappedNodeID, // use mapped iFlytek node ID
								ID:       generateUUID(),
								Label:    input.Reference.OutputName,
								Type:     g.BaseNodeGenerator.convertDataType(input.Reference.DataType),
								Value:    input.Reference.OutputName,
								FileType: "",
							},
						},
						Label: "",
						Value: "",
					},
				},
				Label:      g.determineLabelByID(mappedNodeID, g.nodeTitleMapping), // determine label by priority
				ParentNode: true,
				Value:      mappedNodeID, // use mapped ID
			}
			references = append(references, ref)
		}
	}

	return references
}

// getNodeLabelByID retrieves node label by ID
// Uses BaseNodeGenerator.determineLabelByID for unified processing

// SetIDMapping sets ID mapping if needed
func (g *ClassifierNodeGenerator) SetIDMapping(idMapping map[string]string) {
	// implement ID mapping functionality here if needed
	g.idMapping = idMapping
}

// SetNodeTitleMapping sets node title mapping
func (g *ClassifierNodeGenerator) SetNodeTitleMapping(nodeTitleMapping map[string]string) {
	g.nodeTitleMapping = nodeTitleMapping
}

// generateInputsWithMapping generates inputs with ID mapping
func (g *ClassifierNodeGenerator) generateInputsWithMapping(inputs []models.Input) []IFlytekInput {
	iflytekInputs := make([]IFlytekInput, 0, len(inputs))

	for _, input := range inputs {
		iflytekInput := IFlytekInput{
			ID:         generateUUID(),
			Name:       input.Name,
			NameErrMsg: "",
			Schema: IFlytekSchema{
				Type:    g.BaseNodeGenerator.convertDataType(input.Type),
				Default: input.Default,
			},
			FileType: "",
		}

		// handle variable references with mapped ID
		if input.Reference != nil {
			// get mapped node ID
			mappedNodeID := input.Reference.NodeID
			if g.idMapping != nil {
				if mapped, exists := g.idMapping[input.Reference.NodeID]; exists {
					mappedNodeID = mapped
				}
			}

			iflytekInput.Schema.Value = &IFlytekSchemaValue{
				Type: "ref",
				Content: &IFlytekRefContent{
					Name:   input.Reference.OutputName,
					ID:     generateUUID(),
					NodeID: mappedNodeID,
				},
				ContentErrMsg: "",
			}
		}

		iflytekInputs = append(iflytekInputs, iflytekInput)
	}

	return iflytekInputs
}

// GetClassIDToIntentIDMapping returns classification ID to intent ID mapping
func (g *ClassifierNodeGenerator) GetClassIDToIntentIDMapping() map[string]string {
	return g.classIDToIntentID
}

// SetClassIDToIntentIDMapping sets classification ID to intent ID mapping
func (g *ClassifierNodeGenerator) SetClassIDToIntentIDMapping(mapping map[string]string) {
	g.classIDToIntentID = mapping
}

// GetSupportedTypes returns supported node types
func (g *ClassifierNodeGenerator) GetSupportedTypes() []models.NodeType {
	return []models.NodeType{models.NodeTypeClassifier}
}
