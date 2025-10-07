package import_export

import "errors"

var (
	ErrNoCollectionHandler = errors.New("no collection encoding handler was installed")
	ErrNoRecordsHandler    = errors.New("no records encoding handler was installed")
)
