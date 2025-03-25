package mappers

import (
	"errors"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// Mapper is a generic wrapper around the mapstructure library.
type Mapper struct {
	decoderConfig *mapstructure.DecoderConfig
}

// NewMapper creates a new Mapper instance with default configurations.
func NewMapper() *Mapper {
	return &Mapper{
		decoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				DecodeTimeHookFunc(), // Add time decoding support
			),
			TagName: "json", // Default tag to use for mapping
			Result:  nil,    // Result will be set during decoding
		},
	}
}

// Decode decodes a map into a given struct or object.
func (m *Mapper) Decode(input interface{}, output interface{}) error {
	if reflect.ValueOf(output).Kind() != reflect.Ptr {
		return errors.New("output must be a pointer to a struct")
	}

	m.decoderConfig.Result = output
	decoder, err := mapstructure.NewDecoder(m.decoderConfig)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

// DecodeWithCustomHook decodes a map into a struct with additional custom hooks.
func (m *Mapper) DecodeWithCustomHook(input interface{}, output interface{}, customHooks ...mapstructure.DecodeHookFunc) error {
	if reflect.ValueOf(output).Kind() != reflect.Ptr {
		return errors.New("output must be a pointer to a struct")
	}

	// allHooks := append([]mapstructure.DecodeHookFunc{DecodeTimeHookFunc()}, customHooks...)
	m.decoderConfig.DecodeHook = mapstructure.ComposeDecodeHookFunc(customHooks...)
	m.decoderConfig.Result = output

	decoder, err := mapstructure.NewDecoder(m.decoderConfig)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}
