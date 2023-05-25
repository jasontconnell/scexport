package process

import (
	"fmt"
	"log"
	"sort"

	"github.com/jasontconnell/sitecore/data"
)

func Resolve(pkg *DataPackage, settings Settings, lang data.Language) ([]Group, error) {
	gmap := map[string]Group{}
	for _, item := range pkg.ReportItems {
		tsettings, ok := settings.Templates[item.GetTemplateId()]
		if !ok {
			continue
		}

		gkey := tsettings.Name
		var group Group
		if group, ok = gmap[gkey]; !ok {
			group = Group{Name: gkey}
		}

		item := resolveItem(item, pkg, tsettings, lang)
		for _, b := range item.Blobs {
			group.Blobs = append(group.Blobs, b)
		}

		group.Items = append(group.Items, item)
		gmap[group.Name] = group
	}

	groups := []Group{}
	for _, g := range gmap {
		groups = append(groups, g)
	}
	return groups, nil
}

func resolveReferenceItem(item data.ItemNode, pkg *DataPackage, field string, lang data.Language) (Item, error) {
	if item == nil {
		return Item{}, fmt.Errorf("item is nil")
	}
	gitem := Item{ID: item.GetId().String(), Name: item.GetName(), Path: item.GetPath(), Fields: []Field{}}

	var err error
	if field != ItemNameField {
		itmp := item.GetTemplate()
		fld := itmp.FindField(field)
		if fld == nil {
			return gitem, fmt.Errorf("field not in template %s %v item id: %v", field, item.GetTemplateId(), item.GetId())
		}

		fv := item.GetFieldValue(fld.GetId(), lang)
		if fv == nil {
			return gitem, nil
		}

		fnm := fv.GetName()
		gfld := Field{Name: fnm}
		result, err := ResolveField(fv, fld, item, pkg, FieldSettings{}, lang)
		if err != nil {
			shortval := fv.GetValue()
			if len(shortval) > 100 {
				shortval = string(shortval[:99]) + "..."
			}
			return gitem, fmt.Errorf("couldn't resolve field %v (id: %v) %v (id: %v) Language: %v Field Value: %v (root cause: %v)\n", item.GetName(), item.GetId(), fv.GetName(), fv.GetFieldId(), lang, shortval, err)
		}

		gfld.Value = result.GetValue()
		gfld.CData = result.IsHtml()

		for _, blob := range result.GetBlobs() {
			b := Blob{Id: blob.GetId(), Filename: blob.GetName() + "." + blob.GetExt(), Path: blob.GetPath()}
			for _, attr := range blob.GetAttrs() {
				b.Attrs = append(b.Attrs, Attr{Name: attr.Name, Value: attr.Value})
			}
			gitem.Blobs = append(gitem.Blobs, b)
		}

		gitem.Fields = append(gitem.Fields, gfld)
	}

	sort.Slice(gitem.Fields, func(i, j int) bool {
		return gitem.Fields[i].Name < gitem.Fields[j].Name
	})

	return gitem, err
}

func resolveItem(item data.ItemNode, pkg *DataPackage, tsetting TemplateSettings, lang data.Language) Item {
	gitem := Item{ID: item.GetId().String(), Name: item.GetName(), Path: item.GetPath(), Fields: []Field{}}
	for _, fs := range tsetting.Fields {
		itmp := item.GetTemplate()
		fld := itmp.FindField(fs.Name)
		if fld == nil {
			continue
		}

		fv := item.GetFieldValue(fld.GetId(), lang)
		if fv == nil {
			// nil here means it has no value, it's not an error
			continue
		}

		fnm := fv.GetName()
		if fs.Alias != "" {
			fnm = fs.Alias
		}
		gfld := Field{Name: fnm}
		result, err := ResolveField(fv, fld, item, pkg, fs, lang)
		if err != nil {
			shortval := fv.GetValue()
			if len(shortval) > 100 {
				shortval = string(shortval[:99]) + "..."
			}
			log.Printf("couldn't resolve field %v (id: %v) %v (id: %v) Language: %v Field Value: %v (root cause: %v)\n", item.GetName(), item.GetId(), fv.GetName(), fv.GetFieldId(), lang, shortval, err)
			continue
		}
		gfld.Value = result.GetValue()
		gfld.CData = result.IsHtml()

		for _, blob := range result.GetBlobs() {
			b := Blob{Id: blob.GetId(), Filename: blob.GetName() + "." + blob.GetExt(), Path: blob.GetPath()}
			for _, attr := range blob.GetAttrs() {
				b.Attrs = append(b.Attrs, Attr{Name: attr.Name, Value: attr.Value})
			}
			gitem.Blobs = append(gitem.Blobs, b)
		}

		for _, ref := range result.GetReferences() {
			gfld.Refs = append(gfld.Refs, ref)
		}

		gitem.Fields = append(gitem.Fields, gfld)
	}

	sort.Slice(gitem.Fields, func(i, j int) bool {
		return gitem.Fields[i].Name < gitem.Fields[j].Name
	})

	return gitem
}
