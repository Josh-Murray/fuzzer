package main

import (
	syscall "golang.org/x/sys/unix"
)

/*
 * This file contains struct declarations and helper functions
 * for the feedback/coverage functionality used by the harness.
 */

/*
 * Basic unit of code coverage which forms an execution trace.
 * Generated for every syscall trap.
 * rax: syscall number
 */
type regSet struct {
	rax uint64
}

/*
 * Trace of a single program execution.
 * trace: list of regSet structs generated through a program run.
 */
type execTrace struct {
	trace []regSet
}

/*
 * Grabs registers of interest from a register set.
 * Returns a newly created regSet struct.
 */
func getInterestingRegs(regs *syscall.PtraceRegs) regSet {
	r := regSet{rax: regs.Orig_rax}
	return r
}

/*
 * Compares two regSet structs.
 * Returns whether they contain the same elements.
 */
func sameRegs(r1, r2 regSet) bool {
	return r1.rax == r2.rax
}

/*
 * Compares two execTrace structs.
 * Returns whether they contain the same elements.
 */
func sameTrace(t1, t2 execTrace) bool {
	// Fastpath: length mismatch.
	if len(t1.trace) != len(t2.trace) {
		return false
	}

	for i := 0; i < len(t1.trace); i++ {
		if sameRegs(t1.trace[i], t2.trace[i]) == false {
			return false
		}

	}
	return true
}

/*
 * Checks whether a given execTrace is unique in a slice of execTraces.
 * Returns true if it is unique.
 */
func isUniqueTrace(curExecTrace execTrace, listExecTrace []execTrace) bool {
	for _, t := range listExecTrace {
		if sameTrace(curExecTrace, t) {
			return false
		}
	}
	return true
}
