package server

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/infobloxopen/themis/pdp"
)

type mockResponseWriter struct {
	statusCode int
	msg        string
}

func (w *mockResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (w *mockResponseWriter) Write(b []byte) (int, error) {
	w.msg += string(b)
	return len(b), nil
}

func (w *mockResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *mockResponseWriter) clear() {
	w.statusCode = 0
	w.msg = ""
}

func TestHandleQuery(t *testing.T) {
	var w mockResponseWriter

	handleQuery(&w, nil, []string{}, url.Values{})
	assertEqualWriter(t, mockResponseWriter{
		statusCode: http.StatusNotFound,
		msg:        missingStorageMsg,
	}, w, "On nil storage:")
	w.clear()

	rule := pdp.NewRule("permit", false, pdp.Target{}, nil, pdp.EffectPermit, nil)

	policy := pdp.NewPolicy("first", false, pdp.Target{}, []*pdp.Rule{rule},
		pdp.RuleCombiningAlgs["firstapplicableeffect"], nil, nil)

	root := pdp.NewPolicySet("test", false, pdp.Target{}, []pdp.Evaluable{policy},
		pdp.PolicyCombiningAlgs["firstapplicableeffect"], nil, nil)

	s := pdp.NewPolicyStorage(root, pdp.Symbols{}, nil)

	handleQuery(&w, s, []string{}, url.Values{"depth": []string{"nine"}})
	assertEqualWriter(t, mockResponseWriter{
		statusCode: http.StatusBadRequest,
		msg:        "\"strconv.ParseUint: parsing \\\"nine\\\": invalid syntax\"",
	}, w, "On bad depth:")
	w.clear()

	handleQuery(&w, s, []string{"test", "second"}, url.Values{"depth": []string{"3"}})
	assertEqualWriter(t, mockResponseWriter{
		statusCode: http.StatusNotFound,
		msg:        "\"#73: Path [test second] not found\"",
	}, w, "On not found path:")
	w.clear()

	handleQuery(&w, s, []string{"test", "first"}, url.Values{"depth": []string{"3"}})
	assertEqualWriter(t, mockResponseWriter{
		statusCode: http.StatusOK,
		msg:        "{\"ord\":0,\"id\":\"first\",\"target\":{},\"obligations\":null,\"algorithm\":{\"type\":\"firstApplicableEffectRCA\"},\"rules\":[{\"ord\":0,\"id\":\"permit\",\"target\":{},\"obligations\":null,\"effect\":\"Permit\"}]}",
	}, w, "On found:")
	w.clear()
}

func assertEqualWriter(t *testing.T, expect, got mockResponseWriter, failPrefix string) {
	t.Helper()
	if expect.statusCode != got.statusCode {
		t.Errorf(failPrefix+"expected status %d, got %d", expect.statusCode, got.statusCode)
	}
	expectMsg, gotMsg := strings.TrimSpace(expect.msg), strings.TrimSpace(got.msg)
	if strings.Compare(expectMsg, gotMsg) != 0 {
		t.Errorf(failPrefix+"expected message \"%s\", got \"%s\"", expectMsg, gotMsg)
	}
}
