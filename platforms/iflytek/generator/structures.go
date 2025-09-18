package generator

// IFlytekDSL represents the root structure of iFlytek SparkAgent DSL.
type IFlytekDSL struct {
	FlowMeta IFlytekFlowMeta `yaml:"flowMeta" json:"flowMeta"`
	FlowData IFlytekFlowData `yaml:"flowData" json:"flowData"`
}

// IFlytekFlowMeta contains flow metadata.
type IFlytekFlowMeta struct {
	Name           string `yaml:"name" json:"name"`
	Description    string `yaml:"description" json:"description"`
	AvatarIcon     string `yaml:"avatarIcon" json:"avatarIcon"`
	AvatarColor    string `yaml:"avatarColor" json:"avatarColor"`
	AdvancedConfig string `yaml:"advancedConfig" json:"advancedConfig"`
	DSLVersion     string `yaml:"dslVersion" json:"dslVersion"`
}

// IFlytekFlowData contains flow data.
type IFlytekFlowData struct {
	Nodes []IFlytekNode `yaml:"nodes" json:"nodes"`
	Edges []IFlytekEdge `yaml:"edges" json:"edges"`
}

// IFlytekNode represents an iFlytek SparkAgent node.
type IFlytekNode struct {
	ID               string          `yaml:"id" json:"id"`
	Dragging         bool            `yaml:"dragging" json:"dragging"`
	Selected         bool            `yaml:"selected" json:"selected"`
	Width            float64         `yaml:"width" json:"width"`
	Height           float64         `yaml:"height" json:"height"`
	Position         IFlytekPosition `yaml:"position" json:"position"`
	PositionAbsolute IFlytekPosition `yaml:"positionAbsolute" json:"positionAbsolute"`
	Type             string          `yaml:"type" json:"type"`
	ParentID         *string         `yaml:"parentId,omitempty" json:"parentId,omitempty"`
	Extent           string          `yaml:"extent,omitempty" json:"extent,omitempty"`
	ZIndex           int             `yaml:"zIndex,omitempty" json:"zIndex,omitempty"`
	Draggable        *bool           `yaml:"draggable,omitempty" json:"draggable,omitempty"`
	Data             IFlytekNodeData `yaml:"data" json:"data"`
}

// IFlytekPosition contains position information.
type IFlytekPosition struct {
	X float64 `yaml:"x" json:"x"`
	Y float64 `yaml:"y" json:"y"`
}

// IFlytekNodeData contains node data.
type IFlytekNodeData struct {
	AllowInputReference  bool                   `yaml:"allowInputReference" json:"allowInputReference"`
	AllowOutputReference bool                   `yaml:"allowOutputReference" json:"allowOutputReference"`
	Label                string                 `yaml:"label" json:"label"`
	LabelEdit            bool                   `yaml:"labelEdit,omitempty" json:"labelEdit,omitempty"`
	Status               string                 `yaml:"status" json:"status"`
	NodeMeta             IFlytekNodeMeta        `yaml:"nodeMeta" json:"nodeMeta"`
	Inputs               []IFlytekInput         `yaml:"inputs" json:"inputs"`
	Outputs              []IFlytekOutput        `yaml:"outputs" json:"outputs"`
	References           []IFlytekReference     `yaml:"references,omitempty" json:"references,omitempty"`
	NodeParam            map[string]interface{} `yaml:"nodeParam" json:"nodeParam"`
	Icon                 string                 `yaml:"icon" json:"icon"`
	Description          string                 `yaml:"description" json:"description"`
	Updatable            bool                   `yaml:"updatable" json:"updatable"`

	// Iteration node specific fields
	ParentID       *string          `yaml:"parentId,omitempty" json:"parentId,omitempty"`
	OriginPosition *IFlytekPosition `yaml:"originPosition,omitempty" json:"originPosition,omitempty"`
}

// IFlytekNodeMeta contains node metadata.
type IFlytekNodeMeta struct {
	AliasName string `yaml:"aliasName" json:"aliasName"`
	NodeType  string `yaml:"nodeType" json:"nodeType"`
}

// IFlytekInput defines input structure.
type IFlytekInput struct {
	ID         string        `yaml:"id" json:"id"`
	Name       string        `yaml:"name" json:"name"`
	NameErrMsg string        `yaml:"nameErrMsg" json:"nameErrMsg"`
	Schema     IFlytekSchema `yaml:"schema" json:"schema"`
	FileType   string        `yaml:"fileType" json:"fileType"`
}

// IFlytekOutput defines output structure.
type IFlytekOutput struct {
	ID                  string        `yaml:"id" json:"id"`
	Name                string        `yaml:"name" json:"name"`
	NameErrMsg          string        `yaml:"nameErrMsg" json:"nameErrMsg"`
	Schema              IFlytekSchema `yaml:"schema" json:"schema"`
	Required            bool          `yaml:"required,omitempty" json:"required,omitempty"`
	DeleteDisabled      bool          `yaml:"deleteDisabled,omitempty" json:"deleteDisabled,omitempty"`
	CustomParameterType string        `yaml:"customParameterType,omitempty" json:"customParameterType,omitempty"`
}

// IFlytekSchema contains data schema.
type IFlytekSchema struct {
	Type       string              `yaml:"type" json:"type"`
	Properties []interface{}       `yaml:"properties,omitempty" json:"properties,omitempty"`
	Default    interface{}         `yaml:"default,omitempty" json:"default,omitempty"`
	Value      *IFlytekSchemaValue `yaml:"value,omitempty" json:"value,omitempty"`
}

// IFlytekSchemaValue contains data value.
type IFlytekSchemaValue struct {
	Type          string      `yaml:"type" json:"type"`
	Content       interface{} `yaml:"content,omitempty" json:"content,omitempty"` // Supports IFlytekRefContent and literal values
	ContentErrMsg string      `yaml:"contentErrMsg" json:"contentErrMsg"`
}

// IFlytekRefContent contains reference content.
type IFlytekRefContent struct {
	Name   string `yaml:"name" json:"name"`
	ID     string `yaml:"id" json:"id"`
	NodeID string `yaml:"nodeId" json:"nodeId"`
}

// IFlytekReference contains node reference.
type IFlytekReference struct {
	Label      string             `yaml:"label" json:"label"`
	Value      string             `yaml:"value" json:"value"`
	ParentNode bool               `yaml:"parentNode,omitempty" json:"parentNode,omitempty"`
	Children   []IFlytekReference `yaml:"children,omitempty" json:"children,omitempty"`
	References []IFlytekRefDetail `yaml:"references,omitempty" json:"references,omitempty"`
}

// IFlytekRefDetail contains reference details.
type IFlytekRefDetail struct {
	OriginID string             `yaml:"originId" json:"originId"`
	Children []IFlytekReference `yaml:"children,omitempty" json:"children,omitempty"`
	ID       string             `yaml:"id" json:"id"`
	Label    string             `yaml:"label" json:"label"`
	Type     string             `yaml:"type" json:"type"`
	Value    string             `yaml:"value" json:"value"`
	FileType string             `yaml:"fileType" json:"fileType"`
}

// IFlytekEdge represents a connection edge.
type IFlytekEdge struct {
	Source       string            `yaml:"source" json:"source"`
	Target       string            `yaml:"target" json:"target"`
	SourceHandle string            `yaml:"sourceHandle,omitempty" json:"sourceHandle,omitempty"`
	TargetHandle string            `yaml:"targetHandle,omitempty" json:"targetHandle,omitempty"`
	Type         string            `yaml:"type" json:"type"`
	ID           string            `yaml:"id" json:"id"`
	MarkerEnd    *IFlytekMarkerEnd `yaml:"markerEnd,omitempty" json:"markerEnd,omitempty"`
	Data         *IFlytekEdgeData  `yaml:"data,omitempty" json:"data,omitempty"`
	ZIndex       int               `yaml:"zIndex,omitempty" json:"zIndex,omitempty"`
}

// IFlytekMarkerEnd contains arrow marker.
type IFlytekMarkerEnd struct {
	Color string `yaml:"color" json:"color"`
	Type  string `yaml:"type" json:"type"`
}

// IFlytekEdgeData contains edge data.
type IFlytekEdgeData struct {
	EdgeType string `yaml:"edgeType" json:"edgeType"`
}
