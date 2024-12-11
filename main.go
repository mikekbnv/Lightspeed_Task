package main

import (
	"bufio"
	"container/heap"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"time"
)

const (
	OUTPUT_DIR  = "Temporary"     // Directory to store temporary files for required calculations
	BUFFER_SIZE = 1 * 1024 * 1024 // 1 MB
)

type Config struct {
	filePath    string // Path to the input file
	chunkSizeMb uint64 // Temporary files size in MB
}

type PqItem struct {
	value      uint32        // The value of the item; arbitrary.
	fileReader *bufio.Reader // The poiter to the file reader of specific file.
	index      int           // The index of the item in the heap.
}

// Priority queue implementation for file merging
// Including the methods for heap.Interface
// Len, Less, Swap, Push, Pop
type PriorityQueue []*PqItem

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].value < pq[j].value
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*PqItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// Command line interface for the program
// It takes the input file path(Required) and the chunk size(Optional) as input
func cli() Config {
	help := flag.Bool("h", false, "Display usage information")
	helpLong := flag.Bool("help", false, "Display usage information")
	chunkSizeMb := flag.Uint64("c", 1024, "Set the temporary file chunk size in MB (Default: 1024MB)")
	chunkSizeMbLong := flag.Uint64("chunk", 1024, "Set the temporary file chunk size in MB (Default: 1024MB)")
	filePath := flag.String("f", "", "Input file path (mandatory)")
	filePathLong := flag.String("file", "", "Input file path (mandatory)")

	flag.Parse()

	if *help || *helpLong {
		fmt.Println("Usage: program -f <file-path> [-r <ram-in-gb>]")
		fmt.Println("\nFlags:")
		fmt.Println("  -h, -help          Display usage information")
		fmt.Println("  -c, -chunk         Set the temporary file chunk size in MB (Default: 1024MB)")
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

	if *chunkSizeMb <= 0 || *chunkSizeMbLong <= 0 {
		fmt.Println("Error: -c or -chunk flag must be greater than 0")
		os.Exit(1)
	}
	finalChunkSizeMb := *chunkSizeMb
	if *chunkSizeMbLong != 1024 {
		finalChunkSizeMb = *chunkSizeMbLong
	}

	return Config{
		filePath:    finalFilePath,
		chunkSizeMb: finalChunkSizeMb,
	}
}

// PrintMemoryUsage prints the memory usage of the program
// It prints the memory allocated, total memory allocated, system memory and number of garbage collections
// Can be used to check the memory usage at different stages of the program
func MemoryUsage(stage string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("   Memory Usage [%s]: Alloc = %v MiB, TotalAlloc = %v MiB, Sys = %v MiB, NumGC = %v\n",
		stage, bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC)
}

// bToMb converts bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// Custom function to convert IP address to integer
func IpToInt(ip []byte) uint32 {
	segments := [4]byte{}

	idx := 0
	for i := range ip {
		if ip[i] != '.' {
			segments[idx] = segments[idx]*10 + byte(ip[i]-'0')
		} else {
			idx++
		}
	}

	return uint32(segments[0])<<24 | uint32(segments[1])<<16 | uint32(segments[2])<<8 | uint32(segments[3])
}

// WriteToFile writes the data to the file
// It writes the data in little endian format
// Writes the data to the file as one chunk
// The data is written to the file in binary format
func WriteToFile(fileName string, data []uint32) error {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return nil
	}

	binary.Write(file, binary.LittleEndian, data)
	file.Close()
	return nil
}

// Sorts the provided data, writes it to the file and appends the file name to the fileNames slice for the futher use
func SortAndWriteFileChunk(fileCount int, fileNames *[]string, data []uint32) error {

	sort.Slice(data, func(i, j int) bool { return data[i] < data[j] })

	fileName := fmt.Sprintf(OUTPUT_DIR+"/file-%d.bin", fileCount)
	*fileNames = append(*fileNames, fileName)

	err := WriteToFile(fileName, data)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return nil
	}

	return nil
}

// SplitFileIntoChunks splits the input file into default chunks or proded by user in the command line
// It reads the file line by line and converts the IP address to integer
// It writes the data to the temporary files in binary format
// The temporary files are stored in the OUTPUT_DIR directory
// The file names are stored in the fileNames slice for the further use
func SplitFileIntoChunks(config Config) []string {
	file, err := os.Open(config.filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return nil
	}
	defer file.Close()

	err = os.Mkdir(OUTPUT_DIR, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return nil
	}

	// Info about the file for the progress bar
	// It calculates the file size and the processed bytes to show the progress
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		return nil
	}
	fileSize := fileInfo.Size()

	scanner := bufio.NewScanner(bufio.NewReaderSize(file, BUFFER_SIZE))
	scanner.Buffer(make([]byte, BUFFER_SIZE), BUFFER_SIZE)

	var processedBytes int64 = 0
	var lineCount uint64 = 0
	var fileCount int = 0
	var linesForChunk uint64 = config.chunkSizeMb * 1024 * 1024 / 4
	fileNames := make([]string, 0)
	ips := make([]uint32, linesForChunk)
	var i uint64 = 0

	for scanner.Scan() {
		line := scanner.Bytes()

		if len(line) < 7 || len(line) > 16 {
			continue
		}

		ips[i] = IpToInt(line)

		lineCount++
		i++
		processedBytes += int64(len(line)) + 1

		if lineCount%100000 == 0 {
			progress := float64(processedBytes) * 100 / float64(fileSize)
			fmt.Printf("\rProgress for the file read, sort and write to new files: %.2f%%", progress)
		}

		if i == linesForChunk {
			i = 0

			err = SortAndWriteFileChunk(fileCount, &fileNames, ips)
			if err != nil {
				fmt.Printf("Error writing to file: %v\n", err)
				return nil
			}
			fileCount++
		}
	}

	if i > 0 {
		err = SortAndWriteFileChunk(fileCount, &fileNames, ips[:i])
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			return nil
		}
	}

	ips = nil

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return nil
	}

	return fileNames
}

// MergeSortedFiles merges the sorted files into one final file
// It reads the first IP address from each file and pushes it to the priority queue
// Priority queue max size is the number of files
// At a time it reads the minimum IP address from the priority queue and writes it to the final file
// It reads the next IP address from the file from which the minimum IP address was read
// It continues the process until all the files are read
func MergeSortedFiles(fileNames []string) string {
	fmt.Println("\nMerging files...")
	var ip uint32
	pq := make(PriorityQueue, 0)

	for i := 0; i < len(fileNames); i++ {
		file, err := os.Open(fileNames[i])
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return ""
		}
		defer file.Close()

		bufferReader := bufio.NewReader(file)
		err = binary.Read(bufferReader, binary.LittleEndian, &ip)

		if err == nil {
			heap.Push(&pq, &PqItem{value: ip, fileReader: bufferReader})
		}
	}

	finalFile, err := os.Create(OUTPUT_DIR + "/finalfile.bin")
	if err != nil {
		fmt.Printf("Error creating final file: %v\n", err)
		return ""
	}

	bufferWriter := bufio.NewWriter(finalFile)

	for pq.Len() > 0 {

		item := heap.Pop(&pq).(*PqItem)
		err = binary.Write(bufferWriter, binary.LittleEndian, item.value)
		if err != nil {
			fmt.Printf("Error writing to final file: %v\n", err)
		}

		err = binary.Read(item.fileReader, binary.LittleEndian, &ip)
		if err == nil {
			heap.Push(&pq, &PqItem{value: ip, fileReader: item.fileReader})
		}

	}
	bufferWriter.Flush()
	finalFile.Close()

	return OUTPUT_DIR + "/finalfile.bin"
}

// GetUniqIpAddresses reads the final file and counts the unique IP addresses
// It stores the current and prev IP addresses
// If the current IP address is not equal to the previous IP address, it increments the unique IP address count
func GetUniqIpAddresses(filename string) uint64 {
	fmt.Println("Counting unique IP addresses...")
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error creating final file: %v\n", err)
		return 0
	}
	defer file.Close()

	var uniqueIp uint64 = 0
	var prev uint32 = 0
	var current uint32 = 0
	var first bool = true
	bufferRead := bufio.NewReader(file)

	for {
		err = binary.Read(bufferRead, binary.LittleEndian, &current)
		if err != nil {
			break
		}
		if first || current != prev {
			uniqueIp++
			prev = current
			first = false
		}
	}

	return uniqueIp
}

func main() {
	start := time.Now()

	config := cli()

	fmt.Printf("File: %s\n", config.filePath)
	fmt.Printf("Chunk size: %.2d MB\n", config.chunkSizeMb)

	filenames := SplitFileIntoChunks(config)
	finalfile := MergeSortedFiles(filenames)
	uniqip := GetUniqIpAddresses(finalfile)

	duration := time.Since(start)

	err := os.RemoveAll(OUTPUT_DIR)
	if err != nil {
		fmt.Printf("Error removing directory: %v\n", err)
	}

	
	
	//MemoryUsage("After all operations")
	fmt.Printf("\nUnique IP addresses: %d\n", uniqip)
	fmt.Println("Time taken to count unique ips:", duration)
}
