package printer

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
)

const (
	JSONPrinter = "json"
	YAMLPrinter = "yaml"
)

func NewPrinter(serializer string, writer io.Writer) Printer {
	switch s := serializer; s {
	case JSONPrinter:
		return &jsonPrinter{writer: writer}
	case YAMLPrinter:
		return &yamlPrinter{writer: writer}
	default:
		return &unknownPrinter{writer: writer}
	}
}

type Printer interface {
	Print(v any) error
}

type unknownPrinter struct {
	writer io.Writer
}

func (dp *unknownPrinter) Print(v any) error {
	_, err := fmt.Fprint(dp.writer, v)
	return err

}

type jsonPrinter struct {
	writer io.Writer
}

func (jp *jsonPrinter) Print(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(jp.writer, string(data))
	return err

}

type yamlPrinter struct {
	writer io.Writer
}

func (yp *yamlPrinter) Print(v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(yp.writer, string(data))
	return err

}
