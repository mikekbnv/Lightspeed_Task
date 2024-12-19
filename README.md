# Unique IPv4 Address Counter

## Overview

A high-performance Go program designed to count unique IPv4 addresses in large files. This program optimizes memory usage and processing speed through bit manipulation and concurrent processing.

## Key Features

- **Concurrent Processing:** Utilizes multiple threads for parallel file reading and data processing 
- **Memory Efficient:** Uses bit array representation for storing IP addresses
- **Configurable**: Adjustable thread count via command-line flag
- **Uses POPCNT** instruction for bit counting
- **Employs atomic operations** for thread safety
- Implements **buffered** file reading

### Quick Start

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
| `-f, -file`       | Path to input file (REQUIRED)   | string |    -    |
| `-t, -threads`    | Set number of threads           |  int   |  NumCPU |
| `-h, -help`       | Display usage information       |   -    |    -    |

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

2. **Store Strategy**
   - Uses a ``uint32`` array of size 2^27 (512MB) to store IP addresses
   - Each IP address requires only 1 bit for storage
   - Maps each IP address to a bit:
        - The first 27 bits determine the array index.
        - The last 5 bits determine the bit index within the ``uint32``.

3. **Concurrent Processing**
    - Divides file reading among multiple threads
    - Uses atomic operations for thread-safe bit array updates
     
4. **Unique Counting**
    - Efficient bit counting using hardware instructions
    - Single pass counting after all IPs are processed
   

## Performance Metrics

#### Computational Complexity

| Time Complexity | Space Complexity |
|-----------------|------------------|
| O(n) | O(512MB + b * t) |
\**n - file size; t - num of threads; b - buffer size;*
#### Results for the provided file

| Metric | Value |
|--------|-------|
| **Input File Size** | 106 GB |
| **Total Processed IPs** | 8 000 000 000 |
| **Unique IPs** | 1 000 000 000 |
| **IP Duplication Rate** | 1:8 |

| Num of threads | Time | 
|----------------|------|
| 1               | ~4m |
| 2               | ~2m |
| 4             | ~1m5s |
| 8              | ~45s |
| 16             | ~35s |

 **Note: The most effective thread count is equal to the number of logical cores available on the system.**
 
## System Benchmark

**These results were achieved on a system with:**

- CPU: 16-core processor

- RAM: 16 GB

- Storage: SSD with 3.5 GB/s read speed

**Key Observations:**

- The program's speed was limited by the SSD's maximum read speed.

- On faster storage, the processing time could be further reduced.
