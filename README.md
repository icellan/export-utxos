# Export UTXOs

This Go package allows you to export UTXOs (Unspent Transaction Outputs) for given Bitcoin SV addresses using the WhatsOnChain API. It supports reading addresses from a file, command-line arguments, or standard input.

## Installation

To install the package, use the following command:

```sh
go get github.com/icellan/export-utxos
```

## Usage

### Command-Line Arguments

You can run the program with the following command-line arguments:

- \`-help\`: Show help message.
- \`-file <filename>\`: Specify a file containing Bitcoin SV addresses, one per line.
- \`-output <filename>\`: Specify a file to write the output JSON.

### Examples

#### Example 1: Reading Addresses from a File

Create a file named \`addresses.txt\` with the following content:

```
1LY2M3RCkEVKo82ym1SQ1iZGQhM5Lf5Pkf
1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa
```

Run the program with the file:

```sh
go run main.go -file addresses.txt -output output.json
```

#### Example 2: Reading Address from Command-Line Argument

Run the program with an address as a command-line argument:

```sh
go run main.go 1LY2M3RCkEVKo82ym1SQ1iZGQhM5Lf5Pkf
```

#### Example 3: Reading Addresses from Standard Input

Run the program without any arguments and paste the addresses followed by a blank line:

```sh
go run main.go
Please paste a list of addresses, one per line, followed by a blank line:
1LY2M3RCkEVKo82ym1SQ1iZGQhM5Lf5Pkf
1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa

```

## Output

The output will be a JSON file containing the UTXOs for the given addresses. Here is an example of the output:

```json
[
  {
    "address": "1LY2M3RCkEVKo82ym1SQ1iZGQhM5Lf5Pkf",
    "utxos": [
      {
        "tx_id": "137614ec60dba6aad2f37c469bb3f70d455964f10f565cb9a1874a85b5199466",
        "vout": 0,
        "previous_tx_script": "020a5314df44ccfa5b8e5c5c5b354f397d2590832c40e032099f442b12fca370",
        "satoshis": 312542128
      }
    ]
  }
]
```

## Code Overview

### Main Function

The \`main\` function handles command-line arguments, reads addresses from a file or standard input, and processes the addresses to fetch UTXOs.

### ProcessAddresses Function

The \`ProcessAddresses\` function takes a list of addresses, fetches their UTXOs using the \`fetchUtxosOfAddress\` function, and returns the result.

### fetchUtxosOfAddress Function

The \`fetchUtxosOfAddress\` function fetches UTXOs for a given address from the WhatsOnChain API and converts the response into the \`UtxoList\` model.

### fetchTransactionHex Function

The \`fetchTransactionHex\` function fetches the transaction hex for a given transaction ID from the WhatsOnChain API and caches the result.

## Dependencies

The package relies on the following dependencies:

- \`github.com/bitcoin-sv/go-sdk\`
- \`github.com/libsv/go-bt/v2\`
- \`github.com/libsv/go-bk\`
- \`github.com/pkg/errors\`
- \`golang.org/x/crypto\`

## License

This project is licensed under the MIT License. See the \`LICENSE\` file for details.