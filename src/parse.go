package ubqt

import (
  "fmt"
  v "container/vector"
)

const (
  NAME  = "(name)"
  END   = "(end)"
  LIST  = "(list)"
)

const (
  ART_NAME = iota
  ART_LITERAL
  ART_OPER
  ART_UNARY
  ART_BIN
  ART_TERN
  ART_STMT
  ART_LIST
  ART_FUNK
  ART_THIS
)

type Symbol struct {
  id    string
  lbp   int
  nud   func(s *Token) *Token
  led   func(s *Token, left *Token) *Token
  std   func(s *Token) *Token
  value string
}

type State struct {
  tokens    chan *Token
  current   *Token
  symTable  map[string] *Symbol
  scope     *Scope
}

// Utility

func (s *Symbol) toToken() *Token {
  return &Token{id: s.id, lbp: s.lbp, nud: s.nud, led: s.led, std: s.std, Value: s.value}
}

func listToken(vect *v.Vector) *Token {
  return &Token{id: LIST, Arity: ART_LIST, ttype: TOK_LIST, List: vect}
}

// Scope handling

func newScope(st *State) *Scope {
  st.scope = &Scope{make(map[string] *Token), st.scope}
  return st.scope
}

func (s *Scope) define(n *Token) *Token {
  t := s.defs[n.Value]
  if t != nil {
    panic("Already reserverd: " + n.Value)
  }
  n.reserved = false
  n.nud = func(s *Token) *Token { return s }
  n.led = nil
  n.std = nil
  n.lbp = 0
  n.scope = s
  s.defs[n.Value] = n
  return n
}

func (s *Scope) find(st *State, n string) *Token {
  e := s
  var o *Token
  for {
    o = e.defs[n]
    if o != nil { return o }

    e = e.parent
    if e == nil {
      sym := st.symTable[n]
      if (sym != nil) {
        return sym.toToken()
      } else {
        return st.symTable[NAME].toToken()
      }
    }
  }
  return nil
}

func (s *Scope) pop(st *State) {
  st.scope = s.parent
}

func (s *Scope) reserve(n *Token) {
  if n.Arity != ART_NAME || n.reserved {
    return
  }
  t := s.defs[n.Value]
  if t != nil {
    if t.reserved { return }
    if t.Arity == ART_NAME { panic("Already defined: " + n.Value) }
  }
  n.reserved = true
  s.defs[n.Value] = n
  return
}

// Main parsing functions

func advance(st *State, id string) *Token {
  if len(id) > 0 && st.current.id != id {
    panic("Expected " + id + ", got " + st.current.id)
  }
  if id == END { return nil }

  t := <-st.tokens
  fmt.Printf("__adv %#v\n", t)
  if t.ttype == TOK_EOF {
    st.current = st.symTable[END].toToken()
    return nil
  }
  v := t.Value
  a := t.ttype
  var o *Token
  switch a {
    case TOK_NAME:
      o = st.scope.find(st, v)
    case TOK_OPER:
      o = st.symTable[v].toToken()
      if o == nil {
        panic("Unkown operator: " + v)
      }
    case TOK_STR, TOK_NUM:
      a = ART_LITERAL
      o = st.symTable["(literal)"].toToken()
    default:
      panic("Unexpected token type: " + string(a))
  }
  st.current = o
  st.current.Value = v
  st.current.Arity = a
  return st.current
}

func expression(st *State, rbp int) *Token {
  t := st.current
  advance(st, "")
  fmt.Printf("%#v\n", t.nud)
  left := t.nud(t)

  for rbp < st.current.lbp {
    t = st.current
    advance(st, "")
    left = t.led(t, left)
  }
  return left
}

// Specialized parsing functions

func makeSymbol(symTable map[string] *Symbol, id string, bp int) *Symbol {
  s := symTable[id]
  if s != nil {
    if bp > s.lbp { s.lbp = bp }
  } else {
    s = &Symbol{id:id, lbp:bp}
    s.nud = func(s *Token) *Token { panic("Undefined.") }
    s.led = func(s *Token, left *Token) *Token { panic("Missing operator.") }
    symTable[id] = s
  }
  return s
}

func infixlr(right bool, st *State, id string, bp int, led func(s *Token, left *Token) *Token) *Symbol {
  s := makeSymbol(st.symTable, id, bp)
  if led != nil {
    s.led = led
  } else {
    var rel int
    if right { rel = 1 } else { rel = 0 }
    s.led = func(s *Token, left *Token) *Token {
      s.First = left
      s.Second = expression(st, bp - rel )
      s.Arity = ART_BIN
      return s
    }
  }
  return s
}

func infix(st *State, id string, bp int, led func(s *Token, left *Token) *Token) *Symbol {
  return infixlr(false, st, id, bp, led)
}
func infixr(st *State, id string, bp int, led func(s *Token, left *Token) *Token) *Symbol {
  return infixlr(true, st, id, bp, led)
}

func prefix(st *State, id string, nud func(s *Token) *Token) *Symbol {
  s := makeSymbol(st.symTable, id, 0)
  if nud != nil {
    s.nud = nud
  } else {
    s.nud = func(s *Token) *Token {
      st.scope.reserve(s)
      s.First = expression(st, 70)
      s.Arity = ART_UNARY
      return s
    }
  }
  return s
}

func assignment(st *State, id string) *Symbol {
  return infixr(st, id, 10, func(s *Token, left *Token) *Token {
    if left.id != "." && left.id != "[" && left.Arity != ART_NAME {
      panic("Bad lvalue: " + left.id)
    }
    s.First = left
    s.Second = expression(st, 9)
    s.assignm = true
    s.Arity = ART_BIN
    return s
  })
}

func constant(st *State, id string) *Symbol {
  s := makeSymbol(st.symTable, id, 0)
  s.nud = func(s *Token) *Token {
    st.scope.reserve(s)
    s.Value = st.symTable[s.id].value
    s.Arity = ART_LITERAL
    return s
  }
  s.value = id
  return s
}

func statement(st *State) *Token {
  n := st.current
  fmt.Printf("__stmt %#v\n", n)
  if n.std != nil {
    advance(st, "")
    st.scope.reserve(n)
    return n.std(n)
  }
  vt := expression(st, 0)
  fmt.Printf("__expr %#v\n", vt)
  if !vt.assignm && vt.id != "(" {
    panic("Bad expression statement " + vt.id)
  }
  advance(st, ";")
  return vt
}

func statements(st *State) *Token {
  a := new(v.Vector)
  for {
    fmt.Printf("__stmts %#v\n", st.current)
    if st.current.id == "}" || st.current.id == END { break }
    s := statement(st)
    if s != nil {
      a.Push(s)
    }
  }
  return listToken(a)
}

func stmt(st *State, s string, f func(s *Token) *Token) *Symbol {
  x := makeSymbol(st.symTable, s, 0)
  x.std = f
  return x
}

func block(st *State) *Token {
  t := st.current
  advance(st, "}")
  return t.std(t)
}

func makeState(toks chan *Token) *State {
  symTable := make(map[string] *Symbol)

  makeSymbol(symTable, ":", 0)
  makeSymbol(symTable, ";", 0)
  makeSymbol(symTable, ",", 0)
  makeSymbol(symTable, ")", 0)
  makeSymbol(symTable, "]", 0)
  makeSymbol(symTable, "}", 0)
  makeSymbol(symTable, "else", 0)
  makeSymbol(symTable, END, 0)
  makeSymbol(symTable, NAME, 0)

  st := &State{tokens: toks, symTable:symTable}

  infix(st, "+", 50, nil)
  infix(st, "-", 50, nil)
  infix(st, "*", 60, nil)
  infix(st, "/", 60, nil)
  infix(st, "===", 40, nil)
  infix(st, "!==", 40, nil)
  infix(st, "<=", 40, nil)
  infix(st, ">", 40, nil)
  infix(st, "<", 40, nil)
  infix(st, ">=", 40, nil)

  infix(st, "?", 20, func(s *Token, left *Token) *Token {
    s.First = left
    s.Second = expression(st, 0)
    advance(st, ":")
    s.Third = expression(st, 0)
    s.Arity = ART_TERN
    return s
  })
  infix(st, ".", 80, func(s *Token, left *Token) *Token {
    s.First = left
    if st.current.Arity != ART_NAME {
      panic("Expected a property name after '.'.")
    }
    st.current.Arity = ART_LITERAL
    s.Second = st.current
    s.Arity = ART_BIN
    advance(st, "")
    return s
  })
  infix(st, "[", 80, func(s *Token, left *Token) *Token {
    s.First = left
    s.Second = expression(st, 0)
    s.Arity = ART_BIN
    advance(st, "]")
    return s
  })

  infixr(st, "&&", 30, nil)
  infixr(st, "||", 30, nil)

  prefix(st, "-", nil)
  prefix(st, "!", nil)
  prefix(st, "typeof", nil)
  prefix(st, "(", func(s *Token) *Token {
    e := expression(st, 0)
    advance(st, ")")
    return e
  })

  assignment(st, "=")
  assignment(st, "-=")
  assignment(st, "+=")

  constant(st, "true")
  constant(st, "false")
  constant(st, "null")
  makeSymbol(symTable, "(literal)", 0).nud = func(s *Token) *Token { return s }

  stmt(st, "{", func(s *Token) *Token {
    newScope(st)
    a := statements(st)
    advance(st, "}")
    st.scope.pop(st)
    return a
  })

  stmt(st, "var", func(s *Token) *Token {
    a := new(v.Vector)
    var t *Token
    for {
      n := st.current
      if n.Arity != ART_NAME {
        panic("Expected a new variable name.")
      }
      st.scope.define(n)
      advance(st, "")

      if (st.current.id == "=") {
        t = st.current
        advance(st, "=")
        t.First = n
        t.Second = expression(st, 0)
        t.Arity = ART_BIN
        a.Push(t)
      }

      if st.current.id != "," { break }
      advance(st, ",")
    }
    advance(st, ";")
    return listToken(a)
  })

  stmt(st, "while", func(s *Token) *Token {
    advance(st, "(")
    s.First = expression(st, 0)
    advance(st, ")")
    s.Second = block(st)
    s.Arity = ART_STMT
    return s
  })

  stmt(st, "if", func(s *Token) *Token {
    advance(st, "(")
    s.First = expression(st, 0)
    advance(st, ")")
    s.Second = block(st)

    if st.current.id == "else" {
      st.scope.reserve(st.current)
      advance(st, "else")
      if st.current.id == "if" {
        s.Third = statement(st)
      } else {
        s.Third = block(st)
      }
    } else {
      s.Third = nil
    }
    s.Arity = ART_STMT
    return s
  })

  stmt(st, "break", func(s *Token) *Token {
    advance(st, ";")
    if st.current.id != "}" {
      panic("Unreachable statement: " + st.current.id)
    }
    s.Arity = ART_STMT
    return s
  })

  stmt(st, "return", func(s *Token) *Token {
    if st.current.id != ";" {
      s.First = expression(st, 0)
    }
    advance(st, ";")

    if st.current.id != "}" {
      panic("Unreachable statement: " + st.current.id)
    }
    s.Arity = ART_STMT
    return s
  })

  prefix(st, "function", func(s *Token) *Token {
    a := new(v.Vector)
    newScope(st)
    if st.current.Arity == ART_NAME {
      st.scope.define(st.current)
      s.Value = st.current.Value
      advance(st, "")
    }
    advance(st, "(")

    if st.current.id != ")" {
      for {
        if st.current.Arity != ART_NAME {
          panic("Expected a parameter name: " + st.current.id)
        }
        st.scope.define(st.current)
        a.Push(st.current)
        advance(st, "")
        if st.current.id != "," {
          break
        }
        advance(st, ",")
      }
    }
    s.First = listToken(a)
    advance(st, ")")
    advance(st, "{")
    s.Second = statements(st)
    advance(st, "}")
    s.Arity = ART_FUNK
    st.scope.pop(st)
    return s
  })

  infix(st, ".", 80, func(s *Token, left *Token) *Token {
    a := new(v.Vector)
    if left.id == "." || left.id == "[" {
      s.Arity = ART_TERN
      s.First = left.First
      s.Second = left.Second
      s.Third = listToken(a)
    } else {
      s.Arity = ART_BIN
      s.First = left
      s.Second = listToken(a)
      if (left.Arity != ART_UNARY || left.id != "function") && left.Arity != ART_NAME &&
          left.id != "(" && left.id != "&&" &&left.id != "||" && left.id != "?" {
        panic("Expected a variable name: " + left.id)
      }
    }
    if st.current.id != ")" {
      var p *Token
      for {
        p = expression(st, 0)
        a.Push(p)
        if st.current.id != "," {
          break;
        }
        advance(st, ",")
      }
    }
    advance(st, ")")
    return s
  })

  makeSymbol(symTable, "this", 0).nud = func(s *Token) *Token {
    st.scope.reserve(s)
    s.Arity = ART_THIS
    return s
  }

  fmt.Printf("%#v\n", symTable)
  return st
}

func Parse(toks chan *Token, res chan *Token) {
  st := makeState(toks)
  newScope(st)
  advance(st, "")
  r := statements(st)
  advance(st, END)
  st.scope.pop(st)
  res <- r
}
