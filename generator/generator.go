package generator

import (
	"fmt"
	"go/types"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"

	"github.com/mavolin/dblog/generator/file"
	"github.com/mavolin/dblog/logger"
)

type Generator struct {
	ImportManager *file.ImportManager

	Loggers []logger.Logger

	// iface is the interface for which the
	iface     *types.Interface
	ifaceName string // example.Repository

	outPath, outTypeName string
}

// New creates a new generator that parses the interface named
// typeName found in path and then generates a wrapper type using the passed
// loggers.
//
// path can either be a relative path to the package containing the interface,
// or a Go import path to it.
func New(path, typeName string, ls ...logger.Logger) (*Generator, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedDeps,
	}, path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse '%s': %w", path, err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("path '%s' did not match a directory containing go files", path)
	}

	pkg := pkgs[0]

	g := &Generator{
		ImportManager: file.NewImportManager(),
		Loggers:       ls,
	}

	for _, l := range ls {
		for _, imp := range l.Imports() {
			g.ImportManager.Add(imp)
		}
	}

	s := pkg.Types.Scope()
	r := s.Lookup(typeName)
	if r == nil {
		return nil, fmt.Errorf("found no type named %s in '%s'", typeName, path)
	}

	named, ok := r.Type().(*types.Named)
	if !ok {
		return nil, fmt.Errorf("expected type to be named, but is %T", r.Type())
	}

	g.ImportManager.Add(named.Obj().Pkg().Path())

	g.ifaceName = named.Obj().Pkg().Name() + "." + named.Obj().Name()

	g.iface, ok = named.Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("expected type to be an interface, but is %T", named.Underlying())
	}

	return g, nil
}

// Generate generates the code for a wrapper type named typeName, and writes it
// to outPath.
func (g *Generator) Generate(outPath, typeName string) error {
	ms, err := g.parse()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(outPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", filepath.Dir(outPath), err)
	}

	pkg := filepath.Base(filepath.Dir(outPath))
	contents, err := g.genFile(pkg, typeName, ms)
	if err != nil {
		return err
	}

	err = os.WriteFile(outPath, []byte(contents), 0666)
	if err != nil {
		return fmt.Errorf("failed to write to file '%s': %w", outPath, err)
	}

	return nil
}

func (g *Generator) parse() ([]file.Method, error) {
	ms := make([]file.Method, g.iface.NumMethods())

	for i := 0; i < g.iface.NumMethods(); i++ {
		f := g.iface.Method(i)

		m, err := file.NewMethod(f, g.ImportManager)
		if err != nil {
			return nil, err
		}

		ms[i] = *m
	}

	return ms, nil
}
