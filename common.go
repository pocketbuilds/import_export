package import_export

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/tools/inflector"
)

// borrowed from pocketbase to support older pocketbase versions
// https://github.com/pocketbase/pocketbase/blob/b1f1d19d7f0422a373c6de810f42376a7e62dfa4/tools/osutils/cmd.go#L40
func confirm(message string, fallback bool) bool {
	options := "Y/n"
	if !fallback {
		options = "y/N"
	}

	r := bufio.NewReader(os.Stdin)

	var s string
	for {
		fmt.Fprintf(os.Stderr, "%s (%s) ", message, options)

		s, _ = r.ReadString('\n')

		s = strings.ToLower(strings.TrimSpace(s))

		switch s {
		case "":
			return fallback
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
	}
}

func backupName(name string) string {
	return fmt.Sprintf(
		"%s_%s.zip",
		inflector.Snakecase(name),
		time.Now().UTC().Format("20060102150405"),
	)
}

func sliceToAnySlice[T any](s []T) []any {
	result := make([]any, 0, len(s))
	for _, v := range s {
		result = append(result, v)
	}
	return result
}
