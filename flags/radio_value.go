package flags

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type RadioValue struct {
	value   string
	options []string
}

func NewRadioValue(options ...string) *RadioValue {
	rv := &RadioValue{}
	if len(options) > 0 {
		rv.value = options[0]
	}
	slices.Sort(options)
	rv.options = options
	return rv
}

func (rv *RadioValue) String() string {
	return string(rv.value)
}

func (rv *RadioValue) Set(v string) error {
	rv.value = v
	return nil
}

func (rv *RadioValue) Type() string {
	return "radio"
}

func (rv *RadioValue) Options() []string {
	return rv.options
}

func (rv *RadioValue) Validate() error {
	_, found := slices.BinarySearch(rv.options, rv.value)
	if !found {
		return validation.ErrInInvalid.SetMessage(
			fmt.Sprintf(
				"Must be one of the following values: %s",
				strings.Join(rv.options, ", "),
			),
		)
	}
	return nil
}

func (rv *RadioValue) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &rv.value)
}

func (rv *RadioValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(rv.value)
}
