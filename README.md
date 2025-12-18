[![Go Reference](https://pkg.go.dev/badge/github.com/imflog/xmlcompare.svg)](https://pkg.go.dev/github.com/imflog/xmlcompare)
[![CI](https://github.com/ImFlog/xml-compare/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/ImFlog/xml-compare/actions/workflows/go.yml)

# xmlcompare

A tiny, focused Go library to compare two XML documents for structural equality.

It is designed for tests and validation code where you want to assert that two
XML snippets are the same regardless of child elements order, attribute
order, or incidental whitespace differences.

Key properties:

- Order-independent comparison of child elements
- Attribute order does not matter; names and values must match
- Text nodes are compared with whitespace normalization
- Namespace-aware tag matching (qualified name = `prefix:local` semantics)
- Helpful mismatch messages printed to stdout describing the first difference

## Installation

```bash
go get github.com/imflog/xmlcompare
```

## Quick start

```go
package main

import (
	"fmt"
	xmlcmp "github.com/imflog/xmlcompare"
)

func main() {
	a := `<root><a id="1">hello  world</a><b x="1" y="2"/></root>`
	b := `<root><b y="2" x="1"/><a id="1">hello world</a></root>`

	equal, err := xmlcmp.Equal(a, b)
	if err != nil {
		panic(err)
	}
	fmt.Println(equal) // true
}
```

## API

```go
func Equal(actual, expected string) (bool, error)
```

Parses both XML strings and performs an order‑independent, namespace-aware
comparison. It returns:

- `true, nil` when documents are considered equal
- `false, nil` when a difference is found
- `false, err` if either XML cannot be parsed

On the first difference, a human‑readable explanation is printed to stdout to
aid debugging (see examples below). This is convenient in tests because your
test logs will show exactly what differed.

## What “equal” means here

This library purposefully defines equality in a practical testing‑friendly way:

- Element order is ignored. Siblings are matched by qualified tag name and
  then paired using a similarity heuristic (attributes and child tags) to make
  diagnostics meaningful.
- Attributes are compared by qualified name and value, ignoring attribute
  order. Missing, extra, or different values are reported.
- Text content is compared after whitespace normalization (collapsing runs of
  whitespace to a single space and trimming ends). This avoids false negatives
  from indentation and formatting.
- Namespaces matter. The qualified element name must match, both prefix
  (namespace) and local tag must be the same for a match.

## Examples

Order and whitespace insensitivity:

```go
ok, err := xmlcmp.Equal(
`<root><a id="1">  hello   world  </a><b x="1" y="2"/></root>`,
`<root><b y="2" x="1"/><a id="1">hello world</a></root>`,
)
// ok == true, err == nil
```

Mismatch examples (messages go to stdout):

```go
ok, _ := xmlcmp.Equal(`<root><a/></root>`, `<root><a/><b/></root>`)
// Output (example):
// Missing child at /root/b
// ok == false

ok, _ = xmlcmp.Equal(`<root id="123"/>`, `<root id="999"/>`)
// Output:
// Attribute value differs at /root: @id actual="123" expected="999"
// ok == false

ok, _ = xmlcmp.Equal(
`<ns1:root xmlns:ns1="urn:x"><child/></ns1:root>`,
`<ns2:root xmlns:ns2="urn:y"><child/></ns2:root>`,
)
// Output:
// XML mismatch at /ns1:root: different tags: actual=<ns1:root> expected=<ns2:root>
// ok == false
```

Parsing errors:

```go
ok, err := xmlcmp.Equal(`<root>`, `<root/>`)
// ok == false, err != nil (invalid XML)
```

## Behavior details

- Child matching strategy: For each actual child, candidates in the expected
  document with the same qualified tag are considered. If there is a single
  candidate, it is selected. If multiple candidates exist, a similarity score
  based on attributes, child tag sets, and direct text is used to pick the best
  match before recursing.
- Attributes: comparison is exact on both name and value. Attribute order is
  irrelevant. Namespace declaration attributes (e.g., `xmlns` / `xmlns:prefix`)
  are currently treated like regular attributes during equality checks.
- Text normalization: `strings.Fields` is used to collapse runs of whitespace
  into single spaces and trim at both ends before comparison.

## Limitations and notes

- Only element and text nodes are considered. Comments, processing instructions,
  and CDATA are not explicitly handled and may affect parsing depending on your
  inputs.
- Namespace comparison currently requires that the element’s qualified name
  (prefix plus local) matches between actual and expected; it does not attempt to
  canonicalize or resolve different prefixes that bind to the same URI.
  This is something we will work on in the future.
- The function prints the first detected mismatch to stdout for simplicity, as it is
  intended for use in tests. This could be improved in the future to return a
  structured result.

## Testing

The repository includes unit tests illustrating typical success and failure
cases. Run:

```bash
go test ./...
```

## Version compatibility

- Go: tested with Go 1.25+
- XML parser: [`github.com/beevik/etree`](https://github.com/beevik/etree)

## Contributing

Contributions and ideas are welcome—please open an issue to discuss.

## License

This project is licensed under the MIT License. See `LICENSE` for details.
