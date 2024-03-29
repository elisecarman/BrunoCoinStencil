package miner

import (
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"context"
	"encoding/hex"
	"fmt"
	"math"
)

/*
 *  Brown University, CS1951L, Summer 2021
 *  Designed by: Colby Anderson, Parker Ljung
 */

// Mine waits to be told to mine a block
// or to kill it's thread. If it is asked
// to mine, it selects the transactions
// with the highest priority to add to the
// mining pool. The nonce is then attempted
// to be found unless the miner is stopped.
func (m *Miner) Mine() {
	ctx, cancel := context.WithCancel(context.Background())
	for {
		<-m.PoolUpdated
		cancel()
		if !m.Active.Load() {
			continue
		}
		ctx, cancel = context.WithCancel(context.Background())
		go func(ctx context.Context) {
			if !m.TxP.PriMet() {
				return
			}
			m.Mining.Store(true)
			m.MiningPool = m.NewMiningPool()
			txs := append([]*tx.Transaction{m.GenCBTx(m.MiningPool)}, m.MiningPool...)

			//utils.Debug.Printf("Mine: the Sum input of the first transaction %v of mining pool is %v",
			//	m.MiningPool[0].NameTag(),m.MiningPool[0].SumInputs())
			//utils.Debug.Printf("Mine: the first transaction of mining pool is %v",
			//	m.MiningPool[0].NameTag())
			//utils.Debug.Printf("Mine: the first transaction of the block is %v and has inputs: %v and output %v",
			//	txs[0].NameTag(), txs[0].SumInputs(), txs[0].SumOutputs())
			//utils.Debug.Printf("Mine: the second transaction of the block is %v and has inputs: %v",
			//	txs[1].NameTag(), txs[1].SumInputs())


			b := block.New(m.PrvHsh, txs, m.DifTrg())
			result := m.CalcNonce(ctx, b)
			m.Mining.Store(false)
			if result {
				utils.Debug.Printf("%v mined %v %v", utils.FmtAddr(m.Addr), b.NameTag(), b.Summarize())
				m.SendBlk <- b
				m.HndlBlk(b)
			}
		}(ctx)
	}
	cancel()
}

// Returns boolean to indicate success
func (m *Miner) CalcNonce(ctx context.Context, b *block.Block) bool {
	for i := uint32(0); i < m.Conf.NncLim; i++ {
		select {
		case <-ctx.Done():
			return false
		default:
			b.Hdr.Nonce = i
			if b.SatisfiesPOW(m.DifTrg()) {
				return true
			}
		}
	}
	return false
}

// DifTrg (DifficultyTarget) calculates the
// current difficulty target.
// Returns:
// string the difficulty target as a hex
// string
func (m *Miner) DifTrg() string {
	return m.Conf.InitPOWD
}

// GenCBTx (GenerateCoinbaseTransaction) generates a coinbase
// transaction based off the transactions in the mining pool.
// It does this by adding the fee reward to the minting reward.
// Inputs:
// txs	[]*tx.Transaction the transactions (besides the
// coinbase tx) that the miner is mining to a block
// Returns:
// the coinbase transaction that pays the miner the reward
// for mining the block
// TODO
// 1. calculate how many fees you can collect
// 2. calculate the minting reward
// for the minting reward: it starts as c.InitSubsdy
// and gets cut in half every c.SubsdyHlvRt until
// c.MxHlvgs
// 3. Then a transaction is made with zero inputs
// and with an output paying the miner fees + mint

// some functions/fields/methods that might be helpful
// tx.Deserialize(...)
// proto.NewTx(...)
// m.Conf.Ver
// m.Id.GetPublicKeyBytes()
// proto.NewTxOutpt(...)
// hex.EncodeToString(...)
// m.Conf.DefLckTm
// m.ChnLen.Load()
// c.SubsdyHlvRt
// c.MxHlvgs
// c.InitSubsdy
// t.SumInputs()
// t.SumOutputs()

func (m *Miner) GenCBTx(txs []*tx.Transaction) *tx.Transaction {
	if txs != nil && len(txs) != 0{
		fee := uint32(0)
		for _, t := range txs {
			if t == nil {
				fmt.Printf("ERROR {txs.GenCBTx}:" + " a nil transaction within the list")
				return nil
			} else {
				sumIn := t.SumInputs()
				sumOut := t.SumOutputs()
				fee += sumIn - sumOut
				utils.Debug.Printf("sum of inputs: %v", t.SumInputs())
				utils.Debug.Printf("sum of outputs: %v", t.SumOutputs())
			}
		}

		halves := math.Floor(float64(m.ChnLen.Load() / m.Conf.SubsdyHlvRt))
		mint := float64(0)
		if halves < float64(m.Conf.MxHlvgs){
			mint += float64(m.Conf.InitSubsdy)/ math.Pow(float64(2), halves)   //ToDo: many changes (Pow)
		} else {
			mint += float64(m.Conf.InitSubsdy)/ math.Pow(float64(2), float64(m.Conf.MxHlvgs))
		}
		reward := uint32(mint) + fee
		pubK := hex.EncodeToString(m.Id.GetPublicKeyBytes())
		outpt := proto.NewTxOutpt(reward, pubK)

		trx := proto.NewTx(m.Conf.Ver, []*proto.TransactionInput{}, []*proto.TransactionOutput{outpt}, m.Conf.DefLckTm)
		return tx.Deserialize(trx)

	} else {
		fmt.Printf("ERROR {tp.GenCBTx}: " + "nil transactions:  were given to the function")

	return nil
	}
}

