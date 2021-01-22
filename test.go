package test

import "fmt"

// Note: Lhs -> Names, Rhs -> Values,
//

//                    0: *ast.AssignStmt {
//     44  .  .  .  .  .  .  Lhs: []ast.Expr (len = 2) {
//     45  .  .  .  .  .  .  .  0: *ast.Ident {
//     46  .  .  .  .  .  .  .  .  NamePos: test.go:6:2
//     47  .  .  .  .  .  .  .  .  Name: "a"
//     48  .  .  .  .  .  .  .  .  Obj: *ast.Object {
//     49  .  .  .  .  .  .  .  .  .  Kind: var
//     50  .  .  .  .  .  .  .  .  .  Name: "a"
//     51  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 43)
//     52  .  .  .  .  .  .  .  .  }
//     53  .  .  .  .  .  .  .  }
//     54  .  .  .  .  .  .  .  1: *ast.Ident {
//     55  .  .  .  .  .  .  .  .  NamePos: test.go:6:5
//     56  .  .  .  .  .  .  .  .  Name: "d"
//     57  .  .  .  .  .  .  .  .  Obj: *ast.Object {
//     58  .  .  .  .  .  .  .  .  .  Kind: var
//     59  .  .  .  .  .  .  .  .  .  Name: "d"
//     60  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 43)
//     61  .  .  .  .  .  .  .  .  }
//     62  .  .  .  .  .  .  .  }
//     63  .  .  .  .  .  .  }
//     64  .  .  .  .  .  .  TokPos: test.go:6:7
//     65  .  .  .  .  .  .  Tok: :=
//     66  .  .  .  .  .  .  Rhs: []ast.Expr (len = 2) {
//     67  .  .  .  .  .  .  .  0: *ast.BasicLit {
//     68  .  .  .  .  .  .  .  .  ValuePos: test.go:6:10
//     69  .  .  .  .  .  .  .  .  Kind: INT
//     70  .  .  .  .  .  .  .  .  Value: "1"
//     71  .  .  .  .  .  .  .  }
//     72  .  .  .  .  .  .  .  1: *ast.BasicLit {
//     73  .  .  .  .  .  .  .  .  ValuePos: test.go:6:13
//     74  .  .  .  .  .  .  .  .  Kind: INT
//     75  .  .  .  .  .  .  .  .  Value: "2"
//     76  .  .  .  .  .  .  .  }
//     77  .  .  .  .  .  .  }
//     78  .  .  .  .  .  }
//     79  .  .  .  .  .  1: *ast.DeclStmt {
//     80  .  .  .  .  .  .  Decl: *ast.GenDecl {
//     81  .  .  .  .  .  .  .  TokPos: test.go:7:2
//     82  .  .  .  .  .  .  .  Tok: var
//     83  .  .  .  .  .  .  .  Lparen: -
//     84  .  .  .  .  .  .  .  Specs: []ast.Spec (len = 1) {
//     85  .  .  .  .  .  .  .  .  0: *ast.ValueSpec {
//     86  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 2) {
//     87  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
//     88  .  .  .  .  .  .  .  .  .  .  .  NamePos: test.go:7:6
//     89  .  .  .  .  .  .  .  .  .  .  .  Name: "b"
//     90  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
//     91  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
//     92  .  .  .  .  .  .  .  .  .  .  .  .  Name: "b"
//     93  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 85)
//     94  .  .  .  .  .  .  .  .  .  .  .  .  Data: 0
//     95  .  .  .  .  .  .  .  .  .  .  .  }
//     96  .  .  .  .  .  .  .  .  .  .  }
//     97  .  .  .  .  .  .  .  .  .  .  1: *ast.Ident {
//     98  .  .  .  .  .  .  .  .  .  .  .  NamePos: test.go:7:9
//     99  .  .  .  .  .  .  .  .  .  .  .  Name: "e"
//    100  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
//    101  .  .  .  .  .  .  .  .  .  .  .  .  Kind: var
//    102  .  .  .  .  .  .  .  .  .  .  .  .  Name: "e"
//    103  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 85)
//    104  .  .  .  .  .  .  .  .  .  .  .  .  Data: 0
//    105  .  .  .  .  .  .  .  .  .  .  .  }
//    106  .  .  .  .  .  .  .  .  .  .  }
//    107  .  .  .  .  .  .  .  .  .  }
//    108  .  .  .  .  .  .  .  .  .  Type: *ast.Ident {
//    109  .  .  .  .  .  .  .  .  .  .  NamePos: test.go:7:11
//    110  .  .  .  .  .  .  .  .  .  .  Name: "int"
//    111  .  .  .  .  .  .  .  .  .  }
//    112  .  .  .  .  .  .  .  .  .  Values: []ast.Expr (len = 2) {
//    113  .  .  .  .  .  .  .  .  .  .  0: *ast.BasicLit {
//    114  .  .  .  .  .  .  .  .  .  .  .  ValuePos: test.go:7:17
//    115  .  .  .  .  .  .  .  .  .  .  .  Kind: INT
//    116  .  .  .  .  .  .  .  .  .  .  .  Value: "2"
//    117  .  .  .  .  .  .  .  .  .  .  }
//    118  .  .  .  .  .  .  .  .  .  .  1: *ast.BasicLit {
//    119  .  .  .  .  .  .  .  .  .  .  .  ValuePos: test.go:7:20
//    120  .  .  .  .  .  .  .  .  .  .  .  Kind: INT
//    121  .  .  .  .  .  .  .  .  .  .  .  Value: "5"
//    122  .  .  .  .  .  .  .  .  .  .  }
//    123  .  .  .  .  .  .  .  .  .  }
//    124  .  .  .  .  .  .  .  .  }
//    125  .  .  .  .  .  .  .  }
//    126  .  .  .  .  .  .  .  Rparen: -
//    127  .  .  .  .  .  .  }
//    128  .  .  .  .  .  }

func main() {
	a, d := 1, 2
	var b, e int = 2, 5
	c := uint8(3)
	fmt.Println(a+b, c, d, e)
}
