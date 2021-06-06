package test

import (
	"BrunoCoin/pkg"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"testing"
	"time"
)

func TestGenCBTx(t *testing.T) {
	utils.SetDebug(true)
	genNd := NewGenNd()
	node2 := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNd.Start()
	node2.Start()
	genNd.ConnectToPeer(node2.Addr)
	if peer := genNd.PeerDb.Get(node2.Addr); peer == nil {
		t.Fatal("Seed node did not contain newNode as peer")
	}
	if peer := node2.PeerDb.Get(genNd.Addr); peer == nil {
		t.Fatal("New node did not contain seedNode as peer")
	}
	// Sleep to give time for both nodes to connect
	time.Sleep(1 * time.Second)

	// Checks that both nodes have the same main blockchain,
	// and are both connected
	ChkMnChnCons(t, []*pkg.Node{genNd, node2})
	ChkNdPrs(t, genNd, []*pkg.Node{node2})
	ChkNdPrs(t, node2, []*pkg.Node{genNd})

	// The genesis node sends money to node 2 with a high
	// fee
	genNd.SendTx(100, 100, node2.Id.GetPublicKeyBytes())
	node2.SendTx(100, 100, genNd.Id.GetPublicKeyBytes())
	// Sleep to give time for the transaction to be broadcast,
	// and validated by the other node
	time.Sleep(6 * time.Second)
	node2.StartMiner()
	valid := genNd.Mnr.GenCBTx(nil)
	if valid != nil {
		t.Errorf("Did not account for nil input")
	}
	valid2 := genNd.Mnr.GenCBTx([]*tx.Transaction{})
	if valid2 != nil {
		t.Errorf("Did not account for empty input")
	}

	trxTest := proto.NewTx(0,
		[]*proto.TransactionInput{},
		[]*proto.TransactionOutput{},
		0)

	valid3 := genNd.Mnr.GenCBTx([]*tx.Transaction{tx.Deserialize(trxTest), nil})
	if valid3 != nil {
		t.Errorf("Did not account for empty input")
	}

	inpt1 := proto.NewTxInpt("",0,"", 100)
	inpt2 := proto.NewTxInpt("",0,"", 100)

	outp1 := proto.NewTxOutpt(200, "")
	goodTrx := proto.NewTx(0,
		[]*proto.TransactionInput{inpt1, inpt2},
		[]*proto.TransactionOutput{outp1}, 0)

	basicValidity := genNd.Mnr.GenCBTx([]*tx.Transaction{tx.Deserialize(goodTrx)})

	if basicValidity.SumInputs() != 0{
		t.Errorf("Does not send trx with no inputs")
	}
	if basicValidity.SumOutputs() != 5 {
		t.Errorf("Sent the incorrect reward")
	}


}