package main

import (
	"fmt"

	"horizon-core/internal/chain"
)

func init() {
	chain.ConsensusGate = func(txs []chain.PendingTx) (bool, string) {
		v := RunAngelConsensusBatch(txs)
		if v.Passed {
			return true, fmt.Sprintf("OK critical=%d/%d important=%d/%d root=%s",
				v.Critical, v.CriticalTotal, v.Important, v.ImportantTotal, v.Root[:12])
		}
		return false, fmt.Sprintf("REJECT critical=%d/%d important=%d/%d",
			v.Critical, v.CriticalTotal, v.Important, v.ImportantTotal)
	}
}
