package vvmap

import (
	"fmt"
	"strings"
	"testing"
)

var alwaysLeftResolver DeterministicConflictResolver = func(_ string, left, _ Record) Record { return left }

func TestSetIncrementsVersion(t *testing.T) {
	vm := New("test", alwaysLeftResolver)
	oldVersion := vm.Version()[vm.ID()]
	vm.Set("someKey", "someValue")
	got := vm.Version()[vm.ID()]
	want := oldVersion + 1
	if got != want {
		t.Fatalf("got: %v want: %v", got, want)
	}
}

func TestGetReturnsSetValue(t *testing.T) {
	vm := New("test", alwaysLeftResolver)
	key := "someKey"
	want := "someValue"
	vm.Set(key, "someValue")
	got := vm.Get(key)
	if got != want {
		t.Fatalf("got: %v want: %v", got, want)
	}
}

func TestMergeUsesResolverFunc(t *testing.T) {
	resolverInvoked := false
	testResolver := func(key string, left, right Record) Record {
		resolverInvoked = true
		return left
	}

	alice := New("alice", testResolver)
	bob := New("bob", testResolver)

	alice.Set("lunch", "turkey")
	bob.Set("lunch", "ham")

	alice.Merge(bob.Delta(alice.Version()))
	if resolverInvoked != true {
		t.Fatalf("conflict resolver not invoked")
	}
}

func Example() {
	aliceID := ID("alice")
	bobID := ID("bob")
	timID := ID("tim")

	lexicographicResolver := func(key string, left, right Record) Record {
		leftVal := left.Value.(string)
		rightVal := right.Value.(string)
		if strings.Compare(leftVal, rightVal) > 0 { // choose left if lexicographically greater
			return left
		}
		return right
	}

	alice := New(aliceID, lexicographicResolver)
	bob := New(bobID, lexicographicResolver)
	tim := New(timID, lexicographicResolver)

	// concurrently update everyone -- causes a conflict, should all resolve to "turkey" since
	// lexicographically greatest
	alice.Set("lunch", "turkey")
	bob.Set("lunch", "ham")
	tim.Set("lunch", "chicken")

	// sync alice
	alice.Merge(bob.Delta(alice.Version()))
	alice.Merge(tim.Delta(alice.Version()))

	// sync bob
	bob.Merge(alice.Delta(bob.Version())) // alice is most up-to-date so no need to sync with Tim

	// sync tim
	tim.Merge(alice.Delta(tim.Version()))

	fmt.Println("alice:", alice.Get("lunch"))
	fmt.Println("bob:", bob.Get("lunch"))
	fmt.Println("tim:", tim.Get("lunch"))

	// Output: alice: turkey
	// bob: turkey
	// tim: turkey
}
