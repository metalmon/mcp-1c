package tools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// toolwiring_business_test.go IS THE SAME CONTRACT, for the business toolset of
// this fork (--toolset business|all).
//
// It exists as its own file for one reason: the rows below are this fork's, and
// keeping them out of toolwiring_test.go keeps the next upstream merge a merge of
// upstream's table rather than a three way conflict inside one literal. They are
// appended to upstream's tables by operationalSites() and protocolSites(), so the
// drivers and the census in tool_error_contract_test.go read one enumeration.
//
// SITES ARE NAMED BY (file, source text), never by line number, exactly as
// upstream's rows are, because the file name is what the census matches on.

// unsuccessful1C answers every path with 200 and {"success":false}, which is the
// shape 1C uses when it accepted the call and refused the operation. It is the
// only way to reach the «1C returned unsuccessful … result» sites: a transport
// or status failure leaves through the site above them.
func unsuccessful1C(t *testing.T) *onec.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"success":false}`))
	}))
	t.Cleanup(srv.Close)
	return onec.NewClient(srv.URL, "", "")
}

// businessOops is the detail the stand in 1C returns, so a row that expects the
// cause to be quoted has one string to look for.
const businessOops = `Поле не найдено "Номенклатура"`

func businessOperationalSites() []operationalSite {
	const (
		validCounterparty = `{"name":"ООО Ромашка","inn":"7700000000","kpp":"770001001","counterparty_type":"ЮридическоеЛицо"}`
		validAttachment   = `{"document_ref":"Документ.РеализацияТоваровУслуг:aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",` +
			`"file_name":"scan.pdf","content_base64":"QUJDRA=="}`
		// quantity and price are STRINGS here: onec.IncomingGoodsItem carries the
		// numbers 1C expects as text, and a JSON number would leave through the
		// decode site instead of the one this row drives.
		validIncoming = `{"organization_ref":"a","counterparty_ref":"b","items":[{"nomenclature_ref":"n1","quantity":"1","price":"10"}]}`
	)
	return []operationalSite{
		// contracts.go
		{
			name: "read_contracts live fetch", site: `contracts.go "reading contracts from 1C"`,
			build:   func(t *testing.T) mcp.ToolHandler { return NewReadContractsHandler(envelope1C(t, 500, businessOops)) },
			args:    `{}`,
			heading: headingContractsRead, wants: []string{businessOops},
		},

		// counterparties.go
		{
			name: "read_counterparties live fetch", site: `counterparties.go "reading counterparties from 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewReadCounterpartiesHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{}`,
			heading: headingCounterpartiesRead, wants: []string{businessOops},
		},
		{
			name: "create_counterparty required fields", site: `counterparties.go "name, inn, kpp, counterparty_type are required"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewCreateCounterpartyHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"name":"ООО Ромашка"}`,
			heading: headingCounterpartyCreate, wants: []string{"name, inn, kpp, counterparty_type are required"},
		},
		{
			name: "create_counterparty live post", site: `counterparties.go "creating counterparty in 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewCreateCounterpartyHandler(envelope1C(t, 500, businessOops))
			},
			args:    validCounterparty,
			heading: headingCounterpartyCreate, wants: []string{businessOops},
		},
		{
			name: "create_counterparty unsuccessful", site: `counterparties.go "1C returned unsuccessful create result"`,
			build:   func(t *testing.T) mcp.ToolHandler { return NewCreateCounterpartyHandler(unsuccessful1C(t)) },
			args:    validCounterparty,
			heading: headingCounterpartyCreate, wants: []string{"1C returned unsuccessful create result"},
		},

		// nomenclature_read.go
		{
			name: "read_nomenclature live fetch", site: `nomenclature_read.go "reading nomenclature from 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewReadNomenclatureHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{}`,
			heading: headingNomenclatureRead, wants: []string{businessOops},
		},

		// nomenclature.go
		{
			name: "create_nomenclature name required", site: `nomenclature.go "name is required"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewCreateNomenclatureHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{}`,
			heading: headingNomenclatureCreate, wants: []string{"name is required"},
		},
		{
			name: "create_nomenclature live post", site: `nomenclature.go "creating nomenclature in 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewCreateNomenclatureHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"name":"Болт"}`,
			heading: headingNomenclatureCreate, wants: []string{businessOops},
		},
		{
			name: "create_nomenclature unsuccessful", site: `nomenclature.go "1C returned unsuccessful create result"`,
			build:   func(t *testing.T) mcp.ToolHandler { return NewCreateNomenclatureHandler(unsuccessful1C(t)) },
			args:    `{"name":"Болт"}`,
			heading: headingNomenclatureCreate, wants: []string{"1C returned unsuccessful create result"},
		},

		// organizations.go
		{
			name: "read_organizations live fetch", site: `organizations.go "reading organizations from 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewReadOrganizationsHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{}`,
			heading: headingOrganizationsRead, wants: []string{businessOops},
		},

		// document_attachments.go
		{
			name: "attach_file document_ref required", site: `document_attachments.go "document_ref is required"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewAttachFileToDocumentHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"file_name":"scan.pdf","content_base64":"QUJDRA=="}`,
			heading: headingAttachmentAttach, wants: []string{"document_ref is required"},
		},
		{
			name: "attach_file payload required", site: `document_attachments.go "file_name and content_base64 are required"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewAttachFileToDocumentHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"document_ref":"Документ.РеализацияТоваровУслуг:aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}`,
			heading: headingAttachmentAttach, wants: []string{"file_name and content_base64 are required"},
		},
		{
			name: "attach_file live post", site: `document_attachments.go "attaching file to document in 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewAttachFileToDocumentHandler(envelope1C(t, 500, businessOops))
			},
			args:    validAttachment,
			heading: headingAttachmentAttach, wants: []string{businessOops},
		},
		{
			name: "attach_file unsuccessful", site: `document_attachments.go "1C returned unsuccessful attach result"`,
			build:   func(t *testing.T) mcp.ToolHandler { return NewAttachFileToDocumentHandler(unsuccessful1C(t)) },
			args:    validAttachment,
			heading: headingAttachmentAttach, wants: []string{"1C returned unsuccessful attach result"},
		},

		// document_attachments_read_update.go
		{
			name: "read_attachments document_ref required", site: `document_attachments_read_update.go "document_ref is required"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewReadDocumentAttachmentsHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{}`,
			heading: headingAttachmentsRead, wants: []string{"document_ref is required"},
		},
		{
			name: "read_attachments live fetch", site: `document_attachments_read_update.go "reading document attachments from 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewReadDocumentAttachmentsHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"document_ref":"Документ.РеализацияТоваровУслуг:aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}`,
			heading: headingAttachmentsRead, wants: []string{businessOops},
		},
		{
			name: "attachment_content attachment_ref required", site: `document_attachments_read_update.go "attachment_ref is required"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewGetDocumentAttachmentContentHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{}`,
			heading: headingAttachmentContentRead, wants: []string{"attachment_ref is required"},
		},
		{
			name: "attachment_content live fetch", site: `document_attachments_read_update.go "reading document attachment content from 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewGetDocumentAttachmentContentHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"attachment_ref":"Справочник.РеализацияТоваровУслугПрисоединенныеФайлы:aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}`,
			heading: headingAttachmentContentRead, wants: []string{businessOops},
		},
		{
			name: "attachment_metadata attachment_ref required", site: `document_attachments_read_update.go "attachment_ref is required (update)"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewUpdateDocumentAttachmentMetadataHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"file_name":"scan.pdf"}`,
			heading: headingAttachmentMetaUpdate, wants: []string{"attachment_ref is required"},
		},
		{
			name: "attachment_metadata one field required", site: `document_attachments_read_update.go "at least one field is required"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewUpdateDocumentAttachmentMetadataHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"attachment_ref":"Справочник.РеализацияТоваровУслугПрисоединенныеФайлы:aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}`,
			heading: headingAttachmentMetaUpdate, wants: []string{"at least one field is required: file_name or description"},
		},
		{
			name: "attachment_metadata live post", site: `document_attachments_read_update.go "updating document attachment metadata in 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewUpdateDocumentAttachmentMetadataHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"attachment_ref":"Справочник.РеализацияТоваровУслугПрисоединенныеФайлы:aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","file_name":"scan.pdf"}`,
			heading: headingAttachmentMetaUpdate, wants: []string{businessOops},
		},
		{
			name: "attachment_metadata unsuccessful", site: `document_attachments_read_update.go "1C returned unsuccessful update result"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewUpdateDocumentAttachmentMetadataHandler(unsuccessful1C(t))
			},
			args:    `{"attachment_ref":"Справочник.РеализацияТоваровУслугПрисоединенныеФайлы:aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","file_name":"scan.pdf"}`,
			heading: headingAttachmentMetaUpdate, wants: []string{"1C returned unsuccessful update result"},
		},

		// incoming_goods.go
		{
			name: "incoming_goods refs required", site: `incoming_goods.go "organization_ref and counterparty_ref are required"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewCreateIncomingGoodsDocumentHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{}`,
			heading: headingIncomingGoodsCreate, wants: []string{"organization_ref and counterparty_ref are required"},
		},
		{
			name: "incoming_goods items required", site: `incoming_goods.go "items must contain at least one row"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewCreateIncomingGoodsDocumentHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"organization_ref":"a","counterparty_ref":"b"}`,
			heading: headingIncomingGoodsCreate, wants: []string{"items must contain at least one row"},
		},
		{
			name: "incoming_goods live post", site: `incoming_goods.go "creating incoming goods document in 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewCreateIncomingGoodsDocumentHandler(envelope1C(t, 500, businessOops))
			},
			args:    validIncoming,
			heading: headingIncomingGoodsCreate, wants: []string{businessOops},
		},
		{
			name: "incoming_goods unsuccessful", site: `incoming_goods.go "1C returned unsuccessful create result"`,
			build:   func(t *testing.T) mcp.ToolHandler { return NewCreateIncomingGoodsDocumentHandler(unsuccessful1C(t)) },
			args:    validIncoming,
			heading: headingIncomingGoodsCreate, wants: []string{"1C returned unsuccessful create result"},
		},

		// sales_documents.go
		{
			name: "read_sales_documents live fetch", site: `sales_documents.go "reading sales documents from 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewReadSalesDocumentsHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{}`,
			heading: headingSalesDocumentsRead, wants: []string{businessOops},
		},
		{
			name: "create_sales_document refs required", site: `sales_documents.go "organization_ref and counterparty_ref are required"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewCreateSalesDocumentHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{}`,
			heading: headingSalesDocumentCreate, wants: []string{"organization_ref and counterparty_ref are required"},
		},
		{
			name: "create_sales_document post without items", site: `sales_documents.go "return nil, err (validatePostItems)"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewCreateSalesDocumentHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"organization_ref":"a","counterparty_ref":"b","post":true}`,
			heading: headingSalesDocumentCreate, wants: []string{"items are required when post=true"},
		},
		{
			name: "create_sales_document live post", site: `sales_documents.go "creating sales document in 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewCreateSalesDocumentHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"organization_ref":"a","counterparty_ref":"b"}`,
			heading: headingSalesDocumentCreate, wants: []string{businessOops},
		},
		{
			name: "create_sales_document unsuccessful", site: `sales_documents.go "1C returned unsuccessful create result"`,
			build:   func(t *testing.T) mcp.ToolHandler { return NewCreateSalesDocumentHandler(unsuccessful1C(t)) },
			args:    `{"organization_ref":"a","counterparty_ref":"b"}`,
			heading: headingSalesDocumentCreate, wants: []string{"1C returned unsuccessful create result"},
		},
		{
			name: "update_sales_document ref required", site: `sales_documents.go "ref is required"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewUpdateSalesDocumentHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"comment":"уточнение"}`,
			heading: headingSalesDocumentUpdate, wants: []string{"ref is required"},
		},
		{
			name: "update_sales_document post with a zero price", site: `sales_documents.go "return nil, err (validatePostItems, update)"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewUpdateSalesDocumentHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"ref":"dddddddd-dddd-dddd-dddd-dddddddddddd","post":true,"items":[{"nomenclature_ref":"n1","quantity":1,"price":0}]}`,
			heading: headingSalesDocumentUpdate, wants: []string{"price must be > 0 when post=true"},
		},
		{
			name: "update_sales_document live post", site: `sales_documents.go "updating sales document in 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewUpdateSalesDocumentHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"ref":"dddddddd-dddd-dddd-dddd-dddddddddddd","comment":"уточнение"}`,
			heading: headingSalesDocumentUpdate, wants: []string{businessOops},
		},
		{
			name: "update_sales_document unsuccessful", site: `sales_documents.go "1C returned unsuccessful update result"`,
			build:   func(t *testing.T) mcp.ToolHandler { return NewUpdateSalesDocumentHandler(unsuccessful1C(t)) },
			args:    `{"ref":"dddddddd-dddd-dddd-dddd-dddddddddddd","comment":"уточнение"}`,
			heading: headingSalesDocumentUpdate, wants: []string{"1C returned unsuccessful update result"},
		},

		// sales_invoices.go
		{
			name: "read_sales_invoices live fetch", site: `sales_invoices.go "reading sales invoices from 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewReadSalesInvoicesHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{}`,
			heading: headingSalesInvoicesRead, wants: []string{businessOops},
		},
		{
			name: "create_sales_invoice refs required", site: `sales_invoices.go "organization_ref and counterparty_ref are required"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewCreateSalesInvoiceHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{}`,
			heading: headingSalesInvoiceCreate, wants: []string{"organization_ref and counterparty_ref are required"},
		},
		{
			name: "create_sales_invoice post without items", site: `sales_invoices.go "return nil, err (validatePostItems)"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewCreateSalesInvoiceHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"organization_ref":"a","counterparty_ref":"b","post":true}`,
			heading: headingSalesInvoiceCreate, wants: []string{"items are required when post=true"},
		},
		{
			name: "create_sales_invoice live post", site: `sales_invoices.go "creating sales invoice in 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewCreateSalesInvoiceHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"organization_ref":"a","counterparty_ref":"b"}`,
			heading: headingSalesInvoiceCreate, wants: []string{businessOops},
		},
		{
			name: "create_sales_invoice unsuccessful", site: `sales_invoices.go "1C returned unsuccessful create result"`,
			build:   func(t *testing.T) mcp.ToolHandler { return NewCreateSalesInvoiceHandler(unsuccessful1C(t)) },
			args:    `{"organization_ref":"a","counterparty_ref":"b"}`,
			heading: headingSalesInvoiceCreate, wants: []string{"1C returned unsuccessful create result"},
		},
		{
			name: "update_sales_invoice ref required", site: `sales_invoices.go "ref is required"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewUpdateSalesInvoiceHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"comment":"уточнение"}`,
			heading: headingSalesInvoiceUpdate, wants: []string{"ref is required"},
		},
		{
			name: "update_sales_invoice post with a zero price", site: `sales_invoices.go "return nil, err (validatePostItems, update)"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewUpdateSalesInvoiceHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"ref":"cccccccc-cccc-cccc-cccc-cccccccccccc","post":true,"items":[{"nomenclature_ref":"n1","quantity":1,"price":0}]}`,
			heading: headingSalesInvoiceUpdate, wants: []string{"price must be > 0 when post=true"},
		},
		{
			name: "update_sales_invoice live post", site: `sales_invoices.go "updating sales invoice in 1C"`,
			build: func(t *testing.T) mcp.ToolHandler {
				return NewUpdateSalesInvoiceHandler(envelope1C(t, 500, businessOops))
			},
			args:    `{"ref":"cccccccc-cccc-cccc-cccc-cccccccccccc","comment":"уточнение"}`,
			heading: headingSalesInvoiceUpdate, wants: []string{businessOops},
		},
		{
			name: "update_sales_invoice unsuccessful", site: `sales_invoices.go "1C returned unsuccessful update result"`,
			build:   func(t *testing.T) mcp.ToolHandler { return NewUpdateSalesInvoiceHandler(unsuccessful1C(t)) },
			args:    `{"ref":"cccccccc-cccc-cccc-cccc-cccccccccccc","comment":"уточнение"}`,
			heading: headingSalesInvoiceUpdate, wants: []string{"1C returned unsuccessful update result"},
		},
	}
}

// businessProtocolSites is one row per business constructor: a body that is not
// JSON never became a valid tool invocation, so it leaves as -32602 rather than
// as readable content.
func businessProtocolSites() []protocolSite {
	const notJSON = `not json`
	build := func(f func(*onec.Client) mcp.ToolHandler) func(t *testing.T) mcp.ToolHandler {
		return func(t *testing.T) mcp.ToolHandler { return f(envelope1C(t, 500, businessOops)) }
	}
	rows := []struct {
		name string
		site string
		ctor func(*onec.Client) mcp.ToolHandler
	}{
		{"read_contracts decode", `contracts.go "parsing input"`, NewReadContractsHandler},
		{"read_counterparties decode", `counterparties.go "parsing input"`, NewReadCounterpartiesHandler},
		{"create_counterparty decode", `counterparties.go "parsing input (create)"`, NewCreateCounterpartyHandler},
		{"read_nomenclature decode", `nomenclature_read.go "parsing input"`, NewReadNomenclatureHandler},
		{"create_nomenclature decode", `nomenclature.go "parsing input"`, NewCreateNomenclatureHandler},
		{"read_organizations decode", `organizations.go "parsing input"`, NewReadOrganizationsHandler},
		{"attach_file decode", `document_attachments.go "parsing input"`, NewAttachFileToDocumentHandler},
		{"read_attachments decode", `document_attachments_read_update.go "parsing input"`, NewReadDocumentAttachmentsHandler},
		{"attachment_content decode", `document_attachments_read_update.go "parsing input (content)"`, NewGetDocumentAttachmentContentHandler},
		{"attachment_metadata decode", `document_attachments_read_update.go "parsing input (metadata)"`, NewUpdateDocumentAttachmentMetadataHandler},
		{"incoming_goods decode", `incoming_goods.go "parsing input"`, NewCreateIncomingGoodsDocumentHandler},
		{"read_sales_documents decode", `sales_documents.go "parsing input"`, NewReadSalesDocumentsHandler},
		{"create_sales_document decode", `sales_documents.go "parsing input (create)"`, NewCreateSalesDocumentHandler},
		{"update_sales_document decode", `sales_documents.go "parsing input (update)"`, NewUpdateSalesDocumentHandler},
		{"read_sales_invoices decode", `sales_invoices.go "parsing input"`, NewReadSalesInvoicesHandler},
		{"create_sales_invoice decode", `sales_invoices.go "parsing input (create)"`, NewCreateSalesInvoiceHandler},
		{"update_sales_invoice decode", `sales_invoices.go "parsing input (update)"`, NewUpdateSalesInvoiceHandler},
	}
	out := make([]protocolSite, 0, len(rows))
	for _, r := range rows {
		out = append(out, protocolSite{
			name:  r.name,
			site:  r.site,
			build: build(r.ctor),
			args:  notJSON,
			code:  jsonrpc.CodeInvalidParams,
		})
	}
	return out
}

// TestBusinessUnsuccessfulStandInIsWhatItClaims is the control for
// unsuccessful1C: a stand in that answered anything else would make the
// «unsuccessful … result» rows above pass through a different site.
func TestBusinessUnsuccessfulStandInIsWhatItClaims(t *testing.T) {
	client := unsuccessful1C(t)
	var got struct {
		Success bool `json:"success"`
	}
	if err := client.Get(t.Context(), "/anything", &got); err != nil {
		t.Fatalf("the stand in must answer 200 with a decodable body: %v", err)
	}
	if got.Success {
		t.Fatal("the stand in reports success, so no row can reach the unsuccessful-result site")
	}
	raw, _ := json.Marshal(got)
	t.Logf("stand in answered %s", raw)
}
