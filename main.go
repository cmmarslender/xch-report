package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cmmarslender/go-chia-rpc/pkg/rpc"
	"github.com/cmmarslender/go-chia-rpc/pkg/util"
)

func main() {
	client, err := rpc.NewClient()
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

	log.Printf("Exporting %d total transactions...\n", transactionCount.Count)

	transactions, _, err := client.WalletService.GetTransactions(
		&rpc.GetWalletTransactionsOptions{
			WalletID: 1,
			End: util.IntPtr(transactionCount.Count),
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create("transactions.csv")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"name", "height", "date", "type", "amount"})

	for _, transaction := range transactions.Transactions {
		createdTime := time.Unix(int64(transaction.CreatedAtTime), 0)
		var inOrOut string
		if len(transaction.Removals) == 0 {
			inOrOut = "inbound"
		} else {
			inOrOut = "outbound"
		}
		writer.Write([]string{transaction.Name, fmt.Sprintf("%d", transaction.ConfirmedAtHeight), createdTime.String(), inOrOut, fmt.Sprintf("%.12f", float64(transaction.Amount)/1000000000000)})
	}

	log.Println("Done")
}
