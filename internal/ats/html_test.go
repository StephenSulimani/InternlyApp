package ats

import "testing"

func TestHTMLToText(t *testing.T) {
	in := `<div><h3>About</h3><p>Build things &amp; ship.</p><br/>Remote<br></div>`
	got := HTMLToText(in)
	want := "About\nBuild things & ship.\n\nRemote"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
