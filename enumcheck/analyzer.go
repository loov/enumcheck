package enumcheck

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"sort"
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
	Pkg      *types.Package
	Mode     enumMode
	Type     types.Type
	TypeEnum bool
	Values   []types.Object
	Types    []types.Type
}

func (enum *enum) ContainsType(t types.Type) bool {
	if !enum.TypeEnum {
		return true
	}
	for _, typ := range enum.Types {
		if typ == t {
			return true
		}
	}
	return false
}

func (enum *enum) String() string {
	names := []string{}
	for _, obj := range enum.Values {
		names = append(names, obj.Name())
	}
	for _, typ := range enum.Types {
		names = append(names, types.TypeString(typ, types.RelativeTo(enum.Pkg)))
	}
	return enum.Type.String() + " = {" + strings.Join(names, " | ") + "}"
}

type enumMode byte

const (
	// modeExhaustive, requires default blocks
	modeExhaustive enumMode = 1
	// modeRelaxed, makes default block optional
	modeRelaxed enumMode = 2
)

func (mode enumMode) NeedsDefault() bool {
	return mode == modeExhaustive
}

func contains(tags []string, s string) bool {
	for _, x := range tags {
		if s == x {
			return true
		}
	}
	return false
}

type enumComment struct {
	mode   enumMode
	ignore bool
}

func isEnumcheckComment(comment string) (enumComment, bool) {
	comment = strings.TrimSpace(strings.TrimPrefix(comment, "//"))
	matches := comment == "enumcheck" || strings.HasPrefix(comment, "enumcheck:")
	if !matches {
		return enumComment{}, false
	}

	var c enumComment
	c.mode = modeExhaustive

	args := strings.TrimPrefix(strings.TrimPrefix(comment, "enumcheck"), ":")
	for _, x := range strings.Split(args, ",") {
		switch strings.TrimSpace(x) {
		case "":
		case "exhaustive":
			c.mode = modeExhaustive
		case "relaxed":
			c.mode = modeRelaxed
		case "ignore":
			c.ignore = true
		}
	}

	return c, true
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	pkgEnums := enumSet{}

	addTypeSpec := func(ts *ast.TypeSpec, c enumComment) {
		obj := pass.TypesInfo.Defs[ts.Name]
		pkgEnums[obj.Type()] = &enum{
			Pkg:      obj.Pkg(),
			Type:     obj.Type(),
			TypeEnum: types.IsInterface(obj.Type()),
			Mode:     c.mode,
		}
	}

	// collect checked types
	inspect.Preorder([]ast.Node{
		(*ast.GenDecl)(nil),
	}, func(n ast.Node) {
		gd := n.(*ast.GenDecl)

		var check *enumComment
		if gd.Doc != nil {
			for _, c := range gd.Doc.List {
				if c, ok := isEnumcheckComment(c.Text); ok {
					check = &c
					break
				}
			}
		}

	nextSpec:
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue nextSpec
			}

			if check != nil {
				addTypeSpec(ts, *check)
				continue nextSpec
			}

			if ts.Doc != nil {
				for _, c := range ts.Doc.List {
					if c, ok := isEnumcheckComment(c.Text); ok {
						addTypeSpec(ts, c)
						continue nextSpec
					}
				}
			}

			if ts.Comment != nil {
				for _, c := range ts.Comment.List {
					if c, ok := isEnumcheckComment(c.Text); ok {
						addTypeSpec(ts, c)
						continue nextSpec
					}
				}
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
			enum.Values = append(enum.Values, obj)
		case *types.Var:
			enum.Values = append(enum.Values, obj)
		}
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.ValueSpec:
						typ := pass.TypesInfo.TypeOf(spec.Type)
						enum, check := pkgEnums[typ]
						if !check {
							continue
						}

						for _, value := range spec.Values {
							typ := pass.TypesInfo.TypeOf(value)
							enum.Types = append(enum.Types, typ)
						}
					}
				}
			}
		}
	}

	if len(pkgEnums) > 0 {
		for _, enum := range pkgEnums {
			sort.Slice(enum.Values, func(i, k int) bool {
				return enum.Values[i].Name() < enum.Values[k].Name()
			})
			sort.Slice(enum.Types, func(i, k int) bool {
				return enum.Types[i].String() < enum.Types[k].String()
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

	type ignoreKey struct {
		file *token.File
		line int
	}

	ignore := map[ignoreKey]struct{}{}

	for _, file := range pass.Files {
		for _, group := range file.Comments {
			for _, comment := range group.List {
				if x, ok := isEnumcheckComment(comment.Text); ok {
					if x.ignore {
						file := pass.Fset.File(comment.Pos())
						line := file.Line(comment.Pos())
						ignore[ignoreKey{file: file, line: line}] = struct{}{}
					}
				}

			}
		}
	}

	reportf := func(pos token.Pos, format string, args ...interface{}) {
		file := pass.Fset.File(pos)
		line := file.Line(pos)
		if _, ok := ignore[ignoreKey{file: file, line: line}]; ok {
			return
		}
		pass.Reportf(pos, format, args...)
	}

	// disallow basic literal declarations and assignments
	inspect.WithStack([]ast.Node{
		(*ast.ValueSpec)(nil),
		(*ast.AssignStmt)(nil),
		(*ast.SwitchStmt)(nil),
		(*ast.TypeSwitchStmt)(nil),
		(*ast.ReturnStmt)(nil),
		(*ast.SendStmt)(nil),
		(*ast.CallExpr)(nil),
	}, func(n ast.Node, push bool, stack []ast.Node) bool {
		switch n := n.(type) {
		case *ast.SwitchStmt:
			typ := pass.TypesInfo.TypeOf(n.Tag)
			enum, ok := enums[typ]
			if !ok {
				return false
			}

			foundValues := map[types.Object]struct{}{}
			foundDefault := false
			for _, clause := range n.Body.List {
				clause := clause.(*ast.CaseClause)
				if clause.List == nil {
					foundDefault = true
					continue
				}

				for _, option := range clause.List {
					switch option := option.(type) {
					case *ast.BasicLit:
						reportf(option.Pos(), "implicit conversion of %v to %v", option.Value, typ)
					case *ast.Ident:
						obj := pass.TypesInfo.ObjectOf(option)
						foundValues[obj] = struct{}{}
					case *ast.SelectorExpr:
						obj := pass.TypesInfo.ObjectOf(option.Sel)
						foundValues[obj] = struct{}{}
					case *ast.CompositeLit:
						reportf(option.Pos(), "invalid enum for %v", typ)
					default:
						filePos := pass.Fset.Position(option.Pos())
						fmt.Fprintf(os.Stderr, "%v: enumcheck internal error: unhandled clause type %T\n", filePos, option)
					}
				}
			}

			missing := []string{}
			for _, obj := range enum.Values {
				if _, exists := foundValues[obj]; !exists {
					missing = append(missing, obj.Name())
				}
			}
			if enum.Mode.NeedsDefault() && !foundDefault {
				missing = append(missing, "default")
			}

			if len(missing) > 0 {
				reportf(n.Pos(), "missing cases %v", humaneList(missing))
			}

		case *ast.TypeSwitchStmt:
			var typ types.Type
			switch a := n.Assign.(type) {
			case *ast.AssignStmt:
				if len(a.Rhs) == 1 {
					if a, ok := a.Rhs[0].(*ast.TypeAssertExpr); ok {
						typ = pass.TypesInfo.TypeOf(a.X)
					}
				}
			case *ast.ExprStmt:
				if a, ok := a.X.(*ast.TypeAssertExpr); ok {
					typ = pass.TypesInfo.TypeOf(a.X)
				}
			default:
				return false
			}
			enum, ok := enums[typ]
			if !ok {
				return false
			}

			foundTypes := map[types.Type]struct{}{}
			for _, clause := range n.Body.List {
				clause := clause.(*ast.CaseClause)
				for _, option := range clause.List {
					t := pass.TypesInfo.TypeOf(option)
					if t == nil {
						filePos := pass.Fset.Position(option.Pos())
						fmt.Fprintf(os.Stderr, "%v: enumcheck internal error: unhandled clause type %T\n", filePos, option)
						continue
					}

					foundMatch := false
					for _, typ := range enum.Types {
						if typ == t {
							foundMatch = true
							break
						}
					}
					if !foundMatch {
						reportf(option.Pos(), "implicit conversion of %v to %v", t.String(), enum.Type)
					}

					foundTypes[t] = struct{}{}
				}
			}

			missing := []string{}
			for _, typ := range enum.Types {
				if _, exists := foundTypes[typ]; !exists {
					missing = append(missing, typ.String())
				}
			}

			if len(missing) > 0 {
				reportf(n.Pos(), "missing cases %v", humaneList(missing))
			}

		case *ast.ValueSpec:
			// var x, y EnumType = 123, EnumConst
			typ := pass.TypesInfo.TypeOf(n.Type)
			enum, ok := enums[typ]
			if !ok {
				return false
			}

			for _, rhs := range n.Values {
				if basic, isBasic := rhs.(*ast.BasicLit); isBasic {
					reportf(n.Pos(), "implicit conversion of %v to %v", basic.Value, typ)
					return false
				}
				rhstyp := pass.TypesInfo.TypeOf(rhs)
				if !enum.ContainsType(rhstyp) {
					reportf(n.Pos(), "implicit conversion of %v to %v", rhstyp, typ)
					return false
				}
			}

		case *ast.AssignStmt:
			// var x EnumType
			// x = 123

			// check against right hand side
			check := func(enum *enum, against types.Type, i int) {
				if len(n.Lhs) != len(n.Rhs) {
					// if it's a tuple assignent,
					// then type checker guarantees the assignment
				} else {
					rhs := n.Rhs[i]
					if basic, isBasic := rhs.(*ast.BasicLit); isBasic {
						reportf(n.Pos(), "implicit conversion of %v to %v", basic.Value, against)
					}

					rhstyp := pass.TypesInfo.TypeOf(rhs)
					if !enum.ContainsType(rhstyp) {
						reportf(n.Pos(), "implicit conversion of %v to %v", rhstyp, enum.Type)
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
					enum, ok := enums[obj.Type()]
					if !ok {
						continue
					}
					check(enum, obj.Type(), i)
				case ast.Expr:
					typ := pass.TypesInfo.TypeOf(lhs)
					enum, ok := enums[typ]
					if !ok {
						continue
					}

					check(enum, typ, i)
				default:
					filePos := pass.Fset.Position(n.Pos())
					fmt.Fprintf(os.Stderr, "%v: enumcheck internal error: unhandled assignment type %T\n", filePos, lhs)
				}
			}

			for _, rhs := range n.Rhs {
				if callExpr, ok := rhs.(*ast.CallExpr); ok {
					verifyCallExpr(reportf, pass, enums, callExpr)
					continue
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
				fmt.Fprintf(os.Stderr, "%v: enumcheck internal error: unable to find func decl\n", filePos)
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
					enum, ok := enums[typ]
					if ok {
						ret := n.Results[returnIndex]
						if basic, isBasic := ret.(*ast.BasicLit); isBasic {
							reportf(n.Pos(), "implicit conversion of %v to %v", basic.Value, enum.Type)
						}
						rettyp := pass.TypesInfo.TypeOf(ret)
						if !enum.ContainsType(rettyp) {
							reportf(n.Pos(), "implicit conversion of %v to %v", rettyp, enum.Type)
							return false
						}
					}
					returnIndex++
				}
			}

		case *ast.SendStmt:
			chanType := pass.TypesInfo.TypeOf(n.Chan)
			if named, ok := chanType.(*types.Named); ok {
				chanType = named.Underlying()
			}

			switch typ := chanType.(type) {
			case *types.Chan:
				enum, ok := enums[typ.Elem()]
				if !ok {
					return false
				}
				if basic, isBasic := n.Value.(*ast.BasicLit); isBasic {
					reportf(n.Pos(), "implicit conversion of %v to %v", basic.Value, enum.Type)
				}
				valtyp := pass.TypesInfo.TypeOf(n.Value)
				if !enum.ContainsType(valtyp) {
					reportf(n.Pos(), "implicit conversion of %v to %v", valtyp, enum.Type)
					return false
				}
			default:
				filePos := pass.Fset.Position(n.Pos())
				fmt.Fprintf(os.Stderr, "%v: enumcheck internal error: unhandled SendStmt.Chan type %T\n", filePos, chanType)
				return false
			}

		case *ast.CallExpr:
			verifyCallExpr(reportf, pass, enums, n)

		default:
			filePos := pass.Fset.Position(n.Pos())
			fmt.Fprintf(os.Stderr, "%v: enumcheck internal error: unhandled %T\n", filePos, n)
		}

		return false
	})

	return nil, nil
}

type reportFn func(pos token.Pos, format string, args ...interface{})

func verifyCallExpr(reportf reportFn, pass *analysis.Pass, enums enumSet, n *ast.CallExpr) {
	fn := pass.TypesInfo.TypeOf(n.Fun)
	sig, ok := fn.(*types.Signature)
	if !ok {
		return
	}

	params := sig.Params()
	for i := 0; i < params.Len(); i++ {
		param := params.At(i)
		enum, ok := enums[param.Type()]
		if !ok {
			continue
		}

		arg := n.Args[i]
		if basic, isBasic := arg.(*ast.BasicLit); isBasic {
			reportf(n.Pos(), "implicit conversion of %v to %v", basic.Value, enum.Type)
		}

		argtyp := pass.TypesInfo.TypeOf(arg)
		if !enum.ContainsType(argtyp) {
			reportf(n.Pos(), "implicit conversion of %v to %v", argtyp, enum.Type)
			return
		}
	}
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
