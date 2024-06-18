package main

import (
	"flag"
	"fmt"
	"github.com/SadikSunbul/GO-BlockChain-Simulation/blockchain"
	"log"
	"os"
	"runtime"
	"strconv"
)

func main() {
	defer os.Exit(0)     // programı sonlandırır
	cli := CommandLine{} // Komut satırı işlemleri için kullanılan yapıyı temsil eder.
	cli.run()            // Komut satırı işlemlerini başlatır
}

// CommandLine struct, komut satırı işlemleri için kullanılan yapıyı temsil eder.
type CommandLine struct {
	//blockchain *blockchain.BlockChain // blockchain adında bir BlockChain nesnesi
}

// PrintUsage fonksiyonu, komut satırında kullanıcıya kullanım talimatlarını gösterir.
func (cli *CommandLine) printUsage() {
	fmt.Printf("Usage:\n")
	fmt.Printf(" %-40s : %s\n", "getbalance -address ADDRESS", "Belirtilen adrese ait bakiyeyi görüntüler")
	fmt.Printf(" %-40s : %s\n", "createblockchain -address ADDRESS", "Yeni bir blok zinciri oluşturur ve belirtilen adrese oluşum ödülünü gönderir")
	fmt.Printf(" %-40s : %s\n", "printchain", "Blok zincirindeki tüm blokları yazdırır")
	fmt.Printf(" %-40s : %s\n", "send -from FROM -to TO -amount AMOUNT", "Belirtilen miktarı belirtilen adresten diğer bir adrese gönderir")

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
	iter := chain.Iterator()                   // blok zinciri iteratorunu olustur

	for { // blok zinciri sonuna kadar döngü
		block := iter.Next() // Sıradaki bloğu al

		fmt.Printf("Prev. hash: %x\n", block.PrevHash)              // Blok zincirinden o bloğun önceki hash degerini yazdır
		fmt.Printf("Hash: %x\n", block.Hash)                        // Blok zincirinden o bloğun hash degerini yazdır
		pow := blockchain.NewProof(block)                           // Blok zincirinden o bloğun proof of work degerini al
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate())) // Blok zincirinden o bloğun proof of work degerini yazdır
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createBlockChain(address string) { // blockchain oluşturur
	chain := blockchain.InitBlockChain(address) // blockchain adında bir BlockChain nesnesi
	chain.Database.Close()                      // blok zincirini kapat
	fmt.Println("Finished!")
}

func (cli *CommandLine) getBalance(address string) { // bakiye almak
	chain := blockchain.ContinueBlockChain(address) // blockchain adında bir BlockChain nesnesi
	defer chain.Database.Close()                    // blok zincirini kapat

	balance := 0
	UTXOs := chain.FindUTXO(address) // blok zincirinden o bloğun UTXO degerlerini al

	for _, out := range UTXOs { // blok zincirinden o bloğun UTXO degerlerini döngürecek
		balance += out.Value // blok zincirinden o bloğun UTXO degerlerinin toplamını al
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int) { // para göndermek
	chain := blockchain.ContinueBlockChain(from) // blockchain adında bir BlockChain nesnesi
	defer chain.Database.Close()                 // blok zincirini kapat

	tx := blockchain.NewTransaction(from, to, amount, chain) // Yeni bir işlem oluştur
	chain.AddBlock([]*blockchain.Transaction{tx})            // blok zincirine ekler
	fmt.Println("Success!")
}

func (cli *CommandLine) run() { // komut satırı işlemleri
	cli.validateArgs() // komut satırı argümanlarını dogrular

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)             // getbalance komutunu tanımla
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError) // createblockchain komutunu tanımla
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)                         // send komutunu tanımla
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)             // printchain komutunu tanımla

	getBalanceAddress := getBalanceCmd.String("address", "", "Bakiye almanın adresi")                                 // getbalance komutundaki adres bilgisini tanımla
	createBlockchainAddress := createBlockchainCmd.String("address", "", "Genesis blok ödülünün gönderileceği adres") // createblockchain komutundaki adres bilgisini tanımla
	sendFrom := sendCmd.String("from", "", "Kaynak cüzdan adresi")                                                    // send komutundaki kaynak adresini tanımla
	sendTo := sendCmd.String("to", "", "Hedef cüzdan adresi")                                                         // send komutundaki hedef adresini tanımla
	sendAmount := sendCmd.Int("amount", 0, "Gönderilecek tutar")                                                      // send komutundaki tutarı tanımla

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
}
