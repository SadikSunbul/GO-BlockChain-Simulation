package main

import (
	"github.com/SadikSunbul/GO-BlockChain-Simulation/cli"
	"github.com/SadikSunbul/GO-BlockChain-Simulation/wallet"
	"os"
)

func main() {
	defer os.Exit(0)         // programı sonlandırır
	cli := cli.CommandLine{} // Komut satırı işlemleri için kullanılan yapıyı temsil eder.
	cli.Run()                // Komut satırı işlemlerini başlatır

	w := wallet.MakeWallet()
	w.Address()

	//fmt.Print(wallet.ValidateAddress("1Dg9RoRcMYtcyj3dhXooFyjJubERMXRRwC"))
}
