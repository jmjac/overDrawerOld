package btcState

import (
	"fmt"
	"log"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/jmjac/overDrawer/store"
)

type BtcState struct {
	Height       int64
	BlockHash    string
	Stats        Stats     `json:"-"`
	StatsPerHour []Summary `json:"-"`
	StatsPerDay  []Summary `json:"-"`
	store        *store.Store
	btc          *rpcclient.Client
}

func New(btc *rpcclient.Client, store *store.Store) BtcState {
	b := BtcState{}
	b.store = store
	b.btc = btc
	b.Stats = Stats{}
	b.BlockHash = ""
	b.Height = 0
	return b
}

func LoadBlockchainState(btc *rpcclient.Client, store *store.Store) BtcState {
	b := BtcState{}
	b.btc = btc
	b.store = store
	height, hash := b.store.GetTopBlock()
	blockHash, _ := b.btc.GetBlockHash(height)
	for blockHash.String() != hash {
		err := b.store.RemoveByHeight(int64(height))
		if err != nil {
			log.Fatal(err)
		}
		log.Println("BTC: Block hash in store is different then on chain block hash. Removing block:", height)
		height, hash = b.store.GetTopBlock()
		blockHash, _ = b.btc.GetBlockHash(height)
	}
	b.Height = height
	return b
}

//Save the block, transactions and transaction address mapping to databse
func (b *BtcState) enterBlock(blockHeight int64) (string, error) {
	//TODO: Fix non-standard TX
	blockHash, err := b.btc.GetBlockHash(blockHeight)
	if err != nil {
		return "", err
	}
	block, err := b.btc.GetBlockVerboseTx(blockHash)
	if err != nil {
		return "", err
	}

	now := time.Now()
	hash, err := b.store.GetBlockHash(block.Height)
	fmt.Println("Block height lookup took:", time.Since(now))
	//if Block is in the db
	if err == nil {
		if block.Hash == hash {
			log.Println("BTC: Block already in databse", block.Height)
			b.Height++
			return block.Hash, nil

		} else {
			log.Println("BTC: Repeated height, wrong hash. Removing all wrong blocks from db", hash, block.Hash)
			err := b.store.RemoveByHeight(block.Height)
			b.Height--
			if err != nil {
				return "", err
			}
		}
	}

	var valueTotal float64
	var blockFee float64
	var minedCoins float64
	txCount := int64(len(block.Tx))
	for _, tx := range block.Tx {
		coinbaseTx := false
		var txFee float64
		var valueTx float64
		addressesInOut := make(map[string][]bool)
		for _, vin := range tx.Vin {
			if vin.IsCoinBase() {
				coinbaseTx = true
			} else {
				value, addresses := b.findTxInputAndAddress(vin.Vout, vin.Txid)
				valueTx += value
				txFee += value

				//TODO: The map ma be not needed
				for _, address := range addresses {
					if _, ok := addressesInOut[address]; !ok {
						addressesInOut[address] = make([]bool, 2, 2)
					}
					addressesInOut[address][0] = true

				}
			}
		}

		for _, vout := range tx.Vout {
			if coinbaseTx {
				minedCoins += vout.Value
			} else {
				txFee -= vout.Value
			}

			for _, address := range vout.ScriptPubKey.Addresses {
				if address == "" {
					continue
				}
				if _, ok := addressesInOut[address]; !ok {
					addressesInOut[address] = make([]bool, 2, 2)
				}
				addressesInOut[address][1] = true
			}
		}

		//Fix for transactions with no inputs from
		if txFee > 0 {
			blockFee += txFee
		}
		if valueTx > 0 {
			valueTotal += valueTx
		}
		for address, v := range addressesInOut {
			err := b.store.MapTransactionAddress(tx.Txid, address, v[0], v[1], block.Hash, block.Height)
			if err != nil {
				return "", err
			}
		}
		err := b.store.EnterBTCTransaction(tx.Txid, len(tx.Vin), len(tx.Vout), txFee, block.Hash, valueTx, block.Height)
		if err != nil {
			return "", err
		}
	}

	err = b.store.EnterBTCBlock(block.Height, txCount, valueTotal, minedCoins, blockFee, block.Hash)
	if err != nil {
		return "", err
	}
	log.Println("BTC: Block:", b.Height)
	b.Height++
	return block.Hash, nil
}

func (b BtcState) findTxInputAndAddress(N uint32, txHash string) (float64, []string) {
	hash, err := chainhash.NewHashFromStr(txHash)
	if err != nil {
		log.Fatal(err)
	}
	tx, err := b.btc.GetRawTransactionVerbose(hash)
	if err != nil {
		log.Fatal(err)
	}
	vout := tx.Vout[N]
	value := vout.Value
	addresses := vout.ScriptPubKey.Addresses
	return value, addresses
}

func (b *BtcState) Scan(quit chan bool) {
	for {
		select {
		case <-quit:
			log.Println("Stopping BTC blockchain scan")
			return
		default:
		}

		top, _ := b.btc.GetBlockCount()

		for b.Height == top {
			top, _ = b.btc.GetBlockCount()
			log.Println("Updating BTC stats")
			b.CalculateStats()
			log.Println("BTC Sleeping")
			time.Sleep(time.Second * 30)
		}
		hash, err := b.enterBlock(b.Height)
		if err != nil {
			log.Fatal(err)
		}

		b.BlockHash = hash
		if b.Height%200 == 0 {
			log.Println("Updating BTC stats")
			b.CalculateStats()
		}
	}
}

func (b *BtcState) DeleteBlocks(n int64) {
	top, _ := b.btc.GetBlockCount()
	b.store.RemoveByHeight(top - n)
}

func (b *BtcState) WriteLastNBlocksToDB(n int64) {
	top, _ := b.btc.GetBlockCount()
	for i := top - n; i < top; i++ {
		b.enterBlock(i)
	}
	b.Height = top
}

func (b *BtcState) CalculateStats() (Stats, error) {
	stats := Stats{}
	blocks, err := b.store.GetBTCBlocks(b.Height-60*24*365, b.Height)
	if err != nil {
		return Stats{}, err
	}

	stats.Day = Summary{}
	stats.Hour = Summary{}
	stats.Week = Summary{}
	stats.Month = Summary{}
	stats.Year = Summary{}
	b.StatsPerDay = make([]Summary, 0)
	b.StatsPerHour = make([]Summary, 0)
	var hourlyTx, dayTx int64
	var hourlyValue, dayValue float64
	for i, block := range blocks {
		stats.Year.TxCount += block.TxCount
		stats.Year.Value += block.ValueTotal
		hourlyTx += block.TxCount
		hourlyValue += block.ValueTotal
		dayTx += block.TxCount
		dayValue += block.ValueTotal
		//Month
		if i >= 60*24*(365-30) {
			stats.Month.TxCount += block.TxCount
			stats.Month.Value += block.ValueTotal
		}
		//Week
		if i >= 60*24*(365-7) {
			stats.Week.TxCount += block.TxCount
			stats.Week.Value += block.ValueTotal
		}
		//Day
		if i >= 60*24*(365-1) {
			stats.Day.TxCount += block.TxCount
			stats.Day.Value += block.ValueTotal
		}
		//Hour
		if i >= 60*24*(365)-60 {
			stats.Hour.TxCount += block.TxCount
			stats.Hour.Value += block.ValueTotal
		}

		if (i+1)%60 == 0 {
			b.StatsPerHour = append(b.StatsPerHour, Summary{hourlyTx, hourlyValue})
			hourlyTx = 0
			hourlyValue = 0
		}

		if (i+1)%(60*24) == 0 {
			b.StatsPerDay = append(b.StatsPerDay, Summary{dayTx, dayValue})
			dayTx = 0
			dayValue = 0
		}
	}
	stats.BlockCount = b.Height
	b.Stats = stats
	return stats, nil
}
