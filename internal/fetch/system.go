package fetch

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// diskRoot is the filesystem path used for root disk stats.
// Override with DISK_PATH env var when running inside Docker
// (mount host / as /host and set DISK_PATH=/host).
var diskRoot = func() string {
	if p := os.Getenv("DISK_PATH"); p != "" {
		return p
	}
	return "/"
}()

// dataRoot is an optional chain-data directory for disk stats (e.g. ~/.evmd).
// Set DATA_PATH when the validator home is on a path worth monitoring separately.
var dataRoot = strings.TrimSpace(os.Getenv("DATA_PATH"))

// SystemSnapshot holds OS-level metrics.
type SystemSnapshot struct {
	LoadAvg1  float64
	LoadAvg5  float64
	LoadAvg15 float64
	NumCPU    int
	MemTotal  uint64
	MemAvail  uint64
	SwapTotal uint64
	SwapFree  uint64
	DiskTotal uint64
	DiskUsed  uint64
	DiskAvail uint64
	DataPath  string
	DataTotal uint64
	DataUsed  uint64
	DataAvail uint64
	Err       error
}

// FetchSystem reads /proc/loadavg, /proc/meminfo, and disk stats via syscall.
func FetchSystem() SystemSnapshot {
	s := SystemSnapshot{NumCPU: runtime.NumCPU()}

	if err := readLoadAvg(&s); err != nil {
		s.Err = err
		return s
	}
	if err := readMemInfo(&s); err != nil {
		s.Err = err
		return s
	}
	if err := readDiskAt(diskRoot, &s.DiskTotal, &s.DiskUsed, &s.DiskAvail); err != nil {
		s.Err = err
		return s
	}
	if dataRoot != "" {
		s.DataPath = dataRoot
		if err := readDiskAt(dataRoot, &s.DataTotal, &s.DataUsed, &s.DataAvail); err != nil {
			s.Err = err
		}
	}
	return s
}

func readLoadAvg(s *SystemSnapshot) error {
	path := "/proc/loadavg"
	data, err := os.ReadFile(path)
	recordFileExchange(path, string(data), err)
	if err != nil {
		return err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return fmt.Errorf("unexpected /proc/loadavg format")
	}
	s.LoadAvg1, _ = strconv.ParseFloat(fields[0], 64)
	s.LoadAvg5, _ = strconv.ParseFloat(fields[1], 64)
	s.LoadAvg15, _ = strconv.ParseFloat(fields[2], 64)
	return nil
}

func readMemInfo(s *SystemSnapshot) error {
	path := "/proc/meminfo"
	f, err := os.Open(path)
	if err != nil {
		recordFileExchange(path, "", err)
		return err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		lines = append(lines, line)
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(parts[1], 10, 64)
		switch parts[0] {
		case "MemTotal:":
			s.MemTotal = val * 1024
		case "MemAvailable:":
			s.MemAvail = val * 1024
		case "SwapTotal:":
			s.SwapTotal = val * 1024
		case "SwapFree:":
			s.SwapFree = val * 1024
		}
	}
	if err := sc.Err(); err != nil {
		recordFileExchange(path, strings.Join(lines, "\n"), err)
		return err
	}
	recordFileExchange(path, strings.Join(lines, "\n"), nil)
	return nil
}

func readDiskAt(path string, total, used, avail *uint64) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		recordFSExchange(path, stat, err)
		return err
	}
	bs := uint64(stat.Bsize)
	*total = stat.Blocks * bs
	*used = (stat.Blocks - stat.Bfree) * bs
	*avail = stat.Bavail * bs
	recordFSExchange(path, stat, nil)
	return nil
}

func recordFileExchange(path, content string, err error) {
	ex := Exchange{
		Kind:    "file",
		Method:  "READ",
		URL:     path,
		Request: "(none)",
		OK:      err == nil,
	}
	if err != nil {
		ex.Error = err.Error()
	} else {
		ex.Response = truncateExchangeResponse(content)
	}
	recordTrace(ex)
}

func recordFSExchange(path string, stat syscall.Statfs_t, err error) {
	ex := Exchange{
		Kind:    "fs",
		Method:  "statfs",
		URL:     path,
		Request: "(none)",
		OK:      err == nil,
	}
	if err != nil {
		ex.Error = err.Error()
	} else {
		ex.Response = truncateExchangeResponse(fmt.Sprintf(
			"blocks=%d bfree=%d bavail=%d bsize=%d",
			stat.Blocks, stat.Bfree, stat.Bavail, stat.Bsize))
	}
	recordTrace(ex)
}
