package timetest

import (
	"time"

	"github.com/bouk/monkey"
)

// FreezeTime makes time.Now returns ReferenceTime by monkey-patching the
// function. See UnfreezeTime for the inverse operation. WARNING: This function
// effect isn't scoped, meaning that calling it will affect all tests. If
// running parallel tests, it is recommended to write a TestMain function to do
// the actual call.
func FreezeTime() {
	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2006, 01, 02, 15, 04, 05, 000, time.UTC)
	})
}

// UnfreezeTime makes time.Now returns the current time by removing all
// monkey-patching on the function.
func UnfreezeTime() {
	monkey.Unpatch(time.Now)
}
