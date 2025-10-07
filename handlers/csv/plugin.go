package import_export_csv

import (
	"encoding/csv"
	"encoding/json"
	"io"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

type Plugin struct {
	// Delimiter character to use for the csv.
	//   - default: ","
	Delimiter string `json:"delimiter"`
}

// Name implements xpb.Plugin.
func (p *Plugin) Name() string {
	return "import_export_csv"
}

// This variable will automatically be set at build time by xpb.
var version string

// Version implements xpb.Plugin.
func (p *Plugin) Version() string {
	return version
}

// Description implements xpb.Plugin.
func (p *Plugin) Description() string {
	return "csv encoding extension for import_export"
}

// PreValidate implements xpb.PreValidator.
func (p *Plugin) PreValidate(app core.App) error {
	p.Delimiter = ","
	return nil
}

// Validate implements validation.Validatable.
func (p *Plugin) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.Delimiter,
			validation.Required,
			validation.Length(1, 1),
		),
	)
}

// Init implements xpb.Plugin.
func (p *Plugin) Init(app core.App) error {
	return nil
}

// FileExtension implements import_export.Handler.
func (p *Plugin) FileExtension() string {
	return "csv"
}

// DecodeRecords implements import_export.RecordsHandler.
func (p *Plugin) DecodeRecords(collection *core.Collection, reader io.Reader) ([]*core.Record, error) {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = rune(p.Delimiter[0])
	fields, err := csvReader.Read()
	if err != nil {
		return nil, err
	}
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}
	records := make([]*core.Record, 0, len(rows))
	for _, row := range rows {
		record := core.NewRecord(collection)
		for i, fieldName := range fields {
			field := collection.Fields.GetByName(fieldName)
			if field == nil {
				continue
			}
			var value any
			value = row[i]
			if value == "\"\"" {
				value = nil
			}
			switch field.Type() {
			case core.FieldTypeAutodate:
				date, err := types.ParseDateTime(value)
				if err != nil {
					break
				}
				record.SetRaw(fieldName, date)
			default:
				record.Set(fieldName, value)
			}
		}
		records = append(records, record)
	}
	return records, nil
}

// EncodeRecords implements import_export.RecordsHandler.
func (p *Plugin) EncodeRecords(records []*core.Record, writer io.Writer) error {
	if len(records) == 0 {
		return nil
	}

	collection := records[0].Collection()

	csvWriter := csv.NewWriter(writer)
	csvWriter.Comma = rune(p.Delimiter[0])
	defer csvWriter.Flush()

	fieldNames := []string{}
	for name, f := range collection.Fields.AsMap() {
		switch {
		case f.Type() == core.FieldTypePassword:
			continue
		case name == core.FieldNameTokenKey && collection.IsAuth():
			continue
		default:
			fieldNames = append(fieldNames, name)
		}
	}

	if err := csvWriter.Write(fieldNames); err != nil {
		return err
	}

	for _, record := range records {
		row := []string{}
		for _, field := range fieldNames {
			var value string
			switch v := record.Get(field).(type) {
			case string:
				value = v
			case types.DateTime:
				value = v.String()
			default:
				valueBytes, err := json.Marshal(v)
				if err != nil {
					return err
				}
				value = string(valueBytes)
			}
			row = append(row, value)
		}
		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}
	return nil
}
