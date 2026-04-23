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
	Ref              string `json:"ref"` // UUID string
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

// Nomenclature represents a 1C nomenclature item.
type Nomenclature struct {
	Ref              string `json:"ref"` // UUID string
	Code             string `json:"code"`
	Name             string `json:"name"`
	FullName         string `json:"full_name,omitempty"`
	Article          string `json:"article,omitempty"`
	NomenclatureType string `json:"nomenclature_type,omitempty"`
	Unit             string `json:"unit,omitempty"`
	IsService        bool   `json:"is_service"`
}

// CreateNomenclatureRequest is the request body for creating a nomenclature item.
type CreateNomenclatureRequest struct {
	Name             string `json:"name"`
	FullName         string `json:"full_name,omitempty"`
	Article          string `json:"article,omitempty"`
	NomenclatureType string `json:"nomenclature_type,omitempty"`
	Unit             string `json:"unit,omitempty"`
	IsService        *bool  `json:"is_service,omitempty"`
}

// CreateNomenclatureResult is the response from nomenclature create endpoint.
type CreateNomenclatureResult struct {
	Success      bool         `json:"success"`
	Nomenclature Nomenclature `json:"nomenclature"`
}

// ReadNomenclatureRequest is the request body for nomenclature read endpoint.
type ReadNomenclatureRequest struct {
	Search  string `json:"search,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Code    string `json:"code,omitempty"`
	Ref     string `json:"ref,omitempty"`
	Article string `json:"article,omitempty"`
}

// ReadNomenclatureResult is the response from nomenclature read endpoint.
type ReadNomenclatureResult struct {
	Nomenclature []Nomenclature `json:"nomenclature"`
	Total        int            `json:"total"`
	Truncated    bool           `json:"truncated"`
}

// Organization represents a 1C organization.
type Organization struct {
	Ref  string `json:"ref"` // UUID string
	Code string `json:"code"`
	Name string `json:"name"`
	INN  string `json:"inn,omitempty"`
	KPP  string `json:"kpp,omitempty"`
}

// ReadOrganizationsRequest is the request body for organizations read endpoint.
type ReadOrganizationsRequest struct {
	Search string `json:"search,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Code   string `json:"code,omitempty"`
	Ref    string `json:"ref,omitempty"`
	INN    string `json:"inn,omitempty"`
	KPP    string `json:"kpp,omitempty"`
}

// ReadOrganizationsResult is the response from organizations read endpoint.
type ReadOrganizationsResult struct {
	Organizations []Organization `json:"organizations"`
	Total         int            `json:"total"`
	Truncated     bool           `json:"truncated"`
}

// Contract represents a 1C counterparty contract.
type Contract struct {
	Ref          string `json:"ref"` // UUID string
	Code         string `json:"code"`
	Name         string `json:"name"`
	Number       string `json:"number,omitempty"`
	Date         string `json:"date,omitempty"`
	Organization string `json:"organization,omitempty"`
	Counterparty string `json:"counterparty,omitempty"`
	Currency     string `json:"currency,omitempty"`
}

// ReadContractsRequest is the request body for contracts read endpoint.
type ReadContractsRequest struct {
	Search          string `json:"search,omitempty"`
	Limit           int    `json:"limit,omitempty"`
	Code            string `json:"code,omitempty"`
	Ref             string `json:"ref,omitempty"`
	Number          string `json:"number,omitempty"`
	CounterpartyRef string `json:"counterparty_ref,omitempty"`
	OrganizationRef string `json:"organization_ref,omitempty"`
}

// ReadContractsResult is the response from contracts read endpoint.
type ReadContractsResult struct {
	Contracts []Contract `json:"contracts"`
	Total     int        `json:"total"`
	Truncated bool       `json:"truncated"`
}

// SalesItem represents a line item for sales documents.
type SalesItem struct {
	NomenclatureRef string  `json:"nomenclature_ref"`
	Quantity        float64 `json:"quantity"`
	Price           float64 `json:"price"`
}

// SalesInvoice represents a buyer invoice document.
type SalesInvoice struct {
	Ref          string `json:"ref"` // UUID string
	Number       string `json:"number"`
	Date         string `json:"date,omitempty"`
	Organization string `json:"organization,omitempty"`
	Counterparty string `json:"counterparty,omitempty"`
	Contract     string `json:"contract,omitempty"`
	Amount       string `json:"amount,omitempty"`
	Posted       bool   `json:"posted"`
}

// ReadSalesInvoicesRequest is the request body for invoices read endpoint.
type ReadSalesInvoicesRequest struct {
	Search          string `json:"search,omitempty"`
	Limit           int    `json:"limit,omitempty"`
	Ref             string `json:"ref,omitempty"`
	Number          string `json:"number,omitempty"`
	CounterpartyRef string `json:"counterparty_ref,omitempty"`
	OrganizationRef string `json:"organization_ref,omitempty"`
}

// ReadSalesInvoicesResult is the response from invoices read endpoint.
type ReadSalesInvoicesResult struct {
	Documents []SalesInvoice `json:"documents"`
	Total     int            `json:"total"`
	Truncated bool           `json:"truncated"`
}

// CreateSalesInvoiceRequest is the request body for creating buyer invoice.
type CreateSalesInvoiceRequest struct {
	OrganizationRef string      `json:"organization_ref"`
	CounterpartyRef string      `json:"counterparty_ref"`
	ContractRef     string      `json:"contract_ref,omitempty"`
	Date            string      `json:"date,omitempty"`
	Comment         string      `json:"comment,omitempty"`
	Post            bool        `json:"post"`
	Items           []SalesItem `json:"items,omitempty"`
}

// CreateSalesInvoiceResult is the response from invoice create endpoint.
type CreateSalesInvoiceResult struct {
	Success  bool         `json:"success"`
	Document SalesInvoice `json:"document"`
}

// UpdateSalesInvoiceRequest is the request body for updating buyer invoice.
type UpdateSalesInvoiceRequest struct {
	Ref             string      `json:"ref"`
	OrganizationRef string      `json:"organization_ref,omitempty"`
	CounterpartyRef string      `json:"counterparty_ref,omitempty"`
	ContractRef     string      `json:"contract_ref,omitempty"`
	Date            string      `json:"date,omitempty"`
	Comment         string      `json:"comment,omitempty"`
	Post            bool        `json:"post"`
	Items           []SalesItem `json:"items,omitempty"`
}

// UpdateSalesInvoiceResult is the response from invoice update endpoint.
type UpdateSalesInvoiceResult struct {
	Success  bool         `json:"success"`
	Document SalesInvoice `json:"document"`
}

// SalesDocument represents a sales act/waybill document.
type SalesDocument struct {
	Ref          string `json:"ref"` // UUID string
	Number       string `json:"number"`
	Date         string `json:"date,omitempty"`
	Organization string `json:"organization,omitempty"`
	Counterparty string `json:"counterparty,omitempty"`
	Contract     string `json:"contract,omitempty"`
	Amount       string `json:"amount,omitempty"`
	Posted       bool   `json:"posted"`
}

// ReadSalesDocumentsRequest is the request body for sales documents read endpoint.
type ReadSalesDocumentsRequest struct {
	Search          string `json:"search,omitempty"`
	Limit           int    `json:"limit,omitempty"`
	Ref             string `json:"ref,omitempty"`
	Number          string `json:"number,omitempty"`
	CounterpartyRef string `json:"counterparty_ref,omitempty"`
	OrganizationRef string `json:"organization_ref,omitempty"`
}

// ReadSalesDocumentsResult is the response from sales documents read endpoint.
type ReadSalesDocumentsResult struct {
	Documents []SalesDocument `json:"documents"`
	Total     int             `json:"total"`
	Truncated bool            `json:"truncated"`
}

// CreateSalesDocumentRequest is the request body for creating sales document.
type CreateSalesDocumentRequest struct {
	OrganizationRef string      `json:"organization_ref"`
	CounterpartyRef string      `json:"counterparty_ref"`
	ContractRef     string      `json:"contract_ref,omitempty"`
	InvoiceRef      string      `json:"invoice_ref,omitempty"`
	Date            string      `json:"date,omitempty"`
	Comment         string      `json:"comment,omitempty"`
	Post            bool        `json:"post"`
	Items           []SalesItem `json:"items,omitempty"`
}

// CreateSalesDocumentResult is the response from sales document create endpoint.
type CreateSalesDocumentResult struct {
	Success  bool          `json:"success"`
	Document SalesDocument `json:"document"`
}

// UpdateSalesDocumentRequest is the request body for updating sales document.
type UpdateSalesDocumentRequest struct {
	Ref             string      `json:"ref"`
	OrganizationRef string      `json:"organization_ref,omitempty"`
	CounterpartyRef string      `json:"counterparty_ref,omitempty"`
	ContractRef     string      `json:"contract_ref,omitempty"`
	InvoiceRef      string      `json:"invoice_ref,omitempty"`
	Date            string      `json:"date,omitempty"`
	Comment         string      `json:"comment,omitempty"`
	Post            bool        `json:"post"`
	Items           []SalesItem `json:"items,omitempty"`
}

// UpdateSalesDocumentResult is the response from sales document update endpoint.
type UpdateSalesDocumentResult struct {
	Success  bool          `json:"success"`
	Document SalesDocument `json:"document"`
}
