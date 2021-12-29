package process

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

func WriteBlobs(dir string, bdata []BlobData, settings WriteSettings) error {
	fulldir := filepath.Join(dir, settings.BlobLocation)
	err := os.MkdirAll(fulldir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("couldn't create dir structure %s. %w", dir, err)
	}
	isxml := settings.ContentFormat == "xml"
	for _, b := range bdata {
		if isxml {
			bxml := BlobXml{Id: b.Id.String(), Filename: b.Filename, Length: len(b.Data), Data: base64.StdEncoding.EncodeToString(b.Data)}
			path := filepath.Join(fulldir, bxml.Filename+".xml")
			err = writeBlobXml(path, bxml)
			if err != nil {
				return fmt.Errorf("writing file contents for %s, path: %s. %w", b.Filename, path, err)
			}
		}
	}

	return nil
}

func WriteContent(dir string, groups []Group, settings WriteSettings) error {
	fulldir := filepath.Join(dir, settings.ContentLocation)
	err := os.MkdirAll(fulldir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("couldn't create dir structure %s. %w", dir, err)
	}

	isxml := settings.ContentFormat == "xml"
	for _, g := range groups {

		if isxml {
			path := filepath.Join(fulldir, g.Name+".xml")
			err = writeContentXml(path, g)
			if err != nil {
				return fmt.Errorf("writing file contents for %s, path: %s. %w", g.Name, path, err)
			}
		}

		// write blobs
	}
	return nil
}

func writeBlobXml(fullpath string, b BlobXml) error {
	f, err := os.OpenFile(fullpath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("opening file for write %s. %w", fullpath, err)
	}
	defer f.Close()
	enc := xml.NewEncoder(f)
	enc.Indent(" ", " ")

	return enc.Encode(b)
}

func writeContentXml(fullpath string, g Group) error {
	f, err := os.OpenFile(fullpath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("opening file for write %s. %w", fullpath, err)
	}
	defer f.Close()

	items := []ContentItem{}
	for _, item := range g.Items {
		x := ContentItem{TypeName: g.Name, Name: item.Name}
		x.Fields = []ContentField{}
		for _, f := range item.Fields {
			xf := ContentField{Name: f.Name}
			if f.CData {
				xf.Contents = f.Value
			} else {
				xf.Value = f.Value
			}
			x.Fields = append(x.Fields, xf)
		}

		for _, b := range item.Blobs {
			bref := BlobRef{Id: b.Id.String(), Filename: b.Filename}
			x.Blobs = append(x.Blobs, bref)
		}

		items = append(items, x)
	}

	enc := xml.NewEncoder(f)
	enc.Indent(" ", " ")

	cxml := ContentsXml{ContentItems: items}
	return enc.Encode(cxml)
}
