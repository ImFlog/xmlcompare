package xmlcompare

import (
	"fmt"
	"strings"

	"github.com/beevik/etree"
)

// Equal compares two XML strings for structural equality, ignoring element order.
// It returns true if they are equivalent, false otherwise.
func Equal(actual, expected string) (bool, error) {
	actualDoc := etree.NewDocument()
	if err := actualDoc.ReadFromString(actual); err != nil {
		fmt.Printf("error reading actual XML: %v", err)
		return false, err
	}
	expectedDoc := etree.NewDocument()
	if err := expectedDoc.ReadFromString(expected); err != nil {
		fmt.Printf("error reading expected XML: %v", err)
		return false, err
	}
	return compareElements(actualDoc.Root(), expectedDoc.Root()), nil
}

// compareElements compares two etree.Element trees, ignoring element order.
// When a mismatch occurs, it prints an actionable diff and, when relevant,
// includes the exact unexpected XML subtree found in the actual document
// but not present in the expected one.
func compareElements(actual, expected *etree.Element) bool {
	return compareElementsWithPath(actual, expected, "/"+actual.Tag)
}

// compareElementsWithPath performs a detailed, order-independent comparison.
// On the first difference it prints an explicit explanation. If the actual XML
// contains a subtree that is not present in the expected XML, that exact
// unexpected XML snippet is printed to aid debugging.
func compareElementsWithPath(actual, expected *etree.Element, path string) bool {
	// Tag
	if actual.Tag != expected.Tag {
		fmt.Printf("XML mismatch at %s: different tags: actual=<%s> expected=<%s>\n", path, actual.Tag, expected.Tag)
		return false
	}

	// Attributes (order-independent)
	if !compareAttributes(actual, expected, path) {
		return false
	}

	// Children (order-independent) with smart pairing
	actualChildren := actual.ChildElements()
	expectedChildren := expected.ChildElements()
	used := make([]bool, len(expectedChildren))

	for _, actualChild := range actualChildren {
		// 1) Candidates with the same local tag name
		aLocal := actualChild.Tag
		sameTagIdx := sameTagCandidates(aLocal, expectedChildren, used)
		if len(sameTagIdx) == 0 {
			fmt.Printf("Unexpected child at %s/%s\n", path, aLocal)
			return false
		}

		// 2) If only one candidate, recurse into it
		if len(sameTagIdx) == 1 {
			j := sameTagIdx[0]
			childPath := path + "/" + aLocal
			if compareElementsWithPath(actualChild, expectedChildren[j], childPath) {
				used[j] = true
				continue
			}
			return false
		}

		// 3) Choose the most similar expected child and recurse
		bestIdx := chooseMostSimilar(actualChild, expectedChildren, sameTagIdx)
		if bestIdx >= 0 {
			childPath := path + "/" + aLocal
			if compareElementsWithPath(actualChild, expectedChildren[bestIdx], childPath) {
				used[bestIdx] = true
				continue
			}
			return false // mismatch already reported
		}

		// Fallback (should not happen): report unexpected child
		fmt.Printf("Unexpected child at %s/%s\n", path, aLocal)
		return false
	}

	// Any expected children that remain unmatched are missing in actual
	for j, ec := range expectedChildren {
		if !used[j] {
			fmt.Printf("Missing child at %s/%s\n", path, ec.Tag)
			return false
		}
	}

	// Text content (normalized)
	tAct := normalizeXMLText(actual.Text())
	tExp := normalizeXMLText(expected.Text())
	if tAct != tExp {
		fmt.Printf("Text differs at %s: actual=%q expected=%q\n", path, tAct, tExp)
		return false
	}

	return true
}

// compareAttributes checks attributes ignoring order.
func compareAttributes(actual, expected *etree.Element, path string) bool {
	expAttrMap := make(map[string]string, len(expected.Attr))
	for _, attr := range expected.Attr {
		expAttrMap[attr.Key] = attr.Value
	}

	// Report any unexpected or differing attribute in actual
	for _, attr := range actual.Attr {
		if v, ok := expAttrMap[attr.Key]; !ok {
			fmt.Printf("Unexpected attribute at %s: @%s=\"%s\"\n", path, attr.Key, attr.Value)
			return false
		} else if v != attr.Value {
			fmt.Printf("Attribute value differs at %s: @%s actual=\"%s\" expected=\"%s\"\n", path, attr.Key, attr.Value, v)
			return false
		}
	}

	// Missing checks: attributes in actual but not in expected
	actAttrMap := make(map[string]string, len(actual.Attr))
	for _, attr := range actual.Attr {
		actAttrMap[attr.Key] = attr.Value
	}
	for _, attr := range expected.Attr {
		if _, ok := actAttrMap[attr.Key]; !ok {
			fmt.Printf("Missing attribute at %s: expected @%s=\"%s\"\n", path, attr.Key, attr.Value)
			return false
		}
	}
	return true
}

// attrFullKey formats an attribute's qualified name, including prefix when present.
func attrFullKey(a etree.Attr) string {
	if a.Space != "" {
		return a.Space + ":" + a.Key
	}
	return a.Key
}

// isXMLNSAttr returns true if the attribute is a namespace declaration (xmlns or xmlns:prefix).
func isXMLNSAttr(a etree.Attr) bool {
	return a.Space == "xmlns" || a.Key == "xmlns"
}

// sameTagCandidates returns indices of children having the given local tag and not used.
func sameTagCandidates(local string, candidates []*etree.Element, used []bool) []int {
	idx := make([]int, 0)
	for j, ec := range candidates {
		if used[j] {
			continue
		}
		if ec.Tag == local {
			idx = append(idx, j)
		}
	}
	return idx
}

// chooseMostSimilar picks the candidate with the highest heuristic similarity.
func chooseMostSimilar(a *etree.Element, children []*etree.Element, idxs []int) int {
	bestIdx, bestScore := -1, -1
	for _, j := range idxs {
		if score := similarityScore(a, children[j]); score > bestScore {
			bestScore, bestIdx = score, j
		}
	}
	return bestIdx
}

// normalizeXMLText trims and collapses whitespace so that formatting/indentation
// differences don't cause false negatives when comparing equivalent XML.
func normalizeXMLText(s string) string {
	// Collapse all runs of whitespace to a single space and trim ends.
	// If the result is just an empty string, return empty.
	collapsed := strings.Join(strings.Fields(s), " ")
	return collapsed
}

// similarityScore computes a heuristic similarity between two elements to help
// choose the best candidate for detailed comparison.
func similarityScore(a, b *etree.Element) int {
	score := 0

	// Attribute key/value matches (ignore xmlns, use qualified keys)
	mb := make(map[string]string)
	for _, x := range b.Attr {
		if isXMLNSAttr(x) {
			continue
		}
		mb[attrFullKey(x)] = x.Value
	}
	for _, x := range a.Attr {
		if isXMLNSAttr(x) {
			continue
		}
		if v, ok := mb[attrFullKey(x)]; ok && v == x.Value {
			score += 3
		}
	}

	// Overlap of child tag local names (ignoring order and multiplicity)
	setA := childTagSet(a)
	setB := childTagSet(b)
	for name := range setA {
		if setB[name] {
			score += 1
		}
	}

	// Text equality bonus
	if normalizeXMLText(a.Text()) != "" && normalizeXMLText(a.Text()) == normalizeXMLText(b.Text()) {
		score += 1
	}

	// Direct child text matches by tag name (helps list-like structures generically)
	am := directChildTextMap(a)
	bm := directChildTextMap(b)
	// weigh exact text matches per tag
	for name, va := range am {
		if vb, ok := bm[name]; ok && va == vb && va != "" {
			score += 4
		}
	}

	return score
}

// directChildTextMap returns a map of local child tag name -> normalized text (first occurrence).
func directChildTextMap(el *etree.Element) map[string]string {
	m := make(map[string]string)
	for _, c := range el.ChildElements() {
		name := c.Tag
		if _, exists := m[name]; exists {
			continue
		}
		txt := normalizeXMLText(c.Text())
		m[name] = txt
	}
	return m
}

func childTagSet(el *etree.Element) map[string]bool {
	m := make(map[string]bool)
	for _, c := range el.ChildElements() {
		m[c.Tag] = true
	}
	return m
}
