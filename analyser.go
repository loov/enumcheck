package checkenum

import (
	"go/ast"
	"go/types"
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
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	expectedTypes := map[types.Type][]types.Object{}

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
		expectedTypes[obj.Type()] = []types.Object{}
	})

	scope := pass.Pkg.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		typ := obj.Type()
		expected, check := expectedTypes[typ]
		if !check {
			continue
		}

		switch obj.(type) {
		case *types.Const:
			expectedTypes[typ] = append(expected, obj)
		case *types.Var:
			expectedTypes[typ] = append(expected, obj)
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
			expected, ok := expectedTypes[typ]
			if !ok {
				return
			}

			found := map[types.Object]struct{}{}
			for _, clause := range n.Body.List {
				clause := clause.(*ast.CaseClause) // TODO: can this be anything else?
				if len(clause.List) == 0 {
					pass.Reportf(clause.Pos(), "default literal clause for checked enum")
					continue
				}

				for _, option := range clause.List {
					if _, isBasic := option.(*ast.BasicLit); isBasic {
						pass.Reportf(option.Pos(), "basic literal clause for checked enum")
						continue
					}

					id := option.(*ast.Ident)
					obj := pass.TypesInfo.ObjectOf(id)
					found[obj] = struct{}{}
				}
			}

			missing := []string{}
			for _, obj := range expected {
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
			if _, ok := expectedTypes[typ]; !ok {
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
				if _, ok := expectedTypes[obj.Type()]; !ok {
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
