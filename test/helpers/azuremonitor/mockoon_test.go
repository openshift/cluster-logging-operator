package azuremonitor

import (
	_ "embed"
	"testing"
)

//go:embed test_mockoon.log
var LogData string

func TestParsingLongs(t *testing.T) {
	logs, err := extractStructuredLogs(LogData, "application")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	expect := 8
	i := len(logs)
	if i != expect {
		t.Logf("got: %d, expected: %d", i, expect)
		t.Fail()
	}

}
