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
	Identities       map[string]vrscClient.Identity `json:"-"`
	LockedIdentities []identityWithBalance          `json:"-"`
	Height           int
	Stats            Stats `json:"-"`
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

func (b BlockchainState) saveToDisk() error {
	f, err := os.Create("delete.txt")
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

func LoadBlockchainState(verus vrscClient.Verus, store *store.Store) (BlockchainState, error) {
	f, err := os.ReadFile("state2.json")
	if err != nil {
		return BlockchainState{}, nil
	}
	var b BlockchainState
	err = json.Unmarshal(f, &b)
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
			//b.SaveToDisk()
			//log.Println("State saved")
			return
		default:
		}

		top, _ := b.verus.GetBlockCount()
		b.scanBlock()

		for b.Height == top {
			top, _ = b.verus.GetBlockCount()
			log.Println("Updating stats")
			b.CalculateStats()
			//log.Println("Saving state")
			//b.SaveToDisk()
			//log.Println("State saved")
			select {

			case <-quit:
				log.Println("Stopping blockchain scan")
				//log.Println("Saving state")
				//b.SaveToDisk()
				//log.Println("State saved")
				return
			default:
			}

			log.Println("Sleeping")
			time.Sleep(time.Second * 30)
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

func (b *BlockchainState) DeleteBlocks(n int) {
	top, _ := b.verus.GetBlockCount()
	b.store.RemoveBlocksByHeight(int64(top - n))
}

func (b *BlockchainState) WriteLastNBlocksToDB(n int) {
	top, _ := b.verus.GetBlockCount()
	for i := top - n; i < top; i++ {
		block, err := b.verus.GetBlockFromHeight(i)
		if err != nil {
			log.Fatal(err)
		}
		b.store.EnterBlock(block)
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
	stats.YearPerHour = make([]Summary, 0)
	var hourlyTx, hourlyValue int64
	for i, block := range blocks {
		stats.Year.TxCount += block.TxCount
		stats.Year.Value += block.ValueTotal
		hourlyTx += block.TxCount
		hourlyValue += block.ValueTotal
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
			stats.YearPerHour = append(stats.YearPerHour, Summary{hourlyTx, hourlyValue})
			hourlyTx = 0
			hourlyValue = 0
		}
	}
	stats.BlockCount = b.Height
	b.Stats = stats
	return stats, nil
}
