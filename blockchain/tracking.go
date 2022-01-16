package blockchain

import (
	"bufio"
	"fmt"
	"os"

	"github.com/jmjac/vrscClient"
)

func exploreBlock(block *vrscClient.Block) {
	checkForIdentitiesCreation(block)
	fmt.Println("Block fee:", calculateBlockFee(block))
}

func loadIdentitiesNames(filename string) (map[string]bool, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	identities := make(map[string]bool)
	for scanner.Scan() {
		identities[scanner.Text()] = true
	}

	return identities, nil
}

func checkForIdentitiesCreation(block *vrscClient.Block) map[string]bool {
	newIdentities := make(map[string]bool)
	for _, tx := range block.Tx {
		for _, vout := range tx.Vout {
			if vout.ScriptPubKey.Identityprimary != nil {
				name := vout.ScriptPubKey.Identityprimary.Name
				newIdentities[name] = true
			}
		}
	}
	return newIdentities
}

func calculateBlockFee(block *vrscClient.Block) map[string]float64 {
	fees := make(map[string]float64)
	var in, out int64
	coinbaseTx := false
	for _, tx := range block.Tx {
		for _, vin := range tx.Vin {
			if vin.Coinbase != "" {
				coinbaseTx = true
			}
			in += vin.ValueSat

		}
		for _, vout := range tx.Vout {
			if coinbaseTx {
				coinbaseTx = false
				continue
			}
			out += vout.ValueSat
		}

		if !coinbaseTx {
			fees[tx.Txid] = float64(in-out) / 100000000

		}
	}
	return fees
}

func getIdentities(verus *vrscClient.Verus, start, top int, identites map[string]bool) {
	for i := start; i < top; i++ {
		block, _ := verus.GetBlockFromHeight(i)
		for c := range checkForIdentitiesCreation(block) {
			identites[c] = true
		}
	}
}

func detectMoneyMovement(block *vrscClient.Block) map[string]int64 {
	coinbaseTx := false
	txs := make(map[string]int64)
	for _, tx := range block.Tx {
		var in, out int64
		for _, vin := range tx.Vin {
			if vin.Coinbase != "" {
				coinbaseTx = true
			}
			in += vin.ValueSat
		}

		if !coinbaseTx {
			txs[tx.Txid] = in
		}
		for _, vout := range tx.Vout {
			if coinbaseTx {
				coinbaseTx = false
				continue
			}
			out += vout.ValueSat
		}
	}

	return txs
}
