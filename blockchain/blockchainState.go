package blockchain

import (
	"log"
	"time"

	"github.com/jmjac/overDrawer/store"
	"github.com/jmjac/vrscClient"
)

type BlockchainState struct {
	Identities       map[string]vrscClient.Identity `json:"-"`
	LockedIdentities []identityWithBalance          `json:"-"`
	Height           int64
	BlockHash        string
	Stats            Stats     `json:"-"`
	StatsPerHour     []Summary `json:"-"`
	StatsPerDay      []Summary `json:"-"`
	store            *store.Store
	verus            vrscClient.Verus
}

func New(verus vrscClient.Verus, store *store.Store) BlockchainState {
	b := BlockchainState{}
	b.Identities = make(map[string]vrscClient.Identity)
	b.LockedIdentities = make([]identityWithBalance, 0)
	b.store = store
	b.verus = verus
	b.Stats = Stats{}
	b.Height = 0
	return b
}

func (b *BlockchainState) SetBlockchain(verus vrscClient.Verus) {
	b.verus = verus
}

func LoadBlockchainState(verus vrscClient.Verus, store *store.Store) BlockchainState {
	b := BlockchainState{}
	b.verus = verus
	b.store = store
	b.Identities = make(map[string]vrscClient.Identity)
	b.LockedIdentities = make([]identityWithBalance, 0)
	//Check if the chain diverged
	height, hash := b.store.GetTopBlock()
	blockHash, _ := verus.GetBlockHash(height)
	for blockHash != hash {
		err := b.store.RemoveByHeight(int64(height))
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Block hash in store is different then on chain block hash. Removing block:", height)
		height, hash = b.store.GetTopBlock()
		blockHash, _ = verus.GetBlockHash(height)
	}
	b.Height = height
	return b
}

//Save the block, transactions and transaction address mapping to databse
func (b *BlockchainState) enterBlock(blockHeight int64) (string, error) {
	//TODO: Fix non-standard TX
	//Slow??
	block, err := b.verus.GetBlockFromHeight(blockHeight)
	if err != nil {
		return "", err
	}

	hash, err := b.store.GetBlockHash(block.Height)
	//if Block is in the db
	if err == nil {
		if block.Hash == hash {
			log.Println("Block already in databse", block.Height)
			b.Height++
			return block.Hash, nil

		} else {
			log.Println("Repeated height, wrong hash. Removing all wrong blocks from db", hash, block.Hash)
			err := b.store.RemoveByHeight(block.Height)
			b.Height--
			if err != nil {
				return "", err
			}
		}
	}

	var valueTotal int64
	var blockFee int64
	var minedCoins int64
	txCount := int64(len(block.Tx))
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
				if address == "" {
					continue
				}
				if _, ok := addressesInOut[address]; !ok {
					addressesInOut[address] = make([]bool, 2, 2)
				}
				addressesInOut[address][1] = true
			}

			if vout.ScriptPubKey.Identityprimary != nil {
				name := vout.ScriptPubKey.Identityprimary.Name
				address := vout.ScriptPubKey.Identityprimary.Identityaddress
				if name != "" {
					b.store.EnterIdentity(name, address, block.Height, block.Hash)
				}
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
			err := b.store.MapTransactionAddress(tx.Txid, address, v[0], v[1], block.Hash, block.Height)
			if err != nil {
				return "", err
			}
		}
		err := b.store.EnterTransaction(tx.Txid, len(tx.Vin), len(tx.Vout), txFee, block.Hash, valueTx, block.Height)
		if err != nil {
			return "", err
		}
	}

	err = b.store.EnterBlock(block.Height, txCount, valueTotal, minedCoins, blockFee, block.Hash)
	if err != nil {
		return "", err
	}
	log.Println("Block:", b.Height)
	b.Height++
	return block.Hash, nil
}

func (b *BlockchainState) Scan(quit chan bool) {
	for {
		select {
		case <-quit:
			log.Println("Stopping blockchain scan")
			return
		default:
		}

		top, _ := b.verus.GetBlockCount()

		for b.Height == top {
			top, _ = b.verus.GetBlockCount()
			log.Println("Updating stats")
			b.CalculateStats()
			log.Println("Sleeping")
			time.Sleep(time.Second * 30)
		}
		hash, err := b.enterBlock(b.Height)
		if err != nil {
			log.Fatal(err)
		}

		b.BlockHash = hash
		if b.Height%200 == 0 {
			log.Println("Updating stats")
			b.CalculateStats()
		}
	}
}

type identityWithBalance struct {
	Id      vrscClient.Identity
	Balance int64
}

func (b *BlockchainState) GetLockedIdentities() []identityWithBalance {
	timeLocked := make([]identityWithBalance, 0)
	bc, _ := b.verus.GetBlockCount()
	for i := range b.Identities {
		id, err := b.verus.GetIdentity(i + "@")
		if err != nil {
			continue
		}
		if id.IdentityPrimary.Timelock > int64(bc) {
			balance, err := b.verus.GetAddressBalance([]string{id.IdentityPrimary.Identityaddress})
			if err != nil {
				log.Fatal(err)
			}
			timeLocked = append(timeLocked, identityWithBalance{*id, balance.Balance})
		}

	}
	b.LockedIdentities = timeLocked
	return timeLocked
}

func (b *BlockchainState) DeleteBlocks(n int64) {
	top, _ := b.verus.GetBlockCount()
	b.store.RemoveByHeight(top - n)
}

func (b *BlockchainState) WriteLastNBlocksToDB(n int64) {
	top, _ := b.verus.GetBlockCount()
	for i := top - n; i < top; i++ {
		b.enterBlock(i)
	}
	b.Height = top
}

func (b *BlockchainState) CalculateStats() (Stats, error) {
	stats := Stats{}
	blocks, err := b.store.GetBlocks(b.Height-60*24*365, b.Height)
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
	var hourlyTx, hourlyValue int64
	var dayTx, dayValue int64
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
