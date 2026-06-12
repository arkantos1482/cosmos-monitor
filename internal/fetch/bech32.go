package fetch

import (
	"encoding/hex"
	"strings"

	"github.com/btcsuite/btcd/btcutil/bech32"
)

// Bech32 HRPs for this chain (cosmos-evm default; matches evmd config).
const (
	Bech32PrefixAcc     = "cosmos"
	Bech32PrefixValOper = "cosmosvaloper"
	Bech32PrefixCons    = "cosmosvalcons"
)

// hexToBech32 encodes raw address bytes (hex) as a bech32 string with checksum.
func hexToBech32(hrp, hexStr string) string {
	hexStr = strings.TrimPrefix(strings.ToLower(hexStr), "0x")
	b, err := hex.DecodeString(hexStr)
	if err != nil || len(b) == 0 {
		return ""
	}
	fiveBit, err := bech32.ConvertBits(b, 8, 5, true)
	if err != nil {
		return ""
	}
	s, err := bech32.Encode(hrp, fiveBit)
	if err != nil {
		return ""
	}
	return s
}

// bech32ToBytes decodes a bech32 address to raw bytes (checksum verified).
func bech32ToBytes(bech string) ([]byte, error) {
	hrp, data, err := bech32.Decode(bech)
	if err != nil {
		return nil, err
	}
	if hrp == "" {
		return nil, bech32.ErrInvalidCharacter(' ')
	}
	return bech32.ConvertBits(data, 5, 8, false)
}

// ConsHexToBech32 encodes a CometBFT consensus address hex string as cosmosvalcons bech32.
func ConsHexToBech32(hexStr string) string {
	return hexToBech32(Bech32PrefixCons, hexStr)
}

// ValOperToAcc maps a validator operator bech32 address to the cosmos account
// bech32 with the same underlying bytes (cosmos-evm / ethsecp256k1 validators).
func ValOperToAcc(valoper string) string {
	b, err := bech32ToBytes(valoper)
	if err != nil || len(b) == 0 {
		return ""
	}
	return hexToBech32(Bech32PrefixAcc, hex.EncodeToString(b))
}

// AccBech32ToEVM derives the EVM hex address from a Cosmos account bech32 address.
func AccBech32ToEVM(bech string) string {
	b, err := bech32ToBytes(bech)
	if err != nil || len(b) == 0 {
		return ""
	}
	return "0x" + strings.ToUpper(hex.EncodeToString(b))
}
