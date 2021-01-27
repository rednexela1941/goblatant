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
	"runtime"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

var (
	write = flag.Bool("w", false, "write to source file instead of stdout")

	exit    = 0
	fileSet = token.NewFileSet()

	parserMode parser.Mode = 0

	typeInformation = &types.Info{
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	packageToIdentifier = make(map[string]string)

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

var MainPackage *types.Package

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

	src, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	// parserMode should use |parseComments -- we will sort this out later.
	file, err := parser.ParseFile(fileSet, filename, src, parserMode)
	if err != nil {
		return err
	}

	conf := types.Config{Importer: importer.Default()}

	MainPackage, err = conf.Check(filename, fileSet, []*ast.File{file}, typeInformation)
	if err != nil {
		return err
	}
	result := astutil.Apply(file, walkPre, walkPost)
	// 	ast.Print(fileSet, result) // print the abstract syntax tree.
	cfg := printer.Config{Mode: printerMode, Tabwidth: tabWidth}

	var buf bytes.Buffer
	err = cfg.Fprint(&buf, fileSet, result)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Print(string(buf.Bytes()))

	if *write {
		// make a temporary backup before overwriting original
		bakname, err := backupFile(filename+".", src, perm)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filename, buf.Bytes(), perm)
		if err != nil {
			os.Rename(bakname, filename)
			return err
		}

		err = os.Remove(bakname)
		if err != nil {
			return err
		}
		return nil
	}

	_, err = out.Write(buf.Bytes())

	return err
}

const chmodSupported = runtime.GOOS != "windows"

// backupFile writes data to a new file named filename<number> with permissions perm,
// with <number randomly chosen such that the file name is unique. backupFile returns
// the chosen file name.
func backupFile(filename string, data []byte, perm os.FileMode) (string, error) {
	// create backup file
	f, err := ioutil.TempFile(filepath.Dir(filename), filepath.Base(filename))
	if err != nil {
		return "", err
	}
	bakname := f.Name()
	if chmodSupported {
		err = f.Chmod(perm)
		if err != nil {
			f.Close()
			os.Remove(bakname)
			return bakname, err
		}
	}
	// write data to backup file
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return bakname, err
}

// No need to walk back.
func walkPost(c *astutil.Cursor) bool {
	return true
}

// walk nodes in syntax tree and swap block-scoped assignments for typed declarations.
func walkPre(c *astutil.Cursor) bool {
	n := c.Node()

	if assign, ok := n.(*ast.AssignStmt); ok {
		assignToDecl(c, assign)
	}

	if imp, ok := n.(*ast.ImportSpec); ok {
		if imp.Path != nil {
			fmt.Println("Path", imp.Path.Value)
		}

		fmt.Printf("Import spec %+v\n", imp)
	}

	return true
}

func assignToDecl(c *astutil.Cursor, assign *ast.AssignStmt) {
	if _, ok := c.Parent().(*ast.BlockStmt); !ok {
		return
	}

	if assign.Tok != token.DEFINE {
		return
	}

	decls := make([]*ast.DeclStmt, 0)
	length := len(assign.Lhs)

	for _, l := range assign.Lhs {
		if ident, ok := l.(*ast.Ident); ok {
			if ident.String() == "_" || ident.Obj.Decl != assign {
				continue
			}
			if obj, ok := typeInformation.Defs[ident]; ok {
				// fmt.Printf("%+v\n", obj.Type())
				typstr := types.TypeString(obj.Type(), nil)
				fmt.Println("Type string", typstr)
				decl := makeDecl(assign, ident, typstr, length)
				decls = append(decls, decl)
			}
		}
	}

	assign.Tok = token.ASSIGN // =

	for _, d := range decls {
		if length == 1 {
			c.Replace(d)
			return
		}
		c.InsertBefore(d)
	}
	c.Replace(assign)
}

func makeDecl(assign *ast.AssignStmt, ident *ast.Ident, identType string, length int) *ast.DeclStmt {
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
		Names: []*ast.Ident{
			ast.NewIdent(ident.Name),
		},
		Values: nil,
		Type:   ast.NewIdent(identType),
	}
	gen.Specs[0] = sp
	if length == 1 {
		sp.Values = []ast.Expr{
			assign.Rhs[0],
		}
	}

	decl := &ast.DeclStmt{Decl: gen}
	return decl
}
