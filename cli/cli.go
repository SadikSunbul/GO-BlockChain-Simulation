package cli

import (
	"flag"
	"fmt"
	"github.com/SadikSunbul/GO-BlockChain-Simulation/blockchain"
	"github.com/SadikSunbul/GO-BlockChain-Simulation/wallet"
	"log"
	"os"
	"runtime"
	"strconv"
)

// CommandLine struct, komut satırı işlemleri için kullanılan yapıyı temsil eder.
type CommandLine struct {
	//blockchain *blockchain.BlockChain // blockchain adında bir BlockChain nesnesi
}

// PrintUsage fonksiyonu, komut satırında kullanıcıya kullanım talimatlarını gösterir.
func (cli *CommandLine) printUsage() {
	fmt.Printf("\033[35mUsage:\n\033[0m")
	fmt.Printf(" \033[35m%-40s : %s\n\033[0m", "getbalance -address ADDRESS", "Belirtilen adrese ait bakiyeyi görüntüler")
	fmt.Printf(" \033[35m%-40s : %s\n\033[0m", "createblockchain -address ADDRESS", "Yeni bir blok zinciri oluşturur ve belirtilen adrese oluşum ödülünü gönderir")
	fmt.Printf(" \033[35m%-40s : %s\n\033[0m", "printchain", "Blok zincirindeki tüm blokları yazdırır")
	fmt.Printf(" \033[35m%-40s : %s\n\033[0m", "send -from FROM -to TO -amount AMOUNT", "Belirtilen miktarı belirtilen adresten diğer bir adrese gönderir")
	fmt.Printf(" \033[35m%-40s : %s\n\033[0m", "createwallet", "Yeni bir cüzdan oluşturur")
	fmt.Printf(" \033[35m%-40s : %s\n\033[0m", "listaddresses", "Cüzdan dosyamızdaki adresleri listeleyin\n")

}

// validateArgs fonksiyonu, komut satırı argümanlarını doğrular.
func (cli *CommandLine) validateArgs() {
	// Eğer komut satırında argüman sayısı 2'den az ise (program adı dahil)
	if len(os.Args) < 2 {
		// Kullanım talimatlarını yazdır
		cli.printUsage()
		// Programın çalışmasını sonlandır ve kapat
		runtime.Goexit()
	}
}

func (cli *CommandLine) printChain() {
	chain := blockchain.ContinueBlockChain("") // blockchain adında bir BlockChain nesnesi
	defer chain.Database.Close()               // blok zincirini kapat
	iter := chain.Iterator()                   // blok zinciri iteratorunu oluştur
	fmt.Println()

	for { // blok zinciri sonuna kadar döngü
		block := iter.Next() // Sıradaki bloğu al
		fmt.Println("\033[97m╔══════════════════════════════════════════ BLOCK ═════════════════════════════════════════╗")
		fmt.Printf("║ \033[32m%-10s : %x\033[0m\n", "Hash", block.Hash)
		fmt.Printf("║ \033[32m%-10s : %x\033[0m\n", "Prev. hash", block.PrevHash)
		pow := blockchain.NewProof(block)
		fmt.Printf("║ \033[32m%-10s : %v\033[0m\n", "PoW", strconv.FormatBool(pow.Validate()))
		// Blok zincirinden o bloğun proof of work değerini yazdır
		for _, tx := range block.Transactions {
			fmt.Println("║", tx)
		}
		fmt.Println("\u001B[97m╚═════════════════════════════════════════════════════════════════════════════════════════╝")

		if len(block.PrevHash) == 0 {
			break
		}

	}
}

func (cli *CommandLine) createBlockChain(address string) { // blockchain oluşturur
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not Valid")
	}
	chain := blockchain.InitBlockChain(address)
	chain.Database.Close()
	fmt.Println("Finished!")
}

func (cli *CommandLine) getBalance(address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not Valid")
	}
	chain := blockchain.ContinueBlockChain(address)
	defer chain.Database.Close()

	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := chain.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int) {
	if !wallet.ValidateAddress(to) {
		log.Panic("Address is not Valid")
	}
	if !wallet.ValidateAddress(from) {
		log.Panic("Address is not Valid")
	}
	chain := blockchain.ContinueBlockChain(from)
	defer chain.Database.Close()

	tx := blockchain.NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Println("Success!")
}

func (cli *CommandLine) listAddresses() {
	wallets, _ := wallet.CreateWallets()
	addresses := wallets.GetAllAddress()
	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) CreateWallet() {
	wallets, _ := wallet.CreateWallets()
	address := wallets.AddWallet()
	wallets.SaveFile()
	fmt.Printf("New address is : %s\n", address)
}

func (cli *CommandLine) Run() { // komut satırı işlemleri
	cli.validateArgs() // komut satırı argümanlarını dogrular

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)             // getbalance komutunu tanımla
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError) // createblockchain komutunu tanımla
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)                         // send komutunu tanımla
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)             // printchain komutunu tanımla
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "\033[36mBakiye almanın adresi\033[0m")
	createBlockchainAddress := createBlockchainCmd.String("address[0m", "", "\033[36mGenesis blok ödülünün gönderileceği adres\033[0m")
	sendFrom := sendCmd.String("from", "", "\033[36mKaynak cüzdan adresi\033[0m")
	sendTo := sendCmd.String("to", "", "\033[36mHedef cüzdan adresi\033[0m")
	sendAmount := sendCmd.Int("amount", 0, "\033[36mGönderilecek tutar\033[0m")
	// send komutundaki tutarı tanımla

	switch os.Args[1] { // komut satırı argümanın hangi komut oldugunu bulur
	case "getbalance": // getbalance komutunu çalıştır
		err := getBalanceCmd.Parse(os.Args[2:]) // getbalance komutunu çalıştır
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:]) // createblockchain komutunu çalıştır
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:]) // printchain komutunu çalıştır
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:]) // send komutunu çalıştır
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage() // komut satırı argümanlarını yazdır
		runtime.Goexit() // programın çalışmasını sonlandır
	}

	if getBalanceCmd.Parsed() { // getbalance komutu parse edilirse
		if *getBalanceAddress == "" { // getbalance komutundaki adres bilgisi bos ise
			getBalanceCmd.Usage() // getbalance komutunu yazdır
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress) // getbalance komutunu çalıştır
	}

	if createBlockchainCmd.Parsed() { // createblockchain komutu parse edilirse
		if *createBlockchainAddress == "" { // createblockchain komutundaki adres bilgisi bos ise
			createBlockchainCmd.Usage() // createblockchain komutunu yazdır
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress) // createblockchain komutunu çalıştır
	}

	if printChainCmd.Parsed() { // printchain komutu parse edilirse
		cli.printChain() // printchain komutunu çalıştır
	}

	if sendCmd.Parsed() { // send komutu parse edilirse
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 { // send komutundaki kaynak, hedef ve tutar bilgileri bos ise
			sendCmd.Usage() // send komutunu yazdır
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount) // send komutunu çalıştır
	}

	if createWalletCmd.Parsed() {
		cli.CreateWallet()
	}
	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}
}
