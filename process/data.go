package process

import (
	"github.com/google/uuid"
	"github.com/jasontconnell/sitecore/data"
)

type BlobData struct {
	Id       uuid.UUID
	Data     []byte
	Filename string
}

type Group struct {
	Name  string
	Items []Item
	Blobs []Blob
}

type Item struct {
	ID     string
	Name   string
	Path   string
	Fields []Field
	Blobs  []Blob
}

type Field struct {
	Name  string
	Value string
	CData bool
	Refs  []Item
}

type Blob struct {
	Id       uuid.UUID
	Filename string
	Attrs    []Attr
	Path     string
}

type Attr struct {
	Name, Value string
}

type WriteSettings struct {
	ContentFormat   string
	ContentLocation string
	BlobLocation    string
	WriteBlobs      bool
}

type Settings struct {
	Templates  map[uuid.UUID]TemplateSettings
	References map[uuid.UUID]TemplateSettings
}

type TemplateSettings struct {
	TemplateId uuid.UUID
	Name       string
	Fields     map[string]FieldSettings
	Paths      []string
}

type DataPackage struct {
	ReportItems []data.ItemNode
	Items       data.ItemMap
	RefItems    data.ItemMap
}

type FieldSettings struct {
	Name       string
	Alias      string
	RefField   string
	Properties map[string]interface{}
}

type RefTemplate struct {
	Name  string
	Field string
}
