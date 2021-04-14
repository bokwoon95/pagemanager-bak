package hyp

import (
	"testing"

	"github.com/bokwoon95/pagemanager/testutil"
)

func Test_ParseAttributes(t *testing.T) {
	assertOK := func(t *testing.T, selector string, attributes map[string]string, want Attributes) {
		is := testutil.New(t, testutil.Parallel, testutil.FailFast)
		got := ParseAttributes(selector, attributes)
		is.NoErr(got.ParseErr)
		is.Equal(want, got)
	}
	t.Run("empty", func(t *testing.T) {
		selector := "div"
		assertOK(t, selector, map[string]string{}, Attributes{
			Selector: selector,
			Tag:      "div",
			Dict:     map[string]string{},
		})
	})
	t.Run("body literal", func(t *testing.T) {
		selector := "<b>hello world!</b>"
		assertOK(t, selector, map[string]string{}, Attributes{
			Selector: selector,
			Body:     "<b>hello world!</b>",
			Dict:     map[string]string{},
		})
	})
	t.Run("tag only", func(t *testing.T) {
		selector := "div"
		assertOK(t, selector, map[string]string{}, Attributes{
			Selector: selector,
			Tag:      "div",
			Dict:     map[string]string{},
		})
	})
	t.Run("selector tags, id, classes and attributes", func(t *testing.T) {
		selector := "p#id1.class1.class2.class3#id2[attr1=val1][attr2=val2][attr3=val3][attr4]"
		assertOK(t, selector, map[string]string{}, Attributes{
			Selector: selector,
			Tag:      "p",
			ID:       "id2",
			Class:    "class1 class2 class3",
			Dict: map[string]string{
				"attr1": "val1",
				"attr2": "val2",
				"attr3": "val3",
				"attr4": Enabled,
			},
		})
	})
	t.Run("attributes overwrite selector", func(t *testing.T) {
		selector := "p#id1.class1.class2.class3#id2[attr1=val1][attr2=val2][attr3=val3][attr4]"
		attributes := map[string]string{
			"id":    "id3",
			"class": "class4 class5 class6",
			"attr1": "value-1",
			"attr2": "value-2",
			"attr3": "value-3",
			"attr4": Disabled,
		}
		assertOK(t, selector, attributes, Attributes{
			Selector: selector,
			Tag:      "p",
			ID:       "id3",
			Class:    "class1 class2 class3 class4 class5 class6",
			Dict: map[string]string{
				"attr1": "value-1",
				"attr2": "value-2",
				"attr3": "value-3",
				"attr4": Disabled,
			},
		})
	})
}
