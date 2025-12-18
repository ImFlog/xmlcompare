package xmlcompare

import "testing"

func TestEqual_PassingCase_WhitespaceAndOrderInsensitive(t *testing.T) {
	a := `
    <root>
      <a id="1">  hello   world  </a>
      <b x="1" y="2"/>
      <c><d>text</d></c>
    </root>`
	// Same content but different order, different whitespace, and attribute order flipped
	b := `
    <root>
      <c>
        <d>text</d>
      </c>
      <b y="2" x="1"/>
      <a id="1">hello world</a>
    </root>`

	ok, err := Equal(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected XMLs to be considered equal")
	}
}

func TestEqual_Error_WrongNamespace(t *testing.T) {
	a := `<ns1:root xmlns:ns1="urn:x"><child/></ns1:root>`
	b := `<ns2:root xmlns:ns2="urn:y"><child/></ns2:root>`
	ok, err := Equal(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected mismatch because of different namespaces")
	}
}

func TestEqual_Error_MissingAttribute(t *testing.T) {
	a := `<root id="123"/>`
	b := `<root/>`
	ok, err := Equal(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected mismatch because expected has no @id, actual does")
	}
}

func TestEqual_Error_WrongAttributeValue(t *testing.T) {
	a := `<root id="123"/>`
	b := `<root id="999"/>`
	ok, err := Equal(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected mismatch because attribute values differ")
	}
}

func TestEqual_Error_AdditionalChildTag(t *testing.T) {
	a := `<root><a/><b/></root>`
	b := `<root><a/></root>`
	ok, err := Equal(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected mismatch because of additional child <b>")
	}
}

func TestEqual_Error_MissingChildTag(t *testing.T) {
	a := `<root><a/></root>`
	b := `<root><a/><b/></root>`
	ok, err := Equal(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected mismatch because actual is missing child <b>")
	}
}

func TestEqual_TextDifference(t *testing.T) {
	a := `<root><msg>hello</msg></root>`
	b := `<root><msg>world</msg></root>`
	ok, err := Equal(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected mismatch because text differs")
	}
}

func TestEqual_TextWhitespaceIsNormalized(t *testing.T) {
	a := `<root><msg>  hello   world  </msg></root>`
	b := `<root><msg>hello world</msg></root>`
	ok, err := Equal(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected equality because whitespace should be normalized")
	}
}

func TestEqual_ParsingError(t *testing.T) {
	a := `<root>` // invalid XML
	b := `<root/>`
	ok, err := Equal(a, b)
	if err == nil || ok {
		t.Fatalf("expected parsing error and false, got ok=%v err=%v", ok, err)
	}
}
