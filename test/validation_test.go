package test

import (
	"BrunoCoin/pkg"
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"testing"
	"time"
)

func TestChkBlk(t *testing.T) {
	//setup --> code from TestBootstrapTwoNodesTwoBlocks

	utils.SetDebug(true)

	// Creates two nodes, but only starts the genesis node.
	// Also starts the miner for the genesis node. Sets the
	// mining difficulty target for the genesis node very
	// easy so that blocks can be mined quickly
	genNd := NewGenNd()
	node2 := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNd.Conf.MnrConf.InitPOWD = utils.CalcPOWD(1)
	genNd.Start()
	genNd.StartMiner()

	// Transaction made from genesis node with large fee, so
	// that it is immediately mined
	genNd.SendTx(50, 100, node2.Id.GetPublicKeyBytes())

	// Sleep to give time for the transaction to be mined
	// to a block
	time.Sleep(time.Second * 3)

	// Checks to see that the transaction was mined to a block
	// and that the other node's length is still only 1
	ChkMnChnLen(t, genNd, 2)
	ChkMnChnLen(t, node2, 1)

	// Starts the other node and connects it to the network
	node2.Start()
	genNd.ConnectToPeer(node2.Addr)

	// Sleep to give time for the node to connect to the network
	time.Sleep(time.Second * 1)

	// Bootstrap the node so that it can catch up to the genesis
	// node
	err := node2.Bootstrap()
	if err != nil {
		t.Errorf("Test errored when attempting to" +
			"bootstrap.\n")
	}

	// Sleep to give enough time for the bootstrapping process
	time.Sleep(time.Second * 2)

	// Check to see that both nodes have the same length
	// main chain now
	ChkMnChnLen(t, genNd, 2)
	ChkMnChnLen(t, node2, 2)
	ChkMnChnCons(t, []*pkg.Node{genNd, node2})
	//test for nil
	val := genNd.ChkBlk(nil)
	if val {
		t.Errorf("given a nil input, the output should have been false")
	}


	//copy an initially valid block
	b := genNd.Chain.LastBlock.Block
	bCopy := &block.Block{
		Hdr: b.Hdr,
		Transactions: b.Transactions,
		}

		//test: block does not satisfy proof of work
		bCopy.Hdr.Nonce = 0
		if genNd.ChkBlk(bCopy){
			t.Errorf("the block should not satisfy proof of work")
		}
}

func TestChkTx(t *testing.T) {
	//setup --> code from TestBootstrapTwoNodesTwoBlocks

	utils.SetDebug(true)

	// Creates two nodes, but only starts the genesis node.
	// Also starts the miner for the genesis node. Sets the
	// mining difficulty target for the genesis node very
	// easy so that blocks can be mined quickly
	genNd := NewGenNd()
	node2 := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNd.Conf.MnrConf.InitPOWD = utils.CalcPOWD(1)
	genNd.Start()
	genNd.StartMiner()

	// Transaction made from genesis node with large fee, so
	// that it is immediately mined
	genNd.SendTx(50, 100, node2.Id.GetPublicKeyBytes())

	// Sleep to give time for the transaction to be mined
	// to a block
	time.Sleep(time.Second * 3)

	// Checks to see that the transaction was mined to a block
	// and that the other node's length is still only 1
	ChkMnChnLen(t, genNd, 2)
	ChkMnChnLen(t, node2, 1)

	// Starts the other node and connects it to the network
	node2.Start()
	genNd.ConnectToPeer(node2.Addr)

	// Sleep to give time for the node to connect to the network
	time.Sleep(time.Second * 1)

	// Bootstrap the node so that it can catch up to the genesis
	// node
	err := node2.Bootstrap()
	if err != nil {
		t.Errorf("Test errored when attempting to" +
			"bootstrap.\n")
	}

	// Sleep to give enough time for the bootstrapping process
	time.Sleep(time.Second * 2)

	// Check to see that both nodes have the same length
	// main chain now
	ChkMnChnLen(t, genNd, 2)
	ChkMnChnLen(t, node2, 2)
	ChkMnChnCons(t, []*pkg.Node{genNd, node2})


	//copy a valid transaction
	trx := genNd.Chain.LastBlock.Block.Transactions


		var oneTrx *tx.Transaction
		for _, v := range trx {
			oneTrx = v
			utils.Debug.Printf("%v", v)
		}
	//test t.SumInputs() > 0

	inptCopy := oneTrx.Inputs[0]
	inpt1 := proto.NewTxInpt(
		inptCopy.TransactionHash,
		inptCopy.OutputIndex,
		inptCopy.UnlockingScript,
		0)

	ouptCopy := oneTrx.Outputs[0]
	outp1 := proto.NewTxOutpt(ouptCopy.Amount, ouptCopy.LockingScript)

	trx1 := proto.NewTx(oneTrx.Version,
		[]*proto.TransactionInput{inpt1},
		[]*proto.TransactionOutput{outp1}, oneTrx.LockTime)

	trxD := tx.Deserialize(trx1)

	//test t.SumInputs()> 0
	if genNd.ChkTx(trxD){
		t.Errorf("should have been false where inputs > 0")
	}

	//test t.SumOutput() > 0
	outp2 := proto.NewTxOutpt(0, ouptCopy.LockingScript)
	trx2 := proto.NewTx(oneTrx.Version,
		[]*proto.TransactionInput{inpt1},
		[]*proto.TransactionOutput{outp2}, oneTrx.LockTime)

	trxD2 := tx.Deserialize(trx2)
	if genNd.ChkTx(trxD2){
		t.Errorf(" should have been false where outputs > 0")
	}

	//test for t.SumInputs() > t.SumOutputs()
	inpt3 := proto.NewTxInpt(
		inptCopy.TransactionHash,
		inptCopy.OutputIndex,
		inptCopy.UnlockingScript,
		10)
	outp3 := proto.NewTxOutpt(20, ouptCopy.LockingScript)
	trx3 := proto.NewTx(oneTrx.Version,
		[]*proto.TransactionInput{inpt3},
		[]*proto.TransactionOutput{outp3}, oneTrx.LockTime)
	trxD3 := tx.Deserialize(trx3)
	if genNd.ChkTx(trxD3){
		t.Errorf("should have returned false for inputs < outputs")
	}


}