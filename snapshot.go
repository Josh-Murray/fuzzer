package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
)

/*
 * This file contains the struct declarations and functions
 * for the memory snapshot/restore functionality used by the harness.
 */

/*
 * Struct represting a process memory segment.
 * startAddr: Starting address of the segment.
 * size: Size of the segment.
 * data: The data in the memory segment.
 */
type MemoryRegion struct {
	startAddr uint64
	size      uint64
	data      []byte
}

/*
 * Snapshot of a process state at some point during its execution.
 * pid: Process pid
 * savedRegs: Saved register state of the process.
 * segments: List of MemoryRegion structs for all process memory regions of interest.
 */
type Snapshot struct {
	pid       int
	savedRegs syscall.PtraceRegs
	segments  []MemoryRegion
}

/*
 * Runs the target binary specified by pid until it is in a state desirable
 * for memory snapshotting. The wait status ws is updated along the way.
 * This will run the binary until the first read syscall from fd 0 (stdin)
 */
func setupSnapshotState(pid int, ws *syscall.WaitStatus, is64bit bool) error {
	var err error
	var regs syscall.PtraceRegs
	for {
		// Trace every instruction with ptrace until desired state.
		err = syscall.PtraceSyscall(pid, 0)
		if err != nil {
			return fmt.Errorf("Failed to set up snapshot state (PtraceSingleStep):"+
				"%s", err.Error())

		}

		_, err = syscall.Wait4(pid, ws, syscall.WALL, nil)
		if err != nil {
			return fmt.Errorf("Failed to set up snapshot state (Wait4): "+
				"%s\n", err.Error())

		}

		// If any sort of exit, something has gone wrong.
		if ws.Exited() == true ||
			ws.StopSignal() == syscall.SIGSEGV ||
			ws.StopSignal() == syscall.SIGABRT {
			return fmt.Errorf("Exit during snapshot set up: %v",
				ws.StopSignal())
		}

		err = syscall.PtraceGetRegs(pid, &regs)
		if err != nil {
			return fmt.Errorf("Failed to set up snapshot state (PtraceGetRegs:)"+
				"%s", err.Error())

		}

		// Check if read syscall is made from fd 0.
		// This check is architecture dependent.
		// This is the exit condition for the snapshot set up.
		if is64bit {
			if regs.Orig_rax == 0x0 && regs.Rdi == 0x0 {
				return nil
			}
		} else {
			if regs.Orig_rax == 0x3 && regs.Rbx == 0x0 {
				return nil
			}
		}

	}
}

/*
 * Saves the state of the current process specified by pid into a
 * Snapshot struct, and returns the struct.
 */
func makeSnapshot(pid int) (Snapshot, error) {
	var err error
	var procSnapshot Snapshot
	// The proc filesystem is used to make the snapshot of process
	// memory segments. See PROC(5) for more details.
	path := "/proc/" + strconv.Itoa(pid)

	// /proc/pid/maps gives the memory layout of the process.
	mapFile, err := os.Open(path + "/maps")
	if err != nil {
		log.Fatalf("makeSnapshot failed to open /proc/%d/maps: %s\n",
			pid, err.Error())
		return procSnapshot,
			fmt.Errorf("makeSnapshot failed to open /proc/%d/maps: %s",
				pid, err.Error())

	}
	defer mapFile.Close()

	// /proc/pid/mem gives access to process memory.
	memFile, err := os.Open(path + "/mem")
	if err != nil {
		return procSnapshot,
			fmt.Errorf("makeSnapshot failed to open /proc/%d/maps: %s",
				pid, err.Error())
	}
	defer memFile.Close()

	var mRegions []MemoryRegion
	var addrStart uint64
	var addrEnd uint64
	var perm string

	// Go through the map file line by line- each line is a new memory segment.
	scanner := bufio.NewScanner(mapFile)
	for scanner.Scan() {
		line := scanner.Text()

		// We are interesting in the first 3 elements of a segment entry:
		// the start and ending addresses, and the segment permissions.
		_, err = fmt.Sscanf(line, "%x-%x %5s", &addrStart, &addrEnd, &perm)
		if err != nil {
			return procSnapshot,
				fmt.Errorf("Sscanf failed in makeSnapshot: %s",
					err.Error())
		}

		// Only copy writable memory segments.
		if perm[1] == 'w' {
			mSize := addrEnd - addrStart
			mData := make([]byte, mSize)

			_, err = memFile.ReadAt(mData, int64(addrStart))
			if err != nil {
				return procSnapshot,
					fmt.Errorf("Failed to read memory segment "+
						"in makeSnapshot: %s", err.Error())
			}

			m := MemoryRegion{
				startAddr: addrStart,
				size:      mSize,
				data:      mData,
			}
			mRegions = append(mRegions, m)
		}
	}

	err = scanner.Err()
	if err != nil {
		return procSnapshot,
			fmt.Errorf("Scanner failed in makeSnapshot: %s", err.Error())
	}

	// Save the register set.
	var regs syscall.PtraceRegs
	err = syscall.PtraceGetRegs(pid, &regs)
	if err != nil {
		return procSnapshot,
			fmt.Errorf("Failed to get register set i makeSnapshot: %s",
				err.Error())
	}

	procSnapshot = Snapshot{
		pid:       pid,
		savedRegs: regs,
		segments:  mRegions,
	}
	return procSnapshot, nil
}

/*
 * Restores the state of the process identified by the given Snapshot struct.
 */
func restoreSnapshot(procSnapshot Snapshot) error {
	var err error

	// /proc/pid/mem is used to restore process memory segments.
	path := "/proc/" + strconv.Itoa(procSnapshot.pid) + "/mem"
	memFile, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("restoreSnapshot failed to open /proc/%d/mem: %s",
			procSnapshot.pid, err.Error())
	}

	// Restore saved memory segments
	for _, mRegion := range procSnapshot.segments {
		_, err = memFile.WriteAt(mRegion.data,
			int64(mRegion.startAddr))
		if err != nil {
			return fmt.Errorf("restoreSnapshot failed to restore segment: %s",
				err.Error())
		}
	}

	// Restore saved register set.
	err = syscall.PtraceSetRegs(procSnapshot.pid, &procSnapshot.savedRegs)
	if err != nil {
		return fmt.Errorf("restoreSnapshot failed to restore regs: %s",
			err.Error())
	}

	return nil
}

/*
 * Writes the provided input to the target process's stdin, specified by pid.
 * Uses the /proc filesystem to do so, writing to /proc/pid/fd/0
 */
func writeToProc(pid int, input []byte) error {
	var err error
	path := "/proc/" + strconv.Itoa(pid) + "/fd/0"

	stdinFile, err := syscall.Open(path, syscall.O_WRONLY|syscall.O_CLOEXEC|
		syscall.O_NONBLOCK|syscall.O_SYNC, 0644)
	if err != nil {
		return fmt.Errorf("writeToProc failed to open process stdin: %s",
			err.Error())
	}
	defer syscall.Close(stdinFile)

	_, err = syscall.Write(stdinFile, input)
	// Continue as normal on EAGAIN error.
	if err != nil && err != syscall.EAGAIN {
		return fmt.Errorf("writeToProc failed to write to output: %s",
			err.Error())
	}
	return nil
}
