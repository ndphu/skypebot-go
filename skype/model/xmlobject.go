package model

import "encoding/xml"

type URIObject struct {
	XMLName      xml.Name `xml:"URIObject"`
	Uri          string   `xml:"uri,attr"`
	UrlThumbnail string   `xml:"url_thumbnail,attr"`
	Type         string   `xml:"type,attr"`
	DocId        string   `xml:"doc_id,attr"`
	Width        int      `xml:"width,attr"`
	Height       int      `xml:"height,attr"`
	Text         string   `xml:",chardata"`
	ViewLink     ViewLink
	OriginalName OriginalName
	FileSize     FileSize
	Meta         Meta
}

type ViewLink struct {
	XMLName xml.Name `xml:"a"`
	Href    string   `xml:"href,attr"`
	Link    string   `xml:",chardata"`
}

type OriginalName struct {
	XMLName xml.Name `xml:"OriginalName"`
	Name    string   `xml:"v,attr"`
}

type FileSize struct {
	XMLName xml.Name `xml:"FileSize"`
	Size    int      `xml:"v,attr"`
}

type Meta struct {
	XMLName      xml.Name `xml:"meta"`
	OriginalName string   `xml:"originalName,attr"`
	Type         string   `xml:"type,attr"`
}
