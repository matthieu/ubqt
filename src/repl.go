package main

import (
  "bufio"
  "os"
  "fmt"
  "strings"
  "ubqt"
)

func printTokens(tok *ubqt.Token, pre string) {
  switch tok.Arity {
    case ubqt.ART_LIST:
      println(pre+"[")
      for m := 0; m < tok.List.Len(); m++ {
        if m > 0 { print(", ") }
        printTokens(tok.List.At(m).(*ubqt.Token), pre+"  ")
      }
      println(pre+"]")
    case ubqt.ART_NAME:
      println(pre+"("+tok.Value+")")
    case ubqt.ART_LITERAL:
      println(pre+tok.Value)
    case ubqt.ART_BIN:
      println(pre+"{")
      println(pre+"  value : '"+tok.Value+"'")
      println(pre+"  first :")
      printTokens(tok.First, pre+"    ")
      println(pre+"  second:")
      printTokens(tok.Second, pre+"    ")
      println(pre+"}")
    default:
      fmt.Printf(pre+"%#v\n", tok)
  }
}

func parseString(s string) (res *ubqt.Token) {
  ctok := make(chan *ubqt.Token)
  cpar := make(chan *ubqt.Token)
  go ubqt.Tokenize(strings.NewReader(s), ctok)
  go func() {
    defer func() {
      if r := recover(); r != nil {
        fmt.Printf("Error in parsing: %s\n", r)
        cpar <- nil
      }
    }()
    ubqt.Parse(ctok, cpar)
  }()
  return <- cpar
}

func main() {
  in := bufio.NewReader(os.Stdin)
  for {
    fmt.Print("# ")
    s, _ := in.ReadString('\n')
    toks := parseString(s)
    if toks != nil {
      fmt.Printf("=> %#v\n", toks)
      printTokens(toks, "")
      chunk := ubqt.Compile("<interpr>", toks)
      state := ubqt.NewRunEnv(chunk)
      res := state.Eval()
      println(res.StrLiteral())
    }
  }
}
