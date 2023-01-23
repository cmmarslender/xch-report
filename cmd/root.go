package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chia-network/go-chia-libs/pkg/ptr"
	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "xch-report",
	Short: "Exports a csv of transactions from the official chia wallet",
	Run: func(cmd *cobra.Command, args []string) {
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
			End:      ptr.IntPtr(viper.GetInt("per-page")),
		}

		for {
			transactions, _, err := client.WalletService.GetTransactions(txOpts)
			if err != nil {
				log.Fatal(err)
			}

			for _, transaction := range transactions.Transactions.OrEmpty() {
				if transaction.Amount <= viper.GetUint64("dust-amount") {
					continue
				}
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
			err = bar.Add(viper.GetInt("per-page"))
			cobra.CheckErr(err)

			txOpts.Start = ptr.IntPtr(*txOpts.Start + viper.GetInt("per-page"))
			txOpts.End = ptr.IntPtr(*txOpts.End + viper.GetInt("per-page"))

			if *txOpts.End >= totalTx {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		err = bar.Finish()
		cobra.CheckErr(err)

		log.Println("Done")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	var (
		perPage    int
		dustAmount uint64
	)

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.xch-report.yaml)")
	rootCmd.PersistentFlags().IntVar(&perPage, "per-page", 1000, "Number of results to process per RPC call")
	rootCmd.PersistentFlags().Uint64Var(&dustAmount, "dust-amount", 0, "Amount in mojos to be considered dust, and excluded from the CSV")

	err := viper.BindPFlag("per-page", rootCmd.PersistentFlags().Lookup("per-page"))
	cobra.CheckErr(err)

	err = viper.BindPFlag("dust-amount", rootCmd.PersistentFlags().Lookup("dust-amount"))
	cobra.CheckErr(err)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".xch-report" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".xch-report")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
