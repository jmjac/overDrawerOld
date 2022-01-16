package blockchain

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/jmjac/vrscClient"
)

type BlockchainState struct {
	Identities       map[string]vrscClient.Identity
	LockedIdentities []identityWithBalance
	Height           int
	LastBlocks       []vrscClient.Block
	Stats            Stats
	verus            vrscClient.Verus
}

func New(verus vrscClient.Verus) BlockchainState {
	b := BlockchainState{}
	b.Identities = make(map[string]vrscClient.Identity)
	b.LockedIdentities = make([]identityWithBalance, 0)
	b.LastBlocks = make([]vrscClient.Block, 0)
	b.verus = verus
	b.Height = 0
	return b
}

func (b BlockchainState) SaveToDisk(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	out, err := json.Marshal(b)
	if err != nil {
		return err
	}
	_, err = f.Write(out)
	return err
}

func (b *BlockchainState) SetBlockchain(verus vrscClient.Verus) {
	b.verus = verus
}

func LoadBlockchainState(filename string) (BlockchainState, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return BlockchainState{}, nil
	}
	var b BlockchainState
	err = json.Unmarshal(f, &b)
	return b, err
}

func (b *BlockchainState) Scan() error {
	for {
		top, _ := b.verus.GetBlockCount()
		saved := false
		for b.Height == top {
			//TODO: Change this
			log.Println("Updating stats")
			b.CalculateStats()
			top, _ = b.verus.GetBlockCount()
			time.Sleep(time.Second * 20)
			if !saved {
				log.Println("Saving state")
				b.SaveToDisk("state.json")
				log.Println("State saved")
				saved = true
			}
			log.Println("Sleeping")
		}
		saved = false

		if b.Height%20000 == 0 {
			log.Println("Saving state")
			b.SaveToDisk("state.json")
			log.Println("State saved")
		}
		log.Println("Block: ", b.Height)
		block, err := b.verus.GetBlockFromHeight(b.Height)
		b.LastBlocks = append(b.LastBlocks, *block)
		if len(b.LastBlocks) > 60*24*30 {
			b.LastBlocks = b.LastBlocks[1:]

		}

		if err != nil {
			return err
		}

		for name := range checkForIdentitiesCreation(block) {
			if _, ok := b.Identities[name]; !ok {
				//ch <- identity
			}

			id, err := b.verus.GetIdentity(name + "@")
			if err != nil {
				log.Println("Error:", err)
				log.Println("For identity:", name)
			} else {

				b.Identities[name] = *id
			}

		}
		detectMoneyMovement(block)
		calculateBlockFee(block)
		b.Height++
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

func (b *BlockchainState) TransactionsInLastBlocks(n int) int {
	top, _ := b.verus.GetBlockCount()
	txs := make([]vrscClient.Tx, 0)
	blocks := make([]vrscClient.Block, 0)
	for i := top - n; i < top; i++ {
		block, err := b.verus.GetBlockFromHeight(i)
		if err != nil {
			log.Fatal(err)
		}
		blocks = append(blocks, *block)
		for _, tx := range transactionInBlock(*block) {
			txs = append(txs, tx)
		}
	}
	b.LastBlocks = blocks
	b.Height = top
	return len(txs)
}

func transactionInBlock(block vrscClient.Block) []vrscClient.Tx {
	txs := make([]vrscClient.Tx, 0)
	for _, tx := range block.Tx {
		coinbaseTx := false
		for _, vin := range tx.Vin {
			if vin.Coinbase != "" {
				coinbaseTx = true
				break
			}
		}
		if !coinbaseTx {
			txs = append(txs, tx)
		}
	}
	return txs
}

func (b *BlockchainState) CalculateStats() Stats {
	stats := Stats{}
	monthPerHour := make([]Summary, 0)
	for i, block := range b.LastBlocks {
		movement := detectMoneyMovement(&block)
		var totalMoved int64
		for _, val := range movement {
			totalMoved += val
		}
		count := len(movement)

		//Month
		stats.Month.MoneyMoved += totalMoved
		stats.Month.TransactionCount += count
		//Week
		if i >= len(b.LastBlocks)-60*24*7 {
			stats.Week.MoneyMoved += totalMoved
			stats.Week.TransactionCount += count
		}

		//Day
		if i >= len(b.LastBlocks)-60*24 {
			stats.Day.MoneyMoved += totalMoved
			stats.Day.TransactionCount += count
		}

		//Hour
		if i >= len(b.LastBlocks)-60 {
			stats.Hour.MoneyMoved += totalMoved
			stats.Hour.TransactionCount += count
		}
		if i%60 == 0 && i != 0 {
			monthPerHour = append(monthPerHour, Summary{count, totalMoved})
		}

	}
	stats.MonthPerHour = monthPerHour
	b.Stats = stats
	return stats
}
