package process

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jasontconnell/sitecore/api"
	"github.com/jasontconnell/sitecore/data"
)

func ReadAll(connstr string, settings map[uuid.UUID]TemplateSettings, globalTemplateFilter []string, lang data.Language) ([]data.ItemNode, data.ItemMap, error) {
	items, err := api.LoadItems(connstr)
	if err != nil {
		return nil, nil, fmt.Errorf("loading items %w", err)
	}

	_, m := api.LoadItemMap(items)
	tlist, err := api.LoadTemplates(connstr)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't load templates %w", err)
	}

	tm := api.GetTemplateMap(tlist)
	api.SetTemplates(m, tm)
	if globalTemplateFilter != nil {
		tm = api.FilterTemplateMap(tm, globalTemplateFilter)
	}

	globalFieldFilter := []uuid.UUID{}
	for _, t := range tm {
		for _, fld := range t.GetFields() {
			globalFieldFilter = append(globalFieldFilter, fld.GetId())
		}
	}

	fm := map[uuid.UUID]bool{}
	for tid, setting := range settings {
		tmp, ok := tm[tid]
		if !ok {
			continue
		}

		for _, fld := range setting.Fields {
			result := tmp.FindField(fld.Name)
			if result == nil {
				return nil, nil, fmt.Errorf("couldn't find %s in template %s %v", fld.Name, tmp.GetName(), tmp.GetId())
			}
			fm[result.GetId()] = true
		}
	}
	flds := []uuid.UUID{}
	for k := range fm {
		flds = append(flds, k)
	}

	fvlist, err := api.LoadFilteredFieldValues(connstr, globalFieldFilter, 20)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't load filtered field values. %w", err)
	}

	api.AssignFieldValues(m, fvlist)

	filtered := filterMap(m, settings)

	final := []data.ItemNode{}
	for _, item := range filtered {
		final = append(final, item)
	}

	sort.Slice(final, func(i, j int) bool {
		return final[i].GetName() < final[j].GetName()
	})

	return final, m, nil
}

func filterMap(m data.ItemMap, settings map[uuid.UUID]TemplateSettings) data.ItemMap {
	paths := []string{}
	tm := make(map[uuid.UUID]bool)
	for _, t := range settings {
		paths = append(paths, t.Paths...)
		tm[t.TemplateId] = true
	}

	nm := api.FilterItemMap(m, paths)
	nm = api.FilterItemMapCustom(nm, func(i data.ItemNode) bool {
		_, ok := tm[i.GetTemplateId()]
		return ok
	})

	return nm
}
