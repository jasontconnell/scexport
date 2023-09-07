package process

import (
	"fmt"
	"log"
	"sort"

	"github.com/google/uuid"
	"github.com/jasontconnell/sitecore/api"
	"github.com/jasontconnell/sitecore/data"
)

func ReadAll(connstr, protobufLocation string, settings Settings, lang data.Language) (*DataPackage, error) {
	templateIds := []uuid.UUID{}
	tfm := make(map[uuid.UUID]bool)
	for tid := range settings.References {
		tfm[tid] = true
		templateIds = append(templateIds, tid)
	}

	for tid := range settings.Templates {
		tfm[tid] = true
		templateIds = append(templateIds, tid)
	}

	var pitems []data.ItemNode
	if protobufLocation != "" {
		var perr error
		pitems, perr = api.ReadProtobuf(protobufLocation)
		if perr != nil {
			return nil, fmt.Errorf("can't read protobuf %w", perr)
		}
		log.Println("loaded items from protobuf", len(pitems))
	}

	log.Println("loading items")
	items, err := api.LoadItemsByTemplates(connstr, templateIds)
	if err != nil {
		return nil, fmt.Errorf("loading items %w", err)
	}

	if pitems != nil {
		items = append(items, pitems...)
	}

	log.Println("loaded", len(items), "items")
	_, m := api.LoadItemMap(items)

	log.Println("loading templates")
	tlist, err := api.LoadTemplates(connstr)
	if err != nil {
		return nil, fmt.Errorf("couldn't load templates %w", err)
	}
	log.Println("loaded", len(tlist), "templates")

	tm := api.GetTemplateMap(tlist)
	api.SetStandardValues(m, tm)
	api.SetTemplates(m, tm)

	log.Println("template filter map contains", len(tfm))
	filtered := api.FilterTemplateMapCustom(tm, func(t data.TemplateNode) bool {
		_, ok := tfm[t.GetId()]
		return ok
	})
	log.Println("filtered templates map contains", len(filtered))

	joined := make(map[uuid.UUID]TemplateSettings)
	for _, ts := range settings.Templates {
		joined[ts.TemplateId] = ts
	}

	for _, ts := range settings.References {
		joined[ts.TemplateId] = ts
	}

	fields := []uuid.UUID{}
	for _, stmp := range joined {
		t, ok := filtered[stmp.TemplateId]
		if !ok {
			return nil, fmt.Errorf("template %s not found. %v", stmp.Name, stmp.TemplateId)
		}

		for _, sfld := range stmp.Fields {
			if sfld.Name == ItemNameOutputField {
				continue
			}
			fld := t.FindField(sfld.Name)
			if fld == nil {
				return nil, fmt.Errorf("can't find field %s in template %s %v", sfld.Name, stmp.Name, stmp.TemplateId)
			}
			fields = append(fields, fld.GetId())
		}
	}

	// get file/media fields and create date
	fields = append(fields,
		data.DisplayNameFieldId,
		data.CreateDateFieldId,
		data.BlobFieldId,
		data.AltFieldId,
		data.ExtensionFieldId,
		data.MimeTypeFieldId,
		data.VersionedBlobFieldId,
		data.VersionedAltFieldId,
		data.VersionedExtensionFieldId,
		data.VersionedMimeTypeFieldId,
	)

	log.Println("loading field values with", len(fields), "fields")
	fvlist, err := api.LoadFieldValuesTemplates(connstr, fields, templateIds, 30)
	if err != nil {
		return nil, fmt.Errorf("couldn't load filtered field values. %w", err)
	}

	log.Println("loaded", len(fvlist), "field values")
	api.AssignFieldValues(m, fvlist)

	log.Println("filtering item map, current item count is", len(m))
	filteredItems := filterMap(m, settings.Templates)
	log.Println("filtered item map, new item count is", len(filteredItems))

	log.Println("filtering references, current item count is", len(m))
	filteredRefs := filterMap(m, settings.References)
	log.Println("filtered references map, new item count is", len(filteredRefs))

	reportItems := []data.ItemNode{}
	for _, item := range filteredItems {
		reportItems = append(reportItems, item)
	}

	sort.Slice(reportItems, func(i, j int) bool {
		return reportItems[i].GetName() < reportItems[j].GetName()
	})

	return &DataPackage{reportItems, filteredItems, filteredRefs}, nil
}

func filterMap(m data.ItemMap, tmps map[uuid.UUID]TemplateSettings) data.ItemMap {
	paths := []string{}
	tm := make(map[uuid.UUID]bool)
	for _, t := range tmps {
		paths = append(paths, t.Paths...)
		tm[t.TemplateId] = true
	}

	log.Println("filtering by paths", paths)

	nm := api.FilterItemMap(m, paths)
	nm = api.FilterItemMapCustom(nm, func(i data.ItemNode) bool {
		_, ok := tm[i.GetTemplateId()]
		return ok
	})

	return nm
}
