package panel

import (
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeFeemarket(w Writer, d model.Report) {
	w.Section("3. FEE MARKET")
	w.Em("Chain-wide EIP-1559 fee market — live base fee, demand, and governance parameters.")
	writeFeemarketPage(w, d)
	w.BlankLine()
}
