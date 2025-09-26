package generator

// CozeRootStructure represents the root structure of Coze workflow DSL
type CozeRootStructure struct {
	WorkflowID     string           `yaml:"workflowid" json:"workflowid"`
	Name           string           `yaml:"name" json:"name"`
	Description    string           `yaml:"description" json:"description"`
	Version        string           `yaml:"version" json:"version"`
	CreateTime     int64            `yaml:"createtime" json:"createtime"`
	UpdateTime     int64            `yaml:"updatetime" json:"updatetime"`
	Schema         *CozeSchema      `yaml:"schema" json:"schema"`
	Nodes          []CozeNode       `yaml:"nodes" json:"nodes"`
	Edges          []CozeEdge       `yaml:"edges" json:"edges"`
	Metadata       *CozeMetadata    `yaml:"metadata" json:"metadata"`
	Dependencies   []CozeDependency `yaml:"dependencies" json:"dependencies"`
	ExportFormat   string           `yaml:"exportformat" json:"exportformat"`
	SerializedData string           `yaml:"serializeddata" json:"serializeddata"`
}

// CozeSchema represents the schema section of Coze workflow
type CozeSchema struct {
	Edges    []CozeSchemaEdge `yaml:"edges" json:"edges"`
	Nodes    []CozeSchemaNode `yaml:"nodes" json:"nodes"`
	Versions *CozeVersions    `yaml:"versions" json:"versions"`
}

// CozeSchemaEdge represents edge in schema section
type CozeSchemaEdge struct {
	SourceNodeID string `yaml:"sourceNodeID" json:"sourceNodeID"`
	SourcePortID string `yaml:"sourcePortID,omitempty" json:"sourcePortID,omitempty"`
	TargetNodeID string `yaml:"targetNodeID" json:"targetNodeID"`
	TargetPortID string `yaml:"targetPortID,omitempty" json:"targetPortID,omitempty"`
}

// CozeSchemaNode represents node in schema section
type CozeSchemaNode struct {
	Data   *CozeSchemaNodeData `yaml:"data" json:"data"`
	ID     string              `yaml:"id" json:"id"`
	Meta   *CozeNodeMeta       `yaml:"meta" json:"meta"`
	Type   string              `yaml:"type" json:"type"`
	Blocks []interface{}       `yaml:"blocks,omitempty" json:"blocks,omitempty"` // Added for iteration nodes
	Edges  []interface{}       `yaml:"edges,omitempty" json:"edges,omitempty"`   // Added for iteration internal edges
}

// CozeSchemaNodeData represents node data in schema section
type CozeSchemaNodeData struct {
	NodeMeta          *CozeNodeMetaInfo `yaml:"nodeMeta" json:"nodeMeta"`
	Outputs           interface{}       `yaml:"outputs,omitempty" json:"outputs,omitempty"` // Changed to interface{} to support complex reference structures for iteration nodes
	Inputs            interface{}       `yaml:"inputs,omitempty" json:"inputs,omitempty"`   // Supports different input types through interface
	TriggerParameters []CozeNodeOutput  `yaml:"trigger_parameters,omitempty" json:"trigger_parameters,omitempty"`
	TerminatePlan     string            `yaml:"terminatePlan,omitempty" json:"terminatePlan,omitempty"`
	Version           string            `yaml:"version,omitempty" json:"version,omitempty"` // Specifies node version for compatibility
}

// CozeNode represents a workflow node in Coze format
type CozeNode struct {
	ID      string        `yaml:"id" json:"id"`
	Type    string        `yaml:"type" json:"type"`
	Meta    *CozeNodeMeta `yaml:"meta" json:"meta"`
	Data    *CozeNodeData `yaml:"data" json:"data"`
	Blocks  []interface{} `yaml:"blocks" json:"blocks"`
	Edges   []interface{} `yaml:"edges" json:"edges"`
	Version string        `yaml:"version" json:"version"`
	Size    interface{}   `yaml:"size" json:"size"`
}

// CozeBlockNode represents a node inside iteration blocks with correct field ordering
type CozeBlockNode struct {
	Data *CozeBlockNodeData `yaml:"data" json:"data"` // ✅ data first
	ID   string             `yaml:"id" json:"id"`     // ✅ id after data
	Meta *CozeNodeMeta      `yaml:"meta" json:"meta"` // ✅ meta after id
	Type string             `yaml:"type" json:"type"` // ✅ type last
}

// CozeBlockNodeData represents data for iteration block nodes with correct structure
type CozeBlockNodeData struct {
	Inputs   interface{}       `yaml:"inputs" json:"inputs"`     // ✅ inputs first
	NodeMeta *CozeNodeMetaInfo `yaml:"nodeMeta" json:"nodeMeta"` // ✅ nodeMeta second
	Outputs  interface{}       `yaml:"outputs" json:"outputs"`   // ✅ outputs third
	Version  string            `yaml:"version" json:"version"`   // ✅ version last (inside data)
}

// CozeNodeMeta represents node metadata
type CozeNodeMeta struct {
	CanvasPosition *CozePosition `yaml:"canvasPosition,omitempty" json:"canvasPosition,omitempty"` // FIXED: Add missing canvasPosition field
	Position       *CozePosition `yaml:"position" json:"position"`
}

// CozePosition represents node position
type CozePosition struct {
	X interface{} `yaml:"x" json:"x"`
	Y interface{} `yaml:"y" json:"y"`
}

// CozeNodeData represents node data
type CozeNodeData struct {
	Meta    *CozeNodeMetaInfo `yaml:"meta" json:"meta"`
	Outputs interface{}       `yaml:"outputs" json:"outputs"` // Changed to interface{} to support both []CozeNodeOutput and complex reference structures for iteration nodes
	Inputs  interface{}       `yaml:"inputs" json:"inputs"`
	Size    interface{}       `yaml:"size" json:"size"`
	// LLM node specific configuration
	LLM *CozeLLMConfig `yaml:"llm,omitempty" json:"llm,omitempty"`
}

// CozeNodeMetaInfo represents node meta information
type CozeNodeMetaInfo struct {
	Title       string `yaml:"title" json:"title"`
	Description string `yaml:"description" json:"description"`
	Icon        string `yaml:"icon" json:"icon"`
	SubTitle    string `yaml:"subTitle,omitempty" json:"subTitle,omitempty"`
	Subtitle    string `yaml:"subtitle,omitempty" json:"subtitle,omitempty"`
	MainColor   string `yaml:"maincolor,omitempty" json:"maincolor,omitempty"`
}

// CozeNodeOutput represents node output parameter
type CozeNodeOutput struct {
	Name     string            `yaml:"name" json:"name"`
	Required bool              `yaml:"required" json:"required"`
	Type     string            `yaml:"type" json:"type"`
	Schema   *CozeOutputSchema `yaml:"schema,omitempty" json:"schema,omitempty"` // Schema for array/complex types
}

// CozeOutputSchema represents output schema for array/complex types
type CozeOutputSchema struct {
	Type string `yaml:"type" json:"type"` // Element type for arrays
}

// CozeNodeInputs represents node inputs for end node
type CozeNodeInputs struct {
	InputParameters []CozeInputParameter `yaml:"inputParameters" json:"inputParameters"`
}

// CozeSchemaNodeInputs represents comprehensive node inputs for schema section (includes LLM params)
type CozeSchemaNodeInputs struct {
	InputParameters []CozeInputParameter     `yaml:"inputParameters,omitempty" json:"inputParameters,omitempty"`
	LLMParam        []map[string]interface{} `yaml:"llmParam,omitempty" json:"llmParam,omitempty"`
	SettingOnError  map[string]interface{}   `yaml:"settingOnError,omitempty" json:"settingOnError,omitempty"`
}

// CozeInputParameter represents input parameter for end node
type CozeInputParameter struct {
	Name  string          `yaml:"name" json:"name"`
	Input *CozeInputValue `yaml:"input" json:"input"`
}

// CozeInputValue represents input value with reference
type CozeInputValue struct {
	Type   string           `yaml:"type" json:"type"`                         // Maps unified data types to Coze type format
	Schema *CozeInputSchema `yaml:"schema,omitempty" json:"schema,omitempty"` // Schema definition for iteration inputs
	Value  *CozeInputRef    `yaml:"value" json:"value"`                       // Contains reference content and metadata
}

// CozeInputSchema represents input schema definition
type CozeInputSchema struct {
	Type string `yaml:"type" json:"type"` // Schema type (string, integer, etc.)
}

// CozeInputRef represents input reference
type CozeInputRef struct {
	Content *CozeRefContent `yaml:"content" json:"content"`
	RawMeta *CozeRawMeta    `yaml:"rawMeta,omitempty" json:"rawMeta,omitempty"`
	Type    string          `yaml:"type" json:"type"`
}

// CozeBlockInputs represents inputs for iteration block nodes
type CozeBlockInputs struct {
	InputParameters []CozeBlockInputParameter `yaml:"inputParameters,omitempty" json:"inputParameters,omitempty"`
	LLMParam        []CozeBlockInputParameter `yaml:"llmParam,omitempty" json:"llmParam,omitempty"`
	SettingOnError  *CozeBlockSettingOnError  `yaml:"settingOnError,omitempty" json:"settingOnError,omitempty"`
}

// CozeBlockInputParameter represents input parameter for block nodes
type CozeBlockInputParameter struct {
	Input *CozeBlockInputValue `yaml:"input" json:"input"`
	Name  string               `yaml:"name" json:"name"`
}

// CozeBlockInputValue represents input value for block nodes
type CozeBlockInputValue struct {
	Type    string             `yaml:"type" json:"type"`
	Value   *CozeBlockInputRef `yaml:"value" json:"value"`
	RawMeta *CozeRawMeta       `yaml:"rawMeta,omitempty" json:"rawMeta,omitempty"`
}

// CozeBlockInputRef represents input reference for block nodes
type CozeBlockInputRef struct {
	Content *CozeRefContent `yaml:"content,omitempty" json:"content,omitempty"`
	RawMeta *CozeRawMeta    `yaml:"rawMeta,omitempty" json:"rawMeta,omitempty"`
	Type    string          `yaml:"type" json:"type"`
	// Literal value alternative when not using reference content
	// Use a distinct field name to avoid duplicate tags with Content
	Literal interface{} `yaml:"literal,omitempty" json:"literal,omitempty"`
}

// CozeBlockSettingOnError represents error settings for block nodes
type CozeBlockSettingOnError struct {
	ProcessType int `yaml:"processType" json:"processType"`
	RetryTimes  int `yaml:"retryTimes" json:"retryTimes"`
	TimeoutMs   int `yaml:"timeoutMs" json:"timeoutMs"`
}

// CozeBlockOutput represents output for block nodes
type CozeBlockOutput struct {
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type" json:"type"`
}

// CozeRefContent represents reference content
type CozeRefContent struct {
	BlockID string `yaml:"blockID" json:"blockID"`
	Name    string `yaml:"name" json:"name"`
	Source  string `yaml:"source" json:"source"`
}

// CozeRawMeta represents raw metadata for type mapping
type CozeRawMeta struct {
	Type int `yaml:"type" json:"type"`
}

// CozeEdge represents workflow edge
type CozeEdge struct {
	FromNode string `yaml:"from_node" json:"from_node"`
	FromPort string `yaml:"from_port" json:"from_port"`
	ToNode   string `yaml:"to_node" json:"to_node"`
	ToPort   string `yaml:"to_port" json:"to_port"`
}

// CozeMetadata represents workflow metadata
type CozeMetadata struct {
	ContentType string `yaml:"content_type" json:"content_type"`
	CreatorID   string `yaml:"creator_id" json:"creator_id"`
	Mode        string `yaml:"mode" json:"mode"`
	SpaceID     string `yaml:"space_id" json:"space_id"`
}

// CozeDependency represents workflow dependency
type CozeDependency struct {
	Metadata     *CozeDependencyMeta `yaml:"metadata" json:"metadata"`
	ResourceID   string              `yaml:"resource_id" json:"resource_id"`
	ResourceName string              `yaml:"resource_name" json:"resource_name"`
	ResourceType string              `yaml:"resource_type" json:"resource_type"`
}

// CozeDependencyMeta represents dependency metadata
type CozeDependencyMeta struct {
	NodeType string `yaml:"node_type" json:"node_type"`
}

// CozeVersions represents version information
type CozeVersions struct {
	Loop string `yaml:"loop" json:"loop"`
}

// CozeDataTypeMapping maps unified types to Coze rawMeta type numbers
var CozeDataTypeMapping = map[string]int{
	"string":  1,
	"integer": 2,
	"boolean": 3,
	"float":   4,
	"number":  4,  // number maps to float in coze
	"array":   5,  // array type for Coze format
	"list":    99, // FIXED: list type for iteration outputs uses 99
}

// GetCozeRawMetaType returns the Coze rawMeta type number for a given data type
func GetCozeRawMetaType(dataType string) int {
	if typeNum, exists := CozeDataTypeMapping[dataType]; exists {
		return typeNum
	}
	return 1 // default to string
}

// CozeLLMConfig represents LLM node configuration
type CozeLLMConfig struct {
	Model      *CozeLLMModel      `yaml:"model" json:"model"`
	Parameters *CozeLLMParams     `yaml:"parameters" json:"parameters"`
	Prompt     *CozePromptConfig  `yaml:"prompt" json:"prompt"`
	Context    *CozeContextConfig `yaml:"context,omitempty" json:"context,omitempty"`
}

// CozeLLMModel represents LLM model configuration
type CozeLLMModel struct {
	ID       string `yaml:"id" json:"id"`
	Name     string `yaml:"name" json:"name"`
	Provider string `yaml:"provider" json:"provider"`
	Domain   string `yaml:"domain,omitempty" json:"domain,omitempty"`
	Service  string `yaml:"service,omitempty" json:"service,omitempty"`
}

// CozeLLMParams represents LLM model parameters
type CozeLLMParams struct {
	Temperature float64 `yaml:"temperature" json:"temperature"`
	MaxTokens   int     `yaml:"maxTokens" json:"maxTokens"`
	TopK        int     `yaml:"topK,omitempty" json:"topK,omitempty"`
}

// CozePromptConfig represents prompt configuration
type CozePromptConfig struct {
	SystemTemplate string               `yaml:"systemTemplate" json:"systemTemplate"`
	UserTemplate   string               `yaml:"userTemplate,omitempty" json:"userTemplate,omitempty"`
	Variables      []CozePromptVariable `yaml:"variables,omitempty" json:"variables,omitempty"`
}

// CozePromptVariable represents prompt template variable
type CozePromptVariable struct {
	Name     string `yaml:"name" json:"name"`
	Source   string `yaml:"source" json:"source"`
	NodeID   string `yaml:"nodeID" json:"nodeID"`
	OutputID string `yaml:"outputID" json:"outputID"`
}

// CozeContextConfig represents context configuration
type CozeContextConfig struct {
	Enabled          bool     `yaml:"enabled" json:"enabled"`
	VariableSelector []string `yaml:"variableSelector,omitempty" json:"variableSelector,omitempty"`
}

// CozeIntent represents intent for intent recognition node
type CozeIntent struct {
	Name string `yaml:"name" json:"name"`
}

// CozeLLMParam represents LLM parameters for intent recognition
type CozeLLMParam struct {
	ChatHistoryRound    int             `yaml:"chatHistoryRound" json:"chatHistoryRound"`
	EnableChatHistory   bool            `yaml:"enableChatHistory" json:"enableChatHistory"`
	GenerationDiversity string          `yaml:"generationDiversity" json:"generationDiversity"`
	MaxTokens           int             `yaml:"maxTokens" json:"maxTokens"`
	ModelName           string          `yaml:"modelName" json:"modelName"`
	ModelType           int             `yaml:"modelType" json:"modelType"`
	Prompt              CozePromptValue `yaml:"prompt" json:"prompt"`
	ResponseFormat      int             `yaml:"responseFormat" json:"responseFormat"`
	SystemPrompt        CozePromptValue `yaml:"systemPrompt" json:"systemPrompt"`
	Temperature         float64         `yaml:"temperature" json:"temperature"`
	TopP                float64         `yaml:"topP" json:"topP"`
}

// CozePromptValue represents prompt value structure
type CozePromptValue struct {
	Type  string            `yaml:"type" json:"type"`
	Value CozePromptContent `yaml:"value" json:"value"`
}

// CozePromptContent represents prompt content structure
type CozePromptContent struct {
	Content string `yaml:"content" json:"content"`
	Type    string `yaml:"type" json:"type"`
}

// CozeChatHistorySetting represents chat history settings
type CozeChatHistorySetting struct {
	EnableChatHistory bool `yaml:"enableChatHistory" json:"enableChatHistory"`
	ChatHistoryRound  int  `yaml:"chatHistoryRound" json:"chatHistoryRound"`
}

// CozeIntentInputs represents complete node inputs structure for intent recognition
type CozeIntentInputs struct {
	InputParameters    []CozeInputParameter   `yaml:"inputParameters" json:"inputParameters"`
	ChatHistorySetting CozeChatHistorySetting `yaml:"chatHistorySetting,omitempty" json:"chatHistorySetting,omitempty"`
	Intents            []CozeIntent           `yaml:"intents,omitempty" json:"intents,omitempty"`
	LLMParam           CozeLLMParam           `yaml:"llmParam,omitempty" json:"llmParam,omitempty"`
	Mode               string                 `yaml:"mode,omitempty" json:"mode,omitempty"`
	SettingOnError     CozeSettingOnError     `yaml:"settingOnError" json:"settingOnError"`
}

// CozeSettingOnError represents error handling settings
type CozeSettingOnError struct {
	ProcessType int `yaml:"processType" json:"processType"`
	RetryTimes  int `yaml:"retryTimes" json:"retryTimes"`
	TimeoutMs   int `yaml:"timeoutMs" json:"timeoutMs"`
}
