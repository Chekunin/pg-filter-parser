package pg_filter_parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestModel struct {
	Aaa string `manage:"field0" pg:"pg-field0,pk"`
	Bbb string `manage:"field1" pg:"pg-field1"`
	Ccc string `manage:"field2" pg:"pg-field2"`
	Ddd string `manage:"field3" pg:"pg-field3"`
}

func TestPgFilterParser(t *testing.T) {
	obj := Chain{
		Type: "and",
		Items: []interface{}{
			map[string]interface{}{
				"fieldname": "field0",
				"operator":  "eq",
				"value":     "hi",
			},
			Condition{
				Fieldname: "field1",
				Operator:  "lt",
				Value:     10,
			},
			Chain{
				Type: "or",
				Items: []interface{}{
					Condition{
						Fieldname: "field2",
						Operator:  "eq",
						Value:     "yo",
					},
					Condition{
						Fieldname: "field3",
						Operator:  "eq",
						Value:     "test",
					},
				},
			},
		},
	}
	filter := NewFilter("manage", "pg")
	filterStr, filterParams := filter.ParseFilter((*TestModel)(nil), obj)
	assert.Equal(t, "(pg-field0=? and pg-field1<? and (pg-field2=? or pg-field3=?))", filterStr)
	assert.Equal(t, []interface{}{"hi", 10, "yo", "test"}, filterParams)
}
