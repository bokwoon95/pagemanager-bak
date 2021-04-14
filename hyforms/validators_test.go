package hyforms

import (
	"testing"

	"github.com/bokwoon95/pagemanager/testutil"
)

func Test_Validators(t *testing.T) {
	is := testutil.New(t)
	errs := Validate("pagemanagers", IsURL)
	is.True(errs != nil)
	errs = Validate("http://-10:10/foobar", IsURL)
	is.True(errs != nil)
	errs = Validate("google.come", IsURL)
	is.True(errs == nil)
}
