package parser

import (
  . "container/vector"
  "fmt"
  "unicode"
  . "parsec"
)

const (
  INT = iota
  OPER
)

type Atom struct {
  ttype int
  text  string
}

func intnum() Parser {
  return Lexeme(func(in Vessel) (Output, bool) {
    ret, ok := Many1(Satisfy(unicode.IsDigit))(in)
    if !ok {
      return nil, false
    }
    sret := "";
    var vret = ret.(Vector)
    for i:= 0; i < vret.Len(); i++ {
      sret += string(vret.At(i).(int))
    }
    return Atom{INT, sret}, ok
  })
}

func operator() Parser {
  return Lexeme(func(in Vessel) (Output, bool) {
    ret, ok := OneOf("+-*/")(in)
    if !ok {
      return nil, false
    }
    return Atom{OPER, string(ret.(int))}, ok
  })
}

func operation() Parser {
  return Collect(intnum(), operator(), Recur(calc))
}

func calc() Parser {
  return Any(Try(operation()), intnum())
}

func main() {
  in := new(StringVessel)
  in.SetSpec(Spec{
    "/*", "*/", "//",
    true,
    Any(Satisfy(unicode.IsLetter), OneOf("_")),
    Any(Satisfy(unicode.IsLetter), Satisfy(unicode.IsDigit), OneOf("_?!")),
    OneOf("+-/*"),
    OneOf("+-/*"),
    nil, nil, true})
  in.SetInput("2 + 3 * 4")

  fmt.Printf("Parsing `%s`...\n", in.GetInput())
  out, ok := calc()(in)
  fmt.Println("done")
  if _, unfinished := in.Next(); unfinished {
    fmt.Printf("Incomplete parse: %s\n", out)
    fmt.Println("Parse error.")
    fmt.Printf("Position: %+v\n", in.GetPosition())
    fmt.Printf("State: %+v\n", in.GetState())
    fmt.Printf("Rest: `%s`\n", in.GetInput())
    return
  }

  fmt.Printf("Parsed: %#v\n", ok)
  fmt.Printf("Tree: %#v\n", out)
  fmt.Printf("Rest: %#v\n", in.GetInput())
}
