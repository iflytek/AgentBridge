package generator

// DifyRootStructure represents the Dify root structure
type DifyRootStructure struct {
	App          DifyApp          `yaml:"app"`
	Dependencies []DifyDependency `yaml:"dependencies"`
	Kind         string           `yaml:"kind"`
	Version      string           `yaml:"version"`
	Workflow     DifyWorkflow     `yaml:"workflow"`
}

// DifyApp represents Dify application metadata
type DifyApp struct {
	Name                string `yaml:"name"`
	Description         string `yaml:"description"`
	Icon                string `yaml:"icon"`
	IconBackground      string `yaml:"icon_background"`
	Mode                string `yaml:"mode"`
	UseIconAsAnswerIcon bool   `yaml:"use_icon_as_answer_icon"`
}

// DifyDependency represents a Dify dependency
type DifyDependency struct {
	CurrentIdentifier interface{}  `yaml:"current_identifier"`
	Type              string       `yaml:"type"`
	Value             DifyDepValue `yaml:"value"`
}

// DifyDepValue represents dependency value
type DifyDepValue struct {
	MarketplacePluginUniqueIdentifier string `yaml:"marketplace_plugin_unique_identifier"`
}

// DifyWorkflow represents a Dify workflow
type DifyWorkflow struct {
	ConversationVariables []interface{} `yaml:"conversation_variables"`
	EnvironmentVariables  []interface{} `yaml:"environment_variables"`
	Features              DifyFeatures  `yaml:"features"`
	Graph                 DifyGraph     `yaml:"graph"`
}

// DifyFeatures represents Dify feature configuration
type DifyFeatures struct {
	FileUpload                    DifyFileUpload                    `yaml:"file_upload"`
	OpeningStatement              string                            `yaml:"opening_statement"`
	RetrieverResource             DifyRetrieverResource             `yaml:"retriever_resource"`
	SensitiveWordAvoidance        DifySensitiveWordAvoidance        `yaml:"sensitive_word_avoidance"`
	SpeechToText                  DifySpeechToText                  `yaml:"speech_to_text"`
	SuggestedQuestions            []string                          `yaml:"suggested_questions"`
	SuggestedQuestionsAfterAnswer DifySuggestedQuestionsAfterAnswer `yaml:"suggested_questions_after_answer"`
	TextToSpeech                  DifyTextToSpeech                  `yaml:"text_to_speech"`
}

// DifyFileUpload represents file upload configuration
type DifyFileUpload struct {
	AllowedFileExtensions    []string             `yaml:"allowed_file_extensions"`
	AllowedFileTypes         []string             `yaml:"allowed_file_types"`
	AllowedFileUploadMethods []string             `yaml:"allowed_file_upload_methods"`
	Enabled                  bool                 `yaml:"enabled"`
	FileUploadConfig         DifyFileUploadConfig `yaml:"fileUploadConfig"`
	Image                    DifyImageConfig      `yaml:"image"`
	NumberLimits             int                  `yaml:"number_limits"`
}

// DifyFileUploadConfig represents detailed file upload configuration
type DifyFileUploadConfig struct {
	AudioFileSizeLimit      int `yaml:"audio_file_size_limit"`
	BatchCountLimit         int `yaml:"batch_count_limit"`
	FileSizeLimit           int `yaml:"file_size_limit"`
	ImageFileSizeLimit      int `yaml:"image_file_size_limit"`
	VideoFileSizeLimit      int `yaml:"video_file_size_limit"`
	WorkflowFileUploadLimit int `yaml:"workflow_file_upload_limit"`
}

// DifyImageConfig represents image configuration
type DifyImageConfig struct {
	Enabled         bool     `yaml:"enabled"`
	NumberLimits    int      `yaml:"number_limits"`
	TransferMethods []string `yaml:"transfer_methods"`
}

// DifyRetrieverResource represents retriever resource configuration
type DifyRetrieverResource struct {
	Enabled bool `yaml:"enabled"`
}

// DifySensitiveWordAvoidance represents sensitive word avoidance configuration
type DifySensitiveWordAvoidance struct {
	Enabled bool `yaml:"enabled"`
}

// DifySpeechToText represents speech-to-text configuration
type DifySpeechToText struct {
	Enabled bool `yaml:"enabled"`
}

// DifySuggestedQuestionsAfterAnswer represents suggested questions after answer configuration
type DifySuggestedQuestionsAfterAnswer struct {
	Enabled bool `yaml:"enabled"`
}

// DifyTextToSpeech represents text-to-speech configuration
type DifyTextToSpeech struct {
	Enabled  bool   `yaml:"enabled"`
	Language string `yaml:"language"`
	Voice    string `yaml:"voice"`
}

// DifyGraph represents Dify graph structure
type DifyGraph struct {
	Edges []DifyEdge `yaml:"edges"`
	Nodes []DifyNode `yaml:"nodes"`
}

// DifyEdge represents a Dify connection - field order strictly follows official example
type DifyEdge struct {
	Data         DifyEdgeData `yaml:"data"`
	ID           string       `yaml:"id"`
	Selected     bool         `yaml:"selected,omitempty"`
	Source       string       `yaml:"source"`
	SourceHandle string       `yaml:"sourceHandle"`
	Target       string       `yaml:"target"`
	TargetHandle string       `yaml:"targetHandle"`
	Type         string       `yaml:"type"`
	ZIndex       int          `yaml:"zIndex"`
}

// DifyEdgeData represents Dify connection data - field order strictly follows official example
type DifyEdgeData struct {
	IsInIteration bool   `yaml:"isInIteration,omitempty"`
	IsInLoop      bool   `yaml:"isInLoop"`
	IterationID   string `yaml:"iteration_id,omitempty"`
	SourceType    string `yaml:"sourceType"`
	TargetType    string `yaml:"targetType"`
}

// DifyNode represents a Dify node - field order consistent with official example
type DifyNode struct {
	Data             DifyNodeData `yaml:"data"`
	Height           float64      `yaml:"height,omitempty"`
	ID               string       `yaml:"id"`
	Position         DifyPosition `yaml:"position"`
	PositionAbsolute DifyPosition `yaml:"positionAbsolute"`
	Selected         bool         `yaml:"selected"`
	SourcePosition   string       `yaml:"sourcePosition"`
	TargetPosition   string       `yaml:"targetPosition"`
	Type             string       `yaml:"type"`
	Width            float64      `yaml:"width,omitempty"`

	// Iteration node specific fields
	ParentID   string `yaml:"parentId,omitempty"`
	Draggable  *bool  `yaml:"draggable,omitempty"`  // Use pointer type, only shown when explicitly set
	Selectable *bool  `yaml:"selectable,omitempty"` // Use pointer type, only shown when explicitly set
	ZIndex     int    `yaml:"zIndex,omitempty"`
}

// DifyPosition represents Dify position information
type DifyPosition struct {
	X float64 `yaml:"x"`
	Y float64 `yaml:"y"`
}

// DifyNodeData represents Dify node data - field order strictly follows official example
type DifyNodeData struct {
	// Code node specific fields - in standard example order (code field first)
	Code         string `yaml:"code,omitempty"`
	CodeLanguage string `yaml:"code_language,omitempty"`

	// Common fields - in standard example order
	Desc     string `yaml:"desc"`
	Selected bool   `yaml:"selected"`
	Title    string `yaml:"title"`
	Type     string `yaml:"type"`

	// Iteration node specific fields - in standard example order
	ErrorHandleMode   string   `yaml:"error_handle_mode,omitempty"`
	Height            float64  `yaml:"height,omitempty"`
	IsParallel        *bool    `yaml:"is_parallel,omitempty"`
	IteratorInputType string   `yaml:"iterator_input_type,omitempty"`
	IteratorSelector  []string `yaml:"iterator_selector,omitempty"`
	OutputSelector    []string `yaml:"output_selector,omitempty"`
	OutputType        string   `yaml:"output_type,omitempty"`
	ParallelNums      int      `yaml:"parallel_nums,omitempty"`
	StartNodeID       string   `yaml:"start_node_id,omitempty"`
	Width             float64  `yaml:"width,omitempty"`

	// Iteration internal node specific fields - in standard example order
	IsInIteration bool   `yaml:"isInIteration,omitempty"`
	IsInLoop      *bool  `yaml:"isInLoop,omitempty"`
	IterationID   string `yaml:"iteration_id,omitempty"`

	// Code node output fields
	Outputs interface{} `yaml:"outputs,omitempty"`

	// Code node variable fields
	Variables interface{} `yaml:"variables,omitempty"`

	// LLM node specific fields
	Context        map[string]interface{}   `yaml:"context,omitempty"`
	Model          map[string]interface{}   `yaml:"model,omitempty"`
	PromptTemplate []map[string]interface{} `yaml:"prompt_template,omitempty"`
	Vision         map[string]interface{}   `yaml:"vision,omitempty"`

	// Other fields
	Dependencies string                 `yaml:"dependencies,omitempty"`
	Config       map[string]interface{} `yaml:"config,omitempty"`

	// Conditional node specific fields
	Cases []map[string]interface{} `yaml:"cases,omitempty"`

	// Classifier node specific fields
	Classes               []map[string]interface{} `yaml:"classes,omitempty"`
	Instruction           string                   `yaml:"instruction,omitempty"`  // Single instruction
	Instructions          string                   `yaml:"instructions,omitempty"` // Keep empty string, consistent with Dify instance
	QueryVariableSelector []string                 `yaml:"query_variable_selector,omitempty"`
	Topics                []string                 `yaml:"topics,omitempty"`
}

// DifyVariable represents Dify variable definition - field order consistent with official example
type DifyVariable struct {
	Label     string   `yaml:"label"`
	MaxLength int      `yaml:"max_length,omitempty"`
	Options   []string `yaml:"options"`
	Required  bool     `yaml:"required"`
	Type      string   `yaml:"type"`
	Variable  string   `yaml:"variable"`
}

// DifyOutput represents Dify output definition
type DifyOutput struct {
	Variable      string   `yaml:"variable"`
	ValueSelector []string `yaml:"value_selector"`
	ValueType     string   `yaml:"value_type"`
	Type          string   `yaml:"type,omitempty"`
}
