package import_export_toml

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/BurntSushi/toml"
	"github.com/pocketbase/pocketbase/core"
)

type Plugin struct {
	// Indent to be used for toml collection exports.
	//   - default: "  "
	CollectionIndent string `json:"collection_indent"`
	// Indent to be used for toml record exports.
	//   - default: "  "
	RecordsIndent string `json:"records_indent"`
	// Key for records array, since toml cannot have root level arrays.
	//   - default: "records"
	RecordsArrayKey string `json:"records_array_key"`
}

// Name implements xpb.Plugin.
func (p *Plugin) Name() string {
	return "import_export_toml"
}

// This variable will automatically be set at build time by xpb.
var version string

// Version implements xpb.Plugin.
func (p *Plugin) Version() string {
	return version
}

// Description implements xpb.Plugin.
func (p *Plugin) Description() string {
	return "toml encoding extension for import_export"
}

// PreValidate implements xpb.PreValidator.
func (p *Plugin) PreValidate(app core.App) error {
	p.CollectionIndent = "  "
	p.RecordsIndent = "  "
	p.RecordsArrayKey = "records"
	return nil
}

// Init implements xpb.Plugin.
func (p *Plugin) Init(app core.App) error {
	return nil
}

// FileExtension implements import_export.Handler.
func (p *Plugin) FileExtension() string {
	return "toml"
}

// DecodeRecords implements import_export.RecordsHandler.
func (p *Plugin) DecodeRecords(collection *core.Collection, reader io.Reader) ([]*core.Record, error) {
	var tomlData map[string][]map[string]any
	if _, err := toml.NewDecoder(reader).Decode(&tomlData); err != nil {
		return nil, err
	}
	recordsData, ok := tomlData[p.RecordsArrayKey]
	if !ok {
		return nil, fmt.Errorf("toml records array key \"%s\" does not exist", p.RecordsArrayKey)
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
	jsonBytes, err := json.Marshal(records)
	if err != nil {
		return err
	}
	var recordsData []map[string]any
	if err := json.Unmarshal(jsonBytes, &recordsData); err != nil {
		return err
	}
	encoder := toml.NewEncoder(writer)
	encoder.Indent = p.RecordsIndent
	return encoder.Encode(map[string]any{
		p.RecordsArrayKey: recordsData,
	})
}

// DecodeCollection implements import_export.CollectionHandler.
func (p *Plugin) DecodeCollection(reader io.Reader) (map[string]any, error) {
	var collectionData map[string]any
	if _, err := toml.NewDecoder(reader).Decode(&collectionData); err != nil {
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
	encoder := toml.NewEncoder(writer)
	encoder.Indent = p.CollectionIndent
	return encoder.Encode(collectionData)
}
