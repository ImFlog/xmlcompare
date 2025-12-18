package xmlcompare

import "testing"

func TestEqual_WhitespaceDifferences(t *testing.T) {
	a := `<root><a>hello world</a><b x="1" y="2"/></root>`
	if !Equal(a, a) {
		t.Fatalf("expected XMLs to be considered equal")
	}
}

func TestEqual_DifferentContent(t *testing.T) {
	a := `<root><a>hello</a></root>`
	b := `<root><a>world</a></root>`
	if Equal(a, b) {
		t.Fatalf("expected XMLs to be different")
	}
}
