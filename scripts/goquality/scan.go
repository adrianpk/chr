package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type declaration struct {
	ID         string   `json:"id"`
	Location   string   `json:"location"`
	Signature  string   `json:"signature"`
	Parameters int      `json:"parameters"`
	Results    []string `json:"results"`
	Generated  bool     `json:"generated"`
	Build      []string `json:"build,omitempty"`
}

type sourceScan struct {
	Declarations []declaration       `json:"declarations"`
	Calls        map[string][]string `json:"calls"`
	aliases      map[string]ast.Expr
	defined      map[string]bool
	files        *token.FileSet
	path         string
	packagePath  string
	generated    bool
	build        []string
}

func productionFiles(root string) ([]string, error) {
	command := exec.Command("git", "ls-files", "--cached", "--others", "--exclude-standard", "-z", "--", "*.go")
	command.Dir = root

	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("list production sources: %w", err)
	}

	var paths []string

	for _, path := range strings.Split(string(output), "\x00") {
		if ownedSource(path) {
			paths = append(paths, path)
		}
	}

	sort.Strings(paths)

	return paths, nil
}

func ownedSource(path string) bool {
	if path == "" || strings.HasSuffix(path, "_test.go") || strings.Contains(path, "/testdata/") {
		return false
	}

	for _, prefix := range []string{"cmd/", "internal/", "pkg/", "scripts/"} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return !strings.Contains(path, "/")
}

func scanSources(root string) (sourceScan, error) {
	scan := sourceScan{files: token.NewFileSet(), aliases: map[string]ast.Expr{}, defined: map[string]bool{}, Calls: map[string][]string{}}

	paths, err := productionFiles(root)
	if err != nil {
		return scan, err
	}

	parsed := make(map[string]*ast.File, len(paths))
	for _, path := range paths {
		file, parseErr := parser.ParseFile(scan.files, filepath.Join(root, path), nil, parser.ParseComments)
		if parseErr != nil {
			return scan, fmt.Errorf("parse %s: %w", path, parseErr)
		}

		parsed[path] = file
		scan.collectAliases(path, file)
	}

	for _, path := range paths {
		scan.scanFile(path, parsed[path])
	}

	return scan, nil
}

func (scan *sourceScan) collectAliases(path string, file *ast.File) {
	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if ok {
			scan.collectPackageTypes(filepath.Dir(path), group)
		}
	}
}

func (scan *sourceScan) collectPackageTypes(packagePath string, group *ast.GenDecl) {
	for _, candidate := range group.Specs {
		spec, ok := candidate.(*ast.TypeSpec)
		if !ok {
			continue
		}

		key := packagePath + "/" + spec.Name.Name
		if spec.Assign.IsValid() {
			scan.aliases[key] = spec.Type
		} else {
			scan.defined[key] = true
		}
	}
}

func (scan *sourceScan) scanFile(path string, file *ast.File) {
	scan.path, scan.packagePath = path, filepath.Dir(path)

	scan.generated, scan.build = ast.IsGenerated(file), nil
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if strings.HasPrefix(comment.Text, "//go:build ") || strings.HasPrefix(comment.Text, "// +build ") {
				scan.build = append(scan.build, comment.Text)
			}
		}
	}

	for _, node := range file.Decls {
		scan.scanDeclaration(node)
	}

	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if ok {
			scan.recordCall(call)
		}

		return true
	})
}

func (scan *sourceScan) scanDeclaration(node ast.Decl) {
	if function, ok := node.(*ast.FuncDecl); ok {
		name := function.Name.Name
		if function.Recv != nil {
			name = receiverName(function.Recv.List[0].Type) + "." + name
		}

		scan.addSignature(name, function.Type)
		scan.scanFunctionTypes(name, function.Type)

		if function.Body != nil {
			scan.scanNested(name, function.Body)
		}

		return
	}

	group, ok := node.(*ast.GenDecl)
	if !ok {
		return
	}

	for _, spec := range group.Specs {
		scan.scanSpec(spec)
	}
}

func receiverName(expression ast.Expr) string {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.StarExpr:
		return receiverName(value.X)
	case *ast.IndexExpr:
		return receiverName(value.X)
	case *ast.IndexListExpr:
		return receiverName(value.X)
	default:
		return "unknown"
	}
}

func (scan *sourceScan) scanSpec(spec ast.Spec) {
	switch value := spec.(type) {
	case *ast.TypeSpec:
		scan.scanType(value)
	case *ast.ValueSpec:
		if value.Type != nil {
			scan.scanNested(value.Names[0].Name+"/type", value.Type)
		}

		for index, expression := range value.Values {
			name := value.Names[0].Name
			if index < len(value.Names) {
				name = value.Names[index].Name
			}

			scan.scanNested(name, expression)
		}
	}
}

func (scan *sourceScan) scanType(value *ast.TypeSpec) {
	if function, ok := value.Type.(*ast.FuncType); ok {
		scan.addSignature(value.Name.Name, function)
		scan.scanFunctionTypes(value.Name.Name, function)

		return
	}

	if iface, ok := value.Type.(*ast.InterfaceType); ok {
		for _, field := range iface.Methods.List {
			function, isFunction := field.Type.(*ast.FuncType)
			if isFunction {
				for _, name := range field.Names {
					scan.addSignature(value.Name.Name+"."+name.Name, function)
					scan.scanFunctionTypes(value.Name.Name+"."+name.Name, function)
				}
			} else {
				scan.scanNested(value.Name.Name+"/embedded", field.Type)
			}
		}

		return
	}

	scan.scanNested(value.Name.Name, value.Type)
}

func (scan *sourceScan) scanNested(owner string, node ast.Node) {
	if node == nil {
		return
	}

	index := 0

	ast.Inspect(node, func(child ast.Node) bool {
		if function, ok := child.(*ast.FuncType); ok {
			index++
			scan.addSignature(fmt.Sprintf("%s/function/%d", owner, index), function)
		}

		return true
	})
}

func (scan *sourceScan) scanFunctionTypes(owner string, function *ast.FuncType) {
	if function.Params != nil {
		scan.scanNested(owner+"/parameters", function.Params)
	}

	if function.Results != nil {
		scan.scanNested(owner+"/results", function.Results)
	}
}

func (scan *sourceScan) addSignature(name string, function *ast.FuncType) {
	var results []string

	if function.Results != nil {
		for _, field := range function.Results.List {
			for range fieldCount(field) {
				results = append(results, scan.resultType(field.Type, map[string]bool{}))
			}
		}
	}

	scan.Declarations = append(scan.Declarations, declaration{
		ID: scan.path + "::" + name, Location: fmt.Sprintf("%s:%d", scan.path, scan.files.Position(function.Pos()).Line),
		Signature: render(function), Parameters: parameterCount(function.Params), Results: results,
		Generated: scan.generated, Build: scan.build,
	})
}

func fieldCount(field *ast.Field) int {
	if len(field.Names) == 0 {
		return 1
	}

	return len(field.Names)
}

func parameterCount(fields *ast.FieldList) int {
	count := 0

	if fields != nil {
		for _, field := range fields.List {
			count += fieldCount(field)
		}
	}

	return count
}

func (scan *sourceScan) resultType(expression ast.Expr, seen map[string]bool) string {
	identifier, ok := expression.(*ast.Ident)
	if !ok {
		return render(expression)
	}

	binding := scan.aliasBinding(identifier)
	if binding.defined && (identifier.Name == "bool" || identifier.Name == "error") {
		return "defined " + identifier.Name
	}

	if binding.expression == nil || seen[binding.key] {
		return identifier.Name
	}

	seen[binding.key] = true

	return scan.resultType(binding.expression, seen)
}

func render(node ast.Node) string {
	var output bytes.Buffer
	if err := format.Node(&output, token.NewFileSet(), node); err != nil {
		panic(err)
	}

	return output.String()
}

func (scan *sourceScan) recordCall(call *ast.CallExpr) {
	name := ""

	switch function := call.Fun.(type) {
	case *ast.Ident:
		name = function.Name
	case *ast.SelectorExpr:
		name = function.Sel.Name
	}

	if name != "" {
		location := fmt.Sprintf("%s:%d", scan.path, scan.files.Position(call.Pos()).Line)
		scan.Calls[name] = append(scan.Calls[name], location)
	}
}
