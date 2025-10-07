package import_export

import (
	"path/filepath"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbuilds/import_export/flags"
	import_export_csv "github.com/pocketbuilds/import_export/handlers/csv"
	import_export_json "github.com/pocketbuilds/import_export/handlers/json"
	import_export_toml "github.com/pocketbuilds/import_export/handlers/toml"
	import_export_yml "github.com/pocketbuilds/import_export/handlers/yml"
	"github.com/pocketbuilds/xpb"
	"github.com/spf13/cobra"
)

type Plugin struct {
	// Determines if an automatic database backup should be made prior to an import.
	//   - flag: auto_backup
	//   - default: true
	AutoBackup bool `json:"auto_backup"`
	// Path to directory for collections schema files.
	//   - flag: collections_dir
	//   - default: pb_data/../migrations/collections
	CollectionsDir string `json:"collections_dir"`
	// Encoding to use for collection imports and exports.
	//   - options: json, yml, toml, or any community plugin options installed
	//   - flag: --json, --yml, --toml, etc.
	//   - default: json
	CollectionsEncoding *flags.RadioValue `json:"collections_encoding"`
	// Optional prefix to prepend the commands to avoid possible name collisions.
	//   - default: "" (no prefix)
	CommandPrefix string `json:"command_prefix"`
	// Determines if to include oauth2 config in collections export.
	//   - flag: TODO
	//   - default: false
	IncludeOauth2 bool `json:"include_oauth2"`
	// Path to directory for records data files.
	//   - flag: records_dir
	//   - default: pb_data/../migrations/records
	RecordsDir string `json:"records_dir"`
	// Encoding to use for records imports and exports.
	//   - options: csv, json, yml, toml, or any community plugin options installed
	//   - flag: --csv, --json, --yml, --toml, etc.
	//   - default: csv
	RecordsEncoding *flags.RadioValue `json:"records_encoding"`
	// Determines if record imports should skip validation.
	//   - flag: no_validate
	//   - default: false
	NoValidate bool `json:"no_validate"`
	// Determines if verified state should be overriden.
	//   - options: true, false, null (do not override)
	//   - flag: override_verified
	//   - default: null
	OverrideVerified *flags.OptionalBoolValue `json:"override_verified"`
	// Determines if email visibility should be overriden.
	//   - options: true, false, null (do not override)
	//   - flag: override_email_visibility
	//   - default: null
	OverrideEmailVisibility *flags.OptionalBoolValue `json:"override_email_visibility"`
	// Determines if measures are taken to reduce git diff. Currently, just sets
	// updated to the zero datetime.
	//   - default: false
	ReduceGitDiff bool `json:"reduce_git_diff"`
	// Determines if to include system collections.
	//   - flag: system
	//   - default: false
	System bool `json:"system"`
}

func init() {
	xpb.Register(&Plugin{
		AutoBackup: true,
	})
	csvExt := &import_export_csv.Plugin{}
	xpb.Register(csvExt)
	RegisterHandler(csvExt)
	jsonExt := &import_export_json.Plugin{}
	xpb.Register(jsonExt)
	RegisterHandler(jsonExt)
	tomlExt := &import_export_toml.Plugin{}
	xpb.Register(tomlExt)
	RegisterHandler(tomlExt)
	ymlExt := &import_export_yml.Plugin{}
	xpb.Register(ymlExt)
	RegisterHandler(ymlExt)
}

// Name implements xpb.Plugin.
func (p *Plugin) Name() string {
	return "import_export"
}

// This variable will automatically be set at build time by xpb.
var version string

// Version implements xpb.Plugin.
func (p *Plugin) Version() string {
	return version
}

// Description implements xpb.Plugin.
func (p *Plugin) Description() string {
	return "CLI commands for importing and exporting records and collections"
}

// PreValidate implements xpb.PreValidator.
func (p *Plugin) PreValidate(app core.App) error {
	p.CollectionsDir = filepath.Join(app.DataDir(), "../migrations/collections")
	p.RecordsDir = filepath.Join(app.DataDir(), "../migrations/records")
	collectionHandlers := []string{}
	recordsHandlers := []string{}
	for k, v := range handlers {
		if _, ok := v.(CollectionHandler); ok {
			collectionHandlers = append(collectionHandlers, k)
		}
		if _, ok := v.(RecordsHandler); ok {
			recordsHandlers = append(recordsHandlers, k)
		}
	}
	p.CollectionsEncoding = flags.NewRadioValue(collectionHandlers...)
	p.CollectionsEncoding.Set("json")
	p.RecordsEncoding = flags.NewRadioValue(recordsHandlers...)
	p.RecordsEncoding.Set("csv")
	p.OverrideVerified = flags.NewOptionalBoolValue()
	p.OverrideEmailVisibility = flags.NewOptionalBoolValue()
	return nil
}

// Validate implements validation.Validatable.
func (p *Plugin) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.CollectionsEncoding, validation.Required),
		validation.Field(&p.RecordsEncoding, validation.Required),
	)
}

// Init implements xpb.Plugin.
func (p *Plugin) Init(app core.App) error {
	if app, ok := app.(*pocketbase.PocketBase); ok {
		rootCmd := app.RootCmd
		if p.CommandPrefix != "" {
			rootCmd = &cobra.Command{
				Use:   p.CommandPrefix,
				Short: p.Description(),
			}
			app.RootCmd.AddCommand(rootCmd)
		}
		rootCmd.AddCommand(p.ImportCommand(app))
		rootCmd.AddCommand(p.ExportCommand(app))
	}
	return nil
}

func (p *Plugin) ImportCommand(app core.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import records or collections",
	}
	cmd.AddCommand(p.ImportRecordsCommand(app))
	cmd.AddCommand(p.ImportCollectionsCommand(app))
	return cmd
}

func (p *Plugin) ExportCommand(app core.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export records or collections",
	}
	cmd.AddCommand(p.ExportRecordsCommand(app))
	cmd.AddCommand(p.ExportCollectionsCommand(app))
	return cmd
}
