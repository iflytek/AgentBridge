package parser

import (
	"ai-agents-transformer/internal/models"
	"fmt"
)

// IFlytekReference represents iFlytek SparkAgent reference structure.
type IFlytekReference struct {
	OriginID   string             `yaml:"originId"`
	ID         string             `yaml:"id"`
	Label      string             `yaml:"label"`
	Type       string             `yaml:"type"`
	Value      string             `yaml:"value"`
	FileType   string             `yaml:"fileType"`
	Children   []IFlytekReference `yaml:"children,omitempty"`
	References []IFlytekReference `yaml:"references,omitempty"`
	ParentNode bool               `yaml:"parentNode,omitempty"`
}

// IFlytekReferenceGroup represents iFlytek SparkAgent reference group.
type IFlytekReferenceGroup struct {
	Label      string             `yaml:"label"`
	Value      string             `yaml:"value"`
	ParentNode bool               `yaml:"parentNode,omitempty"`
	Children   []IFlytekReference `yaml:"children,omitempty"`
}

// ReferenceParser parses references.
type ReferenceParser struct {
	variableRefSystem *models.VariableReferenceSystem
}

func NewReferenceParser(variableRefSystem *models.VariableReferenceSystem) *ReferenceParser {
	return &ReferenceParser{
		variableRefSystem: variableRefSystem,
	}
}

// ParseReferences parses reference list.
func (p *ReferenceParser) ParseReferences(referencesData interface{}) ([]*models.VariableReference, error) {
	if referencesData == nil {
		return []*models.VariableReference{}, nil
	}

	referencesList, ok := referencesData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("references data is not a list")
	}

	var allReferences []*models.VariableReference

	for i, refGroupData := range referencesList {
		refGroupMap, ok := refGroupData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("reference group at index %d is not a map", i)
		}

		references, err := p.parseReferenceGroup(refGroupMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse reference group at index %d: %w", i, err)
		}

		allReferences = append(allReferences, references...)
	}

	return allReferences, nil
}

// parseReferenceGroup parses reference group.
func (p *ReferenceParser) parseReferenceGroup(refGroupMap map[string]interface{}) ([]*models.VariableReference, error) {
	children, ok := refGroupMap["children"].([]interface{})
	if !ok {
		return []*models.VariableReference{}, nil
	}

	return p.parseChildrenReferences(children)
}

// parseChildrenReferences parses references from children array
func (p *ReferenceParser) parseChildrenReferences(children []interface{}) ([]*models.VariableReference, error) {
	var references []*models.VariableReference

	for _, childData := range children {
		childRefs, err := p.parseChildReferences(childData)
		if err != nil {
			return nil, err
		}
		references = append(references, childRefs...)
	}

	return references, nil
}

// parseChildReferences parses references from a single child
func (p *ReferenceParser) parseChildReferences(childData interface{}) ([]*models.VariableReference, error) {
	childMap, ok := childData.(map[string]interface{})
	if !ok {
		return []*models.VariableReference{}, nil
	}

	childRefs, ok := childMap["references"].([]interface{})
	if !ok {
		return []*models.VariableReference{}, nil
	}

	return p.parseReferenceArray(childRefs)
}

// parseReferenceArray parses an array of reference data
func (p *ReferenceParser) parseReferenceArray(refs []interface{}) ([]*models.VariableReference, error) {
	var references []*models.VariableReference

	for _, refData := range refs {
		refMap, ok := refData.(map[string]interface{})
		if !ok {
			continue
		}

		ref, err := p.parseIFlytekReference(refMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse reference: %w", err)
		}
		if ref != nil {
			references = append(references, ref)
		}
	}

	return references, nil
}

// parseIFlytekReference parses a single iFlytek SparkAgent reference.
func (p *ReferenceParser) parseIFlytekReference(refMap map[string]interface{}) (*models.VariableReference, error) {
	iflytekRef := &IFlytekReference{}

	// Parse basic fields
	if originId, ok := refMap["originId"].(string); ok {
		iflytekRef.OriginID = originId
	}

	if id, ok := refMap["id"].(string); ok {
		iflytekRef.ID = id
	}

	if label, ok := refMap["label"].(string); ok {
		iflytekRef.Label = label
	}

	if refType, ok := refMap["type"].(string); ok {
		iflytekRef.Type = refType
	}

	if value, ok := refMap["value"].(string); ok {
		iflytekRef.Value = value
	}

	if fileType, ok := refMap["fileType"].(string); ok {
		iflytekRef.FileType = fileType
	}

	// Convert to unified reference format
	return p.convertToUnifiedReference(iflytekRef)
}

// convertToUnifiedReference converts to unified reference format.
func (p *ReferenceParser) convertToUnifiedReference(iflytekRef *IFlytekReference) (*models.VariableReference, error) {
	if iflytekRef.OriginID == "" || iflytekRef.Value == "" {
		return nil, nil // Skip invalid references
	}

	// Create unified variable reference
	unifiedRef := &models.VariableReference{
		Type:       models.ReferenceTypeNodeOutput,
		NodeID:     iflytekRef.OriginID,
		OutputName: iflytekRef.Value,
		DataType:   p.convertDataType(iflytekRef.Type),
	}

	return unifiedRef, nil
}

// convertDataType converts data type.
func (p *ReferenceParser) convertDataType(iflytekType string) models.UnifiedDataType {
	// Use unified mapping system, supports alias processing
	mapping := models.GetDefaultDataTypeMapping()
	return mapping.FromIFlytekType(iflytekType)
}

// ParseInputReferences parses input references.
func (p *ReferenceParser) ParseInputReferences(inputs []interface{}) (map[string]*models.VariableReference, error) {
	references := make(map[string]*models.VariableReference)

	for _, inputData := range inputs {
		if ref := p.parseInputReference(inputData); ref != nil {
			references[ref.inputName] = ref.reference
		}
	}

	return references, nil
}

// inputReferenceResult holds the result of parsing an input reference
type inputReferenceResult struct {
	inputName string
	reference *models.VariableReference
}

// parseInputReference parses a single input reference
func (p *ReferenceParser) parseInputReference(inputData interface{}) *inputReferenceResult {
	inputMap, ok := inputData.(map[string]interface{})
	if !ok {
		return nil
	}

	inputName, ok := inputMap["name"].(string)
	if !ok {
		return nil
	}

	reference := p.extractReferenceFromSchema(inputMap)
	if reference == nil {
		return nil
	}

	return &inputReferenceResult{
		inputName: inputName,
		reference: reference,
	}
}

// extractReferenceFromSchema extracts reference from input schema
func (p *ReferenceParser) extractReferenceFromSchema(inputMap map[string]interface{}) *models.VariableReference {
	schema, ok := inputMap["schema"].(map[string]interface{})
	if !ok {
		return nil
	}

	value, ok := schema["value"].(map[string]interface{})
	if !ok {
		return nil
	}

	return p.parseReferenceValue(value, schema)
}

// parseReferenceValue parses reference value and content
func (p *ReferenceParser) parseReferenceValue(value, schema map[string]interface{}) *models.VariableReference {
	valueType, ok := value["type"].(string)
	if !ok || valueType != "ref" {
		return nil
	}

	content, ok := value["content"].(map[string]interface{})
	if !ok {
		return nil
	}

	return p.createReferenceFromContent(content, schema)
}

// createReferenceFromContent creates reference from content and schema
func (p *ReferenceParser) createReferenceFromContent(content, schema map[string]interface{}) *models.VariableReference {
	ref := &models.VariableReference{
		Type:       models.ReferenceTypeNodeOutput,
		NodeID:     p.extractString(content, "nodeId"),
		OutputName: p.extractString(content, "name"),
		DataType:   p.extractDataType(schema),
	}

	return ref
}

// extractString safely extracts string value from map
func (p *ReferenceParser) extractString(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}

// extractDataType extracts and converts data type from schema
func (p *ReferenceParser) extractDataType(schema map[string]interface{}) models.UnifiedDataType {
	if schemaType, ok := schema["type"].(string); ok {
		return p.convertDataType(schemaType)
	}
	return models.DataTypeString
}
