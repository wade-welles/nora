package shapes

import (
	"fmt"

	"github.com/maja42/gl"
	"github.com/maja42/nora"
	"github.com/maja42/nora/assert"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/color"
	"github.com/maja42/vmath"
)

// Text renders a piece of text with the given font.
// Supports multi-line text.
// Origin = left, baseline. Text height (unscaled) = 1
type Text struct {
	nora.Transform

	font       *nora.Font
	tabWidth   int // in characters
	tabWidthPt float32
	mesh       nora.Mesh

	text   []rune
	bounds vmath.Rectf // calculated

	color color.Color
}

func NewText(font *nora.Font, text string) *Text {
	mat := nora.NewMaterial(shader.COL_TEX_2D)
	mat.AddTextureBinding("sampler", font.TextureKey())

	txt := &Text{
		font:       font,
		tabWidth:   4,
		tabWidthPt: 4 * font.AvgWidth(),
		mesh:       *nora.NewMesh(mat),
		text:       []rune(text),
		color:      color.White,
	}
	txt.ClearTransform()
	txt.update()

	txt.SetColor(color.White)
	return txt
}

func (m *Text) Destroy() {
	m.mesh.Destroy()
}

// FontScaling returns the scaling factor of the underlying font.
// This factor is applied to all font measurements to convert them into model space.
func (m *Text) FontScaling() float32 {
	// Regardless of the font's dimensions, the modelspace height should be 1.
	return 1 / float32(m.font.Height)
}

func (m *Text) update() {
	f := m.font
	scale := m.FontScaling()

	m.bounds.Min[0], m.bounds.Max[0] = 0, 0
	m.bounds.Min[1], m.bounds.Max[1] = float32(f.Ascender)*scale, float32(f.Ascender)*scale // The text's origin is at the baseline. The first line extends upwards.

	vertices := make([]float32, len(m.text)*4*4) // each rune requires 4 vertices; (x, y, u, v) per vertex
	indices := make([]uint16, len(m.text)*6)     // each rune requires 2 triangles

	var origin float32   // X
	var baseline float32 // Y

	vtx := uint16(0)
	idx := 0

	for _, r := range m.text {
		if r == '\r' {
			continue
		}
		if r == '\n' {
			if origin > m.bounds.Max[0] {
				m.bounds.Max[0] = origin
			}
			origin = 0
			baseline -= float32(f.Height) * scale
			continue
		}
		if r == '\t' {
			origin += m.tabWidthPt * scale
			continue
		}

		c, ok := f.Char(r)
		if !assert.True(ok, "Font %s does not contain symbol for rune %s (%v)", f, string(r), r) {
			continue
		}

		/* counter-clockwise
		   3 - 2
		   | / |
		   0 - 1
		*/

		xl := origin + float32(c.Offset[0])*scale
		xr := xl + float32(c.Size[0])*scale

		yt := baseline + float32(c.Offset[1])*scale
		yb := yt - float32(c.Size[1])*scale
		tl, br := f.TexCoord(r)

		copy(vertices[vtx*4:], []float32{
			/*xy*/ xl, yb /*uv*/, tl[0], br[1],
			/*xy*/ xr, yb /*uv*/, br[0], br[1],
			/*xy*/ xr, yt /*uv*/, br[0], tl[1],
			/*xy*/ xl, yt /*uv*/, tl[0], tl[1],
		})
		copy(indices[idx:], []uint16{
			vtx, vtx + 1, vtx + 2,
			vtx + 2, vtx + 3, vtx,
		})

		origin += float32(c.Width) * scale
		vtx += 4
		idx += 6
	}
	if origin > m.bounds.Max[0] {
		m.bounds.Max[0] = origin
	}
	m.bounds.Min[1] = baseline + float32(f.Descender)*scale // account for text the baseline

	// remove unprintable characters (missing runes, new lines, ...):
	vertices = vertices[:vtx*4]
	indices = indices[:idx]

	vertexCount := int(vtx)

	m.mesh.SetVertexData(vertexCount, vertices, indices, gl.TRIANGLES, []string{"position", "texCoord"}, nora.InterleavedBuffer)
}

// Set changes the rendered text.
func (m *Text) Set(text string) {
	m.text = []rune(text)
	m.update()
}

// Get returns the original text.
func (m *Text) Get() string {
	return string(m.text)
}

// Bounds returns the bounding box of the rendered text.
// Includes potential space requirements of ascenders/descenders, even if there are no characters that use them.
func (m *Text) Bounds() vmath.Rectf {
	return m.bounds
}

// Set changes the used font.
func (m *Text) SetFont(font *nora.Font) {
	m.font = font
	m.update()
}

// Font returns the used font.
func (m *Text) Font() *nora.Font {
	return m.font
}

// SetTabWidth changes the width of a tab character (in number-of-characters).
func (m *Text) SetTabWidth(tabWidth int) {
	m.tabWidth = tabWidth
	m.tabWidthPt = float32(tabWidth) * m.font.AvgWidth()
	m.update()
}

// TabWidth returns the width of a tab character (in number-of-characters).
func (m *Text) TabWidth() int {
	return m.tabWidth
}

// SetColor changes the text color.
func (m *Text) SetColor(c color.Color) {
	m.color = c
	m.mesh.Material().Uniform4fColor("color", c)
}

// Color returns the text color.
func (m *Text) Color() color.Color {
	return m.color
}

func (m *Text) Draw(renderState *nora.RenderState) {
	renderState.TransformStack.PushMulRight(m.GetTransform())
	m.mesh.Draw(renderState)
	renderState.TransformStack.Pop()
}

func (m *Text) String() string {
	return fmt.Sprintf("Text(%q/%s)", string(m.text), m.font)
}
