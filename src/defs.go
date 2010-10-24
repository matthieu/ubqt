package ubqt

const OpMask uint32 = uint32(^uint8(3<<6))
const AMask  uint32 = uint32(^uint8(0))
const BMask  uint32 = AMask << 1 + 1
const CMask  uint32 = BMask
const BxMask uint32 = BMask + CMask

const (
  ADD = iota // starting with math
  SUB
  MUL
  DIV
  EQ // proceeds to next instruction if (B == C) == bool(A), otherwise skips it
  LT // proceeds to next instruction if (B < C) == bool(A), otherwise skips it
  LE // proceeds to next instruction if (B <= C) == bool(A), otherwise skips it
  TEST
  TESTSET
  MOD
  POW
  UNM // unary minus
  LOADK // loads the constant in B to A
  LOADBOOL
  MOV // copy the content of B in A
  JMP
  RETURN
)

const (
  NIL = iota
  NUM
  BOOL
  STRING
  ARRAY
  MAP
  FUNK
)

type Value struct {
  Num float32
  Type uint8
  Bool bool
}

func (v Value) Truthy() bool {
  switch v.Type {
    case NUM:
      return true
    case BOOL:
      return v.Bool
    default:
      panic("unknown value type: " + string(v.Type))
  }
  return false
}

