package test

import (
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"encoding/hex"
	"testing"
)

func TestGetUTXO(t *testing.T){
	utils.SetDebug(true)
	genNd := NewGenNd()

	//test if nil
	_, _, bool := genNd.Chain.GetUTXOForAmt(0, "")
	if bool {
		t.Errorf("given no public key the function should have returned false")
	}

	//test amount too big
	_, _, bool2 := genNd.Chain.GetUTXOForAmt(1000000000, "")
	if bool2 {
		t.Errorf("given too big an amount the function should have returned false")
	}

	//test given right info
	info,_,_ :=  genNd.Chain.GetUTXOForAmt(10, hex.EncodeToString(genNd.Id.GetPublicKeyBytes()))
	singleInput := info[0]

	if singleInput.Amt != 100000 &&
		singleInput.OutIdx != 0 {
		t.Errorf("the info was improperly formated")
	}
	utils.Debug.Printf("amt is %v", singleInput.Amt)

}

func TestBlkAdd(t *testing.T){
	utils.SetDebug(true)
	genNd := NewGenNd()
	oldBlkChn := genNd.Chain.Length()

	//test nil
	genNd.Chain.Add(nil)
	newBlkChn := genNd.Chain.Length()
	if oldBlkChn != newBlkChn {
		t.Errorf("the function was given a nil input but still added to the chain")
	}

	inpt10 := proto.NewTxInpt("",0,"", 100)
	inpt20 := proto.NewTxInpt("",0,"", 200)
	outp10 := proto.NewTxOutpt(200, "")
	trx10 := proto.NewTx(0,
		[]*proto.TransactionInput{inpt10, inpt20},
		[]*proto.TransactionOutput{outp10}, genNd.Mnr.Conf.DefLckTm)
	trxD0 := tx.Deserialize(trx10)
	b := block.New(genNd.Chain.LastBlock.Hash(),
		[]*tx.Transaction{trxD0},
		genNd.Mnr.Conf.InitPOWD)

	lastBlock := genNd.Chain.LastBlock

	print(b)
	genNd.Chain.Add(b)

	//test update of hash
	if genNd.Chain.Length() == oldBlkChn {
		t.Errorf("the Hash was not succesfully added")
	}

	//test last block updated
	if lastBlock == genNd.Chain.LastBlock{
		t.Errorf("the last block was not changed")
	}

}
