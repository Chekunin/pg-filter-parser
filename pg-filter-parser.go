package pg_filter_parser

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strings"
)

var sqlFilterMap = map[string]string{
	"eq":       "=",
	"neq":      "<>",
	"lt":       "<",
	"lte":      "<=",
	"gt":       ">",
	"gte":      ">=",
	"contains": "like",
	"in":       "in",
}

type filter struct {
	FromTagName string
	ToTagName   string
}

func NewFilter(fromTagName, toTagName string) *filter {
	return &filter{
		FromTagName: fromTagName,
		ToTagName:   toTagName,
	}
}

type Condition struct {
	Fieldname string
	Operator  string
	Value     interface{}
}

type Chain struct {
	Type  string
	Items []interface{}
}

func (f filter) parseChain(model interface{}, chain Chain) (string, []interface{}) {
	if chain.Type != "and" && chain.Type != "or" {
		panic(errors.New("Unknown connection type"))
	}
	var params []interface{}
	str := "("
	for i, v := range chain.Items {
		if i > 0 {
			str += fmt.Sprintf(" %s ", chain.Type)
		}
		tempStr, tempParams := f.ParseFilter(model, v)
		str += tempStr
		params = append(params, tempParams...)
	}
	str += ")"
	return str, params
}

func (f filter) parseCondition(model interface{}, condition Condition) (string, interface{}) {
	field, has := getFieldWithTagValue(reflect.TypeOf(model), f.FromTagName, condition.Fieldname)
	if !has {
		panic(fmt.Errorf("Field %s not found in model", condition.Fieldname))
	}
	fieldName, has := field.Tag.Lookup(f.ToTagName)
	if !has {
		panic(fmt.Errorf("Tag `%s` in field %s not found", f.ToTagName, fieldName))
	}
	operator, has := sqlFilterMap[condition.Operator]
	if !has {
		panic(fmt.Errorf("Operator `%s` is unknown", condition.Operator))
	}
	return fmt.Sprintf("%s%s?", fieldName, operator), condition.Value
}

func (f filter) ParseFilter(model interface{}, filter interface{}) (string, []interface{}) {
	if model == nil {
		panic(errors.New("Model cannot be nil"))
	}
	var conn Chain
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{ErrorUnused: true, Result: &conn, TagName: f.FromTagName})
	if err := decoder.Decode(filter); err == nil {
		return f.parseChain(model, conn)
	}
	var cond Condition
	decoder, _ = mapstructure.NewDecoder(&mapstructure.DecoderConfig{ErrorUnused: true, Result: &cond, TagName: f.FromTagName})
	if err := decoder.Decode(filter); err == nil {
		tempStr, param := f.parseCondition(model, cond)
		return tempStr, []interface{}{param}
	} else {
		fmt.Println(err)
	}
	panic(errors.New("Unknown filter type"))
}

func getFieldWithTagValue(t reflect.Type, tagName, tagValue string) (reflect.StructField, bool) {
	switch t.Kind() {
	case reflect.Array, reflect.Ptr, reflect.Slice:
		return getFieldWithTagValue(t.Elem(), tagName, tagValue)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if val, ok := f.Tag.Lookup(tagName); ok && strings.Split(val, ",")[0] == tagValue {
				return f, true
			}
		}
	}
	return reflect.StructField{}, false
}
