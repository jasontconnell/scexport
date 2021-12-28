package process

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jasontconnell/scexport/conf"
	"github.com/jasontconnell/sitecore/api"
)

func GetSettings(cfg conf.ExportSettings) (map[uuid.UUID]TemplateSettings, error) {
	tsmap := map[uuid.UUID]TemplateSettings{}

	for _, tscfg := range cfg.Templates {
		id, err := api.TryParseUUID(tscfg.TemplateId)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse template id from settings. %v %w", tscfg.TemplateId, err)
		}

		settings := TemplateSettings{
			TemplateId: id,
			Name:       tscfg.Name,
			Path:       tscfg.Path,
			Fields:     map[string]FieldSettings{},
		}

		for _, fld := range tscfg.Fields {
			pmap := map[string]interface{}{}
			for k, v := range fld.Properties {
				pmap[k] = v
			}
			key := fld.Name
			if fld.Alias != "" {
				key += ":" + fld.Alias
			}
			settings.Fields[key] = FieldSettings{Name: fld.Name, Alias: fld.Alias, Properties: pmap, RefField: fld.RefField}
		}

		tsmap[id] = settings
	}

	return tsmap, nil
}
