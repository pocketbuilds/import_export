package import_export_json

import (
	"encoding/json"
	"io"

	"github.com/pocketbase/pocketbase/core"
)

type Plugin struct {
	// Indent prefix to be used for json collection exports.
	//   - default: "" (no indent prefix)
	CollectionPrefix string `json:"collection_prefix"`
	// Indent to be used for json collection exports.
	//   - default: "\t" (tab)
	CollectionIndent string `json:"collection_indent"`
	// Indent prefix to be used for json record exports.
	//   - default: "" (no indent prefix)
	RecordsPrefix string `json:"records_prefix"`
	// Indent to be used for json records exports.
	//   - default: "" (no indent)
	RecordsIndent string `json:"records_indent"`
}

// Name implements xpb.Plugin.
func (p *Plugin) Name() string {
	return "import_export_json"
}

// This variable will automatically be set at build time by xpb.
var version string

// Version implements xpb.Plugin.
func (p *Plugin) Version() string {
	return version
}

// Description implements xpb.Plugin.
func (p *Plugin) Description() string {
	return "json encoding extension for import_export"
}

// PreValidate implements xpb.PreValidator.
func (p *Plugin) PreValidate(app core.App) error {
	p.CollectionPrefix = ""
	p.CollectionIndent = "\t"
	p.RecordsPrefix = ""
	p.RecordsIndent = ""
	return nil
}

// Init implements xpb.Plugin.
func (p *Plugin) Init(app core.App) error {
	return nil
}

// FileExtension implements import_export.Handler.
func (p *Plugin) FileExtension() string {
	return "json"
}

// DecodeRecords implements import_export.RecordsHandler.
func (p *Plugin) DecodeRecords(collection *core.Collection, reader io.Reader) ([]*core.Record, error) {
	var rawJsonRecords []json.RawMessage
	if err := json.NewDecoder(reader).Decode(&rawJsonRecords); err != nil {
		return nil, err
	}
	records := make([]*core.Record, 0, len(rawJsonRecords))
	for _, raw := range rawJsonRecords {
		record := core.NewRecord(collection)
		if err := json.Unmarshal(raw, &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

// EncodeRecords implements import_export.RecordsHandler.
func (p *Plugin) EncodeRecords(records []*core.Record, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	if p.RecordsPrefix != "" || p.RecordsIndent != "" {
		encoder.SetIndent(p.RecordsPrefix, p.RecordsIndent)
	}
	return encoder.Encode(records)
}

// DecodeCollection implements import_export.CollectionHandler.
func (p *Plugin) DecodeCollection(reader io.Reader) (map[string]any, error) {
	var collection map[string]any
	if err := json.NewDecoder(reader).Decode(&collection); err != nil {
		return nil, err
	}
	return collection, nil
}

// EncodeCollection implements import_export.CollectionHandler.
func (p *Plugin) EncodeCollection(collection *core.Collection, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	if p.CollectionPrefix != "" || p.CollectionIndent != "" {
		encoder.SetIndent(p.CollectionPrefix, p.CollectionIndent)
	}
	return encoder.Encode(collection)
}
