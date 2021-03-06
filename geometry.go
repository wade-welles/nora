package nora

import (
	"github.com/maja42/gl"
	"github.com/maja42/nora/assert"
)

// Geometry represents an object's geometric properties.
// Allows reading/merging/manipulating of the underlying geometry.
type Geometry struct {
	vertexCount      int
	vertices         []float32
	indices          []uint16
	primitiveType    PrimitiveType
	vertexAttributes []string
	bufferLayout     BufferLayout
}

// NewGeometry is equivalent as creating an empty geometry object and calling Set() to fill the initial data.
// It is not required to construct a Geometry object using this constructor.
func NewGeometry(vertexCount int, vertices []float32, indices []uint16, primitiveType PrimitiveType, vertexAttributes []string, bufferLayout BufferLayout) *Geometry {
	g := &Geometry{}
	g.Set(vertexCount, vertices, indices, primitiveType, vertexAttributes, bufferLayout)
	return g
}

// Set replaces the existing geometry with new data.
func (g *Geometry) Set(vertexCount int, vertices []float32, indices []uint16, primitiveType PrimitiveType, vertexAttributes []string, bufferLayout BufferLayout) {
	AssertValidGeometry("", vertexCount, vertices, indices, primitiveType, vertexAttributes)
	g.vertexCount = vertexCount
	g.vertices = vertices
	g.indices = indices
	g.primitiveType = primitiveType
	g.vertexAttributes = vertexAttributes
	g.bufferLayout = bufferLayout
}

// Append merges new geometry at the end of the current one.
func (g *Geometry) Append(vertexCount int, vertices []float32, indices []uint16, primitiveType PrimitiveType, vertexAttributes []string, bufferLayout BufferLayout) *Geometry {
	if g.vertexCount == 0 {
		g.Set(vertexCount, vertices, indices, primitiveType, vertexAttributes, bufferLayout)
		return g
	}
	AssertValidGeometry("", vertexCount, vertices, indices, primitiveType, vertexAttributes)

	assert.True(g.primitiveType == primitiveType, "Incompatible primitive type %s<=>%s", g.primitiveType, primitiveType)
	assert.True(equalStringSlice(g.vertexAttributes, vertexAttributes), "Incompatible vertex attributes: %v <> %v", g.vertexAttributes, vertexAttributes)
	assert.True(g.vertexCount+vertexCount+2 <= 0xFFFF, "Resulting geometry is not indexable by uint16") // +2 for possible degenerate triangles
	assert.True((g.indices != nil) == (indices != nil), "Incompatible indexed-drawing property")

	// Support could be added for some of the following cases:

	if len(vertexAttributes) > 1 { // check buffer layout for compatibility
		assert.True(g.bufferLayout == bufferLayout, "Incompatible buffer layouts")  // solvable by splicing the vertex array
		assert.True(bufferLayout == InterleavedBuffer, "Unsupported buffer layout") // solvable by splicing the vertex array
	}
	assert.True(primitiveType == gl.POINTS || primitiveType == gl.LINES || primitiveType == gl.TRIANGLES || primitiveType == gl.TRIANGLE_STRIP, "Unsupported primitive type") // solvable by adding degenerate triangles

	if primitiveType == gl.TRIANGLE_STRIP {
		// create degenerate triangles by duplicating the two vertices at the merge point
		if indices == nil { // duplicate vertices
			vertexSize := len(vertices) / vertexCount
			lastVertex := g.vertices[len(g.vertices)-vertexSize:]
			firstVertex := vertices[:vertexSize]

			g.vertices = append(g.vertices, lastVertex...)
			g.vertices = append(g.vertices, firstVertex...)
			if g.vertexCount%2 != 0 {
				// add another vertex to keep the winding order consistent
				g.vertices = append(g.vertices, firstVertex...)
			}
		} else { // duplicate indices
			lastIdx := uint16(0) //g.indices[len(g.indices)-1]
			firstIdx := indices[0] + uint16(g.vertexCount)
			g.indices = append(g.indices, lastIdx, firstIdx)
			if g.vertexCount%2 != 0 {
				// add another vertex to keep the winding order consistent
				g.indices = append(g.indices, firstIdx)
			}
		}
		g.vertexCount += 2 + g.vertexCount%2
	}

	firstIdx := len(g.indices)
	g.vertices = append(g.vertices, vertices...)
	g.indices = append(g.indices, indices...)
	// offset indices:
	for i := firstIdx; i < len(g.indices); i++ {
		g.indices[i] += uint16(g.vertexCount)
	}
	g.vertexCount += vertexCount
	return g
}

// AppendGeometry merges new geometry at the end of the current one.
func (g *Geometry) AppendGeometry(other *Geometry) *Geometry {
	g.Append(other.vertexCount, other.vertices, other.indices, other.primitiveType, other.vertexAttributes, other.bufferLayout)
	return g
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// AssertValidGeometry checks if the provided geometry is in-itself valid.
// sProgKey is optional. If provided, the geometry is validated against the currently loaded shader program with that key.
func AssertValidGeometry(sProgKey ShaderProgKey, vertexCount int, vertices []float32, indices []uint16, primitiveType PrimitiveType, vertexAttributes []string) {
	// vertex data must be divisible by vertex count
	indexCount := len(indices)
	vertexSize := 0
	if vertexCount > 0 {
		vertexSize = len(vertices) / vertexCount
		assert.True(float32(len(vertices))/float32(vertexCount) == float32(vertexSize), "VertexCount and vertex data does not match")
	} else {
		assert.True(len(vertices) == 0, "VertexCount and vertex data does not match. Should there be vertices?")
	}
	if vertexCount > 0 {
		// there must be attributes
		assert.True(len(vertexAttributes) > 0, "There are no vertex attributes")
		// vertex size must be big enough for all the attributes
		assert.True(vertexSize >= len(vertexAttributes), "Vertex size is too small to fit all vertex attributes") // zero-size attributes don't exist
	}

	if indexCount > 0 {
		// vertices must be indexable
		assert.True(vertexCount <= 0xFFFF, "Too many vertices to be indexed by uint16")
		// indices must not reference out-of-bounds vertices
		minIdx, maxIdx := vertexCount, -1
		assert.Func(func() bool {
			okay := true
			for _, i := range indices {
				idx := int(i)
				if idx > maxIdx {
					maxIdx = idx
				}
				if idx < minIdx {
					minIdx = idx
				}
				if idx >= vertexCount {
					okay = false
				}
			}
			return okay
		}, "Index array references out-of-bounds vertices")

		// Detect unreferenced vertices (doesn't find all, but many):
		obsolete := vertexCount - 1 - maxIdx
		assert.True(obsolete <= 0, "Obsolete vertices: the last %d vertices are not referenced", obsolete)
		assert.True(minIdx <= 0, "Obsolete vertices: the first %d vertices are not referenced", minIdx)
		assert.True(indexCount >= vertexCount, "Obsolete vertices: not all vertices are referenced") // holes in the middle
	}
	// index count must be divisible by number-of-indices-per-primitive
	effectiveIndexCount := indexCount
	if indexCount == 0 {
		effectiveIndexCount = vertexCount
	}

	assertMsg := "Index count %d is incompatible with primitive type %q"
	switch primitiveType {
	case gl.POINTS:
	case gl.LINE_STRIP:
		assert.True(effectiveIndexCount >= 1, assertMsg, effectiveIndexCount, primitiveType)
	case gl.LINE_LOOP:
	case gl.LINES:
		assert.True(effectiveIndexCount%2 == 0, assertMsg, effectiveIndexCount, primitiveType)
	case gl.TRIANGLE_STRIP:
		assert.True(effectiveIndexCount >= 2, assertMsg, effectiveIndexCount, primitiveType)
	case gl.TRIANGLE_FAN:
		assert.True(effectiveIndexCount >= 2, assertMsg, effectiveIndexCount, primitiveType)
	case gl.TRIANGLES:
		assert.True(effectiveIndexCount%3 == 0, assertMsg, effectiveIndexCount, primitiveType)
	default:
		assert.Fail("Unknown primitive type %q", primitiveType)
	}

	// Validate against shader program

	if sProgKey == "" { // shader unknown --> skip
		return
	}

	sProg, _ := engine.Shaders.resolve(sProgKey)
	if !assert.True(sProg != nil, "Shader %q not loaded", sProgKey) {
		return
	}

	// Check existence of attributes
	expectedVertexSize := 0
	for _, attr := range vertexAttributes {
		typ, ok := sProg.attributeTypes[attr]
		expectedVertexSize += int(vaTypePropertyMapping[typ].components)
		assert.True(ok, "Shader %q does not support vertex attribute %s", attr)
	}
	// Check missing attributes
	if vertexCount > 0 && len(sProg.attributeTypes) > len(vertexAttributes) {
		assert.Fail("Shader %q has %d vertex attributes. Geometry only contains %d", sProgKey, len(sProg.attributeTypes), len(vertexAttributes))
	}

	if vertexCount > 0 {
		// Check size of vertex attributes
		assert.True(vertexSize == expectedVertexSize, "Shader %q has %d elements per vertex. Geometry has %d elements.", sProgKey, expectedVertexSize, vertexSize)
	}
}
