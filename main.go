package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	exit       = 0
	fileSet    = token.NewFileSet()
	parserMode parser.Mode

	Identifiers = []*ast.Ident{}
	info        = &types.Info{
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	tabWidth = 8

	printerMode                          = printer.UseSpaces | printer.TabIndent | printerNormalizeNumbers
	printerNormalizeNumbers printer.Mode = 1 << 30
)

func main() {
	flag.Parse()
	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		switch dir, err := os.Stat(path); {
		case err != nil:
			report(err)
		case dir.IsDir():
			filepath.Walk(path, visitFile)
		default:
			if err := processFile(path, nil, os.Stdout, false); err != nil {
				report(err)
			}
		}
	}
	os.Exit(exit)
}

type VisitorFunc func(n ast.Node) ast.Visitor

func (f VisitorFunc) Visit(n ast.Node) ast.Visitor { return f(n) }

func assingToDecl(assign *ast.AssignStmt) (decl *ast.DeclStmt, e error) {
	gen := &ast.GenDecl{
		Doc:    nil,
		Tok:    token.VAR,
		Lparen: token.NoPos,
		Rparen: token.NoPos,
		Specs:  make([]ast.Spec, 1),
	}
	sp := &ast.ValueSpec{
		Comment: nil,
		Doc:     nil,
		Names:   make([]*ast.Ident, len(assign.Lhs)),
		Values:  make([]ast.Expr, len(assign.Rhs)),
	}
	sp.Comment = nil
	sp.Doc = nil

	// token_incr := 4
	for i, l := range assign.Lhs {
		if ident, ok := l.(*ast.Ident); ok {
			if obj, ok := info.Defs[ident]; ok {
				fmt.Printf("%+v obj \n", obj.Type())
				sp.Type = &ast.Ident{
					Name: obj.Type().String(),
				}
				if i == 0 {
					gen.TokPos = l.Pos()
				}
				r := assign.Rhs[i]
				if expr, ok := r.(ast.Expr); ok {
					sp.Names[i] = ident
					sp.Values[i] = expr
				}
			}
		}
	}

	decl = &ast.DeclStmt{Decl: gen}
	return decl, nil
}

func FindTypes(n ast.Node) ast.Visitor {
	if decl, ok := n.(*ast.AssignStmt); ok {
		gen, err := assingToDecl(decl)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		return nil
	}
	return VisitorFunc(FindTypes)
}

func walkPrev(c *Cursor) bool {
	return false
}

func walkNext(c *Cursor) bool {
	return false
}

func processFile(filename string, in io.Reader, out io.Writer, stdin bool) error {
	var perm os.FileMode = 0644
	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		fi, err := f.Stat()
		if err != nil {
			return err
		}
		in = f
		perm = fi.Mode().Perm()
	}
	fmt.Println(perm)

	src, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	file, err := parser.ParseFile(fileSet, filename, src, parserMode)
	if err != nil {
		return err
	}

	conf := types.Config{Importer: importer.Default()}

	pkg, err := conf.Check(filename, fileSet, []*ast.File{file}, info)
	if err != nil {
		log.Fatal(err) // type error
	}

	astutil.Apply(file)

	ast.Walk(VisitorFunc(FindTypes), file)

	fmt.Println("Defs : ", info.Defs)
	fmt.Println("Types : ", info.Types)
	for id, obj := range info.Defs {
		fmt.Printf("Def : %+v %+v \n", id, obj)

	}
	for k, v := range info.Types {
		fmt.Printf("Type %+v %+v \n", k, v)
	}

	ast.Inspect(file, func(n ast.Node) bool {
		if expr, ok := n.(ast.Expr); ok {
			if tv, ok := info.Types[expr]; ok {
				fmt.Printf("\t\t\t\ttype:  %v\n", tv.Type)
				if tv.Value != nil {
					fmt.Printf("\t\t\t\tvalue: %v\n", tv.Value)
				}
			}
		}
		return true
	})

	fmt.Printf("Package  %q\n", pkg.Path())
	fmt.Printf("Name:    %s\n", pkg.Name())
	fmt.Printf("Imports: %s\n", pkg.Imports())
	fmt.Printf("Scope:   %s\n", pkg.Scope().Names())
	for i := 0; i < pkg.Scope().Len(); i++ {
		c := pkg.Scope().Child(i)
		fmt.Println(c.Names(), "Names")
	}

	for _, ident := range Identifiers {
		fmt.Printf("%+v ident \n", ident)
		if obj, ok := info.Defs[ident]; ok {
			fmt.Printf("%+v obj \n", obj.Type())
		}
	}

	var buf bytes.Buffer
	cfg := printer.Config{Mode: printerMode, Tabwidth: tabWidth}
	err = cfg.Fprint(&buf, fileSet, file)
	if err != nil {
		return nil
	}

	fmt.Print(string(buf.Bytes()))

	return nil
}

func isGoFile(f os.FileInfo) bool {
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

func visitFile(path string, f os.FileInfo, err error) error {
	if err == nil && isGoFile(f) {
		err = processFile(path, nil, os.Stdout, false)
	}
	// Don't complain if a file was deleted in the meantime (i.e.
	// the directory changed concurrently while running gofmt).
	if err != nil && !os.IsNotExist(err) {
		report(err)
	}
	return nil
}

func report(e error) {
	fmt.Fprintln(os.Stderr, e.Error())
	exit = 2
}
