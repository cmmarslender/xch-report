package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/chia-network/go-chia-libs/pkg/ptr"
	"github.com/chia-network/go-chia-libs/pkg/rpc"
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

	log.Printf("Exporting %d total transactions...\n", transactionCount.Count.OrEmpty())

	transactions, _, err := client.WalletService.GetTransactions(
		&rpc.GetWalletTransactionsOptions{
			WalletID: 1,
			End:      ptr.IntPtr(transactionCount.Count.OrEmpty()),
		},
	)

	if err != nil {
		log.Fatal(err)
	}

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

	log.Println("Done")
}
