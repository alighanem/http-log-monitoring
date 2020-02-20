package commonlog_test

import (
	"fmt"
	"testing"

	"github.com/ali.ghanem/http-log-monitoring/commonlog"
)

func TestLogLexer_Parse(t *testing.T) {
	t.Parallel()

	t.Run("valid format", func(t *testing.T) {
		event, err := commonlog.Parse(`66.137.220.245 - - [10/Feb/2020:17:35:21 +0100] "POST /technologies/e-enable/collaborative/bandwidth HTTP/1.0" 200 19072`)
		if err != nil {
			t.Fatal(err)
		}

		t.Log("event", event)
	})

	type testCase struct {
		Line          string
		ExpectedError string
	}

	invalidCases := map[string]testCase{
		"empty line": {
			Line:          "",
			ExpectedError: "empty log line",
		},
		"cannot read date": {
			Line:          `66.137.220.245 rfc user[10/Feb/2020:17:35:21 +0100]`,
			ExpectedError: `reading date: character not found [ - event: host:66.137.220.245|rfc931:rfc|user:user[10/Feb/2020:17:35:21|date:0001-01-01 00:00:00 +0000 UTC|request:|status:0|bytes:0`,
		},
		"invalid date": {
			Line:          `66.137.220.245 rfc user [10/Inv/2020:17:35:21 +0100]`,
			ExpectedError: `invalid date format: parsing time "10/Inv/2020:17:35:21 +0100" as "02/Jan/2006:15:04:05 -0700": cannot parse "Inv/2020:17:35:21 +0100" as "Jan"`,
		},
		"invalid status": {
			Line:          `66.137.220.245 - - [10/Feb/2020:17:35:21 +0100] "POST /technologies/e-enable/collaborative/bandwidth HTTP/1.0" XXX 19072`,
			ExpectedError: `invalid status format: strconv.Atoi: parsing "XXX": invalid syntax`,
		},
		"invalid bytes number": {
			Line:          `66.137.220.245 - - [10/Feb/2020:17:35:21 +0100] "POST /technologies/e-enable/collaborative/bandwidth HTTP/1.0" 200 UUUUU`,
			ExpectedError: `invalid bytes number: strconv.Atoi: parsing "UUUUU": invalid syntax`,
		},
	}

	for name, c := range invalidCases {
		t.Run(name, func(t *testing.T) {
			_, err := commonlog.Parse(c.Line)
			if err == nil {
				t.Fatal("expected error not occurred", "err", c.ExpectedError)
			}

			if c.ExpectedError != err.Error() {
				t.Fatal("different error occurred", "expected", c.ExpectedError, "actual", err.Error())
			}
		})
	}
}

func BenchmarkLogLexer_Parse(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := commonlog.Parse(fmt.Sprintf(`66.%v.220.245 - - [21/Feb/2020:17:35:21 +0100] "POST /technologies/e-enable/collaborative/bandwidth HTTP/1.0" 200 19072`, n))
		if err != nil {
			b.Fatal(err)
		}
	}
}
