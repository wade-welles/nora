package font

// This package loads font-files generated with https://github.com/andryblack/fontbuilder in the NGL XML format.

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"

	"github.com/maja42/vmath"
)

type nglFont struct {
	Type        string         `xml:"type,attr"`
	Description nglDescription `xml:"description"`
	Metrics     nglMetrics     `xml:"metrics"`
	Texture     nglTexture     `xml:"texture"`
	Chars       []nglChar      `xml:"chars>char"`
}

type nglDescription struct {
	Family string `xml:"family,attr"`
	Style  string `xml:"style,attr"`
	Size   int    `xml:"size,attr"`
}

type nglMetrics struct {
	Height    int `xml:"height,attr"`
	Ascender  int `xml:"ascender,attr"`
	Descender int `xml:"descender,attr"`
}

type nglTexture struct {
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
	File   string `xml:"file,attr"`
}

type nglChar struct {
	ID      string `xml:"id,attr"`
	Advance int    `xml:"advance,attr"`
	OffsetX int    `xml:"offset_x,attr"`
	OffsetY int    `xml:"offset_y,attr"`
	RectX   int    `xml:"rect_x,attr"`
	RectY   int    `xml:"rect_y,attr"`
	RectW   int    `xml:"rect_w,attr"`
	RectH   int    `xml:"rect_h,attr"`
}

func Load(path string) (Font, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return Font{}, err
	}

	var ngl nglFont
	if err := xml.Unmarshal(content, &ngl); err != nil || ngl.Type != "NGL" {
		return Font{}, fmt.Errorf("unmarshal ngl xml font: %w", err)
	}

	f := Font{
		Family:    ngl.Description.Family,
		Style:     ngl.Description.Style,
		Size:      ngl.Description.Size,
		Monospace: true,
		Chars:     make(map[rune]Char, len(ngl.Chars)),
		Ascender:  ngl.Metrics.Ascender,
		Descender: ngl.Metrics.Descender,
		Height:    ngl.Metrics.Height,
		Texture:   ngl.Texture.File,
	}

	for i, c := range ngl.Chars {
		runes := []rune(c.ID)
		if len(runes) != 1 { // should not happen
			continue
		}

		if i > 0 && c.Advance != ngl.Chars[i-1].Advance {
			f.Monospace = false
		}

		f.Chars[runes[0]] = Char{
			Width:  c.Advance,
			Offset: vmath.Vec2i{c.OffsetX, c.OffsetY},
			Pos:    vmath.Vec2i{c.RectX, c.RectY},
			Size:   vmath.Vec2i{c.RectW, c.RectH},
		}
	}
	return f, nil
}
