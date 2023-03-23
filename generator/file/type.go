package file

import (
	"fmt"
	"go/types"
	"strconv"
	"strings"
)

type Type interface {
	_type()
	String() string
}

// ============================================================================
// Array
// ======================================================================================

// Array is a [Type] representing arrays and slices.
type Array struct {
	// Len is the length of the array.
	//
	// A negative value indicates that this is a slice.
	Len int64

	Elem Type
}

var _ Type = Array{}

func (t Array) _type() {}

func (t Array) String() string {
	underlying := t.Elem.String()

	var strLen string
	if t.Len >= 0 {
		strLen = strconv.FormatInt(t.Len, 10)
	}

	var b strings.Builder

	var n int
	n += len("[") + len(strLen) + len("]") + len(underlying)
	b.Grow(n)

	b.WriteByte('[')
	b.WriteString(strLen)
	b.WriteByte(']')
	b.WriteString(underlying)
	return b.String()
}

// ============================================================================
// Map
// ======================================================================================

type Map struct {
	Key, Value Type
}

var _ Type = Map{}

func (t Map) _type() {}

func (t Map) String() string {
	var b strings.Builder

	key := t.Key.String()
	value := t.Value.String()

	n := len("map[") + len(key) + len("]") + len(value)
	b.Grow(n)

	b.WriteString("map[")
	b.WriteString(t.Key.String())
	b.WriteByte(']')
	b.WriteString(t.Value.String())
	return b.String()
}

// ============================================================================
// Named
// ======================================================================================

// Named is a [Type] representing all named types, such as structs,
// interfaces, or primitives.
type Named struct {
	// Package is the name of the package this type belongs to, if any.
	//
	// If an import alias was used, Package will be the alias.
	Package string
	// PackagePath is the import path of the package of the type.
	PackagePath string

	// Name is the name of the type.
	Name string
}

func (t Named) _type() {}

func (t Named) String() string {
	var b strings.Builder

	n := len(t.Package) + len(t.Name)
	if t.Package != "" {
		n += len(".")
	}
	b.Grow(n)

	if t.Package != "" {
		b.WriteString(t.Package)
		b.WriteByte('.')
	}
	b.WriteString(t.Name)
	return b.String()
}

// ============================================================================
// Pointer
// ======================================================================================

type Pointer struct {
	Elem Type
}

var _ Type = Pointer{}

func (p Pointer) _type() {}

func (p Pointer) String() string {
	return "*" + p.Elem.String()
}

// ============================================================================
// Func
// ======================================================================================

type Func struct {
	Params  []Type
	Results []Type
}

var _ Type = Func{}

func (f Func) _type() {}

func (f Func) String() string {
	var sb strings.Builder

	sb.WriteString("func(")
	for i, param := range f.Params {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(param.String())
	}
	sb.WriteString(") ")

	if len(f.Results) == 0 {
		return sb.String()
	}

	if len(f.Results) > 1 {
		sb.WriteString("(")
	}
	for i, result := range f.Results {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(result.String())
	}
	if len(f.Results) > 1 {
		sb.WriteString(")")
	}

	return sb.String()
}

// ============================================================================
// Constructor
// ======================================================================================

func NewType(t types.Type, m *ImportManager) (Type, error) {
	switch t := t.(type) {
	case *types.Array:
		return newArrayType(t, m)
	case *types.Basic:
		return newBasicType(t), nil
	case *types.Named:
		return newNamedType(t, m), nil
	case *types.Pointer:
		return newPointerType(t, m)
	case *types.Slice:
		return newSliceType(t, m)
	case *types.Signature:
		return newFuncType(t, m)
	default:
		return nil, fmt.Errorf("unknown type %T", t)
	}
}

func newArrayType(t *types.Array, m *ImportManager) (a Array, err error) {
	a.Len = t.Len()
	a.Elem, err = NewType(t.Elem(), m)
	return a, err
}

func newBasicType(t *types.Basic) Named {
	return Named{Name: t.Name()}
}

func newNamedType(t *types.Named, m *ImportManager) Named {
	n := Named{Name: t.Obj().Name()}
	if pkg := t.Obj().Pkg(); pkg != nil {
		n.PackagePath = pkg.Path()
		n.Package = m.Add(pkg.Path())
	}

	return n
}

func newPointerType(t *types.Pointer, m *ImportManager) (p Pointer, err error) {
	p.Elem, err = NewType(t.Elem(), m)
	return p, err
}

func newSliceType(t *types.Slice, m *ImportManager) (a Array, err error) {
	a.Len = -1
	a.Elem, err = NewType(t.Elem(), m)
	return a, err
}

func newFuncType(t *types.Signature, m *ImportManager) (f Func, err error) {
	for i := 0; i < t.Params().Len(); i++ {
		param := t.Params().At(i)
		typ, err := NewType(param.Type(), m)
		if err != nil {
			return f, err
		}
		f.Params = append(f.Params, typ)
	}

	for i := 0; i < t.Results().Len(); i++ {
		result := t.Results().At(i)
		typ, err := NewType(result.Type(), m)
		if err != nil {
			return f, err
		}
		f.Results = append(f.Results, typ)
	}

	return f, nil
}
