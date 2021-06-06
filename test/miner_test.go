package test

import (
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/miner"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"testing"
)

func TestGenCBTx(t *testing.T) {
	//setup
	utils.SetDebug(true)
	genNd := NewGenNd()

	//test for the nil input
	valid := genNd.Mnr.GenCBTx(nil)
	if valid != nil {
		t.Errorf("Did not account for nil input")
	}

	//Test for the transaction slice being empty
	valid2 := genNd.Mnr.GenCBTx([]*tx.Transaction{})
	if valid2 != nil {
		t.Errorf("Did not account for empty input")
	}

	//setup
	trxTest := proto.NewTx(0,
		[]*proto.TransactionInput{},
		[]*proto.TransactionOutput{},
		0)

	//test for one of the transactions being nil
	valid3 := genNd.Mnr.GenCBTx([]*tx.Transaction{tx.Deserialize(trxTest), nil})
	if valid3 != nil {
		t.Errorf("Did not account for empty input")
	}

	//setup
	inpt1 := proto.NewTxInpt("",0,"", 100)
	inpt2 := proto.NewTxInpt("",0,"", 101)
	outp1 := proto.NewTxOutpt(200, "")
	goodTrx := proto.NewTx(0,
		[]*proto.TransactionInput{inpt1, inpt2},
		[]*proto.TransactionOutput{outp1}, 0)
	basicValidity := genNd.Mnr.GenCBTx([]*tx.Transaction{tx.Deserialize(goodTrx)})


	//test that the returned coinbase Trx has no inputs
	if basicValidity.SumInputs() != 0{
		t.Errorf("Does not send trx with no inputs")
	}

	//test that the correct reward was given for 1 fee and 10 mint reward
	if basicValidity.SumOutputs() != 11 {
		t.Errorf("Sent the incorrect reward")
	}

	//test for reduced mint reward
	genNd.Mnr.SetChnLen(10)
	reducedMint := genNd.Mnr.GenCBTx([]*tx.Transaction{tx.Deserialize(goodTrx)})
	if reducedMint.SumOutputs() != 6 {
		t.Errorf("Sent the incorrect reward")
	}

	//test for reached max halving rate
	genNd.Mnr.SetChnLen(99)
	ReducedMint1 := genNd.Mnr.GenCBTx([]*tx.Transaction{tx.Deserialize(goodTrx)})
	utils.Debug.Printf("%v", ReducedMint1.SumOutputs())
	if ReducedMint1.SumOutputs() != 1.0 {
		t.Errorf("Sent the incorrect reward")
	}

	genNd.Mnr.SetChnLen(100)
	maxReducedMint := genNd.Mnr.GenCBTx([]*tx.Transaction{tx.Deserialize(goodTrx)})
	utils.Debug.Printf("%v", maxReducedMint.SumOutputs())
	if maxReducedMint.SumOutputs() != 1.0 {
		t.Errorf("Sent the incorrect reward")
	}
}

func TestMnrHndlTx(t *testing.T) {
	utils.SetDebug(true)
	genNd := NewGenNd()
	poolSize := genNd.Mnr.TxP.Length()

	//setup
	inpt1 := proto.NewTxInpt("",0,"", 100)
	inpt2 := proto.NewTxInpt("",0,"", 101)
	outp1 := proto.NewTxOutpt(200, "")
	goodTrx := proto.NewTx(0,
		[]*proto.TransactionInput{inpt1, inpt2},
		[]*proto.TransactionOutput{outp1}, 0)

	genNd.Mnr.HndlTx(tx.Deserialize(goodTrx))
	newPoolSize := genNd.Mnr.TxP.Length()

	//test that pool size is updated
	if poolSize == newPoolSize {
		t.Errorf("the pool size was not updated")
	}

	//test the Priority was not met
	if genNd.Mnr.TxP.PriMet(){
		t.Errorf("the pool size was not updated")
	}

	//test the miner is not mining yet
	if genNd.Mnr.Mining.Load(){
		t.Errorf("the pool size was not updated")
	}

	//setup
	inpt10 := proto.NewTxInpt("",0,"", 100)
	inpt20 := proto.NewTxInpt("",0,"", 120)
	outp10 := proto.NewTxOutpt(20, "")
	goodTrx0 := proto.NewTx(0,
		[]*proto.TransactionInput{inpt10, inpt20},
		[]*proto.TransactionOutput{outp10}, 0)
	genNd.Mnr.HndlTx(tx.Deserialize(goodTrx0))

	//test the Priority was met
	if !genNd.Mnr.TxP.PriMet(){
		t.Errorf("the pool size was not updated")
	}

	//test the miner is now mining
	if !genNd.Mnr.Mining.Load(){
		t.Errorf("the pool size was not updated")
	}


	poolSize2 := genNd.Mnr.TxP.Length()
	genNd.Mnr.HndlTx(nil)
	poolSize3 := genNd.Mnr.TxP.Length()

	//test for nil input
	if poolSize2 != poolSize3 {
		t.Errorf("the input was nil, an error should have been thrown")
	}
}

func TestTxPChkTx(t *testing.T) {
	utils.SetDebug(true)
	genNd := NewGenNd()
	poolSize := genNd.Mnr.TxP.Length()
	genNd.ChkTx(nil)

	//test for nil input
	if poolSize != genNd.Mnr.TxP.Length(){
		t.Errorf("the input was nil, an error should have been thrown")
	}

	//setup
	inpt1 := proto.NewTxInpt("",0,"", 100)
	inpt2 := proto.NewTxInpt("",0,"", 101)
	outp1 := proto.NewTxOutpt(200, "")
	Trx1 := proto.NewTx(0,
		[]*proto.TransactionInput{inpt1, inpt2},
		[]*proto.TransactionOutput{outp1}, 0)

	inpt10 := proto.NewTxInpt("",0,"", 100)
	inpt20 := proto.NewTxInpt("",0,"", 200)
	outp10 := proto.NewTxOutpt(200, "")
	Trx10 := proto.NewTx(0,
		[]*proto.TransactionInput{inpt10, inpt20},
		[]*proto.TransactionOutput{outp10}, 0)

	genNd.Mnr.TxP.Add(tx.Deserialize(Trx1))
	genNd.Mnr.TxP.Add(tx.Deserialize(Trx10))


	genNd.Mnr.TxP.ChkTxs([]*tx.Transaction{tx.Deserialize(Trx10)})
	//test removal from Queue
	if genNd.Mnr.TxP.TxQ.Has(tx.Deserialize(Trx10)){
		t.Errorf("the transaction was not properly removed")
	}

	//test count
	if genNd.Mnr.TxP.Ct.Load() == 2 {
		t.Errorf("the count was not properly updated")
	}

	//test priority
	if genNd.Mnr.TxP.CurPri.Load() !=  miner.CalcPri(tx.Deserialize(Trx1)){
		t.Errorf("the priority was not properly updated")
	}


}


func TestMnrHndlChkBlk(t *testing.T) {
	utils.SetDebug(true)
	genNd := NewGenNd()
	poolSize := genNd.Mnr.TxP.Length()

	genNd.Mnr.HndlChkBlk(nil)

	//test for nil input
	if poolSize != genNd.Mnr.TxP.Length(){
		t.Errorf("the input was nil, an error should have been thrown")
	}

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

	genNd.Mnr.TxP.Add(trxD)
	genNd.Mnr.TxP.Add(trxD0)

	genNd.Mnr.HndlChkBlk(b)
	//test removal from Queue
	if genNd.Mnr.TxP.TxQ.Has(tx.Deserialize(trx10)){
		t.Errorf("the transaction was not properly removed")
	}

	//test count
	if genNd.Mnr.TxP.Ct.Load() == 2 {
		t.Errorf("the count was not properly updated")
	}

	//test priority
	if genNd.Mnr.TxP.CurPri.Load() !=  miner.CalcPri(tx.Deserialize(trx1)){
		t.Errorf("the priority was not properly updated")
	}

}


func TestMnrTxPAdd(t *testing.T) {
	utils.SetDebug(true)
	genNd := NewGenNd()
	poolSize := genNd.Mnr.TxP.Length()

	genNd.Mnr.TxP.Add(nil)

	//test for nil input
	if poolSize != genNd.Mnr.TxP.Length(){
		t.Errorf("the input was nil, an error should have been thrown")
	}

	//setup
	inpt1 := proto.NewTxInpt("",0,"", 100)
	inpt2 := proto.NewTxInpt("",0,"", 101)
	outp1 := proto.NewTxOutpt(200, "")
	trx1 := proto.NewTx(0,
		[]*proto.TransactionInput{inpt1, inpt2},
		[]*proto.TransactionOutput{outp1}, 0)
	trxD := tx.Deserialize(trx1)

	genNd.Mnr.TxP.Add(trxD)

	//test removal from Queue
	if !genNd.Mnr.TxP.TxQ.Has(trxD){
		t.Errorf("the transaction was not properly removed")
	}

	//test count
	if genNd.Mnr.TxP.Ct.Load() != 1 {
		t.Errorf("the count was not properly updated")
	}

	//test priority
	if genNd.Mnr.TxP.CurPri.Load() !=  miner.CalcPri(trxD){
		t.Errorf("the priority was not properly updated")
	}
}


func TestCalcPri(t *testing.T) {
	utils.SetDebug(true)
	genNd := NewGenNd()
	poolSize := genNd.Mnr.TxP.Length()

	miner.CalcPri(nil)

	//test for nil input
	if poolSize != 0{
		t.Errorf("the input was nil, an error should have been thrown")
	}

	//setup
	inpt1 := proto.NewTxInpt("",0,"", 0)
	inpt2 := proto.NewTxInpt("",0,"", 0)
	outp1 := proto.NewTxOutpt(0, "")
	trx1 := proto.NewTx(0,
		[]*proto.TransactionInput{inpt1, inpt2},
		[]*proto.TransactionOutput{outp1}, 0)
	trxD := tx.Deserialize(trx1)

	inpt10 := proto.NewTxInpt("",0,"", 5)
	inpt20 := proto.NewTxInpt("",0,"", 200)
	outp10 := proto.NewTxOutpt(10, "")
	trx10 := proto.NewTx(0,
		[]*proto.TransactionInput{inpt10, inpt20},
		[]*proto.TransactionOutput{outp10}, 0)
	trxD0 := tx.Deserialize(trx10)


	//priority of empty transaction return 1
	val := miner.CalcPri(trxD)
	if val != 1 {
		t.Errorf("the priority should have been 1 for 0 fees")
	}

	//calculate large priority
	val2 := miner.CalcPri(trxD0)
	if val2 != 180 {
		t.Errorf("the priority should have been 180 for large fee")
	}
}

func TestMnrHndlBlk(t *testing.T) {
	utils.SetDebug(true)
	genNd := NewGenNd()
	chnSize := genNd.Mnr.ChnLen.Load()

	genNd.Mnr.HndlBlk(nil)

	//test for nil input
	if chnSize != genNd.Mnr.ChnLen.Load(){
		t.Errorf("the input was nil, an error should have been thrown")
	}

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

	genNd.Mnr.TxP.Add(trxD)
	genNd.Mnr.TxP.Add(trxD0)


	oldHash := genNd.Mnr.PrvHsh
	oldChnLen := genNd.Mnr.ChnLen.Load()
	genNd.Mnr.HndlBlk(b)
	//test removal from Queue
	if genNd.Mnr.TxP.TxQ.Has(tx.Deserialize(trx10)){
		t.Errorf("the transaction was not properly removed")
	}

	//test count
	if genNd.Mnr.TxP.Ct.Load() == 2 {
		t.Errorf("the count was not properly updated")
	}

	//test priority
	if genNd.Mnr.TxP.CurPri.Load() !=  miner.CalcPri(tx.Deserialize(trx1)){
		t.Errorf("the priority was not properly updated")
	}

	//test updated hash
	if genNd.Mnr.PrvHsh == oldHash {
		t.Errorf("the previous hash was not properly updated")
	}

	utils.Debug.Printf("old chain length : %v", oldChnLen)
	utils.Debug.Printf("chain length incremented: %v", genNd.Mnr.ChnLen.Load())
	//test updated hash
	if genNd.Mnr.ChnLen.Load() == oldChnLen {
		t.Errorf("the chain length was not properly updated")
	}
}

