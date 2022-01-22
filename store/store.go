package store

import (
	"database/sql"
	"fmt"
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
	//TODO: Fix repeated blocks
	query := "SELECT height, block_hash from blocks where height = $1"
	rows, err := s.db.Query(query, block.Height)
	if err != nil {
		log.Fatal(err)
	}
	var height int64
	var hash string
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&height, &hash)
		if hash == block.Hash {
			log.Println("STORE: Block already in databse", block.Height)
			return nil
		} else {
			log.Println("STORE: Repeated height, wrong hash. Removing all wrong blocks from db", hash, block.Hash)
			s.RemoveBlocksByHeight(block.Height)
		}
	}

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
				log.Fatal("STORE: When mapping addresses got:", err)
			}
		}
		err := s.enterTransaction(tx.Txid, len(tx.Vin), len(tx.Vout), txFee, block.Hash, valueTx)
		if err != nil {
			log.Fatalf("STORE: Got %v in block: %v Tx: %v", err, block.Hash, tx.Txid)
		}
	}

	stmt := `INSERT INTO blocks (height, tx_count, value_total, mined_coins, block_fee, block_hash) VALUES ( $1, $2, $3, $4, $5, $6 )`
	_, err = s.db.Exec(stmt, block.Height, txCount, valueTotal, minedCoins, blockFee, block.Hash)
	return err
}

func (s Store) RemoveBlocksByHeight(height int64) {
	query := "SELECT block_hash FROM blocks WHERE height >= $1"
	rows, err := s.db.Query(query, height)
	if err != nil {
		log.Fatal(err)
	}
	var deleteHash string
	for rows.Next() {
		rows.Scan(&deleteHash)
		log.Println("STORE: Removing block: ", deleteHash)
		s.RemoveBlock(deleteHash)
	}
	rows.Close()
}

//Should remove block from blocks. Transactions from this block from transactions and transaction ids from transaction_address
func (s Store) RemoveBlock(blockHash string) {
	queryBlocks := "DELETE FROM blocks WHERE block_hash = $1"
	queryTransactions := "DELETE FROM transactions WHERE block_hash = $1"
	queryAddresses := "DELETE FROM address_transaction WHERE block_hash = $1"
	num, err := s.db.Exec(queryBlocks, blockHash)
	fmt.Println("STORE: Deleted:", num)
	if err != nil {
		log.Fatal(err)
	}
	num, err = s.db.Exec(queryTransactions, blockHash)
	fmt.Println("STORE: Deleted:", num)
	if err != nil {
		log.Fatal(err)
	}
	num, err = s.db.Exec(queryAddresses, blockHash)
	fmt.Println("STORE: Deleted:", num)
	if err != nil {
		log.Fatal(err)
	}
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

func (s Store) GetBlock(blockHash string) {

}

type blockSummary struct {
	Height     int64
	ValueTotal int64
	TxCount    int64
	MinedCoins int64
}

func (s Store) GetBlocks(start, end int) ([]*blockSummary, error) {
	query := "SELECT height, value_total, mined_coins, tx_count from blocks where height >= $1 AND height <= $2 ORDER BY height ASC"
	rows, err := s.db.Query(query, start, end)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var blocksSummaries []*blockSummary

	for rows.Next() {
		bs := &blockSummary{}
		err := rows.Scan(&bs.Height, &bs.ValueTotal, &bs.MinedCoins, &bs.TxCount)
		if err != nil {
			log.Fatal(err)
		}
		blocksSummaries = append(blocksSummaries, bs)
	}

	return blocksSummaries, nil
}

func (s Store) GetAddressTransactions(address string) {

}

func (s Store) GetTransaction(txHash string) {

}

func (s Store) GetTransactionsWithMinFee(minFee int) {

}

func (s Store) GetTransactionsWithMinValue(minValue int) {

}

func (s Store) GetLastBlocksWithMinNumTX(nBlocks int, minNumTx int) {

}
