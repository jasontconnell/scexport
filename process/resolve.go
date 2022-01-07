package process

import (
	"fmt"
	"log"
	"sort"

	"github.com/google/uuid"
	"github.com/jasontconnell/scexport/conf"
	"github.com/jasontconnell/sitecore/api"
	"github.com/jasontconnell/sitecore/data"
)

func Resolve(items []data.ItemNode, m data.ItemMap, settings map[uuid.UUID]TemplateSettings, lang data.Language) ([]Group, error) {
	gmap := map[string]Group{}
	for _, item := range items {
		tsettings, ok := settings[item.GetTemplateId()]
		if !ok {
			continue
		}

		gkey := tsettings.Name
		var group Group
		if group, ok = gmap[gkey]; !ok {
			group = Group{Name: gkey}
		}

		item := resolveItem(item, m, tsettings, lang)
		group.Items = append(group.Items, item)
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

func getTemplateSettingsMap(settings conf.ExportSettings) (map[uuid.UUID]conf.ExportTemplate, error) {
	lookup := map[uuid.UUID]conf.ExportTemplate{}
	for _, t := range settings.Templates {
		uid, err := api.TryParseUUID(t.TemplateId)
		if err != nil {
			return nil, fmt.Errorf("can't parse template id %s. %w", t.TemplateId, err)
		}
		lookup[uid] = t
	}
	return lookup, nil
}

func resolveReferenceItem(item data.ItemNode, m data.ItemMap, fields []string, lang data.Language) Item {
	gitem := Item{Name: item.GetName(), Fields: []Field{}}
	for _, fn := range fields {
		itmp := item.GetTemplate()
		fld := itmp.FindField(fn)
		if fld == nil {
			continue
		}

		fv := item.GetFieldValue(fld.GetId(), lang)
		if fv == nil {
			// nil here means it has no value, it's not an error
			continue
		}

		fnm := fv.GetName()
		gfld := Field{Name: fnm}
		result, err := ResolveField(fv, fld, item, m, FieldSettings{}, lang)
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
			b := Blob{Id: blob.GetId(), Filename: blob.GetName() + "." + blob.GetExt()}
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

	return gitem
}

func resolveItem(item data.ItemNode, m data.ItemMap, tsetting TemplateSettings, lang data.Language) Item {
	gitem := Item{Name: item.GetName(), Fields: []Field{}}
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
		result, err := ResolveField(fv, fld, item, m, fs, lang)
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
			b := Blob{Id: blob.GetId(), Filename: blob.GetName() + "." + blob.GetExt()}
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
