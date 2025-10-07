package import_export

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func (p *Plugin) ImportCollectionsCommand(app core.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "collections",
		Short:   "import collections from json files in the collections directory",
		Aliases: []string{"collection", "col", "c"},
		Args:    cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&p.CollectionsDir, "collections_dir", p.CollectionsDir, "Path to directory for collections schema json files")
	cmd.Flags().BoolVar(&p.AutoBackup, "auto_backup", p.AutoBackup, "Make an automatic database backup before the import")

	for _, opt := range p.CollectionsEncoding.Options() {
		cmd.Flags().VarPF(p.CollectionsEncoding, opt, "", fmt.Sprintf("%s encoding", opt)).NoOptDefVal = opt
	}
	cmd.MarkFlagsMutuallyExclusive(p.CollectionsEncoding.Options()...)

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		// validate manually to catch changes to config due to cli flags
		if err := p.Validate(); err != nil {
			return err
		}

		decoder, ok := handlers[p.CollectionsEncoding.String()].(CollectionHandler)
		if !ok {
			return ErrNoCollectionHandler
		}

		fmt.Printf("Set to encoding: %s\n", p.CollectionsEncoding)

		if yes := confirm(
			fmt.Sprintf("Do you really want to import collections from %q", p.CollectionsDir),
			false,
		); !yes {
			fmt.Println("The command has been cancelled.")
			return nil
		}

		if p.AutoBackup {
			name := backupName("import_collections")
			fmt.Printf("Making backup %s\n", name)
			err := app.CreateBackup(cmd.Context(), name)
			if err != nil {
				return err
			}
		}

		collections := []map[string]any{}

		err = filepath.Walk(p.CollectionsDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil || info.IsDir() || filepath.Ext(path) != ".json" {
				return err
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}

			collection, err := decoder.DecodeCollection(file)
			if err != nil {
				return err
			}

			if !p.IncludeOauth2 {
				delete(collection, "oauth2") // don't write over oauth2 settings
			}

			collections = append(collections, collection)
			return nil
		})
		if err != nil {
			return err
		}

		if err := app.ImportCollections(collections, true); err != nil {
			return err
		}

		return nil
	}

	return cmd
}
