package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
	"os"
	"runtime"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First Transection from genesis"
)

type BlockChain struct { //Block zıncırını tutar
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct { //BlockChain üzerinde gezmek ıcın kullanılır
	CurrentHash []byte
	Database    *badger.DB
}

func DBexists() bool { //block zıncırın var olup olmadıgını kontrolunu yapıcak
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

/*
ContinueBlockChain :
Bu fonksiyon, mevcut bir blockchain'in varlığını kontrol eder, varsa veritabanını açar, son bloğun hash değerini alır
ve bu bilgileri kullanarak bir BlockChain yapısı oluşturur. Daha sonra bu yapının işaretçisini döndürür. Bu işlem,
mevcut bir blockchain'e devam etmek veya yeni işlemler eklemek için kullanılır.
*/
func ContinueBlockChain(address string) *BlockChain {
	if DBexists() == false { //veritabaının olup olmadıgını kontrolunu yapar
		fmt.Println("Mevcut bir blockchain bulunamadı, bir tane oluşturun!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	opts.Logger = nil

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh")) //son hası alıyoruz
		Handle(err)
		lastHash, err = item.ValueCopy(nil)

		return err
	})
	Handle(err)

	chain := BlockChain{lastHash, db} //mevcut chaını devam etırmek ıcın BlockChaın degerlerını koruyarak eklıyoruz

	return &chain
}

// InitBlockChain BlockChainin başlatılmasını sağlar
func InitBlockChain(address string) *BlockChain {

	var lastHash []byte

	if DBexists() { //verı tabanını var olup olmadıgının kontrolu
		fmt.Printf("Blok zinciri zaten mevcut\n")
		runtime.Goexit()
	}

	//Database baglantısı olusturulur
	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	opts.Logger = nil

	db, err := badger.Open(opts)
	Handle(err)

	//Databasede bir güncelleme ekleme değişiklik işlemi yapılıcaktır
	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData) //CoınbaseTx yanı odulu alıcak kısıyı belırlıyoruz burada onun transectıonı olusturuldu
		genesis := Genesis(cbtx)                 //genesis bloguna buradan gelen transectıonı verdık ve genesis blogu olusturuldu
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize()) //blogu verıtabanına kaydetik
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash) //son hash degerı guncellendi

		lastHash = genesis.Hash

		return err
	})

	Handle(err)
	blockChain := BlockChain{lastHash, db} //LastHash ve database degerlerını vererek bır BlockChaın zıncırı olusturduk
	return &blockChain
}

// AddBlock  block zincirine  blok elememızı saglar
func (chain *BlockChain) AddBlock(transactions []*Transaction) {
	var lastHash []byte                                      //son  blogu tutucak
	err := chain.Database.View(func(txn *badger.Txn) error { //verıtabanında okuma ıslemı yapıca k
		item, err := txn.Get([]byte("lh")) //lh degerını oku
		Handle(err)
		lastHash, err = item.ValueCopy(nil) //degeri kopyala
		return err
	})
	Handle(err)
	newBlock := CreateBlock(transactions, lastHash) //blogu olusturt

	err = chain.Database.Update(func(txn *badger.Txn) error { //verıtabanında bır guncelleme ekleme yapılıcak
		err := txn.Set(newBlock.Hash, newBlock.Serialize()) //yenı blogu verıtabanına ekle
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash) //lh degerını guncelle

		chain.LastHash = newBlock.Hash //mevcut BlockChain nesenesındekı lastHası guncelle cunku artık son uretılen blog newBlock
		return err
	})
	Handle(err)
}

// Iterator :BlockChaın de okuma işlemi yapmak için başlangıç değerlerini atayan kod
func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}
	return iter
}

// Next BlockChaınde gerıye dogru ılerlemeyı saglar ve suankı blogun verılerını gerıye doner
func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	err := iter.Database.View(func(txn *badger.Txn) error { //database den okum yapıcak
		item, err := txn.Get(iter.CurrentHash) //son blogun hası ıle ara son blogun verılerıne erıs
		Handle(err)

		encoderBlock, err := item.ValueCopy(nil) //son blogun verıelrını al
		block = Deserilize(encoderBlock)         //blog verılerını deserılıze et
		return err
	})
	Handle(err)
	iter.CurrentHash = block.PrevHash //yenı blog suankının bır oncekı demıs olduk
	return block                      //gerıye su ankı blogu gerı doner
}

// FindUnspentTransactions : Bu fonksiyon, bir blockchain üzerinde belirli bir adrese gönderilmiş ancak henüz harcanmamış (unspent) işlemleri bulmak için kullanılır.
func (chain *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTxs []Transaction        // Harcanmamış işlemleri tutacak slice
	spentTXOs := make(map[string][]int) // Harcanmış işlemlerin çıktılarını izlemek için kullanılacak map

	iter := chain.Iterator() // Blok zinciri iteratorunu olustur

	for {
		block := iter.Next() // Sıradaki bloğu al

		for _, tx := range block.Transactions { // Bloktaki her işlem için döngü
			txID := hex.EncodeToString(tx.ID) // İşlem ID'sini hex formatına dönüştürerek al

		Outputs:
			for outIdx, out := range tx.Outputs { // İşlemin çıktıları üzerinde döngü
				// Eğer bu çıktı daha önce harcanmışsa atla
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				// Eğer çıktı, belirtilen adrese gönderilmişse
				if out.IsLockedWithKey(pubKeyHash) { //aranan adres tarafından acılıp acılmayacagı kontrol edılır
					unspentTxs = append(unspentTxs, *tx) // Harcanmamış işlemler listesine ekle
				}
			}

			// Coinbase işlemi değilse (yani normal bir transfer işlemi)
			if tx.IsCoinbase() == false {
				// İşlemin girdileri üzerinde döngü
				for _, in := range tx.Inputs {
					// Eğer bu girişin kilidi (unlock) belirtilen adrese açılabiliyorsa
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.ID)                   // Girişin işlem ID'sini alarak hex formatına dönüştür
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out) // Harcanmış işlemler listesine ekle
					}
				}
			}
		}

		// Eğer bloğun önceki hash değeri yoksa (genesis block durumu), iterasyonu sonlandır
		if len(block.PrevHash) == 0 {
			break
		}
	}

	return unspentTxs // Harcanmamış işlemleri içeren slice'i döndür
}

// FindUTXO fonksiyonu, belirtilen bir adrese gönderilmiş ve henüz harcanmamış (UTXO) çıktıları bulmak için kullanılır.
func (chain *BlockChain) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput                                             // Harcanmamış çıktıları (UTXO'ları) tutacak slice oluşturulur
	unspentTransactions := chain.FindUnspentTransactions(pubKeyHash) // Belirtilen adrese gönderilmiş harcanmamış işlemleri bul

	for _, tx := range unspentTransactions { // Her harcanmamış işlem için döngü
		for _, out := range tx.Outputs { // İşlemin çıktıları üzerinde döngü
			if out.IsLockedWithKey(pubKeyHash) { // Çıktı, belirtilen adrese gönderilmişse
				UTXOs = append(UTXOs, out) // UTXO'lar listesine çıktıyı ekle
			}
		}
	}
	return UTXOs // Harcanmamış çıktıları (UTXO'ları) içeren slice'i döndür
}

// FindSpendableOutputs, belirtilen bir adrese gönderilmiş ve henüz harcanmamış çıktıları (UTXO'ları) bulmak için kullanılır.
// Ayrıca, bu çıktılar aracılığıyla belirli bir miktar token transfer edilebilecek çıktıları belirler.
func (chain *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)                   // Harcanmamış çıktıları (UTXO'ları) tutacak map
	unspentTxs := chain.FindUnspentTransactions(pubKeyHash) // Belirtilen adrese gönderilmiş harcanmamış işlemleri bul

	accumulated := 0 // Toplam biriktirilen miktar

Work:
	for _, tx := range unspentTxs { // Her harcanmamış işlem için döngü başlat
		txID := hex.EncodeToString(tx.ID) // İşlem ID'sini hex formatına dönüştür

		for outIdx, out := range tx.Outputs { // İşlemin çıktıları üzerinde döngü
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount { // Çıktı, belirtilen adrese gönderilmiş ve biriktirilen miktar istenilen miktarı aşmamışsa
				accumulated += out.Value                              // Çıktının değerini biriktirilen miktar'a ekle
				unspentOuts[txID] = append(unspentOuts[txID], outIdx) // Harcanmamış çıktıları map'e ekle

				if accumulated >= amount { // Biriktirilen miktar istenilen miktara eşit veya fazla ise
					break Work // İşlemi sonlandır
				}
			}
		}
	}

	return accumulated, unspentOuts // Biriktirilen toplam miktarı ve harcanmamış çıktıları map olarak döndür
}
func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

// SignTransaction fonksiyonu, bir Transaction yapısını imzalar.
// İmzalamak için verilen private anahtar (privKey) kullanılır ve işlemi daha önce yapılmış olan işlemlerle ilişkilendirir.
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction) // Önceki işlemlerin haritasını (map) oluşturur

	// İşlemdeki her girdi için önceki işlemi bulup prevTXs haritasına ekler
	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)        // Girdinin referans verdiği önceki işlemi bulur
		Handle(err)                                     // Hata durumunda işlemi ele alır
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX // Önceki işlemi haritaya (map) ekler (ID'si hex olarak kodlanmış olarak)
	}

	tx.Sign(privKey, prevTXs) // Transaction yapısını imzalar
}

// VerifyTransaction fonksiyonu, bir Transaction yapısının geçerliliğini doğrular.
// Geçerlilik kontrolü için verilen önceki işlemler haritası (prevTXs) kullanılır.
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction) // Önceki işlemlerin haritasını (map) oluşturur

	// İşlemdeki her girdi için önceki işlemi bulup prevTXs haritasına ekler
	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)        // Girdinin referans verdiği önceki işlemi bulur
		Handle(err)                                     // Hata durumunda işlemi ele alır
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX // Önceki işlemi haritaya (map) ekler (ID'si hex olarak kodlanmış olarak)
	}

	return tx.Verify(prevTXs) // Transaction yapısının geçerliliğini doğrular
}
