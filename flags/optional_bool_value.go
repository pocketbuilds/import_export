package flags

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/spf13/cast"
)

type OptionalBoolValue struct {
	value *bool
}

func NewOptionalBoolValue() *OptionalBoolValue {
	return &OptionalBoolValue{}
}

func (v *OptionalBoolValue) String() string {
	return cast.ToString(v.value)
}

func (v *OptionalBoolValue) Set(val string) (err error) {
	v.value = new(bool)
	*v.value, err = cast.ToBoolE(val)
	return
}

func (v *OptionalBoolValue) GetValue() (bool, bool) {
	if v.value == nil {
		return false, false
	}
	return *v.value, true
}

func (v *OptionalBoolValue) Type() string {
	return "optional bool"
}

func (v *OptionalBoolValue) Value() driver.Value {
	if v.value != nil {
		return *v.value
	}
	return nil
}

func (v *OptionalBoolValue) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &v.value)
}

func (v *OptionalBoolValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}
