package ubqt

import (
  "math"
)

type RunEnv struct {
  consts  []*Value
  regs    []*Value
  code    []uint32
}

func NewRunEnv(chunk *Chunk) *RunEnv {
  return &RunEnv{chunk.topfn.consts, make([]*Value, chunk.topfn.maxStack), chunk.topfn.code}
}

func (s *RunEnv) inst(instr uint32) uint8 {
  return uint8(instr & OpMask)
}
func (s *RunEnv) va(instr uint32) uint16 {
  return uint16((instr >> 6) & AMask)
}
func (s *RunEnv) vc(instr uint32) uint16 {
  return uint16((instr >> 23) & CMask)
}
func (s *RunEnv) ra(instr uint32) *Value {
  return s.regs[uint16((instr >> 6) & AMask)]
}
func (s *RunEnv) rb(instr uint32) *Value {
  return s.regs[uint16((instr >> 14) & BMask)]
}
func (s *RunEnv) rc(instr uint32) *Value {
  return s.regs[uint16((instr >> 23) & CMask)]
}
func (s *RunEnv) vbx(instr uint32) int32 {
  return int32((instr >> 14) & BxMask)
}
func (s *RunEnv) kb(instr uint32) *Value {
  return s.consts[uint16((instr >> 14) & BMask)]
}
func (s *RunEnv) rkb(instr uint32) (r *Value) {
  bval := uint16((instr >> 14) & BMask)
  if bval > 250 { r = s.consts[bval] } else { r = s.regs[bval] }
  return
}
func (s *RunEnv) rkc(instr uint32) (r *Value) {
  cval := uint16((instr >> 23) & CMask)
  if cval > 250 { r = s.consts[cval] } else { r = s.regs[cval] }
  return
}

func (s *RunEnv) jmp(i int32) int32 {
  sa := int16(s.code[i] >> 14)
  return int32(sa)
}

func (s *RunEnv) Eval() *Value {
  var pc uint32
  var op uint8
  var va uint16
  i := int32(0)
  L: for {
    pc = s.code[i]
    op = s.inst(pc)
    //fmt.Printf("%s op: %s %s\n ", i, op, inc)
    va = uint16((pc >> 6) & AMask)
    switch op {
      case ADD:
        s.regs[va].Num = s.rb(pc).Num + s.rc(pc).Num
      case SUB:
        s.regs[va].Num = s.rb(pc).Num - s.rc(pc).Num
      case MUL:
        s.regs[va].Num = s.rb(pc).Num * s.rc(pc).Num
      case DIV:
        s.regs[va].Num = s.rb(pc).Num / s.rc(pc).Num
      case MOD:
        s.regs[va].Num = float32(int(s.rb(pc).Num) % int(s.rc(pc).Num))
      case POW:
        s.regs[va].Num = float32(math.Pow(float64(s.rb(pc).Num), float64(s.rc(pc).Num)))
      case UNM:
        s.regs[va].Num = -s.rb(pc).Num
      case EQ:
        i++
        if (s.rb(pc).Num == s.rc(pc).Num) == (va == 1) {
          i += s.jmp(i)
        }
      case LT:
        i++
        if (s.rb(pc).Num < s.rc(pc).Num) == (va == 1) {
          i += s.jmp(i)
        }
      case LE:
        i++
        if (s.rb(pc).Num <= s.rc(pc).Num) == (va == 1) {
          i += s.jmp(i)
        }
      case TESTSET:
        i++
        if (s.rb(pc).Truthy() != (s.vc(pc) == 1)) {
          s.regs[va] = s.rb(pc)
          i += s.jmp(i)
        }
      case TEST:
        i++
        if (s.ra(pc).Truthy() != (s.vc(pc) == 1)) {
          i += s.jmp(i)
        }
      case LOADK:
        s.regs[va] = s.kb(pc)
      case LOADBOOL:
        if s.kb(pc).Num == 0 { s.regs[va].Bool = false } else { s.regs[va].Bool = true }
        if s.vc(pc) > 0 { pc++ };
      case MOV:
        s.regs[va] = s.rb(pc)
      case JMP:
        sa := int16(pc >> 14)
        i += int32(sa)
      case RETURN:
        break L;
      default:
        panic("unknown op code: " + string(op))
    }
    i++
  }
  return s.regs[len(s.regs)-1]
}

//func main() {
//  consts  := [...]*Value{&Value{Type:NUM,Num:2} ,&Value{Type:NUM,Num:5}, &Value{Type:NUM,Num:1}, &Value{Type:NUM,Num:50000}}[0:4]
//  regs    := make([]*Value, 5)
//  code    := make([]uint32, 100005)
//
//  gen := Gen{0, code}
//  gen.PushCode(LOADK, 0, 0) // const 2 in reg 0
//  gen.PushCode(LOADK, 1, 1) // const 5 in reg 1
//  gen.PushCode(LOADK, 2, 2) // const 1 in reg 2
//  gen.PushCode(LOADK, 3, 3) // const 50000 in reg 3
//  gen.PushCode(ADD, 1, 0, 1)
//  gen.PushCode(SUB, 2, 1, 2)
//  gen.PushCode(LE, 1, 2, 3)
//  var offset int16 = -4
//  gen.PushCode(JMP, 0, uint16(offset))
//  gen.PushCode(RETURN)
//
//  bef := time.Nanoseconds()
//  s := RunEnv{consts, regs, code}
//  s.run()
//  aft := time.Nanoseconds()
//  fmt.Printf("result: %s in %s\n", regs[2], float64(aft - bef) / 1000000)
//}
