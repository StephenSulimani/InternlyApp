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
	if errRes.Success != 0 || errRes.Message != "something went wrong" {
		t.Fatalf("error response = %+v", errRes)
	}

	okBody := StringResponse("ok")
	var okRes Response
	if err := json.Unmarshal([]byte(okBody), &okRes); err != nil {
		t.Fatal(err)
	}
	if okRes.Success != 1 || okRes.Message != "ok" {
		t.Fatalf("ok response = %+v", okRes)
	}
}
