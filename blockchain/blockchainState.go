package blockchain

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/jmjac/overDrawer/store"
	"github.com/jmjac/vrscClient"
)

type BlockchainState struct {
	Identities       map[string]vrscClient.Identity
	LockedIdentities []identityWithBalance
	Height           int
	Stats            Stats
	store            *store.Store
	verus            vrscClient.Verus
	filename         string
}

func New(verus vrscClient.Verus, store *store.Store, savefilename string) BlockchainState {
	b := BlockchainState{}
	b.Identities = make(map[string]vrscClient.Identity)
	b.LockedIdentities = make([]identityWithBalance, 0)
	b.store = store
	b.verus = verus
	b.filename = savefilename
	b.Stats = Stats{}
	b.Height = 0
	return b
}

func (b BlockchainState) SaveToDisk() error {
	f, err := os.Create(b.filename)
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

func LoadBlockchainState(filename string, verus vrscClient.Verus, store *store.Store) (BlockchainState, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return BlockchainState{}, nil
	}
	var b BlockchainState
	err = json.Unmarshal(f, &b)
	b.filename = filename
	b.verus = verus
	b.store = store
	return b, err
}

func (b *BlockchainState) scanBlock() error {
	log.Println("Block: ", b.Height)
	block, err := b.verus.GetBlockFromHeight(b.Height)
	if err != nil {
		return err
	}

	for name := range checkForIdentitiesCreation(block) {
		id, err := b.verus.GetIdentity(name + "@")
		if err != nil {
			log.Println("Error:", err)
			log.Println("For identity:", name)
		} else {
			b.Identities[name] = *id
		}

	}
	err = b.store.EnterBlock(block)
	b.Height++
	return err
}

func (b *BlockchainState) Scan(quit chan bool) {
	for {
		select {
		case <-quit:
			log.Println("Stopping blockchain scan")
			b.SaveToDisk()
			log.Println("State saved")
			return
		default:
		}

		top, _ := b.verus.GetBlockCount()
		b.scanBlock()
		saved := false

		for b.Height == top {
			top, _ = b.verus.GetBlockCount()
			if !saved {
				log.Println("Updating stats")
				b.CalculateStats()
				log.Println("Saving state")
				b.SaveToDisk()
				log.Println("State saved")
				saved = true
			}
			select {

			case <-quit:
				log.Println("Stopping blockchain scan")
				log.Println("Saving state")
				b.SaveToDisk()
				log.Println("State saved")
				return
			default:
			}

			log.Println("Sleeping")
			time.Sleep(time.Second * 10)
		}
		saved = false
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

func (b *BlockchainState) WriteLastNBlocksToDB(n int) int {
	top, _ := b.verus.GetBlockCount()
	txs := make([]vrscClient.Tx, 0)
	for i := top - n; i < top; i++ {
		block, err := b.verus.GetBlockFromHeight(i)
		if err != nil {
			log.Fatal(err)
		}

		for _, tx := range transactionInBlock(*block) {
			txs = append(txs, tx)
		}
	}
	b.Height = top
	return len(txs)
}

func transactionInBlock(block vrscClient.Block) []vrscClient.Tx {
	//TODO: Change
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
	//TODO: Implement
	stats := Stats{}
	stats.Day = Summary{100, 1000}
	stats.Hour = Summary{9900, 8300}
	stats.Month = Summary{19500, 14320}
	stats.Week = Summary{12100, 51000}
	stats.BlockCount = b.Height
	b.Stats = stats
	return stats
}
