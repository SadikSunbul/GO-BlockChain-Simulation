package main

import (
	"flag"
	"fmt"
	"github.com/SadikSunbul/GO-BlockChain-Simulation/blockchain"
	"os"
	"runtime"
	"strconv"
)

func main() {
	defer os.Exit(0)
	chain := blockchain.InitBlockChain()

	defer chain.Database.Close()

	cli := CommandLine{chain}
	cli.run()
}

// CommandLine struct, komut satırı işlemleri için kullanılan yapıyı temsil eder.
type CommandLine struct {
	blockchain *blockchain.BlockChain // blockchain adında bir BlockChain nesnesi
}

// PrintUsage fonksiyonu, komut satırında kullanıcıya kullanım talimatlarını gösterir.
func (cli *CommandLine) PrintUsage() {
	fmt.Println("Kullanım:")                                         // Kullanım talimatlarını gösterir
	fmt.Println(" add -block BLOCK_DATA - zincire bir blok ekleyin") // 'add' komutu ile yeni bir blok eklemek için kullanım açıklaması
	fmt.Println(" print - Zincirdeki blokları yazdırır")             // 'print' komutu ile zincirdeki blokları yazdırmak için kullanım açıklaması
}

// validateArgs fonksiyonu, komut satırı argümanlarını doğrular.
func (cli *CommandLine) validateArgs() {
	// Eğer komut satırında argüman sayısı 2'den az ise (program adı dahil)
	if len(os.Args) < 2 {
		// Kullanım talimatlarını yazdır
		cli.PrintUsage()
		// Programın çalışmasını sonlandır ve kapat
		runtime.Goexit()
	}
}

// addBlock metod, yeni bir blok eklemek için kullanılır.
func (cli *CommandLine) addBlock(data string) {
	// cli.blockchain.AddBlock(data) çağrısı, CommandLine struct'ına bağlı blockchain nesnesi üzerinden
	// data parametresi ile belirtilen veriyi içeren yeni bir blok ekler.
	cli.blockchain.AddBlock(data)

	fmt.Println("Added Block!")
}

// printChain metod, blok zincirini ekrana yazdırmak için kullanılır.
func (cli *CommandLine) printChain() {
	iter := cli.blockchain.Iterator()

	// Sonsuz döngü ile blok zincirindeki her bloğu yazdırmaya başlar.
	for {
		// iter.Next() çağrısı, yineleyici ile bir sonraki bloğu alır.
		block := iter.Next()

		fmt.Printf("Önceki Blog Hasi :%x\n", block.PrevHash)
		fmt.Printf("Blog Verisi      :%s\n", block.Data)
		fmt.Printf("Blog Hasi        :%x\n", block.Hash)

		// PoW (Proof of Work) doğrulama işlemi için blockchain.NewProof(block) çağrısı ile yeni bir Proof of Work nesnesi oluşturulur.
		pow := blockchain.NewProof(block)
		// PoW doğrulamasının sonucunu ekrana yazdırır.
		fmt.Printf("Pow:%s\n", strconv.FormatBool(pow.Validate()))

		// Blok zincirinin başlangıç noktasına (genesis block) geldiğinde döngüyü sonlandırır.
		if len(block.PrevHash) == 0 {
			break
		}
		fmt.Println()
	}
}

// run metod, komut satırından alınan argümanlara göre işlemleri yönlendirir.
func (cli *CommandLine) run() {
	// Komut satırı argümanlarını doğrula
	cli.validateArgs()

	// "add" ve "print" komutları için flag setleri oluştur
	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data") // "add" komutu için "block" parametresi tanımla

	// İlk argümana göre işlem seçimi yap
	switch os.Args[1] { //os.Args[1]  bu programın baslatılması sırasında verılen komut dur
	case "add":
		// "add" komutu için gerekli argümanları işle
		err := addBlockCmd.Parse(os.Args[2:])
		blockchain.Handler(err) // Hata varsa Handler fonksiyonu ile işle
	case "print":
		// "print" komutu için gerekli argümanları işle
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.Handler(err) // Hata varsa Handler fonksiyonu ile işle
	default:
		// Bilinmeyen komut ise kullanım talimatlarını göster ve programı sonlandır
		cli.PrintUsage()
		runtime.Goexit()
	}

	// Eğer "add" komutu işlendi ise
	if addBlockCmd.Parsed() {
		// Eğer "block" parametresi boş ise kullanım talimatlarını göster ve programı sonlandır
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		// "block" parametresi ile yeni bir blok ekleyerek işlemi gerçekleştir
		cli.addBlock(*addBlockData)
	}

	// Eğer "print" komutu işlendi ise
	if printChainCmd.Parsed() {
		// Blok zincirini ekrana yazdır
		cli.printChain()
	}
}
