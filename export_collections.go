package import_export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/spf13/cobra"
)

func (p *Plugin) ExportCollectionsCommand(app core.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "collections",
		Short:   "export collections to the collections directory",
		Aliases: []string{"collection", "col", "c"},
		Args:    cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&p.CollectionsDir, "collections_dir", p.CollectionsDir, "Path to directory for collections schema json files")
	cmd.Flags().BoolVar(&p.ReduceGitDiff, "reduce_git_diff", p.ReduceGitDiff, "Set updated to zero time to reduce git diff")

	for _, opt := range p.CollectionsEncoding.Options() {
		cmd.Flags().VarPF(p.CollectionsEncoding, opt, "", fmt.Sprintf("%s encoding", opt)).NoOptDefVal = opt
	}
	cmd.MarkFlagsMutuallyExclusive(p.CollectionsEncoding.Options()...)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// validate manually to catch changes to config due to cli flags
		if err := p.Validate(); err != nil {
			return err
		}

		encoder, ok := handlers[p.CollectionsEncoding.String()].(CollectionHandler)
		if !ok {
			return ErrNoCollectionHandler
		}

		fmt.Printf("Set to encoding: %s\n", p.CollectionsEncoding)

		msg := strings.Join([]string{
			fmt.Sprintf(
				"Do you really want to export all collections to %q?",
				p.CollectionsDir,
			),
			"Warning: This will delete all the contents of the directory!",
		}, "\n")

		if yes := confirm(msg, false); !yes {
			fmt.Println("The command has been cancelled.")
			return nil
		}

		if err := os.RemoveAll(p.CollectionsDir); err != nil {
			return err
		}

		if err := os.MkdirAll(p.CollectionsDir, os.ModePerm); err != nil {
			return err
		}

		collections := []*core.Collection{}
		if err := app.CollectionQuery().All(&collections); err != nil {
			return err
		}

		for _, collection := range collections {
			if !p.System && collection.System {
				continue
			}
			if !p.IncludeOauth2 && collection.IsAuth() {
				collection.OAuth2 = core.OAuth2Config{} // dont export oauth2 config
			}
			if p.ReduceGitDiff {
				collection.Updated = types.DateTime{} // set updated to zero value to reduce git diff
			}

			filepath := filepath.Join(p.CollectionsDir, fmt.Sprintf("%s.%s", collection.Name, encoder.FileExtension()))
			file, err := os.Create(filepath)
			if err != nil {
				return err
			}
			if err := func() (err error) {
				defer func() {
					err = file.Close()
				}()
				return encoder.EncodeCollection(collection, file)
			}(); err != nil {
				return err
			}
		}
		return nil
	}

	return cmd
}
