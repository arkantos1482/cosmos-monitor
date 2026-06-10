package panel

import (
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeFeemarket(w Writer, d model.Report) {
	w.Section("Fee market")
	writeFeemarketPage(w, d)
	w.BlankLine()
}
