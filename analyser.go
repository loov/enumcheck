package checkenum

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "enumcheck",
	Doc:  "check for enum validity",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	FactTypes: []analysis.Fact{
		new(packageEnumsFact),
	},
}

type packageEnumsFact struct {
	enums enumSet
}

func (*packageEnumsFact) AFact() {}
func (pkg *packageEnumsFact) String() string {
	texts := []string{}
	for _, enum := range pkg.enums {
		texts = append(texts, enum.String())
	}
	return strings.Join(texts, ", ")
}

type enumSet map[types.Type]*enum

type enum struct {
	Type     types.Type
	Expected []types.Object
}

func (enum *enum) String() string {
	names := []string{}
	for _, obj := range enum.Expected {
		names = append(names, obj.Name())
	}
	return enum.Type.String() + " = {" + strings.Join(names, " | ") + "}"
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	pkgEnums := enumSet{}

	// collect checked types
	inspect.Preorder([]ast.Node{
		(*ast.TypeSpec)(nil),
	}, func(n ast.Node) {
		ts := n.(*ast.TypeSpec)

		// TODO: determine this based on comments
		if !strings.HasSuffix(ts.Name.Name, "Checked") {
			return
		}

		obj := pass.TypesInfo.Defs[ts.Name]
		pkgEnums[obj.Type()] = &enum{
			Type: obj.Type(),
		}
	})

	scope := pass.Pkg.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		typ := obj.Type()
		enum, check := pkgEnums[typ]
		if !check {
			continue
		}

		switch obj.(type) {
		case *types.Const:
			enum.Expected = append(enum.Expected, obj)
		case *types.Var:
			enum.Expected = append(enum.Expected, obj)
		}
	}

	if len(pkgEnums) > 0 {
		pass.ExportPackageFact(&packageEnumsFact{pkgEnums})
	}

	enums := enumSet{}
	for _, fact := range pass.AllPackageFacts() {
		pkgEnums, ok := fact.Fact.(*packageEnumsFact)
		if !ok {
			continue
		}

		for k, v := range pkgEnums.enums {
			enums[k] = v
		}
	}

	// disallow basic literal declarations and assignments
	inspect.Preorder([]ast.Node{
		(*ast.ValueSpec)(nil),
		(*ast.AssignStmt)(nil),
		(*ast.SwitchStmt)(nil),
	}, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.SwitchStmt:
			typ := pass.TypesInfo.TypeOf(n.Tag)
			enum, ok := enums[typ]
			if !ok {
				return
			}

			found := map[types.Object]struct{}{}
			for _, clause := range n.Body.List {
				clause := clause.(*ast.CaseClause)
				if clause.List == nil {
					pass.Reportf(clause.Pos(), "default literal clause for checked enum")
					continue
				}

				for _, option := range clause.List {
					switch option := option.(type) {
					case *ast.BasicLit:
						pass.Reportf(option.Pos(), "basic literal clause for checked enum")
					case *ast.Ident:
						obj := pass.TypesInfo.ObjectOf(option)
						found[obj] = struct{}{}
					case *ast.SelectorExpr:
						obj := pass.TypesInfo.ObjectOf(option.Sel)
						found[obj] = struct{}{}
					default:
						filePos := pass.Fset.Position(option.Pos())
						fmt.Fprintf(os.Stderr, "%v: checkenum internal error: unhandled clause type %T\n", filePos, option)
					}
				}
			}

			missing := []string{}
			for _, obj := range enum.Expected {
				if _, exists := found[obj]; !exists {
					missing = append(missing, obj.Name())
				}
			}

			if len(missing) > 0 {
				pass.Reportf(n.Pos(), "switch clause missing for %v", humaneList(missing))
			}

		case *ast.ValueSpec:
			// var x, y EnumType = 123, EnumConst
			typ := pass.TypesInfo.TypeOf(n.Type)
			if _, ok := enums[typ]; !ok {
				return
			}

			for _, rhs := range n.Values {
				if _, isBasic := rhs.(*ast.BasicLit); isBasic {
					pass.Reportf(n.Pos(), "basic literal declaration to checked enum")
					return
				}
			}

		case *ast.AssignStmt:
			// var x EnumType
			// x = 123
			for i, lhs := range n.Lhs {
				lhsi, ok := lhs.(*ast.Ident)
				if !ok {
					// TODO: figure out what to do
					continue
				}

				obj := pass.TypesInfo.ObjectOf(lhsi)
				if obj == nil {
					continue
				}
				if _, ok := enums[obj.Type()]; !ok {
					continue
				}

				rhs := n.Rhs[i]
				if _, isBasic := rhs.(*ast.BasicLit); isBasic {
					pass.Reportf(n.Pos(), "basic literal assignment to checked enum")
					return
				}
			}
		}
	})

	return nil, nil
}

func humaneList(list []string) string {
	if len(list) == 0 {
		return ""
	}
	if len(list) == 1 {
		return list[0]
	}

	var s strings.Builder
	for i, item := range list[:len(list)-1] {
		if i > 0 {
			s.WriteString(", ")
		}
		s.WriteString(item)
	}
	s.WriteString(" and ")
	s.WriteString(list[len(list)-1])

	return s.String()
}
