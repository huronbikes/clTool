package clSearch

import (
	"code.google.com/p/go.net/html"
)
/*
   find the first node such that findfn returns a value
 */
func Find(doc *html.Node, findfn func(*html.Node) bool) *html.Node {
	result := findfn(doc)
	var found *html.Node
	if result {
	return doc
	}
	if doc.NextSibling != nil && found == nil {
	found = Find(doc.NextSibling, findfn)
}
if doc.FirstChild != nil && found == nil {
found = Find(doc.FirstChild, findfn)
}
return found
}

/*
  find the first attribute with named by attrKey
 */
func GetAttr(doc *html.Node, attrKey string) *html.Attribute {
	attrs := doc.Attr
	for _, i := range attrs {
		if i.Key == attrKey {
			return &i
		}
	}
	return nil
}

/*
  Returns true if Node doc has an attribute with key attr.Key and value attr.Val
 */

func HasAttr(doc *html.Node, attrKey string, attrVal string) bool {
	testAttr := GetAttr(doc, attrKey)
	if testAttr != nil {
		return testAttr.Val == attrVal
	}
	return false
}
