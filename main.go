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

	"golang.org/x/tools/go/ast/astutil"
)

var (
	exit       = 0
	fileSet    = token.NewFileSet()
	parserMode parser.Mode

	typeInformation = &types.Info{
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	tabWidth                             = 8
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

func assingToDecl(assign *ast.AssignStmt) (decl *ast.DeclStmt, e error) {
	gen := &ast.GenDecl{
		Doc:    nil,
		Tok:    token.VAR,
		Lparen: token.NoPos,
		Rparen: token.NoPos,
		Specs:  make([]ast.Spec, 0),
	}
	sp := &ast.ValueSpec{
		Comment: nil,
		Doc:     nil,
		Names:   make([]*ast.Ident, len(assign.Lhs)),
		Values:  make([]ast.Expr, len(assign.Rhs)),
	}

	// token_incr := 4
	for i, l := range assign.Lhs {
		if ident, ok := l.(*ast.Ident); ok {
			if obj, ok := typeInformation.Defs[ident]; ok {
				sp.Type = &ast.Ident{
					Name: obj.Type().String(),
				}
				if i == 0 {
					gen.TokPos = l.Pos()
				}
				r := assign.Rhs[i]
				sp.Names[i] = ast.NewIdent(ident.Name)
				sp.Values[i] = r
			}
		}
	}

	gen.Specs = append(gen.Specs, sp)

	decl = &ast.DeclStmt{Decl: gen}
	return decl, nil
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

	_, err = conf.Check(filename, fileSet, []*ast.File{file}, typeInformation)
	if err != nil {
		return err
	}

	result := astutil.Apply(file, walkPre, walkPost)

	ast.Print(fileSet, result) // print the abstract syntax tree.
	var buf bytes.Buffer
	cfg := printer.Config{Mode: printerMode, Tabwidth: tabWidth}
	err = cfg.Fprint(&buf, fileSet, result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(buf.Bytes()))
	return nil
}

// No need to walk back.
func walkPost(c *astutil.Cursor) bool {
	return true
}

// walk nodes in syntax tree and swap block-scoped assignments for typed declarations.
func walkPre(c *astutil.Cursor) bool {
	fmt.Println("Cursor! ", c.Name(), c.Node())
	n := c.Node()

	if assign, ok := n.(*ast.AssignStmt); ok {
		if _, ok := c.Parent().(*ast.BlockStmt); ok {
			decl, err := assingToDecl(assign)
			if err != nil {
				fmt.Println(err)
				return true
			}
			c.Replace(decl)
		}
	}
	return true
}
