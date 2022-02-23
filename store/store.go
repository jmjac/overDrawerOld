package store

import (
	"database/sql"
	"fmt"
	"log"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) Store {
	d := Store{}
	d.db = db
	return d
}

func (s Store) EnterBlock(height, txCount, valueTotal, minedCoins, blockFee int64, blockHash string) error {
	stmt := `INSERT INTO blocks (height, tx_count, value_total, mined_coins, block_fee, block_hash) VALUES ( $1, $2, $3, $4, $5, $6 )`
	_, err := s.db.Exec(stmt, height, txCount, valueTotal, minedCoins, blockFee, blockHash)
	return err
}

func (s Store) EnterBTCBlock(height, txCount int64, valueTotal, minedCoins, blockFee float64, blockHash string) error {
	stmt := `INSERT INTO blocks (height, tx_count, value_total, mined_coins, block_fee, block_hash) VALUES ( $1, $2, $3, $4, $5, $6 )`
	_, err := s.db.Exec(stmt, height, txCount, valueTotal, minedCoins, blockFee, blockHash)
	return err
}

func (s Store) RemoveByHeight(height int64) error {
	queryBlocks := "DELETE FROM blocks WHERE height >= $1"
	queryTransactions := "DELETE FROM transactions WHERE height >= $1"
	queryAddresses := "DELETE FROM address_transaction WHERE height >= $1"
	_, err := s.db.Exec(queryBlocks, height)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(queryTransactions, height)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(queryAddresses, height)
	if err != nil {
		return err
	}
	return nil

}

func (s Store) removeBlocksByHeight(height int64) {
	//TODO: Rewrite to use height for faster queries
	query := "SELECT block_hash FROM blocks WHERE height >= $1"
	rows, err := s.db.Query(query, height)
	if err != nil {
		log.Fatal(err)
	}
	var deleteHash string
	for rows.Next() {
		rows.Scan(&deleteHash)
		log.Println("STORE: Removing block: ", deleteHash)
		s.removeBlock(deleteHash)
	}
	rows.Close()
}

//Should remove block from blocks. Transactions from this block from transactions and transaction ids from transaction_address
func (s Store) removeBlock(blockHash string) {
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

func (s Store) EnterTransaction(txHash string, numVin, numVout int, txFee int64, blockHash string, value int64, height int64) error {
	stmt := `INSERT INTO transactions (tx_hash, num_vin, num_vout, tx_fee, block_hash, value, height) VALUES ( $1, $2, $3, $4, $5, $6, $7 )`
	_, err := s.db.Exec(stmt, txHash, numVin, numVout, txFee, blockHash, value, height)
	return err
}

func (s Store) EnterBTCTransaction(txHash string, numVin, numVout int, txFee float64, blockHash string, value float64, height int64) error {
	stmt := `INSERT INTO transactions (tx_hash, num_vin, num_vout, tx_fee, block_hash, value, height) VALUES ( $1, $2, $3, $4, $5, $6, $7 )`
	_, err := s.db.Exec(stmt, txHash, numVin, numVout, txFee, blockHash, value, height)
	return err
}

func (s Store) EnterIdentity(name, address string, height int64, blockHash string) error {
	stmt := `INSERT INTO identities (name, address, height, block_hash) VALUES ( $1, $2, $3, $4)`
	_, err := s.db.Exec(stmt, name, address, height, blockHash)
	return err

}

func (s Store) MapTransactionAddress(txHash, address string, vin, vout bool, blockHash string, height int64) error {
	stmt := "INSERT INTO address_transaction (address, transaction_hash, vin, vout, height, block_hash) VALUES ( $1, $2, $3, $4, $5, $6 )"
	_, err := s.db.Exec(stmt, address, txHash, vin, vout, height, blockHash)

	return err
}

func (s Store) GetTopBlock() (int64, string) {
	query := "SELECT height, block_hash FROM blocks ORDER BY height DESC LIMIT 1"
	row := s.db.QueryRow(query)
	var height int64
	var hash string
	row.Scan(&height, &hash)
	return height, hash
}

func (s Store) GetBlockHash(height int64) (string, error) {
	var blockHash string
	query := "SELECT block_hash from blocks where height = $1"
	row := s.db.QueryRow(query, height)
	err := row.Scan(&blockHash)
	if err != nil {
		return "", err
	}
	return blockHash, nil
}

type blockSummary struct {
	Height     int64
	ValueTotal int64
	TxCount    int64
	MinedCoins int64
}

func (s Store) GetBlocks(start, end int64) ([]*blockSummary, error) {
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

type blockSummaryBTC struct {
	Height     int64
	ValueTotal float64
	TxCount    int64
	MinedCoins float64
}

func (s Store) GetBTCBlocks(start, end int64) ([]*blockSummaryBTC, error) {
	query := "SELECT height, value_total, mined_coins, tx_count from blocks where height >= $1 AND height <= $2 ORDER BY height ASC"
	rows, err := s.db.Query(query, start, end)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var blocksSummaries []*blockSummaryBTC

	for rows.Next() {
		bs := &blockSummaryBTC{}
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
