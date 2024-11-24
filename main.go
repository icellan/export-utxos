package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/icellan/export-utxos/models"
	"github.com/icellan/export-utxos/process"
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
		output = process.Addresses(addresses, func(idx int) {
			fmt.Printf("Fetching UTXOs for address %s (%d out of %d)\r", addresses[idx], idx+1, len(addresses))
		})
	} else if flag.NArg() > 0 {
		// address mode
		address := flag.Arg(0)
		if address == "" {
			fmt.Println("No address given")
			return
		}

		fmt.Printf("Processing 1 address: %s", address)
		output = process.Addresses([]string{address}, func(int) {
			return
		})
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
		output = process.Addresses(addresses, func(idx int) {
			fmt.Printf("Fetching UTXOs for address %s (%d out of %d)\r", addresses[idx], idx+1, len(addresses))
		})
	}

	fmt.Println("\nProcessing done, outputting...")

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
