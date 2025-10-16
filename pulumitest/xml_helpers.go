package pulumitest

import (
	"encoding/xml"
	"strings"
)

// xmlNode represents a generic XML node that preserves all structure, attributes, and content.
// This approach allows us to manipulate XML files (like .csproj and pom.xml) without losing any data.
type xmlNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content []byte     `xml:",innerxml"`
	Nodes   []xmlNode  `xml:",any"`
}

// UnmarshalXML implements the xml.Unmarshaler interface to properly capture attributes.
func (n *xmlNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.XMLName = start.Name
	n.Attrs = start.Attr
	type node xmlNode
	return d.DecodeElement((*node)(n), &start)
}

// MarshalXML implements the xml.Marshaler interface to properly write attributes.
func (n *xmlNode) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = n.XMLName
	start.Attr = n.Attrs
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if len(n.Nodes) > 0 {
		for _, node := range n.Nodes {
			if err := e.Encode(&node); err != nil {
				return err
			}
		}
	} else if len(n.Content) > 0 {
		if err := e.EncodeToken(xml.CharData(n.Content)); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

// findChild finds a child node by tag name.
func findChild(nodes []xmlNode, tagName string) *xmlNode {
	for i := range nodes {
		if nodes[i].XMLName.Local == tagName {
			return &nodes[i]
		}
	}
	return nil
}

// findAllChildren finds all child nodes by tag name.
func findAllChildren(nodes []xmlNode, tagName string) []*xmlNode {
	var results []*xmlNode
	for i := range nodes {
		if nodes[i].XMLName.Local == tagName {
			results = append(results, &nodes[i])
		}
	}
	return results
}

// getNodeTextContent extracts text content from a node's inner XML.
func getNodeTextContent(node *xmlNode) string {
	return strings.TrimSpace(string(node.Content))
}
