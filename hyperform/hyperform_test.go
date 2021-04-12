package hyperform

import (
	"testing"

	"github.com/bokwoon95/pagemanager/testutil"
)

func Test_parseSelector(t *testing.T) {
	assertOK := func(t *testing.T, selector string, attributes map[string]string, want parsedSelector) {
		is := testutil.New(t, testutil.Parallel, testutil.FailFast)
		got, err := parseSelector(selector, attributes)
		is.NoErr(err)
		is.Equal(want, got)
	}
	t.Run("empty", func(t *testing.T) {
		assertOK(t, "", map[string]string{}, parsedSelector{tag: "div", attributes: map[string]string{}})
	})
	t.Run("body literal", func(t *testing.T) {
		assertOK(t, "<b>hello world!</b>", map[string]string{}, parsedSelector{body: "<b>hello world!</b>", attributes: map[string]string{}})
	})
	t.Run("tag only", func(t *testing.T) {
		assertOK(t, "div", map[string]string{}, parsedSelector{tag: "div", attributes: map[string]string{}})
	})
	t.Run("selector tags, id, classes and attributes", func(t *testing.T) {
		assertOK(t,
			"p#id1.class1.class2.class3#id2[attr1=val1][attr2=val2][attr3=val3][attr4]",
			map[string]string{},
			parsedSelector{
				tag:   "p",
				id:    "id2",
				class: "class1 class2 class3",
				attributes: map[string]string{
					"attr1": "val1",
					"attr2": "val2",
					"attr3": "val3",
					"attr4": Enabled,
				},
			},
		)
	})
	t.Run("attributes overwrite selector", func(t *testing.T) {
		assertOK(t,
			"p#id1.class1.class2.class3#id2[attr1=val1][attr2=val2][attr3=val3][attr4]",
			map[string]string{
				"id":    "id3",
				"class": "class4 class5 class6",
				"attr1": "value-1",
				"attr2": "value-2",
				"attr3": "value-3",
				"attr4": Disabled,
			},
			parsedSelector{
				tag:   "p",
				id:    "id3",
				class: "class1 class2 class3 class4 class5 class6",
				attributes: map[string]string{
					"attr1": "value-1",
					"attr2": "value-2",
					"attr3": "value-3",
					"attr4": Disabled,
				},
			},
		)
	})
}
