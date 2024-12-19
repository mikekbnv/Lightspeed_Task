package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/bits"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	POW2_27       = 134217728       // 2^27
	BUFFER_SIZE   = 4 * 1024 * 1024 // 4MB
	BYTES_OVERLAP = 64              // 64 bytes overlap between threads
)

var ips = make([]uint32, POW2_27) // 2^27 * uint32 = 512MB

type Config struct {
	filePath   string // Path to the input file
	numThreads int    // Number of threads
}

// Command line interface for the program
// It takes the input file path(Required) and the num of threads(Optional) as input
func cli() Config {
	help := flag.Bool("h", false, "Display usage information")
	helpLong := flag.Bool("help", false, "Display usage information")
	numThreads := flag.Int("t", runtime.NumCPU(), "Set number of threads (Default: Number of CPU logical cores)")
	numThreadsLong := flag.Int("threads", runtime.NumCPU(), "Set number of threads (Default: Number of CPU logical cores)")
	filePath := flag.String("f", "", "Input file path (mandatory)")
	filePathLong := flag.String("file", "", "Input file path (mandatory)")

	flag.Parse()

	if *help || *helpLong {
		fmt.Println("Usage: program -f <file-path> [-r <ram-in-gb>]")
		fmt.Println("\nFlags:")
		fmt.Println("  -h, -help          Display usage information")
		fmt.Println("  -t, -threads       Set number of threads (Default: Number of CPU logical cores)")
		fmt.Println("  -f, -file          Path to the input file (mandatory)")
		os.Exit(0)
	}

	finalFilePath := *filePath
	if finalFilePath == "" {
		finalFilePath = *filePathLong
	}
	if finalFilePath == "" {
		fmt.Println("Error: -f or -file flag is required")
		os.Exit(1)
	}

	finalNumThreads := *numThreads
	if finalNumThreads == 0 {
		finalNumThreads = *numThreadsLong
	}

	if finalNumThreads < 1 {
		fmt.Println("Error: Thread number must be greater than 0")
		os.Exit(1)
	}

	return Config{
		filePath:   finalFilePath,
		numThreads: finalNumThreads,
	}
}

// Function which calculates the array index and bit index for the given IP address
// Sets the bit in the specified index to 1 using atomic operation to prevent race conditions
func writeIpToUint32Arr(arr []uint32, ip uint32) {
	arrIdx := ip >> 5
	bitIdx := ip & 31

	ipBit := uint32(1 << bitIdx)

	atomic.OrUint32(&arr[arrIdx], ipBit)
}

// Function which calculates the number of unique IP addresses in the given array
// It uses the bits.OnesCount32 function to count the number of set bits in each uint32 element
// bits.OnesCount32 faster than the loop implementation because it uses the POPCNT instruction
func calculateUniqueIpsUint32(arr []uint32) uint32 {
	var count uint32 = 0
	for _, b := range arr {
		count += uint32(bits.OnesCount32(b))
	}
	return count
}

// Function which read the specific part/size of the file and extract the IP addresses
// Converts byte line to uint32 IP address and writes it to the array using writeIpToUint32Arr function
func fileRead(name string, offset int64, bytesPerThread int, errCh chan<- error) {
	file, err := os.Open(name)

	if err != nil {
		errCh <- err
		return
	}
	defer file.Close()

	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		errCh <- err
		return
	}

	scanner := bufio.NewScanner(bufio.NewReaderSize(file, BUFFER_SIZE))
	scanner.Buffer(make([]byte, BUFFER_SIZE), BUFFER_SIZE)

	if offset != 0 {
		// one extra scan to skip partial read from offset
		scanner.Scan()
	}

	readBytes := 0
	for scanner.Scan() && readBytes < bytesPerThread {

		bytesLine := scanner.Bytes()
		readBytes += len(bytesLine) + 1
		lineLength := len(bytesLine)

		if lineLength < 7 || lineLength > 16 {
			continue
		}

		ipUint32 := bytesLineToUint32(bytesLine)
		writeIpToUint32Arr(ips, ipUint32)
	}

	if err := scanner.Err(); err != nil {
		errCh <- err
	}

	errCh <- nil
}

// FUnction which provide the file size in bytes
// Uses for the calculation of the bytes per thread
func getFileSize(name string) (int64, error) {
	file, err := os.Stat(name)
	if err != nil {
		return -1, err
	}
	return file.Size(), nil
}

// Function which converts the byte line to uint32 IP address
func bytesLineToUint32(bytes []byte) uint32 {
	segments := [4]byte{}
	idx := 0
	for _, b := range bytes {
		if b != '.' {
			segments[idx] = segments[idx]*10 + byte(b-'0')
		} else {
			idx++
		}
	}
	return uint32(segments[0])<<24 | uint32(segments[1])<<16 | uint32(segments[2])<<8 | uint32(segments[3])
}

// Worker which servres for the reading specific part of the file
// It reads from the offset to the offset+bytesPerThread+BYTES_OVERLAP bytes
// Using Overlap to prevent the loss of the IP addresses which are in the middle of the threads
func readWorker(id int, wg *sync.WaitGroup, name string, bytesPerThread int, errCh chan<- error) {
	defer wg.Done()
	fileRead(name, int64(max(0, id*bytesPerThread-BYTES_OVERLAP)), bytesPerThread+BYTES_OVERLAP, errCh)
}

// Function which start the reading threads
// It divides the file into the number of threads and starts the reading threads
// Number of threads is equal to the number of CPU cores or the number of threads provided by the user
func processIPFile(config Config) (uint32, []error) {
	threadCount := config.numThreads
	fileSize, err := getFileSize(config.filePath)
	if err != nil {
		return 1, []error{err}
	}

	errCh := make(chan error)
	errDone := make(chan struct{})
	errs := []error{}
	wg := sync.WaitGroup{}

	bytesPerThread := int(fileSize / int64(threadCount))

	go func() {
		for err := range errCh {
			if err != nil {
				errs = append(errs, err)
			}
		}
		errDone <- struct{}{}
	}()

	for i := 0; i < threadCount; i++ {
		wg.Add(1)
		go readWorker(i, &wg, config.filePath, bytesPerThread, errCh)
	}

	wg.Wait()
	close(errCh)
	<-errDone

	return calculateUniqueIpsUint32(ips), errs
}

func main() {
	config := cli()

	start := time.Now()

	unique, errs := processIPFile(config)

	for _, err := range errs {
		if err != nil {
			fmt.Println("Error:", err)
		}
	}

	fmt.Println("Unique ip count =", unique)
	fmt.Println("Elapsed =", time.Since(start))
}
