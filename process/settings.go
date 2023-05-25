package process

import (
	"github.com/google/uuid"
	"github.com/jasontconnell/scexport/conf"
	"github.com/jasontconnell/sitecore/api"
)

func GetSettings(cfg conf.ExportSettings) (Settings, error) {
	tsmap := make(map[uuid.UUID]TemplateSettings)

	rmap := make(map[uuid.UUID]TemplateSettings)
	for _, ref := range cfg.ReferenceTemplates {
		id := api.MustParseUUID(ref.TemplateId)
		r := TemplateSettings{Name: ref.Name, Paths: ref.Paths, TemplateId: id, Fields: getFieldSettingsMap(ref.Fields)}
		rmap[id] = r
	}

	for _, tscfg := range cfg.Templates {
		id := api.MustParseUUID(tscfg.TemplateId)

		settings := TemplateSettings{
			TemplateId: id,
			Name:       tscfg.Name,
			Paths:      tscfg.Paths,
			Fields:     getFieldSettingsMap(tscfg.Fields),
		}

		tsmap[id] = settings
	}

	return Settings{Templates: tsmap, References: rmap}, nil
}

func getFieldSettingsMap(list []conf.ExportField) map[string]FieldSettings {
	m := make(map[string]FieldSettings)
	for _, fld := range list {
		key := fld.Name
		if fld.Alias != "" {
			key += ":" + fld.Alias
		}

		m[key] = FieldSettings{
			Name:     fld.Name,
			Alias:    fld.Alias,
			RefField: fld.RefField,
		}
	}
	return m
}
