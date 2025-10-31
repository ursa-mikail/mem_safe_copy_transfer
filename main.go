package main

import (
	"crypto/rand"
	"fmt"
	"log"
)

//// Method 1 - Clears TWICE
func method1() {
	fmt.Println("=== Method 1: Two Clearing Operations ===")
	
	// Step 1: Allocate space for hiddenBuffers
	const bufferSize = 1024
	hiddenBuffers := make([]byte, bufferSize)
	
	// Step 2: Fill ENTIRE buffer with random bytes
	if _, err := rand.Read(hiddenBuffers); err != nil {
		log.Fatal("Failed to generate random data:", err)
	}
	fmt.Printf("1. Entire buffer filled with random (first 16 bytes): %x\n", hiddenBuffers[:16])
	
	// Step 3: Create memory address with random offset
	memoryAddress := createMemoryAddressingWithRandomOffset(hiddenBuffers)
	fmt.Printf("2. Created memory address with offset: %p\n", memoryAddress)
	
	// Step 4: Copy data to memory address
	dataToCopy := []byte("sensitive data that needs protection")
	copy(memoryAddress[:len(dataToCopy)], dataToCopy)
	fmt.Printf("3. Copied data: %s\n", memoryAddress[:len(dataToCopy)])
	
	// Step 5: Clear memory with random filling 10x - CLEARS TWICE!
	// First clear: the original entire buffer
	clearMemorySecure(hiddenBuffers, len(hiddenBuffers))
	fmt.Printf("4. Cleared entire original buffer (first 16): %x\n", hiddenBuffers[:16])
	
	// Second clear: the working memory (even though it's part of the already-cleared buffer)
	clearMemorySecure(memoryAddress, len(dataToCopy))
	fmt.Printf("5. Cleared working memory again (first 16): %x\n", memoryAddress[:16])
}

func createMemoryAddressingWithRandomOffset(buffer []byte) []byte {
	maxOffset := len(buffer) - 100
	offset := 0
	if maxOffset > 0 {
		offsetBytes := make([]byte, 1)
		rand.Read(offsetBytes)
		offset = int(offsetBytes[0]) % maxOffset
	}
	return buffer[offset:]
}

func clearMemorySecure(mem []byte, length int) {
	for i := 0; i < 10; i++ {
		rand.Read(mem[:length])
	}
	for i := 0; i < length; i++ {
		mem[i] = 0
	}
}

//// Method 2 - Clears ONCE
type SecureMemoryManager struct {
	baseBuffer    []byte
	workingMemory []byte
	bufferSize    int
	dataLength    int
}

func NewSecureMemoryManager(bufferSize int) *SecureMemoryManager {
	// Step 1: Allocate space for hiddenBuffers
	baseBuffer := make([]byte, bufferSize)
	
	// Step 2: Create memory address with random offset FIRST
	workingMemory := createMemoryAddressingWithRandomOffsetStruct(baseBuffer)
	
	// Step 3: Fill ONLY the working memory with random bytes
	if _, err := rand.Read(workingMemory); err != nil {
		log.Fatal("Failed to generate random data:", err)
	}
	fmt.Printf("1. Only working memory filled with random (first 16 bytes): %x\n", workingMemory[:16])
	
	return &SecureMemoryManager{
		baseBuffer:    baseBuffer,
		workingMemory: workingMemory,
		bufferSize:    bufferSize,
		dataLength:    0,
	}
}

func (sm *SecureMemoryManager) CopyData(data []byte) {
	// Step 4: Copy data to working memory
	copy(sm.workingMemory[:len(data)], data)
	sm.dataLength = len(data)
	fmt.Printf("2. Copied data to secure memory: %s\n", sm.workingMemory[:len(data)])
}

func (sm *SecureMemoryManager) ClearSecure() {
	// Step 5: Clear ONLY ONCE - just the working memory that was actually used
	if sm.dataLength > 0 {
		clearMemorySecure(sm.workingMemory, sm.dataLength)
		sm.dataLength = 0
	}
	fmt.Printf("3. Cleared only working memory (first 16 bytes): %x\n", sm.workingMemory[:16])
	// Note: baseBuffer was never touched with sensitive data, so no need to clear it
}

func createMemoryAddressingWithRandomOffsetStruct(buffer []byte) []byte {
	if len(buffer) < 256 {
		return buffer
	}
	
	offsetBytes := make([]byte, 2)
	rand.Read(offsetBytes)
	
	maxOffset := len(buffer) - 128
	offset := (int(offsetBytes[0])<<8 | int(offsetBytes[1])) % maxOffset
	
	if offset < 64 {
		offset = 64
	}
	if offset > maxOffset-64 {
		offset = maxOffset - 64
	}
	
	return buffer[offset:]
}

func method2() {
	fmt.Println("\n=== Method 2: One Clearing Operation ===")
	
	// Create memory manager with allocated space
	memManager := NewSecureMemoryManager(2048)
	
	// Work with the memory
	data := []byte("highly sensitive information")
	memManager.CopyData(data)
	
	// Clear securely when done - only clears ONCE
	memManager.ClearSecure()
}

func main() {
	method1()
	method2()
	
	fmt.Println("\n=== Key Difference Summary ===")
	fmt.Println("Method 1: Clears 2x - entire buffer + working memory")
	fmt.Println("Method 2: Clears 1x  - only working memory (more efficient)")
	fmt.Println("Method 2 is more efficient because it only touches the memory that actually contains sensitive data!")
}

/*
=== Method 1: Two Clearing Operations ===
1. Entire buffer filled with random (first 16 bytes): 6dfea461f0e1f5196bfee643fa4cf253
2. Created memory address with offset: 0xc00007e08f
3. Copied data: sensitive data that needs protection
4. Cleared entire original buffer (first 16): 00000000000000000000000000000000
5. Cleared working memory again (first 16): 00000000000000000000000000000000

=== Method 2: One Clearing Operation ===
1. Only working memory filled with random (first 16 bytes): ce2c2f4d6779ac3ff6f5d9bb6bf7acf5
2. Copied data to secure memory: highly sensitive information
3. Cleared only working memory (first 16 bytes): 00000000000000000000000000000000

=== Key Difference Summary ===
Method 1: Clears 2x - entire buffer + working memory
Method 2: Clears 1x  - only working memory (more efficient)
Method 2 is more efficient because it only touches the memory that actually contains sensitive data!

*/