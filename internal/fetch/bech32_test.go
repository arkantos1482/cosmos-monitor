package fetch

import "testing"

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

func TestAccBech32ToEVM(t *testing.T) {
	bech := "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx"
	evm := AccBech32ToEVM(bech)
	want := "0xEDACCBBFB7DB3278BC72AEEF66CC10A96C272A38"
	if evm != want {
		t.Fatalf("got %q want %q", evm, want)
	}
}
