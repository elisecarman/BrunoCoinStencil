package wallet

import (
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/blockchain"
	"BrunoCoin/pkg/id"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"encoding/hex"
	"fmt"
	"sync"
)

/*
 *  Brown University, CS1951L, Summer 2021
 *  Designed by: Colby Anderson, Kotone Ninagawa
 */

// TxReq (TransactionRequest) that represents
// the minimum amount of information needed
// to make a transaction.
// PubK (PublicKey) represents the serialized
// public key of the person they want to pay.
// Amt (Amount) represents the amount of money
// they want to pay the person.
type TxReq struct {
	PubK []byte
	Amt  uint32
	Fee  uint32
}

// Wallet provides the functionality to make
// transactions from transaction requests and
// send them to the node to be broadcast on
// the network.
// Conf represents the configuration for the
// wallet.
// Id represents the identity of the person
// using the wallet.
// Chain represents the blockchain, as the
// wallet needs to be able to query the chain
// for enough UTXO to fulfill a transaction request.
// SendTx (SendTransaction) is a channel for sending
// fulfilled transaction requests (now in the form of
// a transaction) to the node, in order to be sent
// across the network.
// LmnlTxs (LiminalTransactions) represent the
// transactions that the wallet has made, but that
// do not have enough proof of work on top of them
// to be considered valid by everyone.
// Mut (Mutex) is a mutex for concurrent accesses
// to non-atomic reads/writes for the struct
type Wallet struct {
	Conf    *Config
	Id      id.ID
	Chain   *blockchain.Blockchain
	SendTx  chan *tx.Transaction
	LmnlTxs *LiminalTxs
	Addr    string

	mutex sync.Mutex
}


// SetAddr (SetAddress) sets the address
// of the node in the wallet.
func (w *Wallet) SetAddr(a string) {
	w.mutex.Lock()
	w.Addr = a
	w.mutex.Unlock()
}


// New creates a wallet object.
// Inputs:
// c *Config the configuration
// for the wallet
// id id.ID the id of the node
// chain *blockchain.Blockchain the
// blockchain that the wallet needs a
// references to find UTXO for making transactions.
// Returns:
// *Wallet the new wallet object
func New(c *Config, id id.ID, chain *blockchain.Blockchain) *Wallet {
	if !c.HasWt {
		return nil
	}
	return &Wallet{
		Conf:    c,
		Id:      id,
		Chain:   chain,
		SendTx:  make(chan *tx.Transaction),
		LmnlTxs: NewLmnlTxs(c),
	}
}

// HndlBlk (HandleBlock) is called after a new
// block is added to the main chain. However, the
// inputted block is a "safe block amount" down from
// the top block of the main chain. If the wallet
// has still not seen some of its transactions added
// to the main chain this far down, then it may have
// to resend the transactions out.
// Inputs:
// b *block.Block the block that is "safe block amount"
// down from the last block on the main chain
// TODO
// 1. Make sure to update liminal transactions properly
// be incrementing the priorities, removing duplicates,
// and returning which ones are "old"
// 2. Resend out "old" liminal transactions with the
// lock time incremented (in order to give the transaction
// a different hash since lock time essentially does nothing)
// Tip 1: A formatted debugging message saying which
// address created the transaction and which transaction
// may be helpful
// Tip 2: Remember to do error checking on different variables
// that could be nil

// some helpful functions/methods/fields:
// let t be a transaction object
// w.LmnlTxs.ChkTxs(...)
// w.LmnlTxs.Add(...)
// utils.FmtAddr(...)
// t.NameTag()
// w.SendTx <- ...
func (w *Wallet) HndlBlk(b *block.Block) {
	if b == nil {
		fmt.Printf("ERROR {tp.HndlBlk}: "+
			"nil block was given to the function")
		return
	}
	w.mutex.Lock()
	defer w.mutex.Unlock()
	old, _ := w.LmnlTxs.ChkTxs(b.Transactions)
	for _,o := range old{
		o.LockTime += 1
		w.SendTx <- o
	}
}

// HndlTxReq (HandleTransactionRequest) attempts to
// create a transaction from the request, as well as
// sending this transaction to the node to be forwarded
// on the network. It generates the transaction by first
// asking the blockchain for enough UTXO to construct the
// transaction. At this point, the transaction is made, but
// not valid by the consensus since it is not mined onto the
// main chain and have enough POW on top of it. Therefore,
// we must add it to our liminal transactions (transactions that
// have been made/broadcast but not validated).
// Inputs:
// txR *TxReq a transaction request from the node
// TODO
// 1. Try and find enough UTXO to make the transaction
// 2. If not enough, return
// 3. Make the transaction inputs for the transaction
// from the UTXO
// 4. Make the transaction outputs based on who you
// send money to and if there is change leftover for
// yourself
// 5. Add the transaction to liminal transactions
// 6. Send the transaction to the node to be broadcast
// Tip 1: A formatted debugging message saying which
// address created the transaction and which transaction
// may be helpful
// Tip 2: Remember to do error checking on different variables
// that could be nil
// Tip 3: proto.Transaction is used in the networking code.
// tx.Transaction is used elsewhere. To make a tx.Transaction,
// first make a proto.Transaction and then deserialize it to a
// tx.Transaction
// Tip 4: Remember to account for fees properly!

// some helpful functions/methods/fields:
// let t be a transaction object
// w.Id.GetPublicKeyBytes()
// hex.EncodeToString(...)
// w.Chain.GetUTXOForAmt(...)
// proto.NewTx(...)
// tx.Deserialize(...)
// w.LmnlTxs.Add(...)
// w.SendTx <- ...
// utils.FmtAddr(...)
// t.NameTag()
// t.UTXO.MkSig(...)
// proto.NewTxInpt(...)
// proto.NewTxOutpt(...)
func (w *Wallet) HndlTxReq(txR *TxReq) {
	if txR == nil {
		fmt.Printf("ERROR {tp.HndlTxReq}: " +
			"nil TxReq was given to the function")
		return
	}
	if txR.Amt == 0 {
		fmt.Printf("ERROR {tp.HndlTxReq}: " +
			"incorrect requested amount")
		return
	}
	w.mutex.Lock()
	senderPubK := hex.EncodeToString(w.Id.GetPublicKeyBytes())
	info, change, bool := w.Chain.GetUTXOForAmt(txR.Amt + txR.Fee, senderPubK)
	//if w.Chain.Length() > 1 {  //check on this!
	if !bool {
		return
	} //}
	utils.Debug.Printf("the correct amount was found within wallet %v", utils.FmtAddr(w.Addr))
	var inputs []*proto.TransactionInput
	for _, v := range info{
		unlockScr, err := v.UTXO.MkSig(w.Id)  //What to do with err?
		if err != nil {
			// error message here
		}
		newInput := proto.NewTxInpt(v.TxHsh, v.OutIdx, unlockScr, v.Amt)
		inputs = append(inputs, newInput)
	}
	mainOutput := proto.NewTxOutpt(txR.Amt, hex.EncodeToString(txR.PubK))
	outputs := []*proto.TransactionOutput{mainOutput}
	if change != 0 {
		outputToSelf := proto.NewTxOutpt(change, senderPubK)
		outputs = append(outputs, outputToSelf)
	}

	newTrx := proto.NewTx(w.Conf.TxVer, inputs, outputs, w.Conf.DefLckTm)
	utils.Debug.Printf("w.HndlTrxReq: the Sum input of the first transaction %v of mining pool is %v",
		tx.Deserialize(newTrx).NameTag(),tx.Deserialize(newTrx).SumInputs())
	utils.Debug.Printf("w.HndlTrxReq: the Sum output of the first transaction %v of mining pool is %v",
		tx.Deserialize(newTrx).NameTag(),tx.Deserialize(newTrx).SumOutputs())
	utils.Debug.Printf("the fee of the transaction is %v and the amount is %v. We are sending to %v",
		txR.Fee, txR.Amt, utils.FmtAddr(w.Addr))

	//How many field of fields can we call? D:

	w.LmnlTxs.Add(tx.Deserialize(newTrx))
	w.SendTx <- tx.Deserialize(newTrx)
	utils.Debug.Printf("%v Transaction sent by wallet %v", utils.FmtAddr(w.Addr), tx.Deserialize(newTrx).NameTag())
	w.mutex.Unlock()
}
