package process

import "github.com/google/uuid"

type BlobResult interface {
	GetBlobId() uuid.UUID
	GetItemId() uuid.UUID
	GetName() string
	GetAttrs() []Attr
	GetExt() string
	GetPath() string
}

type HandlerResult interface {
	GetId() string
	GetPath() string
	GetValue() string
	IsHtml() bool
	GetBlobs() []BlobResult
	HasMultiple() bool
	GetReferences() []Item
}

type blobResult struct {
	blobId uuid.UUID
	itemId uuid.UUID
	name   string
	attrs  []Attr
	ext    string
	path   string
}

type handlerResult struct {
	id    string
	path  string
	value string
	blobs []BlobResult
	html  bool
	refs  []Item
}

func (h handlerResult) GetId() string {
	return h.id
}

func (h handlerResult) GetPath() string {
	return h.path
}

func (h handlerResult) GetValue() string {
	return h.value
}

func (h handlerResult) IsHtml() bool {
	return h.html
}

func (h handlerResult) GetBlobs() []BlobResult {
	return h.blobs
}

func (h handlerResult) HasMultiple() bool {
	return len(h.refs) > 0
}

func (h handlerResult) GetReferences() []Item {
	return h.refs
}

func (b blobResult) GetBlobId() uuid.UUID {
	return b.blobId
}

func (b blobResult) GetItemId() uuid.UUID {
	return b.itemId
}

func (b blobResult) GetName() string {
	return b.name
}

func (b blobResult) GetAttrs() []Attr {
	return b.attrs
}
func (b blobResult) GetExt() string {
	return b.ext
}
func (b blobResult) GetPath() string {
	return b.path
}
