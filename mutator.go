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
  rng *rand.Rand

}
// TODO work out configurables, they might be needed here
func createMutator() Mutator {
  r := rand.New(rand.NewSource(time.Now().Unix()))
  return Mutator{rng: r}
}

func (m Mutator)flip_bits(ts * TestCase){
  // flip bit in 1% of bytes, could change to $config% bytes or randRange bytes
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
    // TODO log
    // may want to make this ascii
    ts.input[index] ^= 1 << offset;
  }
}

func (m Mutator)flip_bytes(ts * TestCase){
  // flip 1% of bytes
  size := float64(len(ts.input)) * 0.1
  nbytes := int(size)
  if nbytes == 0 {
    nbytes = 1
  }
  for i:=0; i < nbytes; i++ {
    index := m.rng.Intn(len(ts.input))

    msg := fmt.Sprintf("Mutator performed 'flip_bytes' on byte %d\n", index)
    ts.changes = append(ts.changes, msg)
    // TODO log
    //may want to make this ascii
    ts.input[index] ^= 255;
  }
}

// need to define an input struct that will get passed in here
func (m Mutator)mutate(ts * TestCase){
  nMutations := m.rng.Intn(10);
  // TODO make copy of input
  for i:=0; i < nMutations; i++{
    selection := m.rng.Intn(2);
    // TODO work out configurables, they might be needed here
    switch selection {
    case  0:
      //TODO pass in input
      m.flip_bits(ts)
    case  1:
      m.flip_bytes(ts)
    case  2:
      fmt.Println("[Mutator] Chose 2")
    case  3:
      fmt.Println("[Mutator] Chose 3")
    case  4:
      fmt.Println("[Mutator] Chose 4")
    case  5:
      fmt.Println("[Mutator] Chose 5")
    case  6:
      fmt.Println("[Mutator] Chose 6")
    case  7:
      fmt.Println("[Mutator] Chose 7")
    default:
      fmt.Printf("[WARN] mutator borked")
      //dunno
    }
  }
}
