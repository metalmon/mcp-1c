# Strict Contract v1 (`buh_3_0`)

## Scope
- Current contour only: `mcp-1c (buh_3_0)` + `okta-chat`.
- No legacy UUID-only refs in this contour.

## Typed Refs
- Format: `<metadataFullName>:<uuid>`.
- Attachment format: `attachmentCatalog:<uuid>`.
- Input refs for tools and HTTP adapter must be typed refs.

## Attachment Content Contract
`get_document_attachment_content` returns:

```json
{
  "id": "attachmentCatalog:<uuid>",
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
