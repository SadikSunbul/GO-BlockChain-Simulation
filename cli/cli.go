package cli

import (
	"flag"
	"fmt"
	"github.com/SadikSunbul/GO-BlockChain-Simulation/blockchain"
	"github.com/SadikSunbul/GO-BlockChain-Simulation/network"
	"github.com/SadikSunbul/GO-BlockChain-Simulation/wallet"
	"log"
	"os"
	"runtime"
	"strconv"
)

// CommandLine struct, komut satırı işlemleri için kullanılan yapıyı temsil eder.
type CommandLine struct {
	//blockchain *blockchain.Blockchain // blockchain adında bir Blockchain nesnesi
}

// PrintUsage fonksiyonu, komut satırında kullanıcıya kullanım talimatlarını gösterir.
func (cli *CommandLine) printUsage() {
	fmt.Printf("\033[35mUsage:\n\033[0m")
	fmt.Printf(" \033[35m%-40s : %s\n\033[0m", "getbalance -address ADDRESS", "Belirtilen adrese ait bakiyeyi görüntüler")
	fmt.Printf(" \033[35m%-40s : %s\n\033[0m", "createblockchain -address ADDRESS", "Yeni bir blok zinciri oluşturur ve belirtilen adrese oluşum ödülünü gönderir")
	fmt.Printf(" \033[35m%-40s : %s\n\033[0m", "printchain", "Blok zincirindeki tüm blokları yazdırır")
	fmt.Printf(" \033[35m%-40s : %s\n\033[0m", "send -from FROM -to TO -amount AMOUNT -mine", "Belirli bir miktarda coin gönder. Ardından -mine bayrağı ayarlanır, bu düğüm üzerinde madencilik yap")
	fmt.Printf(" \033[35m%-40s : %s\n\033[0m", "createwallet", "Yeni bir cüzdan oluşturur")
	fmt.Printf(" \033[35m%-40s : %s\n\033[0m", "listaddresses", "Cüzdan dosyamızdaki adresleri listeleyin")
	fmt.Printf(" \033[35m%-40s : %s\n\033[0m", "reindexutxo", "UTXO setini yeniden oluşturur")
	fmt.Printf(" \033[35m%-40s : %s\n\033[0m", "startnode -miner ADDRESS", "NODE_ID ortamında belirtilen kimliğe sahip bir düğüm başlatın. var. -miner madenciliği mümkün kılar")

}

// reindexUTXO fonksiyonu, UTXO setini yeniden oluşturur.
func (cli *CommandLine) reindexUTXO(nodeID string) {
	chain := blockchain.ContinueBlockChain(nodeID) // blockchain adında bir Blockchain nesnesi
	defer chain.Database.Close()                   // blok zincirini kapat
	UTXOSet := blockchain.UTXOSet{chain}           // UTXO setini oluştur
	UTXOSet.Reindex()                              // UTXO setini yeniden oluştur

	count := UTXOSet.CountTransactions()                            // UTXO setindeki işlemleri sayar
	fmt.Printf("Tamamlamak! UTXO kümesinde %d işlem var.\n", count) // UTXO setindeki işlemlerin sayısını ekrana yazdırır
}

func (cli *CommandLine) StartNode(nodeID, minerAddress string) {
	fmt.Printf("Başlangıç Düğümü\n %s\n", nodeID)

	if len(minerAddress) > 0 {
		if wallet.ValidateAddress(minerAddress) {
			fmt.Println("Madencilik açık. Ödülleri alacağınız adres: ", minerAddress)
		} else {
			log.Panic("Yanlış madenci adresi!")
		}
	}
	network.StartServer(nodeID, minerAddress)
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

// printChain fonksiyonu, blok zincirindeki tüm blokları yazdırır
func (cli *CommandLine) printChain(nodeID string) {
	chain := blockchain.ContinueBlockChain(nodeID) // blockchain adında bir Blockchain nesnesi
	defer chain.Database.Close()                   // blok zincirini kapat
	iter := chain.Iterator()                       // blok zinciri iteratorunu oluştur
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

// createBlockChain fonksiyonu, belirtilen adresin blok zincirini oluşturur
func (cli *CommandLine) createBlockChain(address, nodeID string) { // blockchain oluşturur
	if !wallet.ValidateAddress(address) { // adresin dogrulugunu kontrol eder
		log.Panic("\033[31mAddress is not Valid\033[0m")
	}
	chain := blockchain.InitBlockChain(address, nodeID) // adresin blok zincirini oluşturur
	defer chain.Database.Close()                        // blok zincirini kapat

	UTXOSet := blockchain.UTXOSet{chain} // adresin UTXO setini oluşturur
	UTXOSet.Reindex()                    // adresin UTXO setini yeniden oluşturur

	fmt.Println("\u001B[32mFinished!\u001B[0m") // sonlandırılır
}

// getBalance fonksiyonu, belirtilen adresin bakiyesini bulur
func (cli *CommandLine) getBalance(address, nodeID string) {
	if !wallet.ValidateAddress(address) { // adresin dogrulugunu kontrol eder
		log.Panic("\033[31mAddress is not Valid\033[0m")
	}
	chain := blockchain.ContinueBlockChain(nodeID) // adresin blok zincirini okur
	UTXOSet := blockchain.UTXOSet{chain}           // adresin UTXO setini oluşturur
	defer chain.Database.Close()                   // blok zincirini kapat

	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))   // adresin base58 kodunu okur
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]       // adresin ilk 4 karakterini kaldırır
	UTXOs := UTXOSet.FindUnspentTransactions(pubKeyHash) // adresin bakiyesini bulur

	for _, out := range UTXOs { // bakiye döngüsü
		balance += out.Value // bakiyeyi arttırır
	}

	fmt.Printf("\033[32mBalance of %s: %d\u001B[0m\n", address, balance) // bakiye yazdırılır
}

// send fonksiyonu, belirtilen miktarı belirtilen adresten diğer bir adrese gönderir.
func (cli *CommandLine) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !wallet.ValidateAddress(to) {
		log.Panic("Address is not Valid")
	}
	if !wallet.ValidateAddress(from) {
		log.Panic("Address is not Valid")
	}
	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOSet := blockchain.UTXOSet{chain}
	defer chain.Database.Close()

	wallets, err := wallet.CreateWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := blockchain.NewTransaction(&wallet, to, amount, &UTXOSet)
	if mineNow {
		cbTx := blockchain.CoinbaseTx(from, "")
		txs := []*blockchain.Transaction{cbTx, tx}
		block := chain.MineBlock(txs)
		UTXOSet.Update(block)
	} else {
		network.SendTx(network.KnownNodes[0], tx)
		fmt.Println("send tx")
	}

	fmt.Println("Success!")
}

// listAddresses fonksiyonu, cüzdan adreslerini listeler.
func (cli *CommandLine) listAddresses(nodeID string) {
	wallets, _ := wallet.CreateWallets(nodeID) // cüzdan dosyasını okur
	addresses := wallets.GetAllAddress()       // cüzdan adreslerini alır
	for _, address := range addresses {
		fmt.Printf("\033[36m	%s\u001B[0m\n", address)
	}
}

// createWallet fonksiyonu, cüzdan oluşturur.
func (cli *CommandLine) createWallet(nodeID string) {
	wallets, _ := wallet.CreateWallets(nodeID) // cüzdan dosyasını okur
	address := wallets.AddWallet()             // cüzdan adresini oluşturur
	wallets.SaveFile(nodeID)                   // dosyayı kaydeder
	fmt.Printf("\u001B[32mNew address is : %s\u001B[0m\n", address)
}

func (cli *CommandLine) Run() { // komut satırı işlemleri
	cli.validateArgs() // komut satırı argümanlarını dogrular

	nodeID := os.Getenv("NODE_ID") // Set-Item -Path Env:NODE_ID -Value "3000" | set NODE_ID=3000
	/*
			Set-Item -Path Env:NODE_ID -Value "3000"
			Set-Item -Path Env:NODE_ID -Value "4000"
			Set-Item -Path Env:NODE_ID -Value "5000"

			xcopy blocks_3000 blocks_5000
		    xcopy blocks_3000 blocks_4000
		    xcopy blocks_3000 blocks_gen

	*/
	if nodeID == "" {
		fmt.Printf("NODE_ID env ayarlanmadı!")
		runtime.Goexit()
	}

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)             // getbalance komutunu tanımla
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError) // createblockchain komutunu tanımla
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)                         // send komutunu tanımla
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)             // printchain komutunu tanımla
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "\033[36mBakiye almanın adresi\033[0m")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "\033[36mGenesis blok ödülünün gönderileceği adres\033[0m")
	sendFrom := sendCmd.String("from", "", "\033[36mKaynak cüzdan adresi\033[0m")
	sendTo := sendCmd.String("to", "", "\033[36mHedef cüzdan adresi\033[0m")
	sendAmount := sendCmd.Int("amount", 0, "\033[36mGönderilecek tutar\033[0m")
	sendMine := sendCmd.Bool("mine", false, "Aynı düğümde hemen madencilik yapın")
	startNodeMiner := startNodeCmd.String("miner", "", "Madencilik modunu etkinleştirin ve ödülü ADDRESS adresine gönderin")

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
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage() // komut satırı argümanlarını yazdır
		runtime.Goexit() // programın çalışmasını sonlandır
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress, nodeID)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress, nodeID)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}
	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeID)
	}
	if reindexUTXOCmd.Parsed() {
		cli.reindexUTXO(nodeID)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
	}

	if startNodeCmd.Parsed() {
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			startNodeCmd.Usage()
			runtime.Goexit()
		}
		cli.StartNode(nodeID, *startNodeMiner)
	}
}
