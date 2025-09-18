package models

import (
	"time"
)

// UnifiedDSL represents the root structure of the unified DSL standard
type UnifiedDSL struct {
	Version          string           `yaml:"version" json:"version"`
	Metadata         Metadata         `yaml:"metadata" json:"metadata"`
	PlatformMetadata PlatformMetadata `yaml:"platform_metadata" json:"platform_metadata"`
	Workflow         Workflow         `yaml:"workflow" json:"workflow"`
}

// Metadata contains common metadata information
type Metadata struct {
	Name        string    `yaml:"name" json:"name"`
	Description string    `yaml:"description" json:"description"`
	CreatedAt   time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at" json:"updated_at"`
	UIConfig    *UIConfig `yaml:"ui_config,omitempty" json:"ui_config,omitempty"`
}

// UIConfig contains user interface configuration
type UIConfig struct {
	OpeningStatement   string   `yaml:"opening_statement,omitempty" json:"opening_statement,omitempty"`
	SuggestedQuestions []string `yaml:"suggested_questions,omitempty" json:"suggested_questions,omitempty"`
	Icon               string   `yaml:"icon,omitempty" json:"icon,omitempty"`
	IconBackground     string   `yaml:"icon_background,omitempty" json:"icon_background,omitempty"`
}

// PlatformMetadata contains platform-specific metadata
type PlatformMetadata struct {
	IFlytek *IFlytekMetadata `yaml:"iflytek,omitempty" json:"iflytek,omitempty"`
	Dify    *DifyMetadata    `yaml:"dify,omitempty" json:"dify,omitempty"`
	Coze    *CozeMetadata    `yaml:"coze,omitempty" json:"coze,omitempty"`
}

// IFlytekMetadata contains iFlytek platform specific metadata
type IFlytekMetadata struct {
	AvatarIcon     string `yaml:"avatar_icon" json:"avatar_icon"`
	AvatarColor    string `yaml:"avatar_color" json:"avatar_color"`
	AdvancedConfig string `yaml:"advanced_config" json:"advanced_config"`
	DSLVersion     string `yaml:"dsl_version" json:"dsl_version"`
}

// DifyMetadata contains Dify platform specific metadata
type DifyMetadata struct {
	Icon                string `yaml:"icon" json:"icon"`
	IconBackground      string `yaml:"icon_background" json:"icon_background"`
	Mode                string `yaml:"mode" json:"mode"`
	UseIconAsAnswerIcon bool   `yaml:"use_icon_as_answer_icon" json:"use_icon_as_answer_icon"`
	Kind                string `yaml:"kind" json:"kind"`
	AppVersion          string `yaml:"app_version" json:"app_version"`
}

// CozeMetadata contains Coze platform specific metadata
type CozeMetadata struct {
	WorkflowID   string `yaml:"workflow_id" json:"workflow_id"`
	Version      string `yaml:"version" json:"version"`
	CreateTime   int64  `yaml:"create_time" json:"create_time"`
	UpdateTime   int64  `yaml:"update_time" json:"update_time"`
	ContentType  string `yaml:"content_type" json:"content_type"`
	CreatorID    string `yaml:"creator_id" json:"creator_id"`
	SpaceID      string `yaml:"space_id" json:"space_id"`
	Mode         string `yaml:"mode" json:"mode"`
	ExportFormat string `yaml:"export_format" json:"export_format"`
}

// Workflow defines the workflow structure
type Workflow struct {
	Nodes     []Node     `yaml:"nodes" json:"nodes"`
	Edges     []Edge     `yaml:"edges" json:"edges"`
	Variables []Variable `yaml:"variables" json:"variables"`
	Features  *Features  `yaml:"features,omitempty" json:"features,omitempty"`
}

// Features contains feature configuration
type Features struct {
	FileUpload         *FileUploadConfig `yaml:"file_upload,omitempty" json:"file_upload,omitempty"`
	OpeningStatement   string            `yaml:"opening_statement,omitempty" json:"opening_statement,omitempty"`
	SuggestedQuestions []string          `yaml:"suggested_questions,omitempty" json:"suggested_questions,omitempty"`
	SpeechToText       *SpeechConfig     `yaml:"speech_to_text,omitempty" json:"speech_to_text,omitempty"`
	TextToSpeech       *SpeechConfig     `yaml:"text_to_speech,omitempty" json:"text_to_speech,omitempty"`
}

// FileUploadConfig contains file upload configuration
type FileUploadConfig struct {
	Enabled                 bool     `yaml:"enabled" json:"enabled"`
	AllowedFileTypes        []string `yaml:"allowed_file_types" json:"allowed_file_types"`
	AllowedFileExtensions   []string `yaml:"allowed_file_extensions" json:"allowed_file_extensions"`
	AllowedUploadMethods    []string `yaml:"allowed_upload_methods" json:"allowed_upload_methods"`
	FileSizeLimit           int      `yaml:"file_size_limit" json:"file_size_limit"`
	ImageFileSizeLimit      int      `yaml:"image_file_size_limit" json:"image_file_size_limit"`
	AudioFileSizeLimit      int      `yaml:"audio_file_size_limit" json:"audio_file_size_limit"`
	VideoFileSizeLimit      int      `yaml:"video_file_size_limit" json:"video_file_size_limit"`
	WorkflowFileUploadLimit int      `yaml:"workflow_file_upload_limit" json:"workflow_file_upload_limit"`
	BatchCountLimit         int      `yaml:"batch_count_limit" json:"batch_count_limit"`
	NumberLimits            int      `yaml:"number_limits" json:"number_limits"`
}

// SpeechConfig contains speech configuration
type SpeechConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Language string `yaml:"language,omitempty" json:"language,omitempty"`
	Voice    string `yaml:"voice,omitempty" json:"voice,omitempty"`
}

// Variable defines global variable structure
type Variable struct {
	Name        string       `yaml:"name" json:"name"`
	Label       string       `yaml:"label" json:"label"`
	Type        string       `yaml:"type" json:"type"`
	Required    bool         `yaml:"required" json:"required"`
	Default     interface{}  `yaml:"default,omitempty" json:"default,omitempty"`
	Description string       `yaml:"description,omitempty" json:"description,omitempty"`
	Constraints *Constraints `yaml:"constraints,omitempty" json:"constraints,omitempty"`

	// Extended fields to support platform-specific information
	ID                  string        `yaml:"id,omitempty" json:"id,omitempty"`
	Properties          []interface{} `yaml:"properties,omitempty" json:"properties,omitempty"`
	CustomParameterType string        `yaml:"custom_parameter_type,omitempty" json:"custom_parameter_type,omitempty"`
	DeleteDisabled      bool          `yaml:"delete_disabled,omitempty" json:"delete_disabled,omitempty"`
	NameErrMsg          string        `yaml:"name_err_msg,omitempty" json:"name_err_msg,omitempty"`
}

// Constraints defines variable constraints
type Constraints struct {
	MaxLength int           `yaml:"max_length,omitempty" json:"max_length,omitempty"`
	MinLength int           `yaml:"min_length,omitempty" json:"min_length,omitempty"`
	Options   []interface{} `yaml:"options,omitempty" json:"options,omitempty"`
	Pattern   string        `yaml:"pattern,omitempty" json:"pattern,omitempty"`
}

// NodeType represents node type enumeration
type NodeType string

const (
	NodeTypeStart      NodeType = "start"      // Start node
	NodeTypeEnd        NodeType = "end"        // End node
	NodeTypeLLM        NodeType = "llm"        // Large language model node
	NodeTypeCode       NodeType = "code"       // Code execution node
	NodeTypeCondition  NodeType = "condition"  // Conditional branch node
	NodeTypeClassifier NodeType = "classifier" // Classification decision node
	NodeTypeIteration  NodeType = "iteration"  // Iteration node
)

// PlatformType represents platform type enumeration
type PlatformType string

const (
	PlatformIFlytek PlatformType = "iflytek" // iFlytek platform
	PlatformDify    PlatformType = "dify"    // Dify platform
	PlatformCoze    PlatformType = "coze"    // Coze platform
)

// Node represents unified node structure
type Node struct {
	ID             string         `yaml:"id" json:"id"`
	Type           NodeType       `yaml:"type" json:"type"`
	Title          string         `yaml:"title" json:"title"`
	Description    string         `yaml:"description,omitempty" json:"description,omitempty"`
	Position       Position       `yaml:"position" json:"position"`
	Size           Size           `yaml:"size" json:"size"`
	Inputs         []Input        `yaml:"inputs" json:"inputs"`
	Outputs        []Output       `yaml:"outputs" json:"outputs"`
	Config         NodeConfig     `yaml:"config" json:"config"`
	PlatformConfig PlatformConfig `yaml:"platform_config" json:"platform_config"`
}

// Position represents node position coordinates
type Position struct {
	X float64 `yaml:"x" json:"x"`
	Y float64 `yaml:"y" json:"y"`
}

// Size represents node dimensions
type Size struct {
	Width  float64 `yaml:"width" json:"width"`
	Height float64 `yaml:"height" json:"height"`
}

// Input defines node input specification
type Input struct {
	Name        string             `yaml:"name" json:"name"`
	Label       string             `yaml:"label,omitempty" json:"label,omitempty"`
	Type        UnifiedDataType    `yaml:"type" json:"type"`
	Required    bool               `yaml:"required" json:"required"`
	Default     interface{}        `yaml:"default,omitempty" json:"default,omitempty"`
	Description string             `yaml:"description,omitempty" json:"description,omitempty"`
	Reference   *VariableReference `yaml:"reference,omitempty" json:"reference,omitempty"`
	Constraints *Constraints       `yaml:"constraints,omitempty" json:"constraints,omitempty"`
}

// Output defines node output specification
type Output struct {
	Name        string          `yaml:"name" json:"name"`
	Label       string          `yaml:"label,omitempty" json:"label,omitempty"`
	Type        UnifiedDataType `yaml:"type" json:"type"`
	Required    bool            `yaml:"required" json:"required"`
	Description string          `yaml:"description,omitempty" json:"description,omitempty"`
	Default     interface{}     `yaml:"default,omitempty" json:"default,omitempty"`
}

// NodeConfig interface for node configuration (implemented by specific node types)
type NodeConfig interface {
	GetNodeType() NodeType
}

// PlatformConfig contains platform-specific configuration
type PlatformConfig struct {
	IFlytek map[string]interface{} `yaml:"iflytek,omitempty" json:"iflytek,omitempty"`
	Dify    map[string]interface{} `yaml:"dify,omitempty" json:"dify,omitempty"`
	Coze    map[string]interface{} `yaml:"coze,omitempty" json:"coze,omitempty"`
}

// VariableReference represents variable reference
type VariableReference struct {
	Type       ReferenceType   `yaml:"type" json:"type"`
	NodeID     string          `yaml:"node_id,omitempty" json:"node_id,omitempty"`
	OutputName string          `yaml:"output_name,omitempty" json:"output_name,omitempty"`
	DataType   UnifiedDataType `yaml:"data_type" json:"data_type"`
	Value      interface{}     `yaml:"value,omitempty" json:"value,omitempty"`
	Template   string          `yaml:"template,omitempty" json:"template,omitempty"`
}

// ReferenceType represents reference type enumeration
type ReferenceType string

const (
	ReferenceTypeNodeOutput ReferenceType = "node_output" // Node output reference
	ReferenceTypeLiteral    ReferenceType = "literal"     // Literal value
	ReferenceTypeTemplate   ReferenceType = "template"    // Template
)

// Edge represents connection relationship
type Edge struct {
	ID             string         `yaml:"id" json:"id"`
	Source         string         `yaml:"source" json:"source"`
	Target         string         `yaml:"target" json:"target"`
	SourceHandle   string         `yaml:"source_handle,omitempty" json:"source_handle,omitempty"`
	TargetHandle   string         `yaml:"target_handle,omitempty" json:"target_handle,omitempty"`
	Type           EdgeType       `yaml:"type" json:"type"`
	Condition      string         `yaml:"condition,omitempty" json:"condition,omitempty"`
	PlatformConfig PlatformConfig `yaml:"platform_config" json:"platform_config"`
}

// EdgeType represents edge type enumeration
type EdgeType string

const (
	EdgeTypeDefault     EdgeType = "default"     // Default connection
	EdgeTypeConditional EdgeType = "conditional" // Conditional connection
)

// StartConfig defines start node configuration
type StartConfig struct {
	Variables     []Variable `yaml:"variables" json:"variables"`
	IsInIteration bool       `yaml:"is_in_iteration,omitempty" json:"is_in_iteration,omitempty"`
	ParentID      string     `yaml:"parent_id,omitempty" json:"parent_id,omitempty"`
}

func (c StartConfig) GetNodeType() NodeType {
	return NodeTypeStart
}

// EndConfig defines end node configuration
type EndConfig struct {
	OutputMode   string      `yaml:"output_mode" json:"output_mode"` // template/variables
	Template     string      `yaml:"template,omitempty" json:"template,omitempty"`
	StreamOutput bool        `yaml:"stream_output" json:"stream_output"`
	Outputs      []EndOutput `yaml:"outputs,omitempty" json:"outputs,omitempty"`
}

func (c EndConfig) GetNodeType() NodeType {
	return NodeTypeEnd
}

// EndOutput defines end node output configuration
type EndOutput struct {
	Variable      string             `yaml:"variable" json:"variable"`
	ValueSelector []string           `yaml:"value_selector" json:"value_selector"`
	ValueType     UnifiedDataType    `yaml:"value_type" json:"value_type"`
	Reference     *VariableReference `yaml:"reference,omitempty" json:"reference,omitempty"`
}

// LLMConfig defines large language model node configuration
type LLMConfig struct {
	Model         ModelConfig     `yaml:"model" json:"model"`
	Parameters    ModelParameters `yaml:"parameters" json:"parameters"`
	Prompt        PromptConfig    `yaml:"prompt" json:"prompt"`
	Context       *ContextConfig  `yaml:"context,omitempty" json:"context,omitempty"`
	Vision        *VisionConfig   `yaml:"vision,omitempty" json:"vision,omitempty"`
	IsInIteration bool            `yaml:"is_in_iteration,omitempty" json:"is_in_iteration,omitempty"`
	IterationID   string          `yaml:"iteration_id,omitempty" json:"iteration_id,omitempty"`
}

func (c LLMConfig) GetNodeType() NodeType {
	return NodeTypeLLM
}

// ModelConfig defines model configuration
type ModelConfig struct {
	Provider string `yaml:"provider" json:"provider"`
	Name     string `yaml:"name" json:"name"`
	Mode     string `yaml:"mode" json:"mode"` // chat/completion
}

// ModelParameters defines model parameters
type ModelParameters struct {
	Temperature    float64 `yaml:"temperature" json:"temperature"`
	MaxTokens      int     `yaml:"max_tokens" json:"max_tokens"`
	TopK           int     `yaml:"top_k,omitempty" json:"top_k,omitempty"`
	TopP           float64 `yaml:"top_p,omitempty" json:"top_p,omitempty"`
	ResponseFormat int     `yaml:"response_format,omitempty" json:"response_format,omitempty"` // 0=text, 1=markdown, 2=json
}

// PromptConfig defines prompt configuration
type PromptConfig struct {
	SystemTemplate string    `yaml:"system_template,omitempty" json:"system_template,omitempty"`
	UserTemplate   string    `yaml:"user_template,omitempty" json:"user_template,omitempty"`
	Messages       []Message `yaml:"messages,omitempty" json:"messages,omitempty"`
}

// Message represents message structure
type Message struct {
	Role    string `yaml:"role" json:"role"`
	Content string `yaml:"content" json:"content"`
}

// ContextConfig defines context configuration
type ContextConfig struct {
	Enabled          bool     `yaml:"enabled" json:"enabled"`
	VariableSelector []string `yaml:"variable_selector,omitempty" json:"variable_selector,omitempty"`
}

// VisionConfig defines vision configuration
type VisionConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// MemoryConfig defines memory configuration
type MemoryConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
	Window  int  `yaml:"window,omitempty" json:"window,omitempty"` // Memory window size
}

// CodeConfig defines code node configuration
type CodeConfig struct {
	Language      string   `yaml:"language" json:"language"` // python3/javascript
	Code          string   `yaml:"code" json:"code"`
	Dependencies  []string `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	IsInIteration bool     `yaml:"is_in_iteration,omitempty" json:"is_in_iteration,omitempty"`
	IterationID   string   `yaml:"iteration_id,omitempty" json:"iteration_id,omitempty"`
}

func (c CodeConfig) GetNodeType() NodeType {
	return NodeTypeCode
}

// ConditionConfig defines condition branch node configuration
type ConditionConfig struct {
	Cases         []ConditionCase `yaml:"cases" json:"cases"`
	DefaultCase   string          `yaml:"default_case,omitempty" json:"default_case,omitempty"`
	IsInIteration bool            `yaml:"is_in_iteration,omitempty" json:"is_in_iteration,omitempty"`
	IterationID   string          `yaml:"iteration_id,omitempty" json:"iteration_id,omitempty"`
}

func (c ConditionConfig) GetNodeType() NodeType {
	return NodeTypeCondition
}

// ConditionCase represents condition branch
type ConditionCase struct {
	CaseID          string      `yaml:"case_id" json:"case_id"`
	Conditions      []Condition `yaml:"conditions" json:"conditions"`
	LogicalOperator string      `yaml:"logical_operator" json:"logical_operator"` // and/or
	Level           int         `yaml:"level,omitempty" json:"level,omitempty"`   // Branch level, 999 for default branch
}

// Condition defines condition specification
type Condition struct {
	VariableSelector   []string        `yaml:"variable_selector" json:"variable_selector"`
	ComparisonOperator string          `yaml:"comparison_operator" json:"comparison_operator"`
	Value              interface{}     `yaml:"value" json:"value"`
	VarType            UnifiedDataType `yaml:"var_type" json:"var_type"`
}

// ClassifierConfig defines classifier decision node configuration
type ClassifierConfig struct {
	Model         ModelConfig       `yaml:"model" json:"model"`
	Parameters    ModelParameters   `yaml:"parameters" json:"parameters"` // Use unified model parameter structure
	Classes       []ClassifierClass `yaml:"classes" json:"classes"`
	QueryVariable string            `yaml:"query_variable" json:"query_variable"`
	Instructions  string            `yaml:"instructions,omitempty" json:"instructions,omitempty"`
	IsInIteration bool              `yaml:"is_in_iteration,omitempty" json:"is_in_iteration,omitempty"`
	IterationID   string            `yaml:"iteration_id,omitempty" json:"iteration_id,omitempty"`
}

func (c ClassifierConfig) GetNodeType() NodeType {
	return NodeTypeClassifier
}

// ClassifierClass represents classification category
type ClassifierClass struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	IsDefault   bool   `yaml:"is_default,omitempty" json:"is_default,omitempty"` // Indicates if this is the default intent
}

// IterationConfig defines iteration node configuration
type IterationConfig struct {
	Iterator       IteratorConfig       `yaml:"iterator" json:"iterator"`
	Execution      ExecutionConfig      `yaml:"execution" json:"execution"`
	SubWorkflow    SubWorkflowConfig    `yaml:"sub_workflow" json:"sub_workflow"`
	OutputSelector OutputSelectorConfig `yaml:"output_selector" json:"output_selector"`
	OutputType     string               `yaml:"output_type" json:"output_type"`
}

func (c IterationConfig) GetNodeType() NodeType {
	return NodeTypeIteration
}

// IteratorConfig defines iterator configuration
type IteratorConfig struct {
	InputType    string `yaml:"input_type" json:"input_type"`
	SourceNode   string `yaml:"source_node" json:"source_node"`
	SourceOutput string `yaml:"source_output" json:"source_output"`
}

// ExecutionConfig defines execution configuration
type ExecutionConfig struct {
	IsParallel      bool   `yaml:"is_parallel" json:"is_parallel"`
	ParallelNums    int    `yaml:"parallel_nums" json:"parallel_nums"`
	ErrorHandleMode string `yaml:"error_handle_mode" json:"error_handle_mode"`
}

// SubWorkflowConfig defines sub-workflow configuration
type SubWorkflowConfig struct {
	Nodes       []Node `yaml:"nodes" json:"nodes"`
	Edges       []Edge `yaml:"edges" json:"edges"`
	StartNodeID string `yaml:"start_node_id" json:"start_node_id"`
	EndNodeID   string `yaml:"end_node_id,omitempty" json:"end_node_id,omitempty"`
}

// OutputSelectorConfig defines output selector configuration
type OutputSelectorConfig struct {
	NodeID     string `yaml:"node_id" json:"node_id"`
	OutputName string `yaml:"output_name" json:"output_name"`
}

func NewUnifiedDSL() *UnifiedDSL {
	return &UnifiedDSL{
		Version: "1.0.0",
		Metadata: Metadata{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		PlatformMetadata: PlatformMetadata{},
		Workflow: Workflow{
			Nodes:     make([]Node, 0),
			Edges:     make([]Edge, 0),
			Variables: make([]Variable, 0),
		},
	}
}

func NewNode(id string, nodeType NodeType, title string) *Node {
	return &Node{
		ID:       id,
		Type:     nodeType,
		Title:    title,
		Position: Position{X: 0, Y: 0},
		Size:     Size{Width: 244, Height: 118},
		Inputs:   make([]Input, 0),
		Outputs:  make([]Output, 0),
		PlatformConfig: PlatformConfig{
			IFlytek: make(map[string]interface{}),
			Dify:    make(map[string]interface{}),
		},
	}
}

func NewEdge(id, source, target string) *Edge {
	return &Edge{
		ID:     id,
		Source: source,
		Target: target,
		Type:   EdgeTypeDefault,
		PlatformConfig: PlatformConfig{
			IFlytek: make(map[string]interface{}),
			Dify:    make(map[string]interface{}),
		},
	}
}

// IsValidNodeType checks if the node type is valid
func IsValidNodeType(nodeType NodeType) bool {
	validTypes := []NodeType{
		NodeTypeStart,
		NodeTypeEnd,
		NodeTypeLLM,
		NodeTypeCode,
		NodeTypeCondition,
		NodeTypeClassifier,
		NodeTypeIteration,
	}

	for _, validType := range validTypes {
		if nodeType == validType {
			return true
		}
	}
	return false
}

// IsValidPlatformType checks if the platform type is valid
func IsValidPlatformType(platform PlatformType) bool {
	return platform == PlatformIFlytek || platform == PlatformDify || platform == PlatformCoze
}

// GetNodeTypeMapping returns node type mapping
func GetNodeTypeMapping() map[PlatformType]map[NodeType]string {
	return map[PlatformType]map[NodeType]string{
		PlatformIFlytek: {
			NodeTypeStart:      "start_node",
			NodeTypeEnd:        "end_node",
			NodeTypeLLM:        "llm_node",
			NodeTypeCode:       "code_node",
			NodeTypeCondition:  "condition_node",
			NodeTypeClassifier: "classifier_node",
			NodeTypeIteration:  "iteration_node",
		},
		PlatformDify: {
			NodeTypeStart:      "start",
			NodeTypeEnd:        "end",
			NodeTypeLLM:        "llm",
			NodeTypeCode:       "code",
			NodeTypeCondition:  "if-else",
			NodeTypeClassifier: "question-classifier",
			NodeTypeIteration:  "iteration",
		},
		PlatformCoze: {
			NodeTypeStart:      "1",
			NodeTypeEnd:        "2",
			NodeTypeLLM:        "3", // LLM node type identifier
			NodeTypeCode:       "4", // Code execution node type identifier
			NodeTypeCondition:  "5", // Conditional branch node type identifier
			NodeTypeClassifier: "6", // Classifier node type identifier
			NodeTypeIteration:  "7", // Iteration node type identifier
		},
	}
}

// GetReverseNodeTypeMapping returns reverse node type mapping
func GetReverseNodeTypeMapping() map[PlatformType]map[string]NodeType {
	mapping := GetNodeTypeMapping()
	reverse := make(map[PlatformType]map[string]NodeType)

	for platform, typeMap := range mapping {
		reverse[platform] = make(map[string]NodeType)
		for nodeType, platformType := range typeMap {
			reverse[platform][platformType] = nodeType
		}
	}

	return reverse
}

// AddNode adds a node to the workflow
func (dsl *UnifiedDSL) AddNode(node Node) {
	dsl.Workflow.Nodes = append(dsl.Workflow.Nodes, node)
}

// AddEdge adds an edge to the workflow
func (dsl *UnifiedDSL) AddEdge(edge Edge) {
	dsl.Workflow.Edges = append(dsl.Workflow.Edges, edge)
}

// GetNodeByID retrieves a node by its ID
func (dsl *UnifiedDSL) GetNodeByID(id string) *Node {
	for i := range dsl.Workflow.Nodes {
		if dsl.Workflow.Nodes[i].ID == id {
			return &dsl.Workflow.Nodes[i]
		}
	}
	return nil
}

// GetEdgesBySource retrieves edges by source node ID
func (dsl *UnifiedDSL) GetEdgesBySource(sourceID string) []Edge {
	var edges []Edge
	for _, edge := range dsl.Workflow.Edges {
		if edge.Source == sourceID {
			edges = append(edges, edge)
		}
	}
	return edges
}

// GetEdgesByTarget retrieves edges by target node ID
func (dsl *UnifiedDSL) GetEdgesByTarget(targetID string) []Edge {
	var edges []Edge
	for _, edge := range dsl.Workflow.Edges {
		if edge.Target == targetID {
			edges = append(edges, edge)
		}
	}
	return edges
}

// UpdateTimestamp updates the timestamp
func (dsl *UnifiedDSL) UpdateTimestamp() {
	dsl.Metadata.UpdatedAt = time.Now()
}
