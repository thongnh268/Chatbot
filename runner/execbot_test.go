package runner

import (
	"reflect"
	"testing"
)

// basic case, exception case, fixed caseCompareEvent
// input, expected output, actual output

func TestMergeLeadPayload(t *testing.T) {
	var inDst, inSrc *ExecBot
	var exo, aco map[string]string

	inDst = &ExecBot{LeadPayload: map[string]string{"a": "1", "b": "2"}}
	inSrc = &ExecBot{LeadPayload: map[string]string{"a": "2", "c": "3"}}
	exo = map[string]string{"a": "1", "b": "2", "c": "3"}
	mergeLeadPayload(inDst.LeadPayload, inSrc.LeadPayload)
	aco = inDst.LeadPayload
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}
}
