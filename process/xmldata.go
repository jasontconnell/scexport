package process

import "encoding/xml"

type BlobXml struct {
	XMLName  xml.Name `xml:"blob"`
	Id       string   `xml:"id,attr"`
	Filename string   `xml:"filename,attr"`
	Length   int      `xml:"length,attr"`
	Data     string   `xml:",chardata"`
}

type ContentsXml struct {
	XMLName      xml.Name      `xml:"items"`
	ContentItems []ContentItem `xml:"items"`
}

type ContentItem struct {
	XMLName  xml.Name        `xml:"item"`
	TypeName string          `xml:"type,attr,omitempty"`
	Name     string          `xml:"name,attr,omitempty"`
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
}
