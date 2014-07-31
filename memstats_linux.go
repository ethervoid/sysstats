// +build linux

package sysstats

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

// MemStat represents the memory statistics on a linux system
type MemStats map[string]uint64

// getMemStats gets the memory stats of a linux system from the
// file /proc/meminfo
// It returns MemStats (map) with the following keys:
// MemUsed      -  Total size of used memory in kilobytes.
// MemFree      -  Total size of free memory in kilobytes.
// MemTotal     -  Total size of memory in kilobytes.
// Buffers      -  Total size of buffers used from memory in kilobytes.
// Cached       -  Total size of cached memory in kilobytes.
// RealFree     -  Total size of memory is real free (memfree + buffers + cached).
// SwapUsed     -  Total size of swap space is used is kilobytes.
// SwapFree     -  Total size of swap space is free in kilobytes.
// SwapTotal    -  Total size of swap space in kilobytes.
// Swapcached   -  Memory that once was swapped out, is swapped back in but still
//                 also is in the swapfile.
// Active       -  Memory that has been used more recently and usually not
//                 reclaimed unless absolutely necessary.
// Inactive     -  Memory which has been less recently used and is more eligible
//                 to be reclaimed for other purposes.
//
// The following statistics are only available for kernels >= 2.6.
// Slab         -  Total size of memory in kilobytes that used by kernel for data
//                 structure allocations.
// Dirty        -  Total size of memory pages in kilobytes that waits to be
//                 written back to disk.
// Mapped       -  Total size of memory in kilobytes that is mapped by devices or
//                 libraries with mmap.
// Writeback    -  Total size of memory that was written back to disk.
// Committed_AS -  The amount of memory presently allocated on the system.
//
// The following statistic is only available for kernels >= 2.6.9.
// CommitLimit  -  Total amount of memory currently available to be allocated on
//                 the system.
func getMemStats() (memStats MemStats, err error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	memStats = MemStats{}
	re := regexp.MustCompile(`^((?:Mem|Swap)(?:Total|Free)|Buffers|Cached|` +
		`SwapCached|Active|Inactive|Dirty|Writeback|Mapped|Slab|` +
		`Commit(?:Limit|ted_AS)):\s*(\d+)`)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		stat := re.FindStringSubmatch(line)
		if stat == nil {
			// No match
			continue
		}
		key := stat[1]
		value, err := strconv.ParseUint(stat[2], 10, 64)
		if err != nil {
			fmt.Println(err)
			continue
		} else {
			memStats[key] = value
		}
	}

	memStats[`MemUsed`] = memStats[`MemTotal`] - memStats[`MemFree`]
	memStats[`SwapUsed`] = memStats[`SwapTotal`] - memStats[`SwapFree`]
	memStats[`RealFree`] = memStats[`MemFree`] + memStats[`Buffers`] + memStats[`Cached`]

	return memStats, nil
}