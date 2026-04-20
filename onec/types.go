package onec

// ObjectStructure represents the structure of a 1C metadata object.
type ObjectStructure struct {
	Name         string        `json:"name"`
	Synonym      string        `json:"synonym"`
	Attributes   []Attribute   `json:"attributes"`
	TabularParts []TabularPart `json:"tabularParts,omitempty"`
	Dimensions   []Attribute   `json:"dimensions,omitempty"`
	Resources    []Attribute   `json:"resources,omitempty"`
}

// Attribute represents a metadata object attribute.
type Attribute struct {
	Name    string `json:"name"`
	Synonym string `json:"synonym"`
	Type    string `json:"type"`
}

// TabularPart represents a tabular part of a metadata object.
type TabularPart struct {
	Name       string      `json:"name"`
	Attributes []Attribute `json:"attributes"`
}

// QueryRequest is the request body for the query endpoint.
type QueryRequest struct {
	Query      string         `json:"query"`
	Limit      int            `json:"limit"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

// QueryResult is the response from the query endpoint.
type QueryResult struct {
	Columns   []string `json:"columns"`
	Rows      [][]any  `json:"rows"`
	Total     int      `json:"total"`
	Truncated bool     `json:"truncated"`
}

// VersionInfo represents the extension version response.
type VersionInfo struct {
	Version string `json:"version"`
}

// FormStructure represents the structure of a 1C form.
type FormStructure struct {
	Name     string        `json:"name"`
	Title    string        `json:"title"`
	Elements []FormElement `json:"elements"`
	Commands []FormCommand `json:"commands,omitempty"`
	Handlers []FormHandler `json:"handlers,omitempty"`
}

// FormElement represents an element on a 1C form.
type FormElement struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Title    string `json:"title,omitempty"`
	DataPath string `json:"dataPath,omitempty"`
}

// FormCommand represents a form command.
type FormCommand struct {
	Name   string `json:"name"`
	Action string `json:"action"`
}

// FormHandler represents an event handler on a form.
type FormHandler struct {
	Event   string `json:"event"`
	Handler string `json:"handler"`
}

// ValidateQueryRequest is the request body for the validate-query endpoint.
type ValidateQueryRequest struct {
	Query string `json:"query"`
}

// ValidateQueryResult is the response from the validate-query endpoint.
type ValidateQueryResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// EventLogRequest is the request body for the eventlog endpoint.
type EventLogRequest struct {
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Level     string `json:"level,omitempty"`
	User      string `json:"user,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

// EventLogResult is the response from the eventlog endpoint.
type EventLogResult struct {
	Events []EventLogEntry `json:"events"`
	Total  int             `json:"total"`
}

// ConfigurationInfo represents general information about the 1C infobase and configuration.
type ConfigurationInfo struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	Vendor          string `json:"vendor"`
	PlatformVersion string `json:"platform_version"`
	Mode            string `json:"mode"`
}

// EventLogEntry represents a single event log record.
type EventLogEntry struct {
	Date        string `json:"date"`
	Level       string `json:"level"`
	Event       string `json:"event"`
	User        string `json:"user"`
	Computer    string `json:"computer,omitempty"`
	Metadata    string `json:"metadata,omitempty"`
	Data        string `json:"data,omitempty"`
	Comment     string `json:"comment,omitempty"`
	Transaction string `json:"transaction,omitempty"`
}

// Counterparty represents a 1C counterparty.
type Counterparty struct {
	Ref              string `json:"ref"`
	Code             string `json:"code"`
	Name             string `json:"name"`
	INN              string `json:"inn"`
	KPP              string `json:"kpp"`
	CounterpartyType string `json:"counterparty_type,omitempty"`
}

// ReadCounterpartiesRequest is the request body for counterparties read endpoint.
type ReadCounterpartiesRequest struct {
	Search string `json:"search,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Code   string `json:"code,omitempty"`
	Ref    string `json:"ref,omitempty"`
	INN    string `json:"inn,omitempty"`
	KPP    string `json:"kpp,omitempty"`
}

// ReadCounterpartiesResult is the response from counterparties read endpoint.
type ReadCounterpartiesResult struct {
	Counterparties []Counterparty `json:"counterparties"`
	Total          int            `json:"total"`
	Truncated      bool           `json:"truncated"`
}

// CreateCounterpartyRequest is the request body for creating a counterparty.
type CreateCounterpartyRequest struct {
	Name             string `json:"name"`
	INN              string `json:"inn"`
	KPP              string `json:"kpp"`
	CounterpartyType string `json:"counterparty_type"`
}

// CreateCounterpartyResult is the response from the counterparty create endpoint.
type CreateCounterpartyResult struct {
	Success      bool         `json:"success"`
	Counterparty Counterparty `json:"counterparty"`
}
