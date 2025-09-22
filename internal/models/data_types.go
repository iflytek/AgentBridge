// Package models provides core data type definitions and mapping utilities.
package models

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// UnifiedDataType represents standardized data types across different AI platforms.
type UnifiedDataType string

const (
	DataTypeString       UnifiedDataType = "string"         // Text or string data
	DataTypeInteger      UnifiedDataType = "integer"        // Integer numeric values
	DataTypeFloat        UnifiedDataType = "float"          // Floating-point numeric values
	DataTypeNumber       UnifiedDataType = "number"         // Generic numeric values (backward compatibility)
	DataTypeBoolean      UnifiedDataType = "boolean"        // Boolean true/false values
	DataTypeArrayString  UnifiedDataType = "array[string]"  // Array of strings
	DataTypeArrayInteger UnifiedDataType = "array[integer]" // Array of integers
	DataTypeArrayFloat   UnifiedDataType = "array[float]"   // Array of floats
	DataTypeArrayNumber  UnifiedDataType = "array[number]"  // Array of numbers (backward compatibility)
	DataTypeArrayBoolean UnifiedDataType = "array[boolean]" // Array of booleans
	DataTypeArrayObject  UnifiedDataType = "array[object]"  // Array of objects
	DataTypeObject       UnifiedDataType = "object"         // Complex object/map structure
)

// DataTypeMapping defines cross-platform type mapping and alias resolution.
type DataTypeMapping struct {
	UnifiedTypes   []UnifiedDataType          `yaml:"unified_types" json:"unified_types"`
	IFlytekMapping map[UnifiedDataType]string `yaml:"iflytek_mapping" json:"iflytek_mapping"`
	DifyMapping    map[UnifiedDataType]string `yaml:"dify_mapping" json:"dify_mapping"`
	CozeMapping    map[UnifiedDataType]string `yaml:"coze_mapping" json:"coze_mapping"`
	// Alias mappings for backward compatibility and alternative type names
	IFlytekAliases map[string]string `yaml:"iflytek_aliases" json:"iflytek_aliases"`
	DifyAliases    map[string]string `yaml:"dify_aliases" json:"dify_aliases"`
}

// GetDefaultDataTypeMapping returns the standard cross-platform type mapping configuration.
func GetDefaultDataTypeMapping() *DataTypeMapping {
	return &DataTypeMapping{
		UnifiedTypes: []UnifiedDataType{
			DataTypeString, DataTypeInteger, DataTypeFloat, DataTypeNumber, DataTypeBoolean,
			DataTypeArrayString, DataTypeArrayInteger, DataTypeArrayFloat, DataTypeArrayNumber, DataTypeArrayBoolean, DataTypeArrayObject, DataTypeObject,
		},
		IFlytekMapping: map[UnifiedDataType]string{
			DataTypeString:       "string",
			DataTypeInteger:      "integer", // iFlytek integer maps to unified integer
			DataTypeFloat:        "number",  // iFlytek number maps to unified float
			DataTypeNumber:       "integer", // Backward compatibility: generic number -> integer
			DataTypeBoolean:      "boolean",
			DataTypeArrayString:  "array-string",
			DataTypeArrayInteger: "array-integer", // iFlytek supports integer arrays
			DataTypeArrayFloat:   "array-number",  // iFlytek uses number for float arrays
			DataTypeArrayNumber:  "array-number",  // Backward compatibility
			DataTypeArrayBoolean: "array-boolean", // iFlytek supports boolean arrays
			DataTypeArrayObject:  "array-object",
			DataTypeObject:       "object",
		},
		DifyMapping: map[UnifiedDataType]string{
			DataTypeString:       "string",
			DataTypeInteger:      "number", // Dify uses number for both integer and float
			DataTypeFloat:        "number", // Dify uses number for both integer and float
			DataTypeNumber:       "number", // Backward compatibility
			DataTypeBoolean:      "boolean",
			DataTypeArrayString:  "array[string]",
			DataTypeArrayInteger: "array[number]",  // Dify uses number for integers in arrays
			DataTypeArrayFloat:   "array[number]",  // Dify uses number for floats in arrays
			DataTypeArrayNumber:  "array[number]",  // Backward compatibility
			DataTypeArrayBoolean: "array[boolean]", // Dify supports boolean arrays
			DataTypeArrayObject:  "array[object]",
			DataTypeObject:       "object",
		},
		// Coze platform type mapping
		CozeMapping: map[UnifiedDataType]string{
			DataTypeString:       "string",
			DataTypeInteger:      "integer", // Coze supports precise integer type
			DataTypeFloat:        "float",   // Coze supports precise float type
			DataTypeNumber:       "float",   // Backward compatibility: generic number -> float
			DataTypeBoolean:      "boolean",
			DataTypeArrayString:  "array[string]",
			DataTypeArrayInteger: "array[integer]", // Coze supports precise integer arrays
			DataTypeArrayFloat:   "array[float]",   // Coze supports precise float arrays
			DataTypeArrayNumber:  "array[number]",  // Backward compatibility for generic number arrays
			DataTypeArrayBoolean: "array[boolean]", // Coze supports boolean arrays
			DataTypeArrayObject:  "array[object]",
			DataTypeObject:       "object",
		},
		// iFlytek platform type aliases for backward compatibility
		IFlytekAliases: map[string]string{
			"int": "integer", "number": "integer", "bool": "boolean",
			"str": "string", "text": "string", "list": "array-string", "array": "array-string",
		},
		// Dify platform type aliases for backward compatibility
		DifyAliases: map[string]string{
			"integer": "number", "int": "number", "bool": "boolean", "str": "string",
			"text": "string", "list": "array[string]", "array": "array[string]",
			"dict": "object", "map": "object", "float": "number", "double": "number",
		},
	}
}

// ToIFlytekType converts unified type to iFlytek platform-specific type.
func (dtm *DataTypeMapping) ToIFlytekType(unifiedType UnifiedDataType) string {
	if iflytekType, exists := dtm.IFlytekMapping[unifiedType]; exists {
		return iflytekType
	}
	return string(unifiedType)
}

// ToDifyType converts unified type to Dify platform-specific type.
func (dtm *DataTypeMapping) ToDifyType(unifiedType UnifiedDataType) string {
	if difyType, exists := dtm.DifyMapping[unifiedType]; exists {
		return difyType
	}
	return string(unifiedType)
}

// ToCozeType converts unified type to Coze platform-specific type.
func (dtm *DataTypeMapping) ToCozeType(unifiedType UnifiedDataType) string {
	if cozeType, exists := dtm.CozeMapping[unifiedType]; exists {
		return cozeType
	}
	return string(unifiedType)
}

// FromIFlytekType converts iFlytek platform type to unified type with precise type recognition.
func (dtm *DataTypeMapping) FromIFlytekType(iflytekType string) UnifiedDataType {
	// Precise type mapping without aliases for core types
	switch iflytekType {
	case "integer":
		return DataTypeInteger
	case "number":
		return DataTypeFloat
	case "string":
		return DataTypeString
	case "boolean":
		return DataTypeBoolean
	}

	// Check for alias mappings for other type names
	if canonical, exists := dtm.IFlytekAliases[iflytekType]; exists {
		// Recursively resolve aliases
		return dtm.FromIFlytekType(canonical)
	}

	// Perform reverse mapping from platform type to unified type for remaining types
	for unified, iflytek := range dtm.IFlytekMapping {
		if iflytek == iflytekType {
			return unified
		}
	}
	return UnifiedDataType(iflytekType)
}

// FromDifyType converts Dify platform type to unified type with alias resolution.
func (dtm *DataTypeMapping) FromDifyType(difyType string) UnifiedDataType {
	// Check for alias mappings first to resolve type names
	if canonical, exists := dtm.DifyAliases[difyType]; exists {
		difyType = canonical
	}

	// Perform reverse mapping from platform type to unified type
	for unified, dify := range dtm.DifyMapping {
		if dify == difyType {
			return unified
		}
	}
	return UnifiedDataType(difyType)
}

// MapToDifyTypeWithAliases maps any input type to Dify format with comprehensive alias support.
// Attempts multiple resolution strategies before defaulting to string type.
func (dtm *DataTypeMapping) MapToDifyTypeWithAliases(inputType string) string {
	// Try direct unified type mapping
	if difyType, exists := dtm.DifyMapping[UnifiedDataType(inputType)]; exists {
		return difyType
	}

	// Check Dify alias mappings
	if canonical, exists := dtm.DifyAliases[inputType]; exists {
		return canonical
	}

	// Try iFlytek -> unified -> Dify conversion path
	unifiedFromIFlytek := dtm.FromIFlytekType(inputType)
	if string(unifiedFromIFlytek) != inputType {
		return dtm.ToDifyType(unifiedFromIFlytek)
	}

	return "string" // Safe fallback
}

// DataTypeValidator provides validation utilities for data types.
type DataTypeValidator struct {
	mapping *DataTypeMapping
}

// NewDataTypeValidator creates a validator with default mapping configuration.
func NewDataTypeValidator() *DataTypeValidator {
	return &DataTypeValidator{mapping: GetDefaultDataTypeMapping()}
}

// ValidateType checks if the unified data type is supported.
func (v *DataTypeValidator) ValidateType(dataType UnifiedDataType) bool {
	for _, validType := range v.mapping.UnifiedTypes {
		if dataType == validType {
			return true
		}
	}
	return false
}

// ValidateIFlytekType checks if the iFlytek data type is supported.
func (v *DataTypeValidator) ValidateIFlytekType(iflytekType string) bool {
	for _, mappedType := range v.mapping.IFlytekMapping {
		if mappedType == iflytekType {
			return true
		}
	}
	return false
}

// ValidateDifyType checks if the Dify data type is supported.
func (v *DataTypeValidator) ValidateDifyType(difyType string) bool {
	for _, mappedType := range v.mapping.DifyMapping {
		if mappedType == difyType {
			return true
		}
	}
	return false
}

// ConvertValue performs type conversion between unified data types.
func (dtm *DataTypeMapping) ConvertValue(value interface{}, fromType, toType UnifiedDataType) (interface{}, error) {
	if fromType == toType {
		return value, nil
	}

	switch toType {
	case DataTypeString:
		return convertToString(value), nil
	case DataTypeNumber:
		return convertToNumber(value)
	case DataTypeBoolean:
		return convertToBoolean(value), nil
	case DataTypeArrayString:
		return convertToArrayString(value)
	case DataTypeObject:
		return convertToObject(value)
	default:
		return value, nil
	}
}

// Type conversion helper functions

func convertToString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func convertToNumber(value interface{}) (float64, error) {
	if value == nil {
		return 0, nil
	}
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int, int32, int64:
		return float64(v.(int)), nil
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed, nil
		}
		return 0, fmt.Errorf("cannot convert string '%s' to number", v)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to number", value)
	}
}

func convertToBoolean(value interface{}) bool {
	if value == nil {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != "" && v != "false" && v != "0"
	case int, int32, int64:
		return v != 0
	case float32, float64:
		return v != 0.0
	default:
		return true
	}
}

func convertToArrayString(value interface{}) ([]string, error) {
	if value == nil {
		return []string{}, nil
	}
	switch v := value.(type) {
	case []string:
		return v, nil
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = convertToString(item)
		}
		return result, nil
	case string:
		var result []string
		if err := json.Unmarshal([]byte(v), &result); err == nil {
			return result, nil
		}
		return []string{v}, nil
	default:
		return []string{convertToString(value)}, nil
	}
}

func convertToObject(value interface{}) (map[string]interface{}, error) {
	if value == nil {
		return map[string]interface{}{}, nil
	}
	switch v := value.(type) {
	case map[string]interface{}:
		return v, nil
	case string:
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(v), &result); err == nil {
			return result, nil
		}
		return map[string]interface{}{"value": v}, nil
	default:
		return map[string]interface{}{"value": value}, nil
	}
}

// Utility functions for type checking

func IsNumericType(dataType UnifiedDataType) bool {
	return dataType == DataTypeNumber || dataType == DataTypeInteger || dataType == DataTypeFloat
}
func IsStringType(dataType UnifiedDataType) bool { return dataType == DataTypeString }
func IsArrayType(dataType UnifiedDataType) bool {
	return dataType == DataTypeArrayString || dataType == DataTypeArrayInteger ||
		dataType == DataTypeArrayFloat || dataType == DataTypeArrayNumber ||
		dataType == DataTypeArrayBoolean || dataType == DataTypeArrayObject
}
func IsObjectType(dataType UnifiedDataType) bool { return dataType == DataTypeObject }
func IsPrimitiveType(dataType UnifiedDataType) bool {
	return dataType == DataTypeString || dataType == DataTypeNumber ||
		dataType == DataTypeInteger || dataType == DataTypeFloat || dataType == DataTypeBoolean
}

// GetTypeCategory returns the general category of a data type.
func GetTypeCategory(dataType UnifiedDataType) string {
	switch {
	case IsPrimitiveType(dataType):
		return "primitive"
	case IsArrayType(dataType):
		return "array"
	case IsObjectType(dataType):
		return "object"
	default:
		return "unknown"
	}
}

// Platform-specific conversion helpers

var DifyInputTypeMapping = map[string]UnifiedDataType{
	"text-input": DataTypeString,
	"number":     DataTypeNumber,
	"boolean":    DataTypeBoolean,
	"select":     DataTypeString,
	"textarea":   DataTypeString,
}

var IFlytekCustomParameterTypeMapping = map[string]UnifiedDataType{
	"xfyun-file": DataTypeString,
}

// ConvertDifyInputType maps Dify input types to unified types.
func ConvertDifyInputType(difyInputType string) UnifiedDataType {
	if unifiedType, exists := DifyInputTypeMapping[difyInputType]; exists {
		return unifiedType
	}
	return DataTypeString
}

// ConvertIFlytekCustomParameterType maps iFlytek custom parameter types to unified types.
func ConvertIFlytekCustomParameterType(customType string) UnifiedDataType {
	if unifiedType, exists := IFlytekCustomParameterTypeMapping[customType]; exists {
		return unifiedType
	}
	return DataTypeString
}

// GetDifyValueType returns Dify value_type for unified type.
func GetDifyValueType(unifiedType UnifiedDataType) string {
	return GetDefaultDataTypeMapping().ToDifyType(unifiedType)
}

// GetIFlytekSchemaType returns iFlytek schema type for unified type.
func GetIFlytekSchemaType(unifiedType UnifiedDataType) string {
	return GetDefaultDataTypeMapping().ToIFlytekType(unifiedType)
}

// ValidateTypeCompatibility checks if two types are compatible for conversion.
func ValidateTypeCompatibility(sourceType, targetType UnifiedDataType) bool {
	if sourceType == targetType {
		return true
	}

	// Define compatible conversion pairs
	compatiblePairs := map[UnifiedDataType][]UnifiedDataType{
		DataTypeString:  {DataTypeNumber, DataTypeInteger, DataTypeFloat, DataTypeBoolean},
		DataTypeNumber:  {DataTypeString, DataTypeInteger, DataTypeFloat, DataTypeBoolean},
		DataTypeInteger: {DataTypeString, DataTypeNumber, DataTypeFloat, DataTypeBoolean},
		DataTypeFloat:   {DataTypeString, DataTypeNumber, DataTypeInteger, DataTypeBoolean},
		DataTypeBoolean: {DataTypeString, DataTypeNumber, DataTypeInteger, DataTypeFloat},
	}

	if compatibleTypes, exists := compatiblePairs[sourceType]; exists {
		for _, compatibleType := range compatibleTypes {
			if compatibleType == targetType {
				return true
			}
		}
	}

	return false
}

// GetSupportedPlatformTypes returns list of supported platforms.
func GetSupportedPlatformTypes() []PlatformType {
	return []PlatformType{PlatformIFlytek, PlatformDify, PlatformCoze}
}

// GetAllUnifiedTypes returns all available unified data types.
func GetAllUnifiedTypes() []UnifiedDataType {
	return GetDefaultDataTypeMapping().UnifiedTypes
}
