package common

import (
	"github.com/feenlace/mcp-1c/onec"
	"github.com/feenlace/mcp-1c/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterReadTools(s *mcp.Server, onecClient *onec.Client) {
	// Язык запросов 1С («ВЫБРАТЬ») не умеет менять данные, инструмент помечен
	// ReadOnlyHint, и права ему даёт тот пользователь базы, под которым
	// подключились, — не список инструментов. Поэтому он законно живёт среди
	// читающих и остаётся доступен при --access read_only.
	//
	// Зачем он бизнес-набору: восемь остальных читающих отвечают на восемь
	// вопросов, а бухгалтер спрашивает что угодно — в том числе остатки по
	// складу, которых ни один именованный инструмент не знает. Описание
	// execute_query прямо учит виртуальным таблицам регистров
	// (РегистрНакопления.X.Остатки, .Обороты, РегистрСведений.X.СрезПоследних),
	// то есть один инструмент закрывает класс вопросов вместо того, чтобы
	// заводить новый под каждый. Решение владельца 2.11.2026.
	s.AddTool(tools.QueryTool(), tools.NewQueryHandler(onecClient))
	// Три спутника `execute_query`, без которых он почти бесполезен, и это не
	// рассуждение: его СОБСТВЕННОЕ описание велит «вызови validate_query перед
	// выполнением» и «имена полей бери из get_object_structure», а
	// get_metadata_tree отвечает на предшествующий вопрос — какие объекты в
	// этой базе вообще есть. Без них бизнес-пользователь получил инструмент,
	// чьим указаниям не может следовать.
	//
	// Все три помечены ReadOnlyHint, читают СТРУКТУРУ конфигурации, а не
	// данные, и не зависят ни от каталога выгрузки, ни от индекса — то есть
	// доступны там, где у человека ничего, кроме опубликованной базы.
	//
	// Граница набора: бизнес-пользователю достаётся ровно то, что делает
	// `execute_query` пригодным, и ничего больше. Журнал событий, поиск по
	// коду, справка по языку и структура форм остаются у разработчика.
	// Решение владельца 2.11.2026 («get_metadata_tree может тоже надо давать
	// во всех профилях, как execute_query?»).
	s.AddTool(tools.MetadataTool(), tools.NewMetadataHandler(onecClient))
	s.AddTool(tools.ObjectStructureTool(), tools.NewObjectStructureHandler(onecClient))
	s.AddTool(tools.ValidateQueryTool(), tools.NewValidateQueryHandler(onecClient))
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
