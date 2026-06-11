package fetch

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FetchChainStatus loads the minimal CometBFT + REST data for the global status strip.
func FetchChainStatus(rpc, rest string) ChainSnapshot {
	snap := ChainSnapshot{}

	var status statusResp
	if err := doJSON(rpc+"/status", &status); err != nil {
		snap.Err = fmt.Errorf("rpc status: %w", err)
		return snap
	}

	snap.Moniker = status.Result.NodeInfo.Moniker
	snap.BlockHeight = parseInt64(status.Result.SyncInfo.LatestBlockHeight)
	snap.CatchingUp = status.Result.SyncInfo.CatchingUp
	if t, err := time.Parse(time.RFC3339Nano, status.Result.SyncInfo.LatestBlockTime); err == nil {
		snap.LatestBlockTime = t
	}

	var netInfo netInfoResp
	if err := doJSON(rpc+"/net_info", &netInfo); err == nil {
		snap.PeerCount, _ = strconv.Atoi(netInfo.Result.NPeers)
	}

	var bf baseFeeResp
	if err := doJSON(rest+"/cosmos/evm/feemarket/v1/base_fee", &bf); err == nil {
		snap.BaseFee = bf.BaseFee
	}

	return snap
}

// FetchEVMPeerCount loads only net_peerCount for the status strip.
func FetchEVMPeerCount(endpoint string) EVMSnapshot {
	snap := EVMSnapshot{}
	p := evmProbe(endpoint, "net_peerCount", []any{})
	if !p.OK {
		snap.Err = fmt.Errorf("net_peerCount: %s", firstNonEmpty(p.Error, "failed"))
		return snap
	}
	var peerRaw json.RawMessage
	if probeResult(p, &peerRaw) {
		s := strings.Trim(string(peerRaw), `"`)
		if strings.HasPrefix(s, "0x") {
			snap.PeerCount = hexToUint64(s)
		} else if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			snap.PeerCount = v
		}
	}
	return snap
}

// FetchDockerRunning loads only container running state for the status strip.
func FetchDockerRunning(container string) DockerSnapshot {
	client := newDockerClient()
	snap := DockerSnapshot{}

	var insp dockerInspect
	if err := doDockerJSON(client, fmt.Sprintf("http://localhost/containers/%s/json", container), &insp); err != nil {
		snap.Err = err
		return snap
	}
	snap.Running = insp.State.Running
	return snap
}
