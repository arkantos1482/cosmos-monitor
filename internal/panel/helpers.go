package panel

import (
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

const validatorLocalRowClass = "data-table__row--local"

func validatorRowClasses(validators []model.Validator) []string {
	if len(validators) == 0 {
		return nil
	}
	classes := make([]string, len(validators))
	for i, v := range validators {
		if v.IsLocal {
			classes[i] = validatorLocalRowClass
		}
	}
	return classes
}

func writeValidatorSetTable(w Writer, headers []string, rows [][]string, validators []model.Validator) {
	w.TableWithRowClasses(headers, rows, validatorRowClasses(validators))
}

func pmtStatus(d model.Report) string {
	if !d.PMTEnabled {
		return "disabled"
	}
	if d.PMTPoolEmpty {
		return "enabled — pool empty  (no PMT rewards distributing)"
	}
	suffix := ""
	if d.PMTRunway != "" {
		suffix = "  (" + d.PMTRunway + ")"
	}
	return "distributing  " + d.PMTRate + "   pool " + d.PMTBalance + suffix
}
