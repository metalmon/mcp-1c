package onec

import (
	"net/url"
	"strconv"
)

func querySet(q url.Values, key, val string) {
	if val != "" {
		q.Set(key, val)
	}
}

func querySetInt(q url.Values, key string, val int) {
	if val > 0 {
		q.Set(key, strconv.Itoa(val))
	}
}

// NomenclatureQuery builds GET query for /nomenclatures.
func NomenclatureQuery(r ReadNomenclatureRequest) url.Values {
	q := url.Values{}
	querySet(q, "search", r.Search)
	querySetInt(q, "limit", r.Limit)
	querySet(q, "code", r.Code)
	querySet(q, "ref", r.Ref)
	querySet(q, "article", r.Article)
	querySet(q, "after_code", r.AfterCode)
	return q
}

// CounterpartiesQuery builds GET query for /counterparties.
func CounterpartiesQuery(r ReadCounterpartiesRequest) url.Values {
	q := url.Values{}
	querySet(q, "search", r.Search)
	querySetInt(q, "limit", r.Limit)
	querySet(q, "code", r.Code)
	querySet(q, "ref", r.Ref)
	querySet(q, "inn", r.INN)
	querySet(q, "kpp", r.KPP)
	querySet(q, "after_code", r.AfterCode)
	return q
}

// OrganizationsQuery builds GET query for /organizations.
func OrganizationsQuery(r ReadOrganizationsRequest) url.Values {
	q := url.Values{}
	querySet(q, "search", r.Search)
	querySetInt(q, "limit", r.Limit)
	querySet(q, "code", r.Code)
	querySet(q, "ref", r.Ref)
	querySet(q, "inn", r.INN)
	querySet(q, "kpp", r.KPP)
	querySet(q, "after_code", r.AfterCode)
	return q
}

// ContractsQuery builds GET query for /contracts.
func ContractsQuery(r ReadContractsRequest) url.Values {
	q := url.Values{}
	querySet(q, "search", r.Search)
	querySetInt(q, "limit", r.Limit)
	querySet(q, "code", r.Code)
	querySet(q, "ref", r.Ref)
	querySet(q, "number", r.Number)
	querySet(q, "counterparty_ref", r.CounterpartyRef)
	querySet(q, "organization_ref", r.OrganizationRef)
	return q
}

// ReadSalesDocumentsQuery builds GET query for sales list endpoints.
func ReadSalesDocumentsQuery(r ReadSalesDocumentsRequest) url.Values {
	q := url.Values{}
	querySet(q, "search", r.Search)
	querySetInt(q, "limit", r.Limit)
	querySet(q, "ref", r.Ref)
	querySet(q, "number", r.Number)
	querySet(q, "counterparty_ref", r.CounterpartyRef)
	querySet(q, "organization_ref", r.OrganizationRef)
	return q
}

// DocumentAttachmentsQuery builds GET query for /document-attachments.
func DocumentAttachmentsQuery(r ReadDocumentAttachmentsRequest) url.Values {
	q := url.Values{}
	querySet(q, "document_ref", r.DocumentRef)
	querySetInt(q, "limit", r.Limit)
	querySet(q, "search", r.Search)
	return q
}

// DocumentAttachmentContentQuery builds GET query for /document-attachment-content.
func DocumentAttachmentContentQuery(r GetDocumentAttachmentContentRequest) url.Values {
	q := url.Values{}
	querySet(q, "attachment_ref", r.AttachmentRef)
	return q
}
