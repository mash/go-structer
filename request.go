package structer

import (
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"

	"github.com/athom/suitecase"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB, copied from http://golang.org/src/net/http/request.go#L31
)

var attachmentType reflect.Type
var attachmentPointerType reflect.Type

func init() {
	attachmentType = reflect.TypeOf(Attachment{})
	attachmentPointerType = reflect.PtrTo(reflect.TypeOf(Attachment{}))
}

type Attachment struct {
	File   multipart.File
	Header *multipart.FileHeader
}

func (a Attachment) Empty() bool {
	return a.File == nil && a.Header == nil
}

func ToStruct(req *http.Request, parameter interface{}) error {
	err := req.ParseMultipartForm(defaultMaxMemory)
	if err != nil && err != http.ErrNotMultipart {
		return err
	}

	if err := toStruct(req, parameter); err != nil {
		return err
	}
	return nil
}

func toStruct(req *http.Request, parameter interface{}) error {
	v := reflect.ValueOf(parameter).Elem()
	if v.Kind() != reflect.Struct {
		panic("Pass pointer to a struct as parameter")
	}

	values := req.Form
	for i := 0; i < v.NumField(); i++ {
		valueField := v.Field(i)
		typeField := v.Type().Field(i)
		typeType := typeField.Type
		valueName := typeField.Name
		paramName := suitecase.ToSnakeCase(valueName)

		// log.Printf("valueField: %#v\n", valueField)
		// log.Printf("valueField.Interface(): %#v\n", valueField.Interface())
		// log.Printf("valueField.Addr().Interface(): %#v\n", valueField.Addr().Interface())
		// log.Printf("typeField: %#v\n", typeField)
		// log.Printf("type.Kind: %#v\n", typeField.Type.Kind())

		// Recursively fill anonymous structs' fields with req.Form's values
		if typeField.Type.Kind() == reflect.Struct && typeField.Anonymous {
			anonymousStruct := valueField.Addr().Interface()
			if err := toStruct(req, anonymousStruct); err != nil {
				return err
			}
			continue
		}

		if typeType == attachmentType ||
			((typeField.Type.Kind() == reflect.Ptr) && (typeType == attachmentPointerType)) {
			// log.Printf("typeType: %#v, typeField: %#v, valueField: %#v", typeType, typeField, valueField)
			// call ParseMultipartForm before hand to limit memory if we want to
			file, header, err := req.FormFile(paramName)
			if err == http.ErrMissingFile || err == http.ErrNotMultipart {
				continue
			} else if err != nil {
				return err
			} else {
				// log.Printf("file: %#v, header: %#v", file, header)
				attachment := Attachment{
					File:   file,
					Header: header,
				}
				if typeField.Type.Kind() == reflect.Struct {
					valueField.Set(reflect.ValueOf(attachment))
				} else if typeField.Type.Kind() == reflect.Ptr {
					valueField.Set(reflect.ValueOf(&attachment))
				}
			}
			continue
		}

		valueFromRequest := values.Get(paramName)
		if valueFromRequest == "" {
			continue
		}

		switch typeField.Type.Kind() {
		case reflect.String:
			valueField.SetString(valueFromRequest)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.ParseInt(valueFromRequest, 10, 64)
			if err != nil {
				return err
			}
			valueField.SetInt(i)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			i, err := strconv.ParseUint(valueFromRequest, 10, 64)
			if err != nil {
				return err
			}
			valueField.SetUint(i)

		case reflect.Float32, reflect.Float64:
			i, err := strconv.ParseFloat(valueFromRequest, 64)
			if err != nil {
				return err
			}
			valueField.SetFloat(i)

		case reflect.Bool:
			b, err := strconv.ParseBool(valueFromRequest)
			if err != nil {
				return err
			}
			valueField.SetBool(b)

		default:
			continue
		}

	}
	return nil
}
