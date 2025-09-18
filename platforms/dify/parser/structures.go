package parser

// DifyDSL represents the root structure of Dify DSL.
type DifyDSL struct {
	App      DifyApp      `yaml:"app" json:"app"`
	Kind     string       `yaml:"kind" json:"kind"`
	Version  string       `yaml:"version" json:"version"`
	Workflow DifyWorkflow `yaml:"workflow" json:"workflow"`
}

// DifyApp contains application information.
type DifyApp struct {
	Name                string `yaml:"name" json:"name"`
	Description         string `yaml:"description" json:"description"`
	Icon                string `yaml:"icon" json:"icon"`
	IconBackground      string `yaml:"icon_background" json:"icon_background"`
	Mode                string `yaml:"mode" json:"mode"`
	UseIconAsAnswerIcon bool   `yaml:"use_icon_as_answer_icon" json:"use_icon_as_answer_icon"`
}

// DifyWorkflow defines workflow structure.
type DifyWorkflow struct {
	ConversationVariables []interface{} `yaml:"conversation_variables" json:"conversation_variables"`
	EnvironmentVariables  []interface{} `yaml:"environment_variables" json:"environment_variables"`
	Features              DifyFeatures  `yaml:"features" json:"features"`
	Graph                 DifyGraph     `yaml:"graph" json:"graph"`
}

// DifyFeatures contains feature configurations.
type DifyFeatures struct {
	FileUpload                    *DifyFileUploadConfig   `yaml:"file_upload,omitempty" json:"file_upload,omitempty"`
	OpeningStatement              string                  `yaml:"opening_statement,omitempty" json:"opening_statement,omitempty"`
	RetrieverResource             *DifyRetrieverConfig    `yaml:"retriever_resource,omitempty" json:"retriever_resource,omitempty"`
	SensitiveWordAvoidance        *DifyFeatureConfig      `yaml:"sensitive_word_avoidance,omitempty" json:"sensitive_word_avoidance,omitempty"`
	SpeechToText                  *DifyFeatureConfig      `yaml:"speech_to_text,omitempty" json:"speech_to_text,omitempty"`
	SuggestedQuestions            []string                `yaml:"suggested_questions,omitempty" json:"suggested_questions,omitempty"`
	SuggestedQuestionsAfterAnswer *DifyFeatureConfig      `yaml:"suggested_questions_after_answer,omitempty" json:"suggested_questions_after_answer,omitempty"`
	TextToSpeech                  *DifyTextToSpeechConfig `yaml:"text_to_speech,omitempty" json:"text_to_speech,omitempty"`
}

// DifyFileUploadConfig contains file upload configuration.
type DifyFileUploadConfig struct {
	Enabled                  bool              `yaml:"enabled" json:"enabled"`
	AllowedFileExtensions    []string          `yaml:"allowed_file_extensions,omitempty" json:"allowed_file_extensions,omitempty"`
	AllowedFileTypes         []string          `yaml:"allowed_file_types,omitempty" json:"allowed_file_types,omitempty"`
	AllowedFileUploadMethods []string          `yaml:"allowed_file_upload_methods,omitempty" json:"allowed_file_upload_methods,omitempty"`
	FileUploadConfig         *DifyUploadLimits `yaml:"fileUploadConfig,omitempty" json:"fileUploadConfig,omitempty"`
	Image                    *DifyImageConfig  `yaml:"image,omitempty" json:"image,omitempty"`
	NumberLimits             int               `yaml:"number_limits,omitempty" json:"number_limits,omitempty"`
}

// DifyUploadLimits contains upload limit configuration.
type DifyUploadLimits struct {
	AudioFileSizeLimit      int `yaml:"audio_file_size_limit" json:"audio_file_size_limit"`
	BatchCountLimit         int `yaml:"batch_count_limit" json:"batch_count_limit"`
	FileSizeLimit           int `yaml:"file_size_limit" json:"file_size_limit"`
	ImageFileSizeLimit      int `yaml:"image_file_size_limit" json:"image_file_size_limit"`
	VideoFileSizeLimit      int `yaml:"video_file_size_limit" json:"video_file_size_limit"`
	WorkflowFileUploadLimit int `yaml:"workflow_file_upload_limit" json:"workflow_file_upload_limit"`
}

// DifyImageConfig contains image configuration.
type DifyImageConfig struct {
	Enabled         bool     `yaml:"enabled" json:"enabled"`
	NumberLimits    int      `yaml:"number_limits" json:"number_limits"`
	TransferMethods []string `yaml:"transfer_methods" json:"transfer_methods"`
}

// DifyRetrieverConfig contains retriever configuration.
type DifyRetrieverConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// DifyFeatureConfig contains general feature configuration.
type DifyFeatureConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// DifyTextToSpeechConfig contains text-to-speech configuration.
type DifyTextToSpeechConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Language string `yaml:"language,omitempty" json:"language,omitempty"`
	Voice    string `yaml:"voice,omitempty" json:"voice,omitempty"`
}

// DifyGraph contains graph structure.
type DifyGraph struct {
	Edges    []DifyEdge    `yaml:"edges" json:"edges"`
	Nodes    []DifyNode    `yaml:"nodes" json:"nodes"`
	Viewport *DifyViewport `yaml:"viewport,omitempty" json:"viewport,omitempty"`
}

// DifyViewport contains viewport configuration.
type DifyViewport struct {
	X    float64 `yaml:"x" json:"x"`
	Y    float64 `yaml:"y" json:"y"`
	Zoom float64 `yaml:"zoom" json:"zoom"`
}

// DifyEdge represents a connection edge.
type DifyEdge struct {
	ID           string        `yaml:"id" json:"id"`
	Source       string        `yaml:"source" json:"source"`
	Target       string        `yaml:"target" json:"target"`
	SourceHandle string        `yaml:"sourceHandle,omitempty" json:"sourceHandle,omitempty"`
	TargetHandle string        `yaml:"targetHandle,omitempty" json:"targetHandle,omitempty"`
	Type         string        `yaml:"type" json:"type"`
	Data         *DifyEdgeData `yaml:"data,omitempty" json:"data,omitempty"`
	Selected     bool          `yaml:"selected,omitempty" json:"selected,omitempty"`
	ZIndex       int           `yaml:"zIndex,omitempty" json:"zIndex,omitempty"`
}

// DifyEdgeData contains edge data.
type DifyEdgeData struct {
	IsInLoop      bool   `yaml:"isInLoop,omitempty" json:"isInLoop,omitempty"`
	IsInIteration bool   `yaml:"isInIteration,omitempty" json:"isInIteration,omitempty"`
	IterationID   string `yaml:"iteration_id,omitempty" json:"iteration_id,omitempty"`
	SourceType    string `yaml:"sourceType,omitempty" json:"sourceType,omitempty"`
	TargetType    string `yaml:"targetType,omitempty" json:"targetType,omitempty"`
}

// DifyNode represents a node.
type DifyNode struct {
	ID               string        `yaml:"id" json:"id"`
	Type             string        `yaml:"type" json:"type"`
	Position         DifyPosition  `yaml:"position" json:"position"`
	PositionAbsolute *DifyPosition `yaml:"positionAbsolute,omitempty" json:"positionAbsolute,omitempty"`
	Width            float64       `yaml:"width,omitempty" json:"width,omitempty"`
	Height           float64       `yaml:"height,omitempty" json:"height,omitempty"`
	Selected         bool          `yaml:"selected,omitempty" json:"selected,omitempty"`
	Draggable        bool          `yaml:"draggable,omitempty" json:"draggable,omitempty"`
	Selectable       bool          `yaml:"selectable,omitempty" json:"selectable,omitempty"`
	SourcePosition   string        `yaml:"sourcePosition,omitempty" json:"sourcePosition,omitempty"`
	TargetPosition   string        `yaml:"targetPosition,omitempty" json:"targetPosition,omitempty"`
	ZIndex           int           `yaml:"zIndex,omitempty" json:"zIndex,omitempty"`
	ParentID         string        `yaml:"parentId,omitempty" json:"parentId,omitempty"`
	Extent           string        `yaml:"extent,omitempty" json:"extent,omitempty"`
	Data             DifyNodeData  `yaml:"data" json:"data"`
}

// DifyPosition contains position coordinates.
type DifyPosition struct {
	X float64 `yaml:"x" json:"x"`
	Y float64 `yaml:"y" json:"y"`
}

// DifyNodeData contains node data.
type DifyNodeData struct {
	Type      string         `yaml:"type" json:"type"`
	Title     string         `yaml:"title" json:"title"`
	Desc      string         `yaml:"desc,omitempty" json:"desc,omitempty"`
	Selected  bool           `yaml:"selected,omitempty" json:"selected,omitempty"`
	Variables []DifyVariable `yaml:"variables,omitempty" json:"variables,omitempty"`
	Outputs   interface{}    `yaml:"outputs,omitempty" json:"outputs,omitempty"` // Supports []DifyOutput and map[string]interface{}

	// LLM node specific fields
	Model          *DifyModel   `yaml:"model,omitempty" json:"model,omitempty"`
	PromptTemplate []DifyPrompt `yaml:"prompt_template,omitempty" json:"prompt_template,omitempty"`
	Context        *DifyContext `yaml:"context,omitempty" json:"context,omitempty"`
	Vision         *DifyVision  `yaml:"vision,omitempty" json:"vision,omitempty"`

	// Code node specific fields
	Code         string `yaml:"code,omitempty" json:"code,omitempty"`
	CodeLanguage string `yaml:"code_language,omitempty" json:"code_language,omitempty"`

	// Condition node specific fields
	Cases []DifyCase `yaml:"cases,omitempty" json:"cases,omitempty"`

	// Classifier node specific fields
	Classes               []DifyClass `yaml:"classes,omitempty" json:"classes,omitempty"`
	Instruction           string      `yaml:"instruction,omitempty" json:"instruction,omitempty"`
	Instructions          string      `yaml:"instructions,omitempty" json:"instructions,omitempty"`
	QueryVariableSelector []string    `yaml:"query_variable_selector,omitempty" json:"query_variable_selector,omitempty"`
	Topics                []string    `yaml:"topics,omitempty" json:"topics,omitempty"`

	// Iteration node specific fields
	ErrorHandleMode   string   `yaml:"error_handle_mode,omitempty" json:"error_handle_mode,omitempty"`
	IsParallel        bool     `yaml:"is_parallel,omitempty" json:"is_parallel,omitempty"`
	IteratorInputType string   `yaml:"iterator_input_type,omitempty" json:"iterator_input_type,omitempty"`
	IteratorSelector  []string `yaml:"iterator_selector,omitempty" json:"iterator_selector,omitempty"`
	OutputSelector    []string `yaml:"output_selector,omitempty" json:"output_selector,omitempty"`
	OutputType        string   `yaml:"output_type,omitempty" json:"output_type,omitempty"`
	ParallelNums      int      `yaml:"parallel_nums,omitempty" json:"parallel_nums,omitempty"`
	StartNodeID       string   `yaml:"start_node_id,omitempty" json:"start_node_id,omitempty"`

	// Iteration-related identifiers
	IsInIteration bool   `yaml:"isInIteration,omitempty" json:"isInIteration,omitempty"`
	IsInLoop      bool   `yaml:"isInLoop,omitempty" json:"isInLoop,omitempty"`
	IterationID   string `yaml:"iteration_id,omitempty" json:"iteration_id,omitempty"`
}

// DifyVariable defines variable structure.
type DifyVariable struct {
	Label         string   `yaml:"label" json:"label"`
	Variable      string   `yaml:"variable" json:"variable"`
	Type          string   `yaml:"type" json:"type"`
	ValueType     string   `yaml:"value_type,omitempty" json:"value_type,omitempty"`
	ValueSelector []string `yaml:"value_selector,omitempty" json:"value_selector,omitempty"`
	Required      bool     `yaml:"required" json:"required"`
	MaxLength     int      `yaml:"max_length,omitempty" json:"max_length,omitempty"`
	Options       []string `yaml:"options,omitempty" json:"options,omitempty"`
}

// DifyOutput defines output structure.
type DifyOutput struct {
	ValueSelector []string `yaml:"value_selector" json:"value_selector"`
	ValueType     string   `yaml:"value_type" json:"value_type"`
	Variable      string   `yaml:"variable" json:"variable"`
}

// DifyModel contains model configuration.
type DifyModel struct {
	Provider         string                 `yaml:"provider" json:"provider"`
	Name             string                 `yaml:"name" json:"name"`
	Mode             string                 `yaml:"mode" json:"mode"`
	CompletionParams map[string]interface{} `yaml:"completion_params,omitempty" json:"completion_params,omitempty"`
}

// DifyPrompt contains prompt configuration.
type DifyPrompt struct {
	ID   string `yaml:"id" json:"id"`
	Role string `yaml:"role" json:"role"`
	Text string `yaml:"text" json:"text"`
}

// DifyContext contains context configuration.
type DifyContext struct {
	Enabled          bool     `yaml:"enabled" json:"enabled"`
	VariableSelector []string `yaml:"variable_selector,omitempty" json:"variable_selector,omitempty"`
}

// DifyVision contains vision configuration.
type DifyVision struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// DifyCase represents a conditional branch.
type DifyCase struct {
	CaseID          string          `yaml:"case_id" json:"case_id"`
	ID              string          `yaml:"id" json:"id"`
	LogicalOperator string          `yaml:"logical_operator" json:"logical_operator"`
	Conditions      []DifyCondition `yaml:"conditions" json:"conditions"`
}

// DifyCondition represents a condition.
type DifyCondition struct {
	ID                 string   `yaml:"id" json:"id"`
	VariableSelector   []string `yaml:"variable_selector" json:"variable_selector"`
	ComparisonOperator string   `yaml:"comparison_operator" json:"comparison_operator"`
	Value              string   `yaml:"value" json:"value"`
	VarType            string   `yaml:"varType" json:"varType"`
}

// DifyClass represents a classification category.
type DifyClass struct {
	ID   string `yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`
}
