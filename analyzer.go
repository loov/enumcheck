package checkenum

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "checkenum",
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

		if ts.Comment == nil {
			return
		}

		for _, c := range ts.Comment.List {
			if c.Text == "// checkenum" || c.Text == "//checkenum" {
				obj := pass.TypesInfo.Defs[ts.Name]
				pkgEnums[obj.Type()] = &enum{
					Type: obj.Type(),
				}
				return
			}
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
		for _, enum := range pkgEnums {
			sort.Slice(enum.Expected, func(i, k int) bool {
				return enum.Expected[i].Name() < enum.Expected[k].Name()
			})
		}
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
	inspect.WithStack([]ast.Node{
		(*ast.ValueSpec)(nil),
		(*ast.AssignStmt)(nil),
		(*ast.SwitchStmt)(nil),
		(*ast.ReturnStmt)(nil),
	}, func(n ast.Node, push bool, stack []ast.Node) bool {
		switch n := n.(type) {
		case *ast.SwitchStmt:
			typ := pass.TypesInfo.TypeOf(n.Tag)
			enum, ok := enums[typ]
			if !ok {
				return false
			}

			found := map[types.Object]struct{}{}
			for _, clause := range n.Body.List {
				clause := clause.(*ast.CaseClause)
				if clause.List == nil {
					pass.Reportf(clause.Pos(), "%v shouldn't have a default case", typ)
					continue
				}

				for _, option := range clause.List {
					switch option := option.(type) {
					case *ast.BasicLit:
						pass.Reportf(option.Pos(), "implicit conversion of %v to %v", option.Value, typ)
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
				pass.Reportf(n.Pos(), "missing cases %v", humaneList(missing))
			}

		case *ast.ValueSpec:
			// var x, y EnumType = 123, EnumConst
			typ := pass.TypesInfo.TypeOf(n.Type)
			if _, ok := enums[typ]; !ok {
				return false
			}

			for _, rhs := range n.Values {
				if basic, isBasic := rhs.(*ast.BasicLit); isBasic {
					pass.Reportf(n.Pos(), "implicit conversion of %v to %v", basic.Value, typ)
					return false
				}
			}

		case *ast.AssignStmt:
			// var x EnumType
			// x = 123

			// check against right hand side
			check := func(against types.Type, i int) {
				if len(n.Lhs) != len(n.Rhs) {
					// if it's a tuple assignent,
					// then type checker guarantees the assignment
				} else {
					rhs := n.Rhs[i]
					if basic, isBasic := rhs.(*ast.BasicLit); isBasic {
						pass.Reportf(n.Pos(), "implicit conversion of %v to %v", basic.Value, against)
					}
				}
			}

			for i, lhs := range n.Lhs {
				switch lhs := lhs.(type) {
				case *ast.Ident:
					obj := pass.TypesInfo.ObjectOf(lhs)
					if obj == nil {
						continue
					}
					if _, ok := enums[obj.Type()]; !ok {
						continue
					}

					check(obj.Type(), i)
				case ast.Expr:
					typ := pass.TypesInfo.TypeOf(lhs)
					if _, ok := enums[typ]; !ok {
						continue
					}

					check(typ, i)
				default:
					filePos := pass.Fset.Position(n.Pos())
					fmt.Fprintf(os.Stderr, "%v: checkenum internal error: unhandled assignment type %T\n", filePos, lhs)
				}
			}
		case *ast.ReturnStmt:
			// TODO: this probably can be optimized
			var funcDecl *ast.FuncDecl
			for i := len(stack) - 1; i >= 0; i-- {
				decl, ok := stack[i].(*ast.FuncDecl)
				if ok {
					funcDecl = decl
					break
				}
			}
			if funcDecl == nil {
				filePos := pass.Fset.Position(n.Pos())
				fmt.Fprintf(os.Stderr, "%v: checkenum internal error: unable to find func decl\n", filePos)
				return false
			}
			if funcDecl.Type.Results == nil {
				return false
			}

			if funcDecl.Type.Results.NumFields() != len(n.Results) {
				// returning tuples is handled by the compiler
				return false
			}

			returnIndex := 0
			for _, resultField := range funcDecl.Type.Results.List {
				for range resultField.Names {
					typ := pass.TypesInfo.TypeOf(resultField.Type)
					if _, ok := enums[typ]; ok {
						ret := n.Results[returnIndex]
						if basic, isBasic := ret.(*ast.BasicLit); isBasic {
							pass.Reportf(n.Pos(), "implicit conversion of %v to %v", basic.Value, typ)
						}
					}
					returnIndex++
				}
			}
		}

		return false
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
