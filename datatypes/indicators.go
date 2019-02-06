package datatypes

import "time"

// Indicator that is used in alerts, pipelines, etc.
type Indicator struct {
	Id          string `json:"id,omitempty"`
	Type        string `json:"type,omitempty"`
	Value       string `json:"value,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Author      string `json:"author,omitempty"`
	Source      string `json:"source,omitempty"`

	// An indicator would never have probability 0.0, so zero means not
	// specified.
	Probability float32 `json:"probability,omitempty"`

}

// Indicators is a collection of flat indicators. This is now
// "IndicatorDefinitions", but we are probably using it someplace...
type Indicators struct {
	Description string       `json:"description,omitempty"`
	Version     string       `json:"version,omitempty"`
	Indicators  []*Indicator `json:"indicators,omitempty"`
}

// TypeValuePair is a type/value pair. Duh.
type TypeValuePair struct {
	Type  string `json:"type" required:"true"`
	Value string `json:"value" required:"true"`
}

// IOCDefinition is used as the branches and leaves of the tree
type IOCDefinition struct {
	Type  string `json:"type, omitempty"`
	Value string `json:"value, omitempty"`
}

// IndicatorDefinition contains the data comprising an IOCDefinition
type IndicatorDefinition struct {
	ID            string         `json:"id" required:"true"`
	AuthoredDate  time.Time      `json:"authored_date" required:"true"`
	LastModified  time.Time      `json:"last_modified" required:"true"`
	Description   string         `json:"description" required:"true"`
	Source        string         `json:"source" required:"true"`
	Author        string         `json:"author" required:"true"`
	Category      string         `json:"category" required:"true"`
	IOCDefinition *IOCDefinition `json:"ioc_definition" required:"true"`
}

// IndicatorDefinitions is a collection of indicator definitions
type IndicatorDefinitions struct {
	Description string                 `json:"description" required:"true"`
	Version     float64                `json:"version" required:"true"`
	Definitions []*IndicatorDefinition `json:"definitions" required:"true"`
}
