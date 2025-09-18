package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// VariableReferenceSystem manages variable reference parsing and conversion
type VariableReferenceSystem struct {
	// Used for parsing and generating variable reference formats for different platforms
	// OutputMappings maps nodeID -> (oldOutputName -> newOutputName) for reference resolution
	OutputMappings map[string]map[string]string
}

func NewVariableReferenceSystem() *VariableReferenceSystem {
	return &VariableReferenceSystem{
		OutputMappings: make(map[string]map[string]string),
	}
}

// ParseIFlytekReference parses iFlytek Agent variable reference format
// iFlytek Agent uses references structure to represent variable references
func (vrs *VariableReferenceSystem) ParseIFlytekReference(refData map[string]interface{}) (*VariableReference, error) {
	// Check if it's a reference type
	if refType, ok := refData["type"].(string); ok && refType == "ref" {
		// Parse reference content
		if content, ok := refData["content"].(map[string]interface{}); ok {
			nodeID, _ := content["nodeId"].(string)
			outputName, _ := content["name"].(string)

			// Parse data type
			var dataType UnifiedDataType
			if schemaType, ok := content["type"].(string); ok {
				mapping := GetDefaultDataTypeMapping()
				dataType = mapping.FromIFlytekType(schemaType)
			}

			return &VariableReference{
				Type:       ReferenceTypeNodeOutput,
				NodeID:     nodeID,
				OutputName: outputName,
				DataType:   dataType,
			}, nil
		}
	} else if refType == "literal" {
		// Literal value reference
		value := refData["content"]
		return &VariableReference{
			Type:     ReferenceTypeLiteral,
			DataType: DataTypeString, // Default to string type
			Value:    value,
		}, nil
	}

	return nil, fmt.Errorf("unsupported iflytek reference format")
}

// ParseDifyReference parses Dify variable reference format
// Dify uses value_selector array to represent variable references
func (vrs *VariableReferenceSystem) ParseDifyReference(valueSelector []string, valueType string) (*VariableReference, error) {
	if len(valueSelector) < 2 {
		return nil, fmt.Errorf("invalid dify value_selector format")
	}

	nodeID := valueSelector[0]
	outputName := valueSelector[1]

	// Convert data type
	mapping := GetDefaultDataTypeMapping()
	dataType := mapping.FromDifyType(valueType)

	return &VariableReference{
		Type:       ReferenceTypeNodeOutput,
		NodeID:     nodeID,
		OutputName: outputName,
		DataType:   dataType,
	}, nil
}

// ParseTemplateReference parses variable references in templates
// Supports multiple template formats: {{input_01}}, {{#nodeId.variable#}}, {{$nodes.nodeId.output}}
func (vrs *VariableReferenceSystem) ParseTemplateReference(template string) ([]*VariableReference, error) {
	var references []*VariableReference

	// Parse different template formats
	iflytekRefs := vrs.parseIFlytekTemplateReferences(template)
	references = append(references, iflytekRefs...)

	difyRefs := vrs.parseDifyTemplateReferences(template)
	references = append(references, difyRefs...)

	unifiedRefs := vrs.parseUnifiedTemplateReferences(template)
	references = append(references, unifiedRefs...)

	return references, nil
}

// parseIFlytekTemplateReferences parses iFlytek template format: {{variable_name}}
func (vrs *VariableReferenceSystem) parseIFlytekTemplateReferences(template string) []*VariableReference {
	var references []*VariableReference

	pattern := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := pattern.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) <= 1 {
			continue
		}

		varName := strings.TrimSpace(match[1])
		if vrs.isSimpleVariableReference(varName) {
			ref := &VariableReference{
				Type:     ReferenceTypeTemplate,
				Template: match[0],
				DataType: DataTypeString,
			}
			references = append(references, ref)
		}
	}

	return references
}

// isSimpleVariableReference checks if it's a simple variable reference
func (vrs *VariableReferenceSystem) isSimpleVariableReference(varName string) bool {
	return !strings.Contains(varName, ".") && !strings.Contains(varName, "#")
}

// parseDifyTemplateReferences parses Dify template format: {{#nodeId.variable#}}
func (vrs *VariableReferenceSystem) parseDifyTemplateReferences(template string) []*VariableReference {
	pattern := regexp.MustCompile(`\{\{#([^#]+)\.([^#]+)#\}\}`)
	return vrs.parseNodeOutputReferences(template, pattern)
}

// parseUnifiedTemplateReferences parses Unified DSL format: {{$nodes.nodeId.output}}
func (vrs *VariableReferenceSystem) parseUnifiedTemplateReferences(template string) []*VariableReference {
	pattern := regexp.MustCompile(`\{\{\$nodes\.([^.]+)\.([^}]+)\}\}`)
	return vrs.parseNodeOutputReferences(template, pattern)
}

// parseNodeOutputReferences parses node output references with given pattern
func (vrs *VariableReferenceSystem) parseNodeOutputReferences(template string, pattern *regexp.Regexp) []*VariableReference {
	var references []*VariableReference

	matches := pattern.FindAllStringSubmatch(template, -1)
	for _, match := range matches {
		if len(match) <= 2 {
			continue
		}

		nodeID := strings.TrimSpace(match[1])
		outputName := strings.TrimSpace(match[2])

		ref := &VariableReference{
			Type:       ReferenceTypeNodeOutput,
			NodeID:     nodeID,
			OutputName: outputName,
			DataType:   DataTypeString,
			Template:   match[0],
		}
		references = append(references, ref)
	}

	return references
}

// ToIFlytekReference converts to iFlytek Agent reference format
func (vrs *VariableReferenceSystem) ToIFlytekReference(ref *VariableReference) (map[string]interface{}, error) {
	switch ref.Type {
	case ReferenceTypeNodeOutput:
		mapping := GetDefaultDataTypeMapping()
		iflytekType := mapping.ToIFlytekType(ref.DataType)

		return map[string]interface{}{
			"type": "ref",
			"content": map[string]interface{}{
				"nodeId": ref.NodeID,
				"name":   ref.OutputName,
				"type":   iflytekType,
			},
			"contentErrMsg": "",
		}, nil

	case ReferenceTypeLiteral:
		return map[string]interface{}{
			"type":          "literal",
			"content":       ref.Value,
			"contentErrMsg": "",
		}, nil

	default:
		return nil, fmt.Errorf("unsupported reference type for iflytek: %s", ref.Type)
	}
}

// ToDifyReference converts to Dify reference format
func (vrs *VariableReferenceSystem) ToDifyReference(ref *VariableReference) ([]string, string, error) {
	if ref.Type != ReferenceTypeNodeOutput {
		return nil, "", fmt.Errorf("only node output references supported for dify format")
	}

	valueSelector := []string{ref.NodeID, ref.OutputName}
	mapping := GetDefaultDataTypeMapping()
	valueType := mapping.ToDifyType(ref.DataType)

	return valueSelector, valueType, nil
}

// ToUnifiedTemplate converts to unified DSL template format
func (vrs *VariableReferenceSystem) ToUnifiedTemplate(ref *VariableReference) (string, error) {
	switch ref.Type {
	case ReferenceTypeNodeOutput:
		return fmt.Sprintf("{{$nodes.%s.%s}}", ref.NodeID, ref.OutputName), nil
	case ReferenceTypeLiteral:
		return fmt.Sprintf("%v", ref.Value), nil
	case ReferenceTypeTemplate:
		if ref.Template != "" {
			return ref.Template, nil
		}
		return fmt.Sprintf("{{%s}}", ref.OutputName), nil
	default:
		return "", fmt.Errorf("unsupported reference type: %s", ref.Type)
	}
}

// ToIFlytekTemplate converts to iFlytek Agent template format
func (vrs *VariableReferenceSystem) ToIFlytekTemplate(ref *VariableReference) (string, error) {
	switch ref.Type {
	case ReferenceTypeNodeOutput:
		return fmt.Sprintf("{{%s}}", ref.OutputName), nil
	case ReferenceTypeLiteral:
		return fmt.Sprintf("%v", ref.Value), nil
	case ReferenceTypeTemplate:
		if ref.Template != "" && strings.Contains(ref.Template, "{{") {
			return ref.Template, nil
		}
		return fmt.Sprintf("{{%s}}", ref.OutputName), nil
	default:
		return "", fmt.Errorf("unsupported reference type: %s", ref.Type)
	}
}

// ToDifyTemplate converts to Dify template format
func (vrs *VariableReferenceSystem) ToDifyTemplate(ref *VariableReference) (string, error) {
	switch ref.Type {
	case ReferenceTypeNodeOutput:
		return fmt.Sprintf("{{#%s.%s#}}", ref.NodeID, ref.OutputName), nil
	case ReferenceTypeLiteral:
		return fmt.Sprintf("%v", ref.Value), nil
	case ReferenceTypeTemplate:
		if ref.Template != "" && strings.Contains(ref.Template, "#") {
			return ref.Template, nil
		}
		return fmt.Sprintf("{{#%s.%s#}}", ref.NodeID, ref.OutputName), nil
	default:
		return "", fmt.Errorf("unsupported reference type: %s", ref.Type)
	}
}

// ReplaceTemplateReferences replaces variable references in template
func (vrs *VariableReferenceSystem) ReplaceTemplateReferences(template string, fromPlatform, toPlatform PlatformType) (string, error) {
	// Parse all references in template
	references, err := vrs.ParseTemplateReference(template)
	if err != nil {
		return "", fmt.Errorf("failed to parse template references: %w", err)
	}

	result := template

	// Replace each reference
	for _, ref := range references {
		var newTemplate string

		switch toPlatform {
		case PlatformIFlytek:
			newTemplate, err = vrs.ToIFlytekTemplate(ref)
		case PlatformDify:
			newTemplate, err = vrs.ToDifyTemplate(ref)
		default:
			newTemplate, err = vrs.ToUnifiedTemplate(ref)
		}

		if err != nil {
			return "", fmt.Errorf("failed to convert template reference: %w", err)
		}

		// Replace reference in original template
		if ref.Template != "" {
			result = strings.ReplaceAll(result, ref.Template, newTemplate)
		}
	}

	return result, nil
}

// ValidateReference validates the validity of variable reference
func (vrs *VariableReferenceSystem) ValidateReference(ref *VariableReference, dsl *UnifiedDSL) error {
	if ref == nil {
		return fmt.Errorf("reference is nil")
	}

	switch ref.Type {
	case ReferenceTypeNodeOutput:
		return vrs.validateNodeOutputReference(ref, dsl)
	case ReferenceTypeLiteral:
		return vrs.validateLiteralReference(ref)
	case ReferenceTypeTemplate:
		return vrs.validateTemplateReference(ref)
	default:
		return fmt.Errorf("unknown reference type: %s", ref.Type)
	}
}

// validateNodeOutputReference validates node output reference
func (vrs *VariableReferenceSystem) validateNodeOutputReference(ref *VariableReference, dsl *UnifiedDSL) error {
	if err := vrs.validateNodeID(ref.NodeID, dsl); err != nil {
		return err
	}

	return vrs.validateOutputName(ref.NodeID, ref.OutputName, dsl)
}

// validateNodeID validates if node ID exists
func (vrs *VariableReferenceSystem) validateNodeID(nodeID string, dsl *UnifiedDSL) error {
	if nodeID == "" {
		return fmt.Errorf("node ID is empty")
	}

	node := dsl.GetNodeByID(nodeID)
	if node == nil {
		return fmt.Errorf("referenced node not found: %s", nodeID)
	}

	return nil
}

// validateOutputName validates if output name exists in node
func (vrs *VariableReferenceSystem) validateOutputName(nodeID, outputName string, dsl *UnifiedDSL) error {
	if outputName == "" {
		return fmt.Errorf("output name is empty")
	}

	node := dsl.GetNodeByID(nodeID)
	for _, output := range node.Outputs {
		if output.Name == outputName {
			return nil
		}
	}

	return fmt.Errorf("referenced output not found: %s.%s", nodeID, outputName)
}

// validateLiteralReference validates literal reference
func (vrs *VariableReferenceSystem) validateLiteralReference(ref *VariableReference) error {
	if ref.Value == nil {
		return fmt.Errorf("literal value is nil")
	}
	return nil
}

// validateTemplateReference validates template reference
func (vrs *VariableReferenceSystem) validateTemplateReference(ref *VariableReference) error {
	if ref.Template == "" {
		return fmt.Errorf("template is empty")
	}
	return nil
}

// SerializeReference serializes variable reference
func (vrs *VariableReferenceSystem) SerializeReference(ref *VariableReference) ([]byte, error) {
	return json.Marshal(ref)
}

// DeserializeReference deserializes variable reference
func (vrs *VariableReferenceSystem) DeserializeReference(data []byte) (*VariableReference, error) {
	var ref VariableReference
	err := json.Unmarshal(data, &ref)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize reference: %w", err)
	}
	return &ref, nil
}

// ConvertVariableReferenceAdvanced performs advanced variable reference format conversion
func ConvertVariableReferenceAdvanced(ref *VariableReference, platform PlatformType) string {
	vrs := NewVariableReferenceSystem()

	switch platform {
	case PlatformIFlytek:
		if template, err := vrs.ToIFlytekTemplate(ref); err == nil {
			return template
		}
	case PlatformDify:
		if template, err := vrs.ToDifyTemplate(ref); err == nil {
			return template
		}
	default:
		if template, err := vrs.ToUnifiedTemplate(ref); err == nil {
			return template
		}
	}

	// Fallback to simple format
	return fmt.Sprintf("{{%s.%s}}", ref.NodeID, ref.OutputName)
}

// ExtractReferencesFromIFlytekNode extracts variable references from iFlytek Agent node
func (vrs *VariableReferenceSystem) ExtractReferencesFromIFlytekNode(nodeData map[string]interface{}) ([]*VariableReference, error) {
	var references []*VariableReference

	// Extract from inputs
	inputRefs := vrs.extractIFlytekInputReferences(nodeData)
	references = append(references, inputRefs...)

	// Extract from node parameters
	paramRefs := vrs.extractIFlytekNodeParamReferences(nodeData)
	references = append(references, paramRefs...)

	return references, nil
}

// extractIFlytekInputReferences extracts references from inputs section
func (vrs *VariableReferenceSystem) extractIFlytekInputReferences(nodeData map[string]interface{}) []*VariableReference {
	var references []*VariableReference

	inputs, ok := nodeData["inputs"].([]interface{})
	if !ok {
		return references
	}

	for _, input := range inputs {
		inputMap, ok := input.(map[string]interface{})
		if !ok {
			continue
		}

		if ref := vrs.parseIFlytekInputReference(inputMap); ref != nil {
			references = append(references, ref)
		}
	}

	return references
}

// parseIFlytekInputReference parses a single input reference
func (vrs *VariableReferenceSystem) parseIFlytekInputReference(inputMap map[string]interface{}) *VariableReference {
	schema, ok := inputMap["schema"].(map[string]interface{})
	if !ok {
		return nil
	}

	value, ok := schema["value"].(map[string]interface{})
	if !ok {
		return nil
	}

	ref, err := vrs.ParseIFlytekReference(value)
	if err != nil {
		return nil
	}

	return ref
}

// extractIFlytekNodeParamReferences extracts references from nodeParam templates
func (vrs *VariableReferenceSystem) extractIFlytekNodeParamReferences(nodeData map[string]interface{}) []*VariableReference {
	var references []*VariableReference

	nodeParam, ok := nodeData["nodeParam"].(map[string]interface{})
	if !ok {
		return references
	}

	// Extract from systemTemplate
	if systemTemplate, ok := nodeParam["systemTemplate"].(string); ok && systemTemplate != "" {
		if templateRefs, err := vrs.ParseTemplateReference(systemTemplate); err == nil {
			references = append(references, templateRefs...)
		}
	}

	// Extract from template
	if template, ok := nodeParam["template"].(string); ok && template != "" {
		if templateRefs, err := vrs.ParseTemplateReference(template); err == nil {
			references = append(references, templateRefs...)
		}
	}

	return references
}

// ExtractReferencesFromDifyNode extracts variable references from Dify node
func (vrs *VariableReferenceSystem) ExtractReferencesFromDifyNode(nodeData map[string]interface{}) ([]*VariableReference, error) {
	var references []*VariableReference

	// Extract from variables
	varRefs := vrs.extractDifyVariableReferences(nodeData)
	references = append(references, varRefs...)

	// Extract from prompt templates
	templateRefs := vrs.extractDifyPromptTemplateReferences(nodeData)
	references = append(references, templateRefs...)

	return references, nil
}

// extractDifyVariableReferences extracts references from variables section
func (vrs *VariableReferenceSystem) extractDifyVariableReferences(nodeData map[string]interface{}) []*VariableReference {
	var references []*VariableReference

	variables, ok := nodeData["variables"].([]interface{})
	if !ok {
		return references
	}

	for _, variable := range variables {
		varMap, ok := variable.(map[string]interface{})
		if !ok {
			continue
		}

		if ref := vrs.parseVariableReference(varMap); ref != nil {
			references = append(references, ref)
		}
	}

	return references
}

// parseVariableReference parses a single variable reference
func (vrs *VariableReferenceSystem) parseVariableReference(varMap map[string]interface{}) *VariableReference {
	valueSelector, ok := varMap["value_selector"].([]interface{})
	if !ok {
		return nil
	}

	selector := vrs.convertToStringSlice(valueSelector)
	if len(selector) < 2 {
		return nil
	}

	valueType := "string"
	if vt, ok := varMap["value_type"].(string); ok {
		valueType = vt
	}

	ref, err := vrs.ParseDifyReference(selector, valueType)
	if err != nil {
		return nil
	}

	return ref
}

// convertToStringSlice converts []interface{} to []string
func (vrs *VariableReferenceSystem) convertToStringSlice(values []interface{}) []string {
	var result []string
	for _, v := range values {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// extractDifyPromptTemplateReferences extracts references from prompt templates
func (vrs *VariableReferenceSystem) extractDifyPromptTemplateReferences(nodeData map[string]interface{}) []*VariableReference {
	var references []*VariableReference

	promptTemplate, ok := nodeData["prompt_template"].([]interface{})
	if !ok {
		return references
	}

	for _, prompt := range promptTemplate {
		promptMap, ok := prompt.(map[string]interface{})
		if !ok {
			continue
		}

		text, ok := promptMap["text"].(string)
		if !ok || text == "" {
			continue
		}

		templateRefs, err := vrs.ParseTemplateReference(text)
		if err == nil {
			references = append(references, templateRefs...)
		}
	}

	return references
}

// RegisterOutputMapping registers an output name mapping for a specific node
// This allows the system to automatically map output names during reference resolution
func (vrs *VariableReferenceSystem) RegisterOutputMapping(nodeID, oldOutputName, newOutputName string) {
	if vrs.OutputMappings == nil {
		vrs.OutputMappings = make(map[string]map[string]string)
	}
	
	if vrs.OutputMappings[nodeID] == nil {
		vrs.OutputMappings[nodeID] = make(map[string]string)
	}
	
	vrs.OutputMappings[nodeID][oldOutputName] = newOutputName
}

// ResolveOutputName resolves an output name using registered mappings
// Returns the mapped name if a mapping exists, otherwise returns the original name
func (vrs *VariableReferenceSystem) ResolveOutputName(nodeID, outputName string) string {
	if vrs.OutputMappings != nil {
		if nodeMapping, exists := vrs.OutputMappings[nodeID]; exists {
			if mappedName, mapped := nodeMapping[outputName]; mapped {
				return mappedName
			}
		}
	}
	return outputName
}
