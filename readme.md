# XCH Report

This uses transaction data from your wallet. Anything that doesn't show up in your wallet transaction list wont be captured by this tool. I suggest looking at [Chia Tracker](https://chiatracker.com) for a more complete look at your transaction history.

If your chia root is not the default, make sure its set in your env before running:

```bash
export CHIA_ROOT=<path to chia root>
```

Then, just run `go run main.go` and it will generate `transactions.csv` file for wallet ID 1
