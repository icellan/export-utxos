package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/bitcoin-sv/go-sdk/chainhash"
	"github.com/icellan/export-utxos/models"
	"github.com/libsv/go-bt/v2"
)

func main() {
	help := flag.Bool("help", false, "Show help")
	inputFile := flag.String("file", "", "File to read")
	outputFile := flag.String("output", "", "File to write output to")

	flag.Parse()

	if help != nil && *help {
		fmt.Println("Usage: main [-help] [-file <filename>] [<address>]")
		return
	}

	var output models.Output

	if inputFile != nil && *inputFile != "" {
		// file mode, read addresses from file and process
		f, err := os.Open(*inputFile)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}

		// read addresses from file
		addresses := make([]string, 0)
		for {
			var address string
			_, err = fmt.Fscanln(f, &address)
			if err != nil {
				break
			}

			addresses = append(addresses, address)
		}

		if len(addresses) == 0 {
			fmt.Println("No addresses found in file")
			return
		}

		// process the addresses
		output = ProcessAddresses(addresses)
	} else if flag.NArg() > 0 {
		// address mode
		address := flag.Arg(0)
		if address == "" {
			fmt.Println("No address given")
			return
		}

		output = ProcessAddresses([]string{address})
	} else {
		// no file and no address given, ask the user to paste a list of addresses
		fmt.Println("Please paste a list of addresses, one per line, followed by a blank line:")

		// read addresses from stdin
		addresses := make([]string, 0)
		for {
			var address string
			_, err := fmt.Scanln(&address)
			if err != nil {
				break
			}

			if address == "" {
				break
			}

			addresses = append(addresses, address)
		}

		if len(addresses) == 0 {
			fmt.Println("No addresses given")
			return
		}

		// process the addresses
		output = ProcessAddresses(addresses)
	}

	// marshal the output to JSON
	outputJSON, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	if outputFile != nil && *outputFile != "" {
		fmt.Printf("Writing output to %s\n", *outputFile)
		// write the output to a file if specified
		if err = os.WriteFile(*outputFile, outputJSON, 0644); err != nil {
			fmt.Println("Error writing output file:", err)
			return
		}

		fmt.Println("Output written successfully")
	} else {
		// print the output
		fmt.Println(string(outputJSON))
	}
}

func ProcessAddresses(addresses []string) models.Output {
	output := models.Output{}

	// do something with the address
	fmt.Printf("Processing %d addresses: %v\n", len(addresses), addresses)

	for idx, address := range addresses {
		addressUtxos, err := fetchUtxosOfAddress(address, idx, len(addresses))
		if err != nil {
			fmt.Println("Error fetching UTXOs:", err)
			return nil
		}

		// do something with the UTXOs
		output = append(output, addressUtxos)
	}

	fmt.Println("\nProcessing done")

	return output
}

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

var txCache = make(map[string]string)

func fetchUtxosOfAddress(address string, idx, numberOf int) (*models.UtxoList, error) {
	time.Sleep(350 * time.Millisecond) // overcome rate limit 3 RPS
	fmt.Printf("Fetching UTXOs for address %s (%d out of %d)\r", address, idx+1, numberOf)

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
		tx, err := bt.NewTxFromString(txHex)
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
