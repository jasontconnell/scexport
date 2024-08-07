package process

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

func WriteContent(groups []Group, settings WriteSettings) error {
	fulldir := settings.ContentLocation
	err := os.MkdirAll(fulldir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("couldn't create dir structure %s. %w", fulldir, err)
	}

	isxml := settings.ContentFormat == "xml"
	for _, g := range groups {
		if isxml {
			path := filepath.Join(fulldir, g.Name+"."+settings.ContentFormat)
			err = writeContentXml(path, g)
			if err != nil {
				return fmt.Errorf("writing file contents for %s, path: %s. %w", g.Name, path, err)
			}
		}
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
	items := []ContentItem{}
	for _, item := range g.Items {
		x := ContentItem{ID: item.ID, TypeName: g.Name, Name: item.Name, Path: item.Path}

		xflds := []ContentField{}
		for _, f := range item.Fields {
			xf := ContentField{Name: f.Name}
			if f.CData {
				xf.Contents = f.Value
			} else {
				xf.Value = f.Value
			}
			for _, ref := range f.Refs {
				xref := ContentItem{ID: ref.ID, Name: ref.Name, Path: ref.Path}
				xrefflds := []ContentField{}
				for _, xreffld := range ref.Fields {
					if xreffld.Value != "" && xreffld.Name != "" {
						xrefflds = append(xrefflds, ContentField{Name: xreffld.Name, Value: xreffld.Value})
					}
				}
				if len(xrefflds) > 0 {
					xref.Fields = &xrefflds
				}
				xf.Refs = append(xf.Refs, xref)
			}
			xflds = append(xflds, xf)
		}

		if len(xflds) > 0 {
			x.Fields = &xflds
		}

		var bloblist []BlobRef
		for _, b := range item.Blobs {
			bref := BlobRef{ItemId: b.ItemId.String(), BlobId: b.BlobId.String(), Filename: b.Filename, Path: b.Path}
			bloblist = append(bloblist, bref)
		}
		if len(bloblist) > 0 {
			x.Blobs = &bloblist
		}

		items = append(items, x)
	}

	f, err := os.OpenFile(fullpath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("opening file for write %s. %w", fullpath, err)
	}
	defer f.Close()

	enc := xml.NewEncoder(f)
	enc.Indent(" ", " ")

	cxml := ContentsXml{ContentItems: items}
	return enc.Encode(cxml)
}
