# xml-compare

A tiny Go library to compare two XML documents.

## Install

    go get github.com/imflog/xml-compare

## Usage

```go

import (
    xmlcmp "github.com/imflog/xml-compare"
)

func main() {
    equal := xmlcmp.Equal("<a>x</a>", "<a>x</a>")
    fmt.Println(equal) // true
}
```

# Roadmap
