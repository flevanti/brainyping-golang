package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

//
// THIS IS USED TO UNDERSTAND IF IT IS SAFE TO UPDATE SUB ELEMENTS OF AN ARRAY/SLICE/MAP CONCURRENTLY
//
// LOOKS LIKE EACH INVIVIDUAL VALUE HAS ITS OWN POINTER THAT IS HANDLED INDIVIDUALLY SO AS LONG AS THERE'S ONLY ONE
// GOROUTINE UPDATING EACH SUB ELEMENT IT IS SAFE
//

//OUTPUT OF THE SCRIPT:
//[1000000 1000000 1000000 1000000 1000000 1000000 1000000 1000000 1000000 1000000]
//10000000
//3861581

var metadata []int
var wg sync.WaitGroup
var wgStartGoRoutines sync.WaitGroup //this to make sure that all goroutines work as much as possible in the same moment
var counterWithAtomic int32          //shouldn't suffer race condition
var counterNotAtomic int32           //should suffer race condition a lot....

func main() {
	wgStartGoRoutines.Add(1) // raise the flag 🏁
	for i := 0; i < 10; i++ {
		wg.Add(1)
		metadata = append(metadata, 0)
		go goroutine(i)
	}

	//ready..set..Go!
	wgStartGoRoutines.Done() //lower the flag 🏁

	//wait for go routines to finish...
	wg.Wait()

	//show some info
	fmt.Printf("%v\n", metadata)
	fmt.Println(counterWithAtomic)
	fmt.Println(counterNotAtomic)

}

func goroutine(i int) {
	defer wg.Done()
	wgStartGoRoutines.Wait()
	for ii := 0; ii < 1000000; ii++ {
		atomic.AddInt32(&counterWithAtomic, 1)
		counterNotAtomic++
		metadata[i]++
	}
}
