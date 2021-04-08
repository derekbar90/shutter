package main

// Analyze disk usage of state.gob file

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/brainbot-com/shutter/shuttermint/keyper"
	"github.com/brainbot-com/shutter/shuttermint/keyper/observe"
)

type storedState struct {
	State     *keyper.State
	Shutter   *observe.Shutter
	MainChain *observe.MainChain
}

type DummyWriter struct {
	Size int
}

func (w *DummyWriter) Write(p []byte) (n int, err error) {
	w.Size += len(p)
	return len(p), nil
}

func gobsize(st storedState) int {
	dw := DummyWriter{}
	enc := gob.NewEncoder(&dw)
	err := enc.Encode(&st)
	if err != nil {
		panic(err)
	}
	return dw.Size
}

func report(id string, full int, st storedState) {
	size := gobsize(st)
	percent := 100.0 * float64(size) / float64(full)
	fmt.Printf("%16s: %10d   %5.1f\n", id, size, percent)
}

func main() {
	gobpath := "state.gob"
	gobfile, err := os.Open(gobpath)
	if err != nil {
		panic(err)
	}
	dec := gob.NewDecoder(gobfile)
	st := storedState{}
	err = dec.Decode(&st)
	if err != nil {
		panic(err)
	}

	full := gobsize(st)
	report("full", full, st)
	report("state", full, storedState{State: st.State})

	report("main", full, storedState{MainChain: st.MainChain})
	report("shutter full", full, storedState{Shutter: st.Shutter})
	report("shutter batches", full, storedState{Shutter: &observe.Shutter{Batches: st.Shutter.Batches}})
	report("shutter eons", full, storedState{Shutter: &observe.Shutter{Eons: st.Shutter.Eons}})
}
