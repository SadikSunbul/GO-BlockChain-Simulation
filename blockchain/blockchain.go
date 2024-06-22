package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	dbPath      = "./tmp/blocks_%s"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct { //Block zıncırını tutar
	LastHash []byte
	Database *badger.DB
}

//type BlockChainIterator struct { //Blockchain üzerinde gezmek ıcın kullanılır
//	CurrentHash []byte
//	Database    *badger.DB
//}

func DBexists(path string) bool { //block zıncırın var olup olmadıgını kontrolunu yapıcak
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
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
func ContinueBlockChain(nodeId string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeId)
	if DBexists(path) == false { //veritabaının olup olmadıgını kontrolunu yapar
		fmt.Println("Mevcut bir blockchain bulunamadı, bir tane oluşturun!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(path)
	opts.Dir = path
	opts.ValueDir = path
	opts.Logger = nil

	db, err := openDB(path, opts)
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
func InitBlockChain(address, nodeId string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeId)

	if DBexists(path) { //verı tabanını var olup olmadıgının kontrolu
		fmt.Printf("Blok zinciri zaten mevcut\n")
		runtime.Goexit()
	}
	var lastHash []byte
	//Database baglantısı olusturulur
	opts := badger.DefaultOptions(path)
	opts.Dir = path
	opts.ValueDir = path
	opts.Logger = nil

	db, err := openDB(path, opts)
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
func (chain *BlockChain) AddBlock(block *Block) {
	err := chain.Database.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}

		blockData := block.Serialize()
		err := txn.Set(block.Hash, blockData)
		Handle(err)

		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ := item.ValueCopy(nil)

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.ValueCopy(nil)

		lastBlock := Deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = txn.Set([]byte("lh"), block.Hash)
			Handle(err)
			chain.LastHash = block.Hash
		}

		return nil
	})
	Handle(err)
}

func (chain *BlockChain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		if chain.VerifyTransaction(tx) != true {
			log.Panic("Invalid Transaction")
		}
	}

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(nil)

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.ValueCopy(nil)

		lastBlock := Deserialize(lastBlockData)

		lastHeight = lastBlock.Height

		return err
	})
	Handle(err)

	newBlock := CreateBlock(transactions, lastHash, lastHeight+1)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})
	Handle(err)

	return newBlock
}

func (chain *BlockChain) GetBestHeight() int {
	var lastBlock Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ := item.ValueCopy(nil)

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.ValueCopy(nil)

		lastBlock = *Deserialize(lastBlockData)

		return nil
	})
	Handle(err)

	return lastBlock.Height
}

func (chain *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("Block is not found")
		} else {
			blockData, _ := item.ValueCopy(nil)

			block = *Deserialize(blockData)
		}
		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

func (chain *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte

	iter := chain.Iterator()

	for {
		block := iter.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks
}

// FindUTXO fonksiyonu, belirtilen bir adrese gönderilmiş ve henüz harcanmamış (UTXO) çıktıları bulmak için kullanılır.
func (chain *BlockChain) FindUTXO() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXO
}

// FindTransaction fonksiyonu, belirtilen bir işlem ID'sine sahip olan işlemi blok zincirinde bulur.
// ID, işlemin benzersiz tanımlayıcısıdır (genellikle işlemin hash değeri olarak kullanılır).
func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator() // Blok zinciri üzerinde bir iterator oluşturur

	for {
		block := iter.Next() // Sonraki bloğu alır

		for _, tx := range block.Transactions { // Bloğün işlemleri üzerinde döner
			if bytes.Compare(tx.ID, ID) == 0 { // İşlem ID'si belirtilen ID'ye eşitse
				return *tx, nil // İşlemi bulduğunda işlemi ve nil hatasını döndürür
			}
		}

		if len(block.PrevHash) == 0 { // Eğer bloğun önceki hash değeri yoksa (genesis blok)
			break // Döngüyü sonlandır
		}
	}

	return Transaction{}, errors.New("Transaction does not exist") // İşlem bulunamazsa hata döndürür
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

	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction) // Önceki işlemlerin haritasını (map) oluşturur

	// İşlemdeki her girdi için önceki işlemi bulup prevTXs haritasına ekler
	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)        // Girdinin referans verdiği önceki işlemi bulur
		Handle(err)                                     // Hata durumunda işlemi ele alır
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX // Önceki işlemi haritaya (map) ekler (ID'si hex olarak kodlanmış olarak)
	}

	return tx.Verify(prevTXs) // Transaction yapısının geçerliliğini doğrular
}

func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	retryOpts.Truncate = true
	db, err := badger.Open(retryOpts)
	return db, err
}

func openDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				log.Println("database unlocked, value log truncated")
				return db, nil
			}
			log.Println("could not unlock database:", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}
