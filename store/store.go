package store

import (
	"database/sql"
	"log"

	"github.com/jmjac/vrscClient"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) Store {
	d := Store{db}
	return d
}

//Save the block, transactions and transaction address mapping to databse
func (s Store) EnterBlock(block *vrscClient.Block) error {
	var valueTotal int64
	var blockFee int64
	var minedCoins int64
	txCount := len(block.Tx)
	for _, tx := range block.Tx {
		coinbaseTx := false
		var txFee int64
		var valueTx int64
		addressesInOut := make(map[string][]bool)
		for _, vin := range tx.Vin {
			if vin.Coinbase != "" {
				coinbaseTx = true
			}
			valueTx += vin.ValueSat
			txFee += vin.ValueSat

			if _, ok := addressesInOut[vin.Address]; !ok {
				addressesInOut[vin.Address] = make([]bool, 2, 2)
			}
			addressesInOut[vin.Address][0] = true
		}

		for _, vout := range tx.Vout {
			if coinbaseTx {
				minedCoins += vout.ValueSat
			} else {
				txFee -= vout.ValueSat
			}

			for _, address := range vout.ScriptPubKey.Addresses {
				if _, ok := addressesInOut[address]; !ok {
					addressesInOut[address] = make([]bool, 2, 2)
				}
				addressesInOut[address][1] = true
			}
		}

		for _, vjs := range tx.Vjoinsplit {
			valueTx -= vjs.VpubNew

		}

		//Fix for transactions with no inputs from
		if txFee > 0 {
			blockFee += txFee
		}
		if valueTx > 0 {
			valueTotal += valueTx
		}
		for address, v := range addressesInOut {
			err := s.mapTransactionAddress(tx.Txid, address, v[0], v[1], block.Hash)
			if err != nil {
				log.Fatal("When mapping addresses got:", err)
			}
		}
		err := s.enterTransaction(tx.Txid, len(tx.Vin), len(tx.Vout), txFee, block.Hash, valueTx)
		if err != nil {
			log.Fatalf("Got %v in block: %v Tx: %v", err, block.Hash, tx.Txid)
		}
	}

	stmt := `INSERT INTO blocks (height, tx_count, value_total, mined_coins, block_fee, block_hash) VALUES ( $1, $2, $3, $4, $5, $6 )`
	_, err := s.db.Exec(stmt, block.Height, txCount, valueTotal, minedCoins, blockFee, block.Hash)
	return err
}

//Should remove block from blocks. Transactions from this block from transactions and transaction ids from transaction_address
func (s Store) RemoveBlock(blockHash string) {

}

func (s Store) enterTransaction(txHash string, numVin, numVout int, txFee int64, blockHash string, value int64) error {
	stmt := `INSERT INTO transactions (tx_hash, num_vin, num_vout, tx_fee, block_hash, value) VALUES ( $1, $2, $3, $4, $5, $6 )`
	_, err := s.db.Exec(stmt, txHash, numVin, numVout, txFee, blockHash, value)
	return err
}

func (s Store) mapTransactionAddress(txHash, address string, vin, vout bool, blockHash string) error {
	stmt := "INSERT INTO address_transaction (address, transaction_hash, vin, vout, block_hash) VALUES ( $1, $2, $3, $4, $5 )"
	_, err := s.db.Exec(stmt, address, txHash, vin, vout, blockHash)
	return err
}

func (s Store) GetAddressTransactions(address string) {

}

func (s Store) GetTransaction(txHash string) {

}

func (s Store) GetTransactionsWithMinFee(minFee int) {

}

func (s Store) GetTransactionsWithMinValue(minValue int) {

}

func (s Store) GetBlock(blockHash string) {

}

func (s Store) GetBlocks(start, end int) {

}

func (s Store) GetLastBlocksWithMinNumTX(nBlocks int, minNumTx int) {

}
