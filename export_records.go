package import_export

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func (p *Plugin) ExportRecordsCommand(app core.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "records",
		Short:   "export records csv files to the records directory",
		Aliases: []string{"record", "rec", "r"},
		Args:    cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&p.RecordsDir, "records_dir", p.RecordsDir, "Path to directory for records csv files")

	for _, opt := range p.RecordsEncoding.Options() {
		cmd.Flags().VarPF(p.RecordsEncoding, opt, "", fmt.Sprintf("%s encoding", opt)).NoOptDefVal = opt
	}
	cmd.MarkFlagsMutuallyExclusive(p.RecordsEncoding.Options()...)

	collectionNames := []string{}
	cmd.Flags().StringSliceVar(&collectionNames, "collection", collectionNames, "Collections to inlcude in the import, otherwise imports all")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// validate manually to catch changes to config due to cli flags
		if err := p.Validate(); err != nil {
			return err
		}

		encoder, ok := handlers[p.RecordsEncoding.String()].(RecordsHandler)
		if !ok {
			return ErrNoRecordsHandler
		}

		fmt.Printf("Set to encoding: %s\n", p.RecordsEncoding)

		allCollections := []*core.Collection{}
		collectionsQuery := app.CollectionQuery()
		if len(collectionNames) > 0 {
			collectionsQuery.AndWhere(dbx.In(
				"name", sliceToAnySlice(collectionNames)...,
			))
		}
		if !p.System {
			collectionsQuery.AndWhere(dbx.HashExp{
				"system": false,
			})
		}
		if err := collectionsQuery.All(&allCollections); err != nil {
			return err
		}

		if len(collectionNames) > 0 && len(allCollections) == 0 {
			return fmt.Errorf("collection(s) do not exist: %s", strings.Join(collectionNames, ", "))
		}

		if len(allCollections) == 0 {
			return fmt.Errorf("no collections to export records")
		}

		if len(collectionNames) > 0 && len(allCollections) != len(collectionNames) {
			notExisting := slices.Collect(func(yield func(string) bool) {
				for _, name := range collectionNames {
					if !slices.ContainsFunc(allCollections, func(c *core.Collection) bool {
						return c.Name == name
					}) {
						if !yield(name) {
							return
						}
					}
				}
			})
			return fmt.Errorf("collection(s) do not exist: %s", strings.Join(notExisting, ", "))
		}

		msg := strings.Join([]string{
			fmt.Sprintf(
				"Do you really want to export records from all collections to %q?",
				p.RecordsDir,
			),
			"Warning: This will delete all the contents of the directory!",
		}, "\n")

		if len(collectionNames) > 0 {
			msg = strings.Join([]string{
				fmt.Sprintf(
					"Do you really want to export records from the listed collections to %q?",
					p.RecordsDir,
				),
				fmt.Sprintf(
					"Collections: %s",
					strings.Join(collectionNames, ", "),
				),
			}, "\n")
		}

		if yes := confirm(msg, false); !yes {
			fmt.Println("The command has been cancelled.")
			return nil
		}

		if len(collectionNames) == 0 {
			if err := os.RemoveAll(p.RecordsDir); err != nil {
				return err
			}
		}

		if err := os.MkdirAll(p.RecordsDir, os.ModePerm); err != nil {
			return err
		}

		for _, collection := range allCollections {

			if collection.IsView() {
				continue
			}

			records, err := app.FindAllRecords(collection)
			if err != nil {
				return err
			}

			filename := fmt.Sprintf("%s.%s", collection.Name, p.RecordsEncoding)

			if err := func() (err error) {
				file, err := os.Create(filepath.Join(p.RecordsDir, filename))
				if err != nil {
					return err
				}
				defer func() {
					err = file.Close()
				}()
				return encoder.EncodeRecords(records, file)
			}(); err != nil {
				return err
			}
		}

		return nil
	}

	return cmd
}
