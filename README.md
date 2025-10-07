# Import Export

An [xpb](https://github.com/pocketbuilds/xpb) plugin for [Pocketbase](https://pocketbase.io/) that provides a suite of cli commands for importing and exporting records and collections.

## Installation

1. [Install XPB](https://docs.pocketbuilds.com/installing-xpb).
2. [Use the builder](https://docs.pocketbuilds.com/using-the-builder):

```sh
xpb build --with github.com/pocketbuilds/import_export@latest
```

## Plugin Config

```toml
# pocketbuilds.toml

[import_export]
# Determines if an automatic database backup should be made prior to an import.
#   - flag: auto_backup
#   - default: true
auto_backup = true
# Path to directory for collections schema files.
#   - flag: collections_dir
#   - default: pb_data/../migrations/collections
collections_dir = ""
# Encoding to use for collection imports and exports.
#   - options: json, yml, toml, or any community plugin options installed
#   - flag: --json, --yml, --toml, etc.
#   - default: json
collections_encoding = "json"
# Optional prefix to prepend the commands to avoid possible name collisions.
#   - default: "" (no prefix)
command_prefix = ""
# Determines if to include oauth2 config in collections export.
#   - flag: TODO
#   - default: false
include_oauth2 = false
# Path to directory for records data files.
#   - flag: records_dir
#   - default: pb_data/../migrations/records
records_dir = ""
# Encoding to use for records imports and exports.
#   - options: csv, json, yml, toml, or any community plugin options installed
#   - flag: --csv, --json, --yml, --toml, etc.
#   - default: csv
records_encoding = "csv"
# Determines if record imports should skip validation.
#   - flag: no_validate
#   - default: false
no_validate = false
# Determines if verified state should be overriden.
#   - options: true, false, null (do not override)
#   - flag: override_verified
#   - default: null
override_verified = true
# Determines if email visibility should be overriden.
#   - options: true, false, null (do not override)
#   - flag: override_email_visibility
#   - default: null
override_email_visibility = true
# Determines if measures are taken to reduce git diff. Currently, just sets
# updated to the zero datetime.
#   - default: false
reduce_git_diff = false
# Determines if to include system collections.
#   - flag: system
#   - default: false
system = false

[import_export_csv]
# Delimiter character to use for the csv.
#   - default: ","
delimiter = ","

[import_export_json]
# Indent prefix to be used for json collection exports.
#   - default: "" (no indent prefix)
collection_prefix = ""
# Indent to be used for json collection exports.
#   - default: "\t" (tab)
collection_indent = "\t"
# Indent prefix to be used for json record exports.
#   - default: "" (no indent prefix)
records_prefix = ""
# Indent to be used for json records exports.
#   - default: "" (no indent)
records_indent = ""

[import_export_toml]
# Indent to be used for toml collection exports.
#   - default: "  "
collection_indent = "  "
# Indent to be used for toml record exports.
#   - default: "  "
records_indent = "  "
# Key for records array, since toml cannot have root level arrays.
#   - default: "records"
records_array_key = "records"

[import_export_yml]
# Number of spaces to use for indentation in yaml collection exports.
#   - default: 2
collection_indent = 2
# Number of spaces to use for indentation in yaml record exports.
#   - default: 2
records_indent = 2
```

## Creating Community Encoding Handler
1. Look at the examples in handlers/ directory.
2. Create struct that implements the xpb.Plugin interface as well as the import_export.RecordsHandler and/or import_export.CollectionHandler interfaces.
3. Register the plugin and handler on `init()`:
```go
    func init() {
        myPlugin := &Plugin{}
        xpb.Register(myPlugin)
        import_export.RegisterHandler(myPlugin)
    }
```
