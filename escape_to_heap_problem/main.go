package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"syscall"
	"unsafe"
)

// lockedBuffer represents a memory-locked buffer with minimal heap escape
type lockedBuffer struct {
	addr   unsafe.Pointer
	size   int
	locked bool
}

// newLockedBuffer allocates and locks memory with careful attention to escape analysis
func newLockedBuffer(size int) *lockedBuffer {
	// Use mmap for page-aligned memory
	data, err := syscall.Mmap(-1, 0, size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANON|syscall.MAP_PRIVATE)
	if err != nil {
		return nil
	}

	buf := &lockedBuffer{
		addr: unsafe.Pointer(&data[0]),
		size: size,
	}

	// Convert to slice for mlock without letting it escape
	slice := *(*[]byte)(unsafe.Pointer(&struct {
		addr unsafe.Pointer
		len  int
		cap  int
	}{buf.addr, size, size}))

	if err := syscall.Mlock(slice); err != nil {
		syscall.Munmap(data)
		return nil
	}

	buf.locked = true
	return buf
}

// writeRandom writes random data without letting the buffer escape
func (b *lockedBuffer) writeRandom() error {
	// Create a slice that doesn't escape
	slice := *(*[]byte)(unsafe.Pointer(&struct {
		addr unsafe.Pointer
		len  int
		cap  int
	}{b.addr, b.size, b.size}))
	
	_, err := rand.Read(slice)
	return err
}

// copyTo copies data to another buffer without escapes
func (b *lockedBuffer) copyTo(dest *lockedBuffer) {
	if b.size != dest.size {
		return
	}
	
	srcSlice := *(*[]byte)(unsafe.Pointer(&struct {
		addr unsafe.Pointer
		len  int
		cap  int
	}{b.addr, b.size, b.size}))
	
	destSlice := *(*[]byte)(unsafe.Pointer(&struct {
		addr unsafe.Pointer
		len  int
		cap  int
	}{dest.addr, dest.size, dest.size}))
	
	copy(destSlice, srcSlice)
}

// display shows the content without letting the data escape the stack
func (b *lockedBuffer) display(label string) {
	// Use a non-escaping slice conversion
	slice := *(*[]byte)(unsafe.Pointer(&struct {
		addr unsafe.Pointer
		len  int
		cap  int
	}{b.addr, b.size, b.size}))
	
	fmt.Printf("%s - Addr: %p | Data: ", label, b.addr)
	for i := 0; i < b.size; i++ {
		fmt.Printf("%02x ", slice[i])
	}
	fmt.Println()
}

// clear zeros the memory
func (b *lockedBuffer) clear() {
	slice := *(*[]byte)(unsafe.Pointer(&struct {
		addr unsafe.Pointer
		len  int
		cap  int
	}{b.addr, b.size, b.size}))
	
	for i := range slice {
		slice[i] = 0
	}
}

// close releases the memory
func (b *lockedBuffer) close() {
	if b.locked {
		slice := *(*[]byte)(unsafe.Pointer(&struct {
			addr unsafe.Pointer
			len  int
			cap  int
		}{b.addr, b.size, b.size}))
		syscall.Munlock(slice)
	}
	b.clear()
	
	// Convert back to proper slice for munmap
	data := *(*[]byte)(unsafe.Pointer(&struct {
		addr unsafe.Pointer
		len  int
		cap  int
	}{b.addr, b.size, b.size}))
	syscall.Munmap(data)
}

// noEscape is a compiler hint to prevent escape
//go:nosplit
//go:nocheckptr
func noEscape(p unsafe.Pointer) unsafe.Pointer {
	return p
}

func main() {
	fmt.Println("=== Secure Memory with Escape Analysis Considerations ===")
	
	const size = 16
	
	// Allocate buffers
	buf1 := newLockedBuffer(size)
	if buf1 == nil {
		log.Fatal("Failed to allocate buffer 1")
	}
	defer buf1.close()
	
	buf2 := newLockedBuffer(size)
	if buf2 == nil {
		log.Fatal("Failed to allocate buffer 2")
	}
	defer buf2.close()
	
	fmt.Println("\n1. Initial random data in Buffer 1:")
	if err := buf1.writeRandom(); err != nil {
		log.Fatal(err)
	}
	buf1.display("Buffer 1")
	buf2.display("Buffer 2")
	
	fmt.Println("\n2. After copying to Buffer 2:")
	buf1.copyTo(buf2)
	buf1.display("Buffer 1")
	buf2.display("Buffer 2")
	
	fmt.Println("\n3. After random fill both:")
	buf1.writeRandom()
	buf2.writeRandom()
	buf1.display("Buffer 1")
	buf2.display("Buffer 2")
	
	fmt.Println("\n=== Complete ===")
}

/*
=== Secure Memory with Escape Analysis Considerations ===

1. Initial random data in Buffer 1:
Buffer 1 - Addr: 0x7b853fb69000 | Data: d9 c3 56 64 c1 29 b9 5e 5b ac 00 a2 45 73 82 9f 
Buffer 2 - Addr: 0x7b853fb68000 | Data: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 

2. After copying to Buffer 2:
Buffer 1 - Addr: 0x7b853fb69000 | Data: d9 c3 56 64 c1 29 b9 5e 5b ac 00 a2 45 73 82 9f 
Buffer 2 - Addr: 0x7b853fb68000 | Data: d9 c3 56 64 c1 29 b9 5e 5b ac 00 a2 45 73 82 9f 

3. After random fill both:
Buffer 1 - Addr: 0x7b853fb69000 | Data: 16 fd 23 52 ad 82 70 0b 9d f2 07 e0 53 23 93 99 
Buffer 2 - Addr: 0x7b853fb68000 | Data: c2 c0 ce f3 09 2a ef f1 05 f7 84 d8 54 72 3d 20 

=== Complete ===
*/