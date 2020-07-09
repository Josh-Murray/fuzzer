package main

import (
  "fmt"
  "time"
  "math/rand"
)

type Mutator struct {
  // TODO suppose this lad will need an input channel to recieve from
  // TODO suppose this lad will need a channel to output from
  /* TODO add configurables here
   * rng seed
   * num mutations
   * etc
   */
  out_chan chan TestCase
  rng *rand.Rand

}
// TODO work out configurables, they might be needed here
func createMutator(out chan TestCase) Mutator {
  r := rand.New(rand.NewSource(time.Now().Unix()))
  return Mutator{rng: r, out_chan: out}
}

func (m Mutator)flip_bits(ts * TestCase){
  // flip bit in 10% of bytes, could change to $config% bytes or randRange bytes
  size := float64(len(ts.input)) * 0.1
  nbytes := int(size)
  if nbytes == 0 {
    nbytes = 1
  }
  for i:=0; i < nbytes; i++ {
    index := m.rng.Intn(len(ts.input))
    offset := m.rng.Intn(8)
    msg := fmt.Sprintf("Mutator performed 'flip_bits' on byte %d, bit %d\n", index, offset)
    ts.changes = append(ts.changes, msg)
    ts.input[index] ^= 1 << offset;
  }
}

func (m Mutator)flip_bytes(ts * TestCase){
  // flip 10% of bytes
  size := float64(len(ts.input)) * 0.1
  nbytes := int(size)
  if nbytes == 0 {
    nbytes = 1
  }
  for i:=0; i < nbytes; i++ {
    index := m.rng.Intn(len(ts.input))

    msg := fmt.Sprintf("Mutator performed 'flip_bytes' on byte %d\n", index)
    ts.changes = append(ts.changes, msg)
    ts.input[index] ^= 255;
  }
}

/* should monitor the probability this gets called to ensure the input
 * pool doesnt converge to 0 length inputs
 */
func (m Mutator)delete_slice(ts * TestCase){
  // used len too much, use a variable instead
  length := len(ts.input)
  if length == 0{
    return
  }
  start := m.rng.Intn(length-1);
  size := float64(length) * 0.2
  end :=  start + m.rng.Intn(int(size))
  if end > length {
    end = length
  }
  msg := fmt.Sprintf("Mutator performed 'delete_slice' on input[%d:%d]\n", start,end)
  ts.changes = append(ts.changes, msg)
  ts.input = append(ts.input[:start], ts.input[end:]...)
}

/* should monitor the probability this gets called to ensure the input
 * pool doesnt become too long
 */
func (m Mutator) duplicate_slice(ts * TestCase){
  // used len too much, use a variable instead
  length := len(ts.input)

  start := m.rng.Intn(length-1);
  size := float64(length) * 0.2
  end :=  start + m.rng.Intn(int(size))
  if end > length {
    end = length
  }
  // this can probably be cleaned up
  tmp := make([]byte, len(ts.input[end:]))
  copy(tmp, ts.input[end:])
  msg := fmt.Sprintf("Mutator performed 'duplicate_slice' on '%s'\n", string(ts.input[start:end]))
  ts.changes = append(ts.changes, msg)
  ts.input = append(ts.input[:end], ts.input[start:end]...)
  ts.input = append(ts.input, tmp...)
}

// TODO add a int16 and int32 equivalent of this
func (m Mutator) interesting_byte(ts * TestCase){
  if len(ts.input) == 0 {
    return
  }
  interesting := []int8{-127, -1, 0, 1, 127, '{', '}', ',', '<', '>'}
  val := interesting[m.rng.Intn(len(interesting))]
  pos := m.rng.Intn(len(ts.input))
  msg := fmt.Sprintf("Mutator performed 'interesting_byte' inserting int8 %d (%c) on byte %d\n", val, byte(val), pos);
  ts.changes = append(ts.changes, msg)
  ts.input[pos] = byte(val)
}


func (m Mutator)mutate(ts * TestCase){
  nMutations := m.rng.Intn(8);
  nMutations = 3;
  for i:=0; i < nMutations; i++{
    selection := m.rng.Intn(5);
    // TODO work out configurables, they might be needed here
    switch selection {
    case  0:
      m.flip_bits(ts)
    case  1:
      m.flip_bytes(ts)
    case  2:
      m.delete_slice(ts)
    case  3:
      m.duplicate_slice(ts)
    case  4:
      m.interesting_byte(ts)
    default:
      fmt.Printf("[WARN] mutator borked")
      //dunno
    }
  }
  m.out_chan <- *ts
}
