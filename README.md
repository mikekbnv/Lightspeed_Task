# Unique IPv4 Address Counter

## Overview

A  Go program designed to count unique IPv4 addresses in large files (100GB+). This tool optimizes memory usage and processing time by leveraging binary conversion, merge sort, and priority queue techniques.

## Key Features

- **Massive File Processing**: Handles files of 100GB+ size using chunk-based processing
- **Compression**: Compresses IPv4 addresses from `string` to `uint32` for space efficiency
- **Configuration**: Customizable chunk size via command-line flags
- **Sorting**: Efficiently merges chunks using a priority queue

## Quick Start

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd unique-ip-counter
   ```

2. Build the executable:
   ```bash
   go build -o unique-ip-counter main.go
   ```

### Usage

#### Command-line Flags

| Flag              | Description                     | Type   | Default |
|:------------------|:--------------------------------|:------:|:-------:|
| `-f, -file`      | Path to input file (REQUIRED)   | string |    -    |
| `-c, -chunk`     | Temporary file chunk size in MB |  int   |  1024   |
| `-h, -help`      | Display usage information       |   -    |    -    |

#### Example Commands

```bash
# Basic usage
./unique-ip-counter -f /path/to/large-ip-file.txt

# Custom chunk size
./unique-ip-counter -f /path/to/large-ip-file.txt -c 512
```

## Algorithm Deep Dive

### Core Processing Steps

1. **Compression**
   - Transform IPv4 addresses from `string` to `uint32`
   - Reduces storage from 16 bytes (string) to 4 bytes (uint32)

2. **Chunk Processing**
   - Splits large input file into manageable chunks
   - Sorts and writes chunks to temporary binary files

3. **Merge Strategy**
   - Uses priority queue for efficient chunk merging
   - Produces a final sorted binary file

4. **Unique Counting**
   - Sequentially reads sorted binary file
   - Compares adjacent addresses to count unique entries

## Performance Metrics

#### Space Efficiency
- **String Format**: `0.0.0.0\n` = 8 bytes; `255.255.255.255\n` = 16 bytes
- **Integer Format**: `uint32` = 4 bytes
- **Compression Ratio**: Up to 4x reduction

#### Memory Profile
- **Estimated Memory Usage**: 3-4x chunk size
- Accounts for buffering and sorting

### Computational Complexity

| Operation | Time Complexity | Space Complexity |
|-----------|-----------------|------------------|
| Sorting | O(n log n) | O(m) |
| Unique Count | O(n) | O(b) |
| Overall | O(n + n log n) | O(4 × m + b) |
\**n - file size; m - chunk size; b - buffer size;*
## Results for the provided file

| Metric | Value |
|--------|-------|
| **Input File Size** | 106 GB |
| **Compressed Size** | 29.8 GB |
| **Total Processed IPs** | 8 000 000 000 |
| **Unique IPs** | 1 000 000 000 |
| **IP Duplication Rate** | 1:8 |
| **Processing Time (106 GB)** | ~48 mins |

---

**Note**:  
These results are based on a regular system (16GB RAM) with a default chunk size of **1024 MB** for the program. Actual processing time and resource usage may vary depending on the system/program specifications.

---

## Future Improvements

- [ ] Add multithreading support for read/write/sort
- [ ] Develop configuration file support
