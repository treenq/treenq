package gen

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"text/template"
	"unicode"

	"github.com/treenq/treenq/pkg/vel"
)

var ErrStructRequired = errors.New("api input/output must be only struct")

// ClientGen defines api client generator
// it doesn't support the following:
// - anonymous nested struct
type ClientGen struct {
	meta ApiClientDesc
}

//go:embed templates/client.tpl
var clientTemplate string

//go:embed templates/call.tpl
var callTemplate string

func New(clientDesc ClientDesc, meta []vel.HandlerMeta) (*ClientGen, error) {
	desc := make([]ApiDesc, len(meta))
	dataTypeSet := make(map[string]struct{}, len(meta)*2)

	var err error
	for i := range meta {
		dataTypes := make([]DataType, 0, len(meta)*2)
		desc[i], err = MakeApiDesc(meta[i])
		if err != nil {
			return nil, err
		}

		name := desc[i].Input.Name
		if _, ok := dataTypeSet[name]; !ok && len(desc[i].Input.Fields) > 0 {
			dataTypes = append(dataTypes, desc[i].Input)
			dataTypeSet[name] = struct{}{}
		}
		name = desc[i].Output.Name
		if _, ok := dataTypeSet[name]; !ok && len(desc[i].Output.Fields) > 0 {
			dataTypes = append(dataTypes, desc[i].Output)
			dataTypeSet[name] = struct{}{}
		}

		for j := range desc[i].Input.Fields {
			types, err := collectStructs(desc[i].Input.Fields[j], dataTypeSet)
			if err != nil {
				return nil, err
			}
			dataTypes = append(dataTypes, types...)
		}
		for j := range desc[i].Output.Fields {
			types, err := collectStructs(desc[i].Output.Fields[j], dataTypeSet)
			if err != nil {
				return nil, err
			}
			dataTypes = append(dataTypes, types...)
		}

		desc[i].DataTypes = dataTypes
	}
	return &ClientGen{
		meta: ApiClientDesc{
			Client: clientDesc,
			Apis:   desc,
		},
	}, nil
}

func collectStructs(field Field, dataTypeSet map[string]struct{}) ([]DataType, error) {
	dataTypes := make([]DataType, 0)

	if field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Map {
		field.Type = field.Type.Elem()
	}
	if field.Type.Kind() == reflect.Pointer {
		field.Type = field.Type.Elem()
	}
	if field.Type.Kind() == reflect.Struct {
		subTypes, err := collectTypes(field, dataTypeSet)
		if err != nil {
			return nil, err
		}
		dataTypes = append(dataTypes, subTypes...)
	}

	return dataTypes, nil
}

func collectTypes(field Field, dataTypeSet map[string]struct{}) ([]DataType, error) {
	dataTypes := make([]DataType, 0)
	subType, err := extractDataType(field.Type)
	if err != nil {
		return nil, err
	}
	name := subType.Name
	if _, ok := dataTypeSet[name]; !ok && len(subType.Fields) > 0 {
		dataTypes = append(dataTypes, subType)
		dataTypeSet[name] = struct{}{}

		for _, subField := range subType.Fields {
			if subField.Type.Kind() == reflect.Struct {
				subTypes, err := collectTypes(subField, dataTypeSet)
				if err != nil {
					return nil, err
				}

				dataTypes = append(dataTypes, subTypes...)
			}
			if subField.Type.Kind() == reflect.Slice || subField.Type.Kind() == reflect.Map || subField.Type.Kind() == reflect.Pointer {
				if subField.Type.Elem().Kind() == reflect.Pointer {
					subField.Type = subField.Type.Elem()
				}
				if subField.Type.Elem().Kind() == reflect.Struct {
					subField.Type = subField.Type.Elem()
					subTypes, err := collectTypes(subField, dataTypeSet)
					if err != nil {
						return nil, err
					}

					dataTypes = append(dataTypes, subTypes...)
				}
			}
		}
	}

	return dataTypes, nil
}

func MakeApiDesc(meta vel.HandlerMeta) (ApiDesc, error) {
	inputReflectType := reflect.TypeOf(meta.Input)
	inputType, err := extractDataType(inputReflectType)
	if err != nil {
		return ApiDesc{}, err
	}
	outputReflectType := reflect.TypeOf(meta.Output)
	outputType, err := extractDataType(outputReflectType)
	if err != nil {
		return ApiDesc{}, err
	}

	return ApiDesc{
		Input:       inputType,
		Output:      outputType,
		OperationID: meta.OperationID,
		FuncName:    Capitalize(meta.OperationID),
	}, nil
}

// Helper function to extract fields from a struct type
func extractDataType(t reflect.Type) (DataType, error) {
	var fields []Field

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		typeName := field.Type.String()
		if field.Type.Kind() == reflect.Struct {
			typeName = field.Type.Name()
		}
		if field.Type.Kind() == reflect.Pointer && field.Type.Elem().Kind() == reflect.Struct {
			typeName = "*" + field.Type.Elem().Name()
		}
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
			typeName = "[]" + field.Type.Elem().Name()
		}
		if field.Type.Kind() == reflect.Map && field.Type.Elem().Kind() == reflect.Struct {
			typeName = "map[" + field.Type.Key().Name() + "]" + field.Type.Elem().Name()
		}

		fields = append(fields, Field{
			Name:     field.Name,
			Type:     field.Type,
			TypeName: typeName,
			JsonTag:  field.Tag.Get("json"),
		})
	}

	name := t.Name()
	if len(fields) == 0 {
		name = ""
	}
	return DataType{
		Name:   name,
		Fields: fields,
	}, nil
}

type ApiClientDesc struct {
	Client ClientDesc
	Apis   []ApiDesc
}

type ClientDesc struct {
	TypeName    string
	PackageName string
}

type ApiDesc struct {
	Input       DataType
	Output      DataType
	OperationID string
	FuncName    string
	DataTypes   []DataType
}

type DataType struct {
	Name   string
	Fields []Field
	// OtherTypes defines a list of types required to generate the fields
	OtherTypes []DataType
}

type Field struct {
	Name     string
	Type     reflect.Type
	TypeName string
	JsonTag  string
}

func (g *ClientGen) Generate(w io.Writer) error {
	pipe := bytes.NewBuffer(nil)
	clientTpl, err := template.New("clientTpl").Parse(clientTemplate)
	if err != nil {
		return err
	}

	callTpl, err := template.New("callTpl").Parse(callTemplate)
	if err != nil {
		return err
	}

	if err := clientTpl.Execute(pipe, g.meta); err != nil {
		return err
	}

	if err := callTpl.Execute(pipe, g.meta); err != nil {
		return err
	}
	nonFormatted := pipe.Bytes()
	f, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}
	if _, err := f.Write(nonFormatted); err != nil {
		return err
	}
	defer f.Close()

	cmdStr := fmt.Sprintf("goimports < %s", f.Name())

	formatted, err := exec.Command("sh", "-c", cmdStr).Output()
	if err != nil {
		return err
	}

	if _, err := w.Write(formatted); err != nil {
		return err
	}

	return nil
}

func Capitalize(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	s = string(r)
	return s
}
