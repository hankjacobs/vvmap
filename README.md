# vv [![GoDoc](https://godoc.org/github.com/hankjacobs/vv?status.png)](https://godoc.org/github.com/hankjacobs/vv)

vv is a [Go](http://www.golang.org) implementation of a delta-based CRDT map as written about in ["∆-CRDTs: Making δ-CRDTs Delta-Based"](http://nova-lincs.di.fct.unl.pt/system/publication_files/files/000/000/666/original/a12-van_der_linde.pdf?1483708753), ["Dotted Version Vectors: Logical Clocks for Optimistic Replication"](https://arxiv.org/pdf/1011.5808.pdf), and Tyler McMullen's excellent talk ["The Anatomy of a Distributed System"](https://www.infoq.com/presentations/health-distributed-system).  

Usage:

```go
package main

import (
	"fmt"
	"strings"
	"testing"
 	"github.com/hankjacobs/vv"
)

func main() {
	lexicographicResolver := func(key string, left, right vv.Record) bool {
		leftVal := left.Value.(string)
		rightVal := right.Value.(string)
		return strings.Compare(leftVal, rightVal) > 0 // choose left if lexicographically greater
	}

	alice := vv.New("alice", lexicographicResolver)
	bob := vv.New("bob", lexicographicResolver)
	tim := vv.New("tim", lexicographicResolver)

	// concurrently update everyone -- causes a conflict, should all resolve to "turkey" since
	// lexicographically greatest
	alice.Set("lunch", "turkey")
	bob.Set("lunch", "ham")
	tim.Set("lunch", "chicken")

	// get records that Bob has but Alice doesn't
	delta := bob.Delta(alice.Version())
	alice.Merge(delta)

	// get records that Tim has but Alice doesn't
	delta = bob.Delta(alice.Version())
	alice.Merge(delta)

	// sync bob
	bob.Merge(alice.Delta(bob.Version())) // alice is most up-to-date so no need to sync with Tim

	// sync tim
	tim.Merge(alice.Delta(tim.Version()))

	fmt.Println("alice:", alice.Get("lunch"))
	fmt.Println("bob:", bob.Get("lunch"))
	fmt.Println("tim:", tim.Get("lunch"))
}
```
