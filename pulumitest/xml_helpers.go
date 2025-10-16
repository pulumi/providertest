package pulumitest

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// xmlNode represents a simplified XML element tree for editing project files like
// .csproj and pom.xml while preserving element names, attributes, child elements,
// and text-oriented inner XML used by this package.
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
		if err := encodeInnerXML(e, n.Content); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

func encodeInnerXML(e *xml.Encoder, content []byte) error {
	wrapped := make([]byte, 0, len(content)+13)
	wrapped = append(wrapped, []byte("<root>")...)
	wrapped = append(wrapped, content...)
	wrapped = append(wrapped, []byte("</root>")...)

	dec := xml.NewDecoder(bytes.NewReader(wrapped))

	inRoot := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to decode inner XML: %w", err)
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			if !inRoot && tok.Name.Local == "root" {
				inRoot = true
				continue
			}
		case xml.EndElement:
			if inRoot && tok.Name.Local == "root" {
				return nil
			}
		}

		if inRoot {
			if err := e.EncodeToken(tok); err != nil {
				return err
			}
		}
	}
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
