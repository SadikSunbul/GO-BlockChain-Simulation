package blockchain

import (
	"fmt"
	"github.com/dgraph-io/badger"
)

const (
	dbPath = "./tmp/blocks"
)

type BlockChain struct { //Block zıncırını tutar
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct { //BlockChain üzerinde gezmek ıcın kullanılır
	CurrentHash []byte
	Database    *badger.DB
}

// InitBlockChain BlockChainin başlatılmasını sağlar
func InitBlockChain() *BlockChain {

	var lastHash []byte

	//Database baglantısı olusturulur
	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	opts.Logger = nil

	db, err := badger.Open(opts)
	Handler(err)

	//Databasede bir güncelleme ekleme değişiklik işlemi yapılıcaktır
	err = db.Update(func(txn *badger.Txn) error {

		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound { //lh yok ise gir
			fmt.Print("Mevcut blockchain bulunamadı\n")
			genesis := Genesis() //genesıs blogu uret
			fmt.Println("Genesis proved")
			err = txn.Set(genesis.Hash, genesis.Serilize()) //genesis blogu hası ne karsılık ablogun verıelrının byte halını verıtabanına yaz
			Handler(err)
			err = txn.Set([]byte("lh"), genesis.Hash) //lh a son blog olan genesisi yaz
			lastHash = genesis.Hash                   //son blogun degerını doldur
			return err
		} else { //lh degerı var ıse
			item, err := txn.Get([]byte("lh")) //lh son eklenen blogu oku
			Handler(err)
			lastHash, err = item.ValueCopy(nil) //degeri kopyala
			return err
		}
	})
	Handler(err)
	blockChain := BlockChain{lastHash, db} //LastHash ve database degerlerını vererek bır BlockChaın zıncırı olusturduk
	return &blockChain
}

// AddBlock  block zincirine  blok elememızı saglar
func (chain *BlockChain) AddBlock(data string) {
	var lastHash []byte                                      //son  blogu tutucak
	err := chain.Database.View(func(txn *badger.Txn) error { //verıtabanında okuma ıslemı yapıca k
		item, err := txn.Get([]byte("lh")) //lh degerını oku
		Handler(err)
		lastHash, err = item.ValueCopy(nil) //degeri kopyala
		return err
	})
	Handler(err)
	newBlock := CreateBlock(data, lastHash) //blogu olusturt

	err = chain.Database.Update(func(txn *badger.Txn) error { //verıtabanında bır guncelleme ekleme yapılıcak
		err := txn.Set(newBlock.Hash, newBlock.Serilize()) //yenı blogu verıtabanına ekle
		Handler(err)
		err = txn.Set([]byte("lh"), newBlock.Hash) //lh degerını guncelle

		chain.LastHash = newBlock.Hash //mevcut BlockChain nesenesındekı lastHası guncelle cunku artık son uretılen blog newBlock
		return err
	})
	Handler(err)
}

func (chain *BlockChain) Iterator() *BlockChainIterator { //BlockChaın de okuma işlemi yapmak için başlangıç değerlerini atayan kod
	iter := &BlockChainIterator{chain.LastHash, chain.Database}
	return iter
}

// Next BlockChaınde gerıye dogru ılerlemeyı saglar ve suankı blogun verılerını gerıye doner
func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	err := iter.Database.View(func(txn *badger.Txn) error { //database den okum yapıcak
		item, err := txn.Get(iter.CurrentHash) //son blogun hası ıle ara son blogun verılerıne erıs
		Handler(err)

		encoderBlock, err := item.ValueCopy(nil) //son blogun verıelrını al
		block = Deserilize(encoderBlock)         //blog verılerını deserılıze et
		return err
	})
	Handler(err)
	iter.CurrentHash = block.PrevHash //yenı blog suankının bır oncekı demıs olduk
	return block                      //gerıye su ankı blogu gerı doner
}
