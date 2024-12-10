# Unique IPv4 Address Counter

## Overview

A  Go program designed to efficiently count unique IPv4 addresses in extremely large files (100GB+). This tool optimizes memory usage and processing time by leveraging binary conversion, merge sort, and priority queue techniques.

## Key Features

- **Massive File Processing**: Handles files of 100GB+ size using chunk-based processing
- **Memory Optimization**: Converts IPv4 addresses to binary (`uint32`) for space efficiency
- **Validation**: Skips invalid or empty lines
- **Configuration**: Customizable chunk size via command-line flags
- **Sorting**: Efficiently merges chunks using a priority queue
- **Garbage Collection Support**: Optimized memory management

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
| `-f, --file`      | Path to input file (REQUIRED)   | string |    -    |
| `-c, --chunk`     | Temporary file chunk size in MB |  int   |  1024   |
| `-h, --help`      | Display usage information       |   -    |    -    |

#### Example Commands

```bash
# Basic usage
./unique-ip-counter -f /path/to/large-ip-file.txt

# Custom chunk size
./unique-ip-counter -f /path/to/large-ip-file.txt -c 512
```

## Algorithm Deep Dive

### Core Processing Steps

1. **Binary Conversion**
   - Transforms IPv4 addresses from text to `uint32`
   - Reduces storage from 16 bytes (text) to 4 bytes (binary)

2. **Chunk Processing**
   - Splits large input file into manageable chunks
   - Sorts and writes chunks to temporary binary files

3. **Merge Strategy**
   - Uses priority queue for efficient chunk merging
   - Produces a globally sorted binary file

4. **Unique Counting**
   - Sequentially reads sorted binary file
   - Compares adjacent addresses to count unique entries

## Performance Metrics

#### Space Efficiency
- **Text Format**: 0.0.0.0\n = 8 bytes
- **Binary Format**: `uint32` = 4 bytes
- **Compression Ratio**: Up to 4x reduction

#### Memory Profile
- **Estimated Memory Usage**: 3-4x chunk size
- Accounts for buffering, sorting, and garbage collection

### Computational Complexity

| Operation | Time Complexity | Space Complexity |
|-----------|-----------------|------------------|
| Sorting | O(n log n) | O(chunk size) |
| Unique Count | O(n) | O(buffer size) |
| Overall | O(n + n log n) | O(4 × chunk size + buffer size) |

## Validation Criteria

- IPv4 address length between 7-15 bytes
- Covers range from 0.0.0.0 to 255.255.255.255

## Results

### Sample Large File Processing

| Metric | Value |
|--------|-------|
| **Input File Size** | 106 GB |
| **Compressed Size** | 29.8 GB |
| **Total Processed IPs** | 8 billion |
| **Unique IPs** | 1 billion |
| **IP Duplication Rate** | 8x |
| **Processing Time (106 GB)** | 40-45 minutes |
| **Processing Time (10 GB)** | 2-3 minutes |

## 🛠 Future Improvements

- [ ] Add multithreading support for read/write
- [ ] Develop configuration file support
