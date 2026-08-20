package ats

import "testing"

func TestHTMLToText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "headings and paragraphs",
			in:   `<div><h3>About</h3><p>Build things &amp; ship.</p><br/>Remote<br></div>`,
			want: "About\nBuild things & ship.\n\nRemote",
		},
		{
			name: "strips style blocks",
			in:   `<style>.foo{color:red}</style><p>Hello</p>`,
			want: "Hello",
		},
		{
			name: "strips inline tags",
			in:   `<p>Join <strong>Acme</strong> and <a href="https://acme.com">apply</a>.</p>`,
			want: "Join Acme and apply.",
		},
		{
			name: "list items",
			in:   `<ul><li>First</li><li>Second</li></ul>`,
			want: "First\nSecond",
		},
		{
			name: "encoded html tags",
			in:   `&lt;p&gt;Already encoded&lt;/p&gt;`,
			want: "Already encoded",
		},
		{
			name: "html comments",
			in:   `<!-- hidden --><p>Visible</p>`,
			want: "Visible",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := HTMLToText(tc.in)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
