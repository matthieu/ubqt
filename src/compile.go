package ubqt

import (
  "fmt"
  "strconv"
)

type Gen struct {
  cpos  uint32
  kpos  uint16
  fn    *Funk
  names map[string] uint16
}

func (gen *Gen) newReg() uint16 {
  ret := gen.fn.maxStack
  gen.fn.maxStack++
  return uint16(ret)
}

func (gen *Gen) pushCode(op uint8, regs... uint16) *Gen {
  opx := uint32(op)
  var ax, bx, cx uint32
  if len(regs) > 0 { ax = uint32(regs[0]) } else { ax = uint32(0) }
  if len(regs) > 1 { bx = uint32(regs[1]) } else { bx = uint32(0) }
  if len(regs) > 2 { cx = uint32(regs[2]) } else { cx = uint32(0) }

  var code uint32
  switch uint32(op) {
    case ADD, SUB, MUL, DIV, EQ, LT, LE:
      code = opx + ax << 6 + bx << 14 + cx << 23
    case LOADK, MOV:
      code = opx + ax << 6 + bx << 14
    case JMP:
      code = opx + bx << 14 + cx << 23
    case RETURN:
      code = uint32(opx)
    default:
      panic("unknown op code: " + string(op))
  }
  gen.fn.code[gen.cpos] = code
  gen.cpos++
  return gen
}

func (gen *Gen) pushConst(v *Value) uint16 {
  // TODO check if the const already exists
  gen.fn.consts[gen.kpos] = v
  k := gen.kpos
  r := gen.newReg()
  gen.pushCode(LOADK, r, k)
  gen.kpos++
  return r
}

func Compile(sourceName string, tok *Token) *Chunk {
  chunk := NewChunk()
  chunk.topfn = funk(sourceName, tok)
  return chunk
}

func funk(sourceName string, tok *Token) *Funk {
  code := make([]uint32, 1024) // reserving 4k for code
  consts := make([]*Value, 50) // 50 constant pointers
  names   := make(map[string] uint16, 255) // 255 names
  funk := &Funk{sourceName: sourceName, code: code, consts: consts, maxStack: 0}
  gen  := &Gen{0, 0, funk, names}
  token(tok, gen)
  return funk
}

func token(tok *Token, gen *Gen) uint16 {
  switch tok.Arity {
    case ART_LIST:
      vlen := tok.List.Len()
      if vlen == 0 { panic("Empty token list") }
      for m := 0; m < vlen; m++ {
        t := token(tok.List.At(m).(*Token), gen)
        if m == vlen-1 { return t }
      }
    case ART_NAME:
      name := gen.names[tok.Value]
      if name == 0 { panic("Unknown name: " + tok.Value) }
    case ART_LITERAL:
      return literal(tok, gen)
    case LOADK, ART_BIN:
      return binOp(tok, gen)
    default:
      panic("Unknown token! " + fmt.Sprintf("%#v", tok))
  }
  panic("Unreachable: " + strconv.Itoa(tok.Arity))
}

func binOp(tok *Token, gen *Gen) uint16 {
  ridx := token(tok.Second, gen)
  if tok.Value == "=" {
    if tok.First.Arity != ART_NAME { panic("Expected name as left value in assignment") }
    name := tok.First.Value
    gen.names[name] = ridx
    return ridx
  }

  lidx := token(tok.First, gen)
  var code uint8
  switch tok.Value {
    case "+": code = ADD
    case "-": code = SUB
    case "*": code = MUL
    case "/": code = DIV
    default:
      panic("Unknown operator: " + tok.Value)
  }
  r := gen.newReg()
  gen.pushCode(code, r, lidx, ridx)
  return r
}

func literal(tok *Token, gen *Gen) uint16 {
  var val *Value
  if tok.Value[0] == 34 {
    val = &Value{Str: tok.Value[1:len(tok.Value)-1]}
  } else {
    f, err := strconv.Atof32(tok.Value)
    if err != nil { panic("Number overflow: " + tok.Value) }
    val = &Value{Num: f}
  }
  return gen.pushConst(val)
}
