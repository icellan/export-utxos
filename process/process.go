package process

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bitcoin-sv/go-sdk/chainhash"
	"github.com/bitcoin-sv/go-sdk/script"
	"github.com/bitcoin-sv/go-sdk/transaction"
	"github.com/icellan/export-utxos/models"
)

func Addresses(addresses []string, progressFunc func(int)) (models.Output, error) {
	output := models.Output{}

	// first check whether all the addresses are valid
	for _, address := range addresses {
		_, err := script.NewAddressFromString(address)
		if err != nil {
			return nil, fmt.Errorf("invalid address: %s", address)
		}
	}

	for idx, address := range addresses {
		progressFunc(idx)
		addressUtxos, err := fetchUtxosOfAddress(address)
		if err != nil {
			return nil, fmt.Errorf("error fetching UTXOs: %v", err)
		}

		// do something with the UTXOs
		output = append(output, addressUtxos)
	}

	return output, nil
}

var txCache = make(map[string]string)

/*
	WhatsOnChain API response:
	{
		"address":"1LY2M3RCkEVKo82ym1SQ1iZGQhM5Lf5Pkf",
		"script":"020a5314df44ccfa5b8e5c5c5b354f397d2590832c40e032099f442b12fca370",
		"result":[
			{
				"height":863675,
				"tx_pos":0,
				"tx_hash":"137614ec60dba6aad2f37c469bb3f70d455964f10f565cb9a1874a85b5199466",
				"value":312542128,
				"isSpentInMempoolTx":false,
				"status":"confirmed"
			},
			.........
		]
	}
*/

func fetchUtxosOfAddress(address string) (*models.UtxoList, error) {
	time.Sleep(350 * time.Millisecond) // overcome rate limit 3 RPS

	// GET https://api.whatsonchain.com/v1/bsv/<network>/address/<address>/unspent/all
	resp, err := http.Get(fmt.Sprintf("https://api.whatsonchain.com/v1/bsv/main/address/%s/unspent/all", address))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// unmarshal the response into a WoCResult
	var wocResult models.WoCResult
	err = json.Unmarshal(body, &wocResult)
	if err != nil {
		return nil, err
	}

	// convert the WoCUtxos into utxos
	utxos := make([]models.Utxos, 0, len(wocResult.Result))
	for _, wocUtxo := range wocResult.Result {
		hash, err := chainhash.NewHashFromHex(wocUtxo.TxHash)
		if err != nil {
			return nil, err
		}

		// get the transaction hex
		txHex, err := fetchTransactionHex(wocUtxo.TxHash)
		if err != nil {
			return nil, err
		}

		// decode the transaction
		tx, err := transaction.NewTransactionFromHex(txHex)
		if err != nil {
			return nil, err
		}

		utxos = append(utxos, models.Utxos{
			TxID:             *hash,
			Vout:             wocUtxo.TxPos,
			PreviousTxScript: tx.Outputs[wocUtxo.TxPos].LockingScript.String(),
			Satoshis:         wocUtxo.Value,
		})
	}

	return &models.UtxoList{
		Address: address,
		Utxos:   utxos,
	}, nil
}

func fetchTransactionHex(txIDHex string) (string, error) {
	time.Sleep(350 * time.Millisecond) // overcome rate limit 3 RPS

	// check the cache
	if txHex, ok := txCache[txIDHex]; ok {
		return txHex, nil
	}

	resp, err := http.Get(fmt.Sprintf("https://api.whatsonchain.com/v1/bsv/main/tx/%s/hex", txIDHex))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// add to cache
	txCache[txIDHex] = string(body)

	return txCache[txIDHex], nil
}
