package hy

import (
	"fmt"
	"regexp"
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
			"attr1": `value-1`,
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

func Test_Txt(t *testing.T) {
	is := testutil.New(t, testutil.Parallel)
	div := H("div", nil, Txt(`<div><b>Hello!</b></div>`))
	html, err := Marshal(nil, div)
	is.NoErr(err)
	fmt.Println(html)
	re := regexp.MustCompile(
		`(?:[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_` + "`" +
			`{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`,
	)
	is.True(re.MatchString("abc@a.my"))
	re = regexp.MustCompile(`\+(9[976]\d|8[987530]\d|6[987]\d|5[90]\d|42\d|3[875]\d|2[98654321]\d|9[8543210]|8[6421]|6[6543210]|5[87654321]|4[987654310]|3[9643210]|2[70]|7|1)\d{1,14}$`)
	is.True(re.MatchString("+6562420960"))
	is.True(re.MatchString("+6591528794"))
	is.True(re.MatchString("+6596697695"))
	re = regexp.MustCompile(`(9[976]\d|8[987530]\d|6[987]\d|5[90]\d|42\d|3[875]\d|2[98654321]\d|9[8543210]|8[6421]|6[6543210]|5[87654321]|4[987654310]|3[9643210]|2[70]|7|1)\d{1,14}$`)
	is.True(re.MatchString("62420960"))
	is.True(re.MatchString("91528794"))
	is.True(re.MatchString("96697695"))
	is.True(re.MatchString("333"))
	html, err = Marshal(nil, H("div", Attr{"class": `value" id="value`, "id": "hey"}, Txt(55)))
	is.NoErr(err)
	fmt.Println(html)
}
