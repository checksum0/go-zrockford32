package zrockford32

import "flag"

type Value []byte

var _ flag.Value = (*Value)(nil)

func (v *Value) String() string {
	return StdEncoding.EncodeToString(*v)
}

func (v *Value) Set(s string) error {
	b, err := StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}

	*v = b
	return nil
}

var _ flag.Getter = (*Value)(nil)

func (v *Value) Get() interface{} {
	return []byte(*v)
}
