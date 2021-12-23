package process

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jasontconnell/scexport/conf"
	"github.com/jasontconnell/sitecore/api"
	"github.com/jasontconnell/sitecore/data"
)

type Group struct {
	Name  string
	Items []Item
}

type Item struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name  string
	Value string
	CData bool
}

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

		gitem := Item{Name: item.GetName(), Fields: []Field{}}
		vfields := item.GetLatestVersionFields(lang)
		for _, fv := range vfields {
			fld := item.GetTemplate().GetField(fv.GetFieldId())
			if fld == nil {
				continue
			}
			fs, ok := tsettings.Fields[fld.GetName()]
			if !ok {
				// not exporting
				continue
			}

			gfld := Field{Name: fv.GetName()}
			result, err := ResolveField(fv, fld, item, m, fs, lang)
			if err != nil {
				log.Println(err)
			}
			gfld.Value = result.value
			gfld.CData = result.html

			gitem.Fields = append(gitem.Fields, gfld)
		}

		group.Items = append(group.Items, gitem)
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
