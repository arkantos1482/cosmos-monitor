package fetch

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// diskRoot is the filesystem path used for disk stats.
// Override with DISK_PATH env var when running inside Docker
// (mount host / as /host and set DISK_PATH=/host).
var diskRoot = func() string {
	if p := os.Getenv("DISK_PATH"); p != "" {
		return p
	}
	return "/"
}()

// SystemSnapshot holds OS-level metrics.
type SystemSnapshot struct {
	LoadAvg1  float64
	LoadAvg5  float64
	LoadAvg15 float64
	MemTotal  uint64
	MemAvail  uint64
	DiskTotal uint64
	DiskUsed  uint64
	Err       error
}

// FetchSystem reads /proc/loadavg, /proc/meminfo, and disk stats via syscall.
func FetchSystem() SystemSnapshot {
	s := SystemSnapshot{}

	if err := readLoadAvg(&s); err != nil {
		s.Err = err
		return s
	}
	if err := readMemInfo(&s); err != nil {
		s.Err = err
		return s
	}
	if err := readDisk(&s); err != nil {
		s.Err = err
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
		}
	}
	if err := sc.Err(); err != nil {
		recordFileExchange(path, strings.Join(lines, "\n"), err)
		return err
	}
	recordFileExchange(path, strings.Join(lines, "\n"), nil)
	return nil
}

func readDisk(s *SystemSnapshot) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(diskRoot, &stat); err != nil {
		recordFSExchange(diskRoot, stat, err)
		return err
	}
	bs := uint64(stat.Bsize)
	s.DiskTotal = stat.Blocks * bs
	s.DiskUsed = (stat.Blocks - stat.Bfree) * bs
	recordFSExchange(diskRoot, stat, nil)
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
