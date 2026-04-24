# Strict Contract v1 (`buh_3_0`)

## Scope
- Current contour only: `mcp-1c (buh_3_0)` + `okta-chat`.
- No legacy UUID-only refs in this contour.

## Typed Refs
- Format: `<metadataFullName>:<uuid>` (same rule for documents and catalogs).
- Attachment refs use the catalog’s full metadata name, for example `Справочник.<DocumentName>ПрисоединенныеФайлы:<uuid>`. The exact catalog name comes from `read_document_attachments` / attach responses (`ref` field).
- Input refs for tools and HTTP adapter must be typed refs. The former `attachmentCatalog:<uuid>` alias is not accepted.

## Attachment Content Contract
`get_document_attachment_content` returns:

```json
{
  "id": "Справочник.<ИмяКаталогаПрисоединенныхФайлов>:<uuid>",
  "name": "file.ext",
  "mime_type": "application/pdf",
  "size_bytes": 12345,
  "encoding": "base64",
  "content": "<base64>",
  "injection": {
    "mode": "inline_text | multimodal_image | file_reference"
  },
  "contract_version": "1.0"
}
```

## Routing Rules
- `inline_text` -> text content part.
- `multimodal_image` -> image content part.
- `file_reference` -> file/extractor pipeline.
