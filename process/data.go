package process

import (
	"github.com/google/uuid"
)

type BlobData struct {
	Id       uuid.UUID
	Data     []byte
	Filename string
}

type BlobResult interface {
	GetId() uuid.UUID
	GetName() string
	GetAttrs() []Attr
	GetExt() string
}

type HandlerResult interface {
	GetValue() string
	GetBlobs() []BlobResult
}

type HandlerMultiResult interface {
	GetResults() []HandlerResult
}

type Group struct {
	Name  string
	Items []Item
	Blobs []Blob
}

type Item struct {
	Name   string
	Fields []Field
	Blobs  []Blob
}

type Field struct {
	Name  string
	Value string
	CData bool
}

type Blob struct {
	Id       uuid.UUID
	Filename string
	Attrs    []Attr
}

type Attr struct {
	Name, Value string
}

type WriteSettings struct {
	ContentFormat   string
	ContentLocation string
	BlobLocation    string
}

type TemplateSettings struct {
	TemplateId uuid.UUID
	Name       string
	Fields     map[string]FieldSettings
	Path       string
}

type FieldSettings struct {
	Name       string
	Properties map[string]interface{}
	RefField   string
	Alias      string
}
