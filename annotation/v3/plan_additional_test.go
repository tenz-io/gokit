package annotation

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type planAddress struct {
	Street string `validate:"required"`
}

type planEnvelope struct {
	Address planAddress
}

func TestPlanFor_NestedPathsDoNotDependOnCacheOrder(t *testing.T) {
	planCache.Delete(reflect.TypeOf(planAddress{}))
	planCache.Delete(reflect.TypeOf(planEnvelope{}))

	_, err := PlanFor(&planAddress{})
	require.NoError(t, err)

	err = Validate(&planEnvelope{})
	require.Error(t, err)
	errs, ok := AsErrors(err)
	require.True(t, ok)
	require.Len(t, errs, 1)
	assert.Equal(t, "Address.Street", errs[0].Field)

	root, err := PlanFor(&planAddress{})
	require.NoError(t, err)
	require.Len(t, root.Fields(), 1)
	assert.Equal(t, "Street", root.Fields()[0].Path)
}

func TestPlanFor_ConfigErrorsNeverPoisonCache(t *testing.T) {
	type invalidRuleConfig struct {
		Count int `validate:"gt=not-a-number"`
	}
	typ := reflect.TypeOf(invalidRuleConfig{})
	planCache.Delete(typ)

	for i := 0; i < 2; i++ {
		p, err := PlanFor(&invalidRuleConfig{})
		assert.Nil(t, p)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not-a-number")
	}
	_, cached := planCache.Load(typ)
	assert.False(t, cached)
}

func TestPlan_MetadataTraversalAndCache(t *testing.T) {
	type child struct {
		Value string `validate:"required"`
	}
	type request struct {
		ID        int    `bind:"uri,name=user_id" json:"ignored"`
		Search    string `bind:"query" json:"q,omitempty"`
		Token     string `bind:"header" protobuf:"bytes,1,opt,name=token_pb"`
		Upload    []byte `bind:"file,name=payload"`
		Form      string `bind:"form"`
		ProtoOnly string `protobuf:"bytes,2,opt,name=proto_name,json=protoName"`
		Fallback  bool
		When      time.Time `validate:"required"`
		Child     child
		hidden    string `validate:"required"`
	}

	p, err := PlanFor(&request{})
	require.NoError(t, err)
	again, err := PlanFor(&request{})
	require.NoError(t, err)
	assert.Same(t, p, again)
	require.Len(t, p.Fields(), 9, "unexported fields must not enter the plan")

	byName := map[string]*Field{}
	for _, f := range p.Fields() {
		byName[f.Name] = f
	}
	assert.Equal(t, "user_id", byName["ID"].BindName)
	assert.Equal(t, "q", byName["Search"].BindName)
	assert.Equal(t, "token_pb", byName["Token"].BindName)
	assert.Equal(t, "proto_name", byName["ProtoOnly"].BindName)
	assert.Equal(t, "Fallback", byName["Fallback"].BindName)
	assert.Equal(t, BindURI, byName["ID"].BindSource)
	assert.Equal(t, BindQuery, byName["Search"].BindSource)
	assert.Equal(t, BindHeader, byName["Token"].BindSource)
	assert.Equal(t, BindFile, byName["Upload"].BindSource)
	assert.Equal(t, BindForm, byName["Form"].BindSource)
	assert.True(t, byName["When"].IsRequired())
	assert.False(t, byName["Fallback"].IsRequired())
	assert.Empty(t, byName["When"].children, "time.Time is a validation leaf")

	query := p.FieldsBySource(BindQuery)
	require.Len(t, query, 1)
	assert.Equal(t, "Search", query[0].Name)
	assert.Len(t, p.FieldsBySource(BindNone), 4)

	var paths []string
	p.Walk(func(f *Field) bool {
		paths = append(paths, f.Path)
		return true
	})
	assert.Contains(t, paths, "Child.Value")
	assert.NotContains(t, paths, "hidden")

	var stopped []string
	p.Walk(func(f *Field) bool {
		stopped = append(stopped, f.Path)
		return len(stopped) < 2
	})
	assert.Len(t, stopped, 2)
}

func TestPlanFor_ArgumentsRecursiveTypesAndTagParsing(t *testing.T) {
	var nilRequest *planEnvelope
	cases := []any{nil, planEnvelope{}, nilRequest, new(int)}
	for _, input := range cases {
		p, err := PlanFor(input)
		assert.Nil(t, p)
		var invalid *invalidArgError
		require.ErrorAs(t, err, &invalid)
		assert.NotEmpty(t, invalid.Error())
	}

	type node struct {
		Value string `validate:"required"`
		Next  *node
	}
	p, err := PlanFor(&node{})
	require.NoError(t, err)
	require.Len(t, p.Fields(), 2)
	assert.Empty(t, p.Fields()[1].children, "recursive type expansion must terminate")

	sf := reflect.StructField{Tag: `bind:"query,header,name=x"`}
	assert.Equal(t, BindQuery, resolveBindSource(sf), "first declared source wins deterministically")
	assert.Equal(t, BindNone, resolveBindSource(reflect.StructField{}))

	tag := reflect.StructTag(`bind:"query,name=x,bare,,key=value=tail"`)
	assert.Equal(t, map[string]string{
		"query": "", "name": "x", "bare": "", "key": "value=tail",
	}, tagMap(tag, tagBind))
	assert.Empty(t, tagMap(reflect.StructTag(""), tagBind))

	assert.Equal(t, "Parent.Child", joinPath("Parent", "Child"))
	assert.Nil(t, peelPtr(reflect.TypeOf(0)))
	assert.Equal(t, reflect.TypeOf(struct{}{}), peelPtr(reflect.TypeOf(&struct{}{})))
}
