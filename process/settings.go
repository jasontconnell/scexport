package process

import (
	"fmt"

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

	bsettings := BlobSettings{}
	for _, c := range cfg.BlobSettings.CustomFields {
		uid, err := uuid.Parse(c)
		if err != nil {
			return Settings{}, fmt.Errorf("couldn't parse uuid %s. %w", c, err)
		}
		bsettings.CustomFields = append(bsettings.CustomFields, uid)
	}

	return Settings{Templates: tsmap, References: rmap, BlobSettings: bsettings}, nil
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
