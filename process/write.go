package process

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func Write(dir string, groups []Group, settings WriteSettings) error {
	fulldir := filepath.Join(dir, settings.ContentLocation)
	err := os.MkdirAll(fulldir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("couldn't create dir structure %s. %w", dir, err)
	}

	isxml := settings.ContentFormat == "xml"
	for _, g := range groups {

		if isxml {
			path := filepath.Join(fulldir, g.Name+".xml")
			err = writeXml(path, g)
			if err != nil {
				return fmt.Errorf("writing file contents for %s, path: %s. %w", g.Name, path, err)
			}
		}
	}
	return nil
}

func writeXml(fullpath string, g Group) error {
	f, err := os.OpenFile(fullpath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("opening file for write %s. %w", fullpath, err)
	}
	enc := xml.NewEncoder(f)
	enc.Indent(" ", " ")
	// head := xml.
	// err = enc.EncodeToken(xml.Header)
	// if err != nil {
	// 	return fmt.Errorf("couldn't encode xml header %w", err)
	// }

	start := xml.StartElement{Name: xml.Name{Local: "items"}}
	err = enc.EncodeToken(start)
	if err != nil {
		return fmt.Errorf("couldn't encode start root %v %w", start, err)
	}

	for _, item := range g.Items {
		selem := xml.StartElement{Name: xml.Name{Local: "item"}}
		selem.Attr = append(selem.Attr, xml.Attr{Name: xml.Name{Local: "name"}, Value: item.Name})
		enc.EncodeToken(selem)

		for _, fld := range item.Fields {
			cleanName := strip(fld.Name)
			fsel := xml.StartElement{Name: xml.Name{Local: cleanName}}
			err = enc.EncodeToken(fsel)
			if err != nil {
				return fmt.Errorf("encoding start field element %s <%s>. %w", fld.Name, cleanName, err)
			}

			// if fld.CData {
			// 	cd := xml.CharData("<![CDATA[")
			// 	err = enc.EncodeToken(cd)
			// 	log.Println(err)
			// }
			v := xml.CharData(fld.Value)
			err = enc.EncodeToken(v)

			// if fld.CData {
			// 	cd := xml.CharData("]]>")
			// 	err = enc.EncodeToken(cd)
			// 	log.Println(err)
			// }

			if err != nil {
				return fmt.Errorf("encoding char data. %w", err)
			}

			feel := xml.EndElement{Name: xml.Name{Local: cleanName}}
			enc.EncodeToken(feel)
		}

		eelem := xml.EndElement{Name: xml.Name{Local: "item"}}
		enc.EncodeToken(eelem)
	}

	end := xml.EndElement{Name: xml.Name{Local: "items"}}
	enc.EncodeToken(end)

	return enc.Flush()
}

var stripreg *regexp.Regexp = regexp.MustCompile("[^a-zA-Z0-9\\-]+")

func strip(str string) string {
	return stripreg.ReplaceAllString(str, "")
}
