package types

import (
	"encoding/json"
	"testing"
)

func TestResponseHelpers(t *testing.T) {
	errBody := ErrorResponse("something went wrong")
	var errRes Response
	if err := json.Unmarshal([]byte(errBody), &errRes); err != nil {
		t.Fatal(err)
	}
	if errRes.Success || errRes.Message != "something went wrong" {
		t.Fatalf("error response = %+v", errRes)
	}

	okBody := SuccessResponse("ok")
	var okRes Response
	if err := json.Unmarshal([]byte(okBody), &okRes); err != nil {
		t.Fatal(err)
	}
	if !okRes.Success || okRes.Message != "ok" {
		t.Fatalf("ok response = %+v", okRes)
	}
}
