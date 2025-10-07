package import_export

import (
	"io"

	"github.com/pocketbase/pocketbase/core"
)

type Handler interface {
	FileExtension() string
}

type RecordsHandler interface {
	Handler
	EncodeRecords(records []*core.Record, writer io.Writer) error
	DecodeRecords(collection *core.Collection, reader io.Reader) ([]*core.Record, error)
}

type CollectionHandler interface {
	Handler
	EncodeCollection(collection *core.Collection, writer io.Writer) error
	DecodeCollection(reader io.Reader) (map[string]any, error)
}

var handlers = map[string]Handler{}

func RegisterHandler(h Handler) {
	handlers[h.FileExtension()] = h
}
