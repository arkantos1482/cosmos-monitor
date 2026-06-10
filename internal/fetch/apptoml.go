package fetch

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// AppTomlGasConfig holds node-local gas acceptance settings from app.toml.
type AppTomlGasConfig struct {
	MinGasPrices      string
	EVMMinTip         string
	MempoolPriceLimit string
	MaxTxGasWanted    string
	Path              string
	OK                bool
}

var (
	appTomlQuotedRE = regexp.MustCompile(`^([a-zA-Z0-9_.-]+)\s*=\s*"([^"]*)"`)
	appTomlBareRE   = regexp.MustCompile(`^([a-zA-Z0-9_.-]+)\s*=\s*(\S+)`)
)

// AppTomlPath returns the app.toml path (APPTOML_PATH env or first existing default).
func AppTomlPath() string {
	if p := os.Getenv("APPTOML_PATH"); p != "" {
		return p
	}
	for _, path := range []string{
		"/home/ubuntu/.evmd/config/app.toml",
		"/data/config/app.toml",
	} {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return "/home/ubuntu/.evmd/config/app.toml"
}

// FetchAppTomlGasConfig parses fee-related keys from the node app.toml.
func FetchAppTomlGasConfig() AppTomlGasConfig {
	path := AppTomlPath()
	cfg := AppTomlGasConfig{Path: path}

	f, err := os.Open(path)
	if err != nil {
		return cfg
	}
	defer f.Close()

	inEVMMempool := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.Trim(line, "[]")
			inEVMMempool = section == "evm.mempool"
			continue
		}
		key, val, ok := parseTomlKV(line)
		if !ok {
			continue
		}
		switch {
		case key == "minimum-gas-prices":
			cfg.MinGasPrices = val
		case key == "min-tip":
			cfg.EVMMinTip = val
		case key == "max-tx-gas-wanted":
			cfg.MaxTxGasWanted = val
		case inEVMMempool && key == "price-limit":
			cfg.MempoolPriceLimit = val
		}
	}
	cfg.OK = cfg.MinGasPrices != "" || cfg.EVMMinTip != "" ||
		cfg.MempoolPriceLimit != "" || cfg.MaxTxGasWanted != ""
	return cfg
}

func parseTomlKV(line string) (key, val string, ok bool) {
	if m := appTomlQuotedRE.FindStringSubmatch(line); len(m) == 3 {
		return m[1], m[2], true
	}
	if m := appTomlBareRE.FindStringSubmatch(line); len(m) == 3 {
		return m[1], strings.Trim(m[2], `"'`), true
	}
	return "", "", false
}
