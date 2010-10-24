package ubqt

type Gen struct {
  pos   uint32
  code  []uint32
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
  gen.code[gen.pos] = code
  gen.pos++;
  return gen
}

func Compile(sourceName string, tok *Token) *Chunk {
  chunk := NewChunk()
  chunk.topfn = compileToken(sourceName, tok)
  return chunk
}

func compileToken(sourceName string, tok *Token) *Funk {
  code := make([]uint32, 1024) // reserving 4k for code
  gen  := Gen{0, code}
//  switch tok.Arity {
//    case ubqt.ART_LIST:
//    case ubqt.ART_NAME:
//    case ubqt.ART_LITERAL:
//    case ubqt.ART_BIN:
//    default:
//      panic("Unknown token! " + fmt.Sprintf("%#v", tok))
//  }
  funk := &Funk{sourceName: sourceName, code: gen.code}
  return funk
}

