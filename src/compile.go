package ubqt

import (
  "fmt"
  "strconv"
)

type Gen struct {
  cpos  uint32
  kpos  uint32
  fn    *Funk
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

func (gen *Gen) pushConst(v *Value) uint32 {
  gen.fn.consts[gen.kpos] = v
  gen.kpos++
  return gen.kpos
}

func Compile(sourceName string, tok *Token) *Chunk {
  chunk := NewChunk()
  chunk.topfn = funk(sourceName, tok)
  return chunk
}

func funk(sourceName string, tok *Token) *Funk {
  code := make([]uint32, 1024) // reserving 4k for code
  consts := make([]*Value, 50) // 50 constant pointers
  funk := &Funk{sourceName: sourceName, code: code, consts: consts}
  gen  := &Gen{0, 0, funk}
  token(tok, gen)
  return funk
}

func token(tok *Token, gen *Gen) {
  switch tok.Arity {
    case ART_LIST:
    case ART_NAME:
    case ART_LITERAL:
      literal(tok, gen)
    case ART_BIN:
      binOp(tok, gen)
    default:
      panic("Unknown token! " + fmt.Sprintf("%#v", tok))
  }
}

func binOp(tok *Token, gen *Gen) {
}

func literal(tok *Token, gen *Gen) uint32 {
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
