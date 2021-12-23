package process

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jasontconnell/scexport/conf"
	"github.com/jasontconnell/sitecore/api"
)

type WriteSettings struct {
	ContentFormat   string
	ContentLocation string
	BlobLocation    string
}

type TemplateSettings struct {
	TemplateId uuid.UUID
	Name       string
	Fields     map[string]FieldSettings
	Path       string
}

type FieldSettings struct {
	Name       string
	Properties map[string]interface{}
	RefField   string
}

func GetSettings(cfg conf.ExportSettings) (map[uuid.UUID]TemplateSettings, error) {
	tsmap := map[uuid.UUID]TemplateSettings{}

	for _, tscfg := range cfg.Templates {
		id, err := api.TryParseUUID(tscfg.TemplateId)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse template id from settings. %w", tscfg.TemplateId)
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
			settings.Fields[fld.Name] = FieldSettings{Name: fld.Name, Properties: pmap, RefField: fld.RefField}
		}

		tsmap[id] = settings
	}

	return tsmap, nil
}
