package common

import (
	"github.com/feenlace/mcp-1c/onec"
	"github.com/feenlace/mcp-1c/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterReadTools(s *mcp.Server, onecClient *onec.Client) {
	s.AddTool(tools.ReadCounterpartiesTool(), tools.NewReadCounterpartiesHandler(onecClient))
	s.AddTool(tools.ReadNomenclatureTool(), tools.NewReadNomenclatureHandler(onecClient))
	s.AddTool(tools.ReadOrganizationsTool(), tools.NewReadOrganizationsHandler(onecClient))
	s.AddTool(tools.ReadContractsTool(), tools.NewReadContractsHandler(onecClient))
	s.AddTool(tools.ReadSalesInvoicesTool(), tools.NewReadSalesInvoicesHandler(onecClient))
	s.AddTool(tools.ReadSalesDocumentsTool(), tools.NewReadSalesDocumentsHandler(onecClient))
	s.AddTool(tools.ReadDocumentAttachmentsTool(), tools.NewReadDocumentAttachmentsHandler(onecClient))
	s.AddTool(tools.GetDocumentAttachmentContentTool(), tools.NewGetDocumentAttachmentContentHandler(onecClient))
}

func RegisterWriteTools(s *mcp.Server, onecClient *onec.Client) {
	s.AddTool(tools.CreateCounterpartyTool(), tools.NewCreateCounterpartyHandler(onecClient))
	s.AddTool(tools.CreateNomenclatureTool(), tools.NewCreateNomenclatureHandler(onecClient))
	s.AddTool(tools.CreateSalesInvoiceTool(), tools.NewCreateSalesInvoiceHandler(onecClient))
	s.AddTool(tools.UpdateSalesInvoiceTool(), tools.NewUpdateSalesInvoiceHandler(onecClient))
	s.AddTool(tools.CreateSalesDocumentTool(), tools.NewCreateSalesDocumentHandler(onecClient))
	s.AddTool(tools.UpdateSalesDocumentTool(), tools.NewUpdateSalesDocumentHandler(onecClient))
	s.AddTool(tools.CreateIncomingGoodsDocumentTool(), tools.NewCreateIncomingGoodsDocumentHandler(onecClient))
	s.AddTool(tools.AttachFileToDocumentTool(), tools.NewAttachFileToDocumentHandler(onecClient))
	s.AddTool(tools.UpdateDocumentAttachmentMetadataTool(), tools.NewUpdateDocumentAttachmentMetadataHandler(onecClient))
}
