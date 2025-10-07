package import_export_yml

import (
	"encoding/json"
	"io"

	"github.com/pocketbase/pocketbase/core"
	"gopkg.in/yaml.v3"
)

type Plugin struct {
	// Number of spaces to use for indentation in yaml collection exports.
	//   - default: 2
	CollectionIndent int `json:"collection_indent"`
	// Number of spaces to use for indentation in yaml record exports.
	//   - default: 2
	RecordsIndent int `json:"records_indent"`
}

// Name implements xpb.Plugin.
func (p *Plugin) Name() string {
	return "import_export_yml"
}

// This variable will automatically be set at build time by xpb.
var version string

// Version implements xpb.Plugin.
func (p *Plugin) Version() string {
	return version
}

// Description implements xpb.Plugin.
func (p *Plugin) Description() string {
	return "yaml encoding extension for import_export"
}

// PreValidate implements xpb.PreValidator.
func (p *Plugin) PreValidate(app core.App) error {
	p.CollectionIndent = 2
	p.RecordsIndent = 2
	return nil
}

// Init implements xpb.Plugin.
func (p *Plugin) Init(app core.App) error {
	return nil
}

// FileExtension implements import_export.Handler.
func (p *Plugin) FileExtension() string {
	return "yml"
}

// DecodeRecords implements import_export.RecordsHandler.
func (p *Plugin) DecodeRecords(collection *core.Collection, reader io.Reader) ([]*core.Record, error) {
	var recordsData []map[string]any
	if err := yaml.NewDecoder(reader).Decode(&recordsData); err != nil {
		return nil, err
	}
	records := make([]*core.Record, 0, len(recordsData))
	for _, data := range recordsData {
		record := core.NewRecord(collection)
		record.Load(data)
		records = append(records, record)
	}
	return records, nil
}

// EncodeRecords implements import_export.RecordsHandler.
func (p *Plugin) EncodeRecords(records []*core.Record, writer io.Writer) error {
	var recordsData []map[string]any
	for _, record := range records {
		recordsData = append(recordsData, record.PublicExport())
	}
	encoder := yaml.NewEncoder(writer)
	encoder.SetIndent(p.RecordsIndent)
	return encoder.Encode(recordsData)
}

// DecodeCollection implements import_export.CollectionHandler.
func (p *Plugin) DecodeCollection(reader io.Reader) (map[string]any, error) {
	var collectionData map[string]any
	if err := yaml.NewDecoder(reader).Decode(&collectionData); err != nil {
		return nil, err
	}
	return collectionData, nil
}

// EncodeCollection implements import_export.CollectionHandler.
func (p *Plugin) EncodeCollection(collection *core.Collection, writer io.Writer) error {
	jsonBytes, err := json.Marshal(collection)
	if err != nil {
		return err
	}
	var collectionData map[string]any
	if err := json.Unmarshal(jsonBytes, &collectionData); err != nil {
		return err
	}
	encoder := yaml.NewEncoder(writer)
	encoder.SetIndent(p.CollectionIndent)
	return encoder.Encode(collectionData)
}
