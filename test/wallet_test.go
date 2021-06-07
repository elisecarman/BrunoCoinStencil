package test

import (
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"testing"
)

func TestWaChkTxs(t *testing.T){
	utils.SetDebug(true)
	genNd := NewGenNd()

	//setup
	inpt1 := proto.NewTxInpt("",0,"", 100)
	inpt2 := proto.NewTxInpt("",0,"", 101)
	outp1 := proto.NewTxOutpt(200, "")
	trx1 := proto.NewTx(0,
		[]*proto.TransactionInput{inpt1, inpt2},
		[]*proto.TransactionOutput{outp1}, 0)
	trxD := tx.Deserialize(trx1)

	inpt10 := proto.NewTxInpt("",0,"", 100)
	inpt20 := proto.NewTxInpt("",0,"", 200)
	outp10 := proto.NewTxOutpt(200, "")
	trx10 := proto.NewTx(0,
		[]*proto.TransactionInput{inpt10, inpt20},
		[]*proto.TransactionOutput{outp10}, 0)
	trxD0 := tx.Deserialize(trx10)

	genNd.Wallet.LmnlTxs.Add(trxD0)

	//test nil input
	genNd.Wallet.LmnlTxs.ChkTxs([]*tx.Transaction{nil})
	if genNd.Wallet.LmnlTxs.TxQ.Len() != 1{
		t.Errorf("the function was given a nil input and should not" +
			"have updated the liminal txs pool")
	}

	//test removal
	genNd.Wallet.LmnlTxs.ChkTxs([]*tx.Transaction{trxD0})
	if genNd.Wallet.LmnlTxs.TxQ.Has(trxD0){
		t.Errorf("duplicate were not removed ")
	}


	//test removal above priority threshold
	genNd.Wallet.LmnlTxs.TxRplyThresh = 1
	genNd.Wallet.LmnlTxs.Add(trxD)
	genNd.Wallet.LmnlTxs.Add(trxD0)
	genNd.Wallet.LmnlTxs.ChkTxs([]*tx.Transaction{})
	if genNd.Wallet.LmnlTxs.TxQ.Has(trxD) {
		t.Errorf("priority above 1 should have been removed")
	}
}

func TestAdd(t *testing.T){
	utils.SetDebug(true)
	genNd := NewGenNd()

	inpt1 := proto.NewTxInpt("",0,"", 100)
	inpt2 := proto.NewTxInpt("",0,"", 101)
	outp1 := proto.NewTxOutpt(200, "")
	trx1 := proto.NewTx(0,
		[]*proto.TransactionInput{inpt1, inpt2},
		[]*proto.TransactionOutput{outp1}, 0)
	trxD := tx.Deserialize(trx1)

	oldQueue := genNd.Wallet.LmnlTxs.TxQ.Len()
	//test nil input
	genNd.Wallet.LmnlTxs.Add(nil)
	if oldQueue != genNd.Wallet.LmnlTxs.TxQ.Len(){
		t.Errorf("a nil input was given to the function. The Heap should not change")
	}

	//test simple addition
	genNd.Wallet.LmnlTxs.Add(trxD)
	if !genNd.Wallet.LmnlTxs.TxQ.Has(trxD){
		t.Errorf("the element was not properly added to the Queue")
	}

}

func TestWaHndlBlk(t *testing.T) {
	//setup
	utils.SetDebug(true)
	genNd := NewGenNd()
	genNd.Start()
	//setup
	inpt1 := proto.NewTxInpt("",0,"", 100)
	inpt2 := proto.NewTxInpt("",0,"", 101)
	outp1 := proto.NewTxOutpt(200, "")
	trx1 := proto.NewTx(0,
		[]*proto.TransactionInput{inpt1, inpt2},
		[]*proto.TransactionOutput{outp1}, 0)
	trxD := tx.Deserialize(trx1)

	inpt10 := proto.NewTxInpt("",0,"", 100)
	inpt20 := proto.NewTxInpt("",0,"", 200)
	outp10 := proto.NewTxOutpt(200, "")
	trx10 := proto.NewTx(0,
		[]*proto.TransactionInput{inpt10, inpt20},
		[]*proto.TransactionOutput{outp10}, 0)
	trxD0 := tx.Deserialize(trx10)
	b := block.New("",
		[]*tx.Transaction{trxD0},
		"")

	genNd.Wallet.LmnlTxs.Add(trxD0)

	//test nil input
	genNd.Wallet.HndlBlk(nil)
	if genNd.Wallet.LmnlTxs.TxQ.Len() != 1{
		t.Errorf("the function was given a nil input and should not" +
			"have updated the liminal txs pool")
	}

	//test removal
	genNd.Wallet.HndlBlk(b)
	if genNd.Wallet.LmnlTxs.TxQ.Has(trxD0){
		t.Errorf("duplicate were not removed ")
	}


	//test removal above priority threshold
	genNd.Wallet.LmnlTxs.TxRplyThresh = 1
	genNd.Wallet.LmnlTxs.Add(trxD)
	genNd.Wallet.LmnlTxs.Add(trxD0)
	genNd.Wallet.HndlBlk(b)
	if genNd.Wallet.LmnlTxs.TxQ.Has(trxD) {
		t.Errorf("priority above 1 should have been removed")
	}
}

func TestWaHndlTxReq(t *testing.T){
	//setup
	utils.SetDebug(true)
	genNd := NewGenNd()
	genNd.Start()
	//setup

	oldQueue := genNd.Wallet.LmnlTxs.TxQ.Len()
	//test nil input
	genNd.Wallet.HndlTxReq(nil)
	if oldQueue != genNd.Wallet.LmnlTxs.TxQ.Len(){
		t.Errorf("a nil input was given to the function. The Heap should not change")
	}


	////testing addition to liminal pool
	//if genNd.Wallet.LmnlTxs.TxQ.Len() != 1{
	//	t.Errorf("the element was not properly added to the liminal pool")
	//}
	//
	//
	//tx := genNd.Wallet.LmnlTxs.TxQ.RemAbv(0)
	//////test correct inputs
	//if tx[0].SumInputs() != 100000{
	//	t.Errorf("the inputs were not properly formatted")
	//}
	//
	//////test correct outputs amount
	//if tx[0].SumOutputs() != 99995{
	//	t.Errorf("the outputs were not properly formatted")
	//}
	//
	//////test output: change and amt sent
	//if len(tx[0].Outputs) != 2{
	//	t.Errorf("the outputs were not properly formatted")
	//}
	//
	//////test correct version
	//if tx[0].Version != genNd.Wallet.Conf.TxVer{
	//	t.Errorf("the wrong version was inputted")
	//}
	//
	//if tx[0].LockTime != genNd.Wallet.Conf.DefLckTm{
	//	t.Errorf("the wrong locktime was inputted")
	//}

}