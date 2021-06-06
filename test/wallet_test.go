package test

import (
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
		utils.Debug.Printf("fake debug", trxD)
		t.Errorf("ahh ")
	}
}

func TestWaHndlBlk(t *testing.T) {
	//setup
	utils.SetDebug(true)
	//genNd := NewGenNd()

}