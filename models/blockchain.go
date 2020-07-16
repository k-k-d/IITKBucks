package models

import (
	"github.com/dryairship/IITKBucks/config"
)

type blockchain struct {
	Chain                    []Block
	UnusedTransactionOutputs OutputMap
	PendingTransactions      TransactionMap
	CurrentTarget            Hash
	CurrentBlockReward       Coins
	TransactionAdded         chan bool
	UserOutputs              map[User][]TransactionIdIndexPair
}

var blockchainInstance *blockchain

func Blockchain() *blockchain {
	if blockchainInstance == nil {
		target, err := HashFromHexString(config.INITIAL_TARGET)
		if err != nil {
			panic(err)
		}
		blockchainInstance = &blockchain{
			Chain:                    make([]Block, 0),
			UnusedTransactionOutputs: make(OutputMap),
			PendingTransactions:      make(TransactionMap),
			CurrentTarget:            target,
			CurrentBlockReward:       Coins(config.INITIAL_BLOCK_REWARD),
			TransactionAdded:         make(chan bool),
			UserOutputs:              make(map[User][]TransactionIdIndexPair),
		}
	}
	return blockchainInstance
}

func (blockchain *blockchain) AppendBlock(block Block) {
	block.SaveToFile()
	block.Transactions = nil
	blockchain.Chain = append(blockchain.Chain, block)
}

func (blockchain *blockchain) IsTransactionPending(transactionHash Hash) bool {
	_, exists := blockchain.PendingTransactions[transactionHash]
	return exists
}

func (blockchain *blockchain) IsTransactionValid(transaction *Transaction) (bool, Coins) {
	outputDataHash := transaction.CalculateOutputDataHash()
	sumOfInputs := Coins(0)

	var pair TransactionIdIndexPair
	for _, input := range transaction.Inputs {
		pair.TransactionId = input.TransactionId
		pair.Index = input.OutputIndex

		output, exists := blockchain.UnusedTransactionOutputs[pair]
		if !exists || !input.Signature.Unlock(&output, &pair, &outputDataHash) {
			return false, 0
		}

		sumOfInputs += output.Amount
	}

	sumOfOutputs := transaction.Outputs.GetSumOfAmounts()
	return sumOfOutputs <= sumOfInputs, sumOfInputs - sumOfOutputs
}

func (blockchain *blockchain) IsBlockValid(block *Block) bool {
	if block.Index != uint32(len(blockchain.Chain)+1) {
		return false
	}

	parentIndex := block.Index - 1

	if blockchain.Chain[parentIndex].Timestamp > block.Timestamp {
		return false
	}

	if blockchain.Chain[parentIndex].GetHash() != block.ParentHash {
		return false
	}

	if block.Target != blockchain.CurrentTarget {
		return false
	}

	if block.BodyHash != block.GetBodyHash() {
		return false
	}

	if !block.GetHash().IsLessThan(blockchain.CurrentTarget) {
		return false
	}

	var totalFee Coins
	for i := range block.Transactions {
		if i == 0 {
			continue
		}
		valid, txnFee := blockchain.IsTransactionValid(&block.Transactions[i])
		if !valid {
			return false
		}
		totalFee += txnFee
	}

	return totalFee >= block.Transactions[0].Outputs.GetSumOfAmounts()
}

func (blockchain *blockchain) AddTransaction(transaction Transaction) {
	blockchain.TransactionAdded <- true
	blockchain.PendingTransactions[transaction.CalculateHash()] = transaction
}

func (blockchain *blockchain) ProcessBlock(block Block) {
	for _, txn := range block.Transactions {
		delete(blockchain.PendingTransactions, txn.CalculateHash())

		for _, input := range txn.Inputs {
			txidIndexPair := TransactionIdIndexPair{
				TransactionId: input.TransactionId,
				Index:         input.OutputIndex,
			}
			delete(blockchain.UnusedTransactionOutputs, txidIndexPair)
		}

		txidIndexPair := TransactionIdIndexPair{
			TransactionId: txn.Id,
			Index:         0,
		}
		for i, output := range txn.Outputs {
			txidIndexPair.Index = uint32(i)
			blockchain.UnusedTransactionOutputs[txidIndexPair] = output
			if blockchain.UserOutputs[output.Recipient] == nil {
				blockchain.UserOutputs[output.Recipient] = make([]TransactionIdIndexPair, 0)
			}
			blockchain.UserOutputs[output.Recipient] = append(blockchain.UserOutputs[output.Recipient], txidIndexPair)
		}
	}
}
