package models

import "github.com/bitcoin-sv/go-sdk/chainhash"

type Output []*UtxoList

type UtxoList struct {
	Address string  `json:"address"`
	Utxos   []Utxos `json:"utxos"`
}

type Utxos struct {
	TxID             chainhash.Hash `json:"tx_id"`
	Vout             uint32         `json:"vout"`
	PreviousTxScript string         `json:"previous_tx_script"`
	Satoshis         uint64         `json:"satoshis"`
}

type WoCResult struct {
	Address string     `json:"address"`
	Script  string     `json:"script"`
	Result  []WoCUtxos `json:"result"`
}

type WoCUtxos struct {
	Height             uint32 `json:"height"`
	TxPos              uint32 `json:"tx_pos"`
	TxHash             string `json:"tx_hash"`
	Value              uint64 `json:"value"`
	IsSpentInMempoolTx bool   `json:"isSpentInMempoolTx"`
	Status             string `json:"status"`
}
