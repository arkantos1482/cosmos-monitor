package fetch

import (
	"strings"
	"testing"
)

func TestHexToBech32RoundTrip(t *testing.T) {
	hexAddr := "31aec3d55f45aa21f7efbcc4257ea9f56c9ad300"
	bech := hexToBech32(Bech32PrefixCons, hexAddr)
	if bech == "" {
		t.Fatal("hexToBech32 returned empty")
	}
	got := bech32ToHex(bech)
	if got != hexAddr {
		t.Fatalf("round trip: got %q want %q", got, hexAddr)
	}
}

func TestConsHexToBech32(t *testing.T) {
	hexAddr := "31aec3d55f45aa21f7efbcc4257ea9f56c9ad300"
	bech := ConsHexToBech32(hexAddr)
	if bech == "" || !strings.HasPrefix(bech, Bech32PrefixCons+"1") {
		t.Fatalf("ConsHexToBech32: got %q", bech)
	}
}

func TestValOperToAcc(t *testing.T) {
	valoper := "cosmosvaloper1vmr9wxpldngnh0tvpr8h2pk2aycts3v7z8pdxh"
	want := "cosmos1vmr9wxpldngnh0tvpr8h2pk2aycts3v78n4c2y"
	if got := ValOperToAcc(valoper); got != want {
		t.Fatalf("ValOperToAcc = %q want %q", got, want)
	}
}

func TestAccBech32ToEVM(t *testing.T) {
	bech := "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx"
	evm := AccBech32ToEVM(bech)
	want := "0xEDACCBBFB7DB3278BC72AEEF66CC10A96C272A38"
	if evm != want {
		t.Fatalf("got %q want %q", evm, want)
	}
}
