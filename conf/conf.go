package conf

import "github.com/jasontconnell/conf"

type Config struct {
	ConnectionString string `json:"connectionString"`
	ProtobufLocation string `json:"protobufLocation"`
}

type WriteSettings struct {
	ContentFormat   string `json:"contentFormat"`
	ContentLocation string `json:"contentLocation"`
	BlobLocation    string `json:"blobLocation"`
	WriteBlobs      bool   `json:"writeBlobs"`
}

type ExportSettings struct {
	FilterLanguage     string           `json:"filterLanguage"`
	Templates          []ExportTemplate `json:"templates"`
	ReferenceTemplates []ExportTemplate `json:"referenceTemplates"`
	Output             WriteSettings    `json:"output"`
}

type ExportTemplate struct {
	Name       string        `json:"name"`
	TemplateId string        `json:"templateId"`
	Paths      []string      `json:"paths"`
	Fields     []ExportField `json:"fields"`
}

type ExportField struct {
	Name       string                 `json:"name"`
	Alias      string                 `json:"alias"`
	RefField   string                 `json:"refTemplates"`
	Properties map[string]interface{} `json:"properties"`
}

func LoadConfig(fn string) (Config, error) {
	cfg := Config{}
	err := conf.LoadConfig(fn, &cfg)
	return cfg, err
}

func LoadExportSettings(fn string) (ExportSettings, error) {
	settings := ExportSettings{}
	err := conf.LoadConfig(fn, &settings)
	return settings, err
}
