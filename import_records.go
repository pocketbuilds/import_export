package import_export

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/spf13/cobra"
)

func (p *Plugin) ImportRecordsCommand(app core.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "records",
		Short:   "import records from csv files in the records directory",
		Aliases: []string{"record", "rec", "r"},
		Args:    cobra.ExactArgs(0),
	}

	collectionNames := []string{}
	var noDelete bool

	cmd.Flags().StringVar(&p.RecordsDir, "records_dir", p.RecordsDir, "Path to directory for records csv files")
	cmd.Flags().BoolVar(&p.AutoBackup, "auto_backup", p.AutoBackup, "Make an automatic database backup before the import")
	cmd.Flags().StringSliceVar(&collectionNames, "collection", collectionNames, "Collections to inlcude in the import, otherwise imports all")
	cmd.Flags().Var(p.OverrideVerified, "override_verified", "Determines override value of verfied state for auth records")
	cmd.Flags().Var(p.OverrideEmailVisibility, "override_email_visibility", "Determines override value of email visibility for auth records")
	cmd.Flags().BoolVar(&p.NoValidate, "no_validate", p.NoValidate, "Determines if record imports should skip validation")
	cmd.Flags().BoolVar(&noDelete, "no_delete", noDelete, "Determines if existing records should not be deleted")

	for _, opt := range p.RecordsEncoding.Options() {
		cmd.Flags().VarPF(p.RecordsEncoding, opt, "", fmt.Sprintf("%s encoding", opt)).NoOptDefVal = opt
	}
	cmd.MarkFlagsMutuallyExclusive(p.RecordsEncoding.Options()...)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// validate manually to catch changes to config due to cli flags
		if err := p.Validate(); err != nil {
			return err
		}

		decoder, ok := handlers[p.RecordsEncoding.String()].(RecordsHandler)
		if !ok {
			return ErrNoRecordsHandler
		}

		fmt.Printf("Set to encoding: %s\n", p.RecordsEncoding)

		if _, err := os.Stat(p.RecordsDir); err != nil {
			return err
		}

		msg := strings.Join([]string{
			fmt.Sprintf(
				"Do you really want to import records from data files in %q?",
				p.RecordsDir,
			),
		}, "\n")

		if len(collectionNames) > 0 {
			msg = strings.Join([]string{
				fmt.Sprintf(
					"Do you really want to import records to the listed collections to %q?",
					p.RecordsDir,
				),
				fmt.Sprintf(
					"Collections: %s",
					strings.Join(collectionNames, ", "),
				),
			}, "\n")
		}

		if !noDelete {
			msg += "\nWarning this will delete all current records in these collections!"
		}

		if yes := confirm(msg, false); !yes {
			fmt.Println("The command has been cancelled.")
			return nil
		}

		if p.AutoBackup {
			name := backupName("import_records")
			fmt.Printf("Making backup %s\n", name)
			err := app.CreateBackup(cmd.Context(), name)
			if err != nil {
				return err
			}
		}

		return filepath.Walk(p.RecordsDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil || info.IsDir() || filepath.Ext(path) != fmt.Sprintf(".%s", decoder.FileExtension()) {
				return err
			}

			collectionName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))

			if len(collectionNames) != 0 && !slices.Contains(collectionNames, collectionName) {
				return nil
			}

			collection, err := app.FindCollectionByNameOrId(collectionName)
			if err != nil {
				return err
			}

			if !noDelete {
				if _, err := app.DB().Delete(collectionName, nil).Execute(); err != nil {
					return err
				}
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			records, err := decoder.DecodeRecords(collection, file)
			if err != nil {
				return err
			}

			fmt.Printf(
				"Importing %d to collection %s.\n",
				len(records),
				collection.Name,
			)

			for _, record := range records {
				record.MarkAsNew()
				if record.Id == "" {
					record.Id, err = security.RandomStringByRegex(`[a-z0-9]{15}`)
					if err != nil {
						return err
					}
				}
				if record.Collection().IsAuth() {
					record.Set(core.FieldNamePassword, security.RandomString(30))
					record.RefreshTokenKey()
					if raw, ok := record.GetRaw(core.FieldNamePassword).(*core.PasswordFieldValue); ok {
						raw.Plain = ""
					}
					if verified, ok := p.OverrideVerified.GetValue(); ok {
						record.SetVerified(verified)
					}
					if visibility, ok := p.OverrideEmailVisibility.GetValue(); ok {
						record.SetEmailVisibility(visibility)
					}
				}
				if p.NoValidate {
					if err := app.SaveNoValidate(record); err != nil {
						return err
					}
				} else {
					if err := app.Save(record); err != nil {
						return err
					}
				}
			}

			return nil
		})
	}

	return cmd
}
