package ubqt

import (
  "fmt"
  "io"
  scn "scanner"
  v "container/vector"
)

const (
  TOK_NAME = iota
  TOK_STR
  TOK_NUM
  TOK_OPER
  TOK_LIST
  TOK_EOF
)

type Scope struct {
  defs    map[string] *Token
  parent  *Scope
}

type Token struct {
  id        string
  lbp       int
  ttype     int
  Value     string
  Line      int
  Column    int
  Arity     int
  reserved  bool
  assignm   bool
  scope     *Scope
  First     *Token
  Second    *Token
  Third     *Token
  List      *v.Vector
  nud       func(s *Token) *Token
  led       func(s *Token, left *Token) *Token
  std       func(s *Token) *Token
}

func newToken(ttype int, value string, line int, col int) *Token {
  return &Token{ttype: ttype, Value: value, Line: line, Column: col}
}

func Tokenize(src io.Reader, out chan *Token) {
  var s scn.Scanner
  s.Mode = scn.ScanIdents | scn.ScanFloats | scn.SkipComments | scn.ScanStrings
  s.Init(src)
  var tok int
  for {
    tok = s.Scan()
    fmt.Printf("tok %#v %#v\n", tok, s.TokenText())
    var ttype int
    switch tok {
      case scn.Ident  : ttype = TOK_NAME
      case scn.Float  : ttype = TOK_NUM
      case scn.Int    : ttype = TOK_NUM
      case scn.String : ttype = TOK_STR
      case scn.EOF    : ttype = TOK_EOF
      default         : ttype = TOK_OPER
    }
    fmt.Printf("tt %#v\n", ttype)
    out <- newToken(ttype, s.TokenText(), s.Position.Line, s.Position.Column)
    if tok == scn.EOF { break; }
  }
}
