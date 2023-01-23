package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chia-network/go-chia-libs/pkg/ptr"
	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/schollz/progressbar/v3"
)

const (
	perPage int = 1000
)

func main() {
	client, err := rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig())
	if err != nil {
		log.Fatalln(err.Error())
	}

	transactionCount, _, err := client.WalletService.GetTransactionCount(
		&rpc.GetWalletTransactionCountOptions{
			WalletID: 1,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	totalTx := transactionCount.Count.OrEmpty()
	log.Printf("Exporting %d total transactions...\n", totalTx)

	file, err := os.Create("transactions.csv")
	if err != nil {
		log.Fatalln(err.Error())
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{"name", "height", "date", "type", "amount"})
	if err != nil {
		log.Fatalln(err.Error())
	}

	bar := progressbar.Default(int64(totalTx))
	txOpts := &rpc.GetWalletTransactionsOptions{
		WalletID: 1,
		Start:    ptr.IntPtr(0),
		End:      ptr.IntPtr(perPage),
	}

	for {
		transactions, _, err := client.WalletService.GetTransactions(txOpts)
		if err != nil {
			log.Fatal(err)
		}

		for _, transaction := range transactions.Transactions.OrEmpty() {
			createdTime := transaction.CreatedAtTime
			var inOrOut string
			if len(transaction.Removals) == 0 {
				inOrOut = "inbound"
			} else {
				inOrOut = "outbound"
			}
			err = writer.Write([]string{transaction.Name.String(), fmt.Sprintf("%d", transaction.ConfirmedAtHeight), createdTime.String(), inOrOut, fmt.Sprintf("%.12f", float64(transaction.Amount)/1000000000000)})
			if err != nil {
				log.Fatalln(err.Error())
			}
		}
		bar.Add(perPage)

		txOpts.Start = ptr.IntPtr(*txOpts.Start + perPage)
		txOpts.End = ptr.IntPtr(*txOpts.End + perPage)

		if *txOpts.End >= totalTx {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	log.Println("Done")
}
