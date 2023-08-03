package process

import "encoding/xml"

type BlobXml struct {
	XMLName  xml.Name       `xml:"blob"`
	Id       string         `xml:"id,attr"`
	Filename string         `xml:"filename,attr"`
	Length   int            `xml:"length,attr"`
	Fields   []BlobFieldXml `xml:"fields>field,omitempty"`
	Data     BlobDataXml    `xml:"data"`
}

type BlobDataXml struct {
	XMLName xml.Name `xml:"data"`
	Data    string   `xml:",cdata"`
}

type BlobFieldXml struct {
	XMLName xml.Name `xml:"field"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

type ContentsXml struct {
	XMLName      xml.Name      `xml:"items"`
	ContentItems []ContentItem `xml:"item"`
}

type ContentItem struct {
	XMLName  xml.Name        `xml:"item"`
	ID       string          `xml:"id,attr"`
	TypeName string          `xml:"type,attr,omitempty"`
	Name     string          `xml:"name,attr,omitempty"`
	Path     string          `xml:"path,attr"`
	Fields   *[]ContentField `xml:"fields>field"`
	Blobs    *[]BlobRef      `xml:"blobrefs>blob,omitempty"`
}

type ContentField struct {
	XMLName  xml.Name      `xml:"field"`
	Name     string        `xml:"name,attr,omitempty"`
	Value    string        `xml:"value,attr,omitempty"`
	Contents string        `xml:",cdata"`
	Refs     []ContentItem `xml:"refs,omitempty"`
}

type BlobRef struct {
	XMLName  xml.Name `xml:"blob"`
	Id       string   `xml:"id,attr"`
	Filename string   `xml:"filename,attr"`
	Path     string   `xml:"path,attr"`
}
