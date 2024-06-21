package blockchain

import (
	"bytes"
	"encoding/hex"
	"github.com/dgraph-io/badger"
	"log"
)

var (
	utxoPrefix   = []byte("utxo-")
	prefixLength = len(utxoPrefix)
)

type UTXOSet struct {
	Blockchain *BlockChain
}

// FindSpendableOutputs, belirtilen bir adrese gönderilmiş ve henüz harcanmamış çıktıları (UTXO'ları) bulmak için kullanılır.
// Ayrıca, bu çıktılar aracılığıyla belirli bir miktar token transfer edilebilecek çıktıları belirler.
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	// Kullanılmamış çıkışları saklamak için bir harita oluşturuyoruz
	unspentOuts := make(map[string][]int)
	// Toplam biriktirilen miktarı izlemek için bir değişken tanımlıyoruz
	accumulated := 0
	// Veritabanına erişim için Blockchain'in veritabanı bağlantısını alıyoruz
	db := u.Blockchain.Database

	// Veritabanında sorgu yapmak için bir View işlevi başlatıyoruz
	err := db.View(func(txn *badger.Txn) error {
		// Badger için varsayılan Iterator seçeneklerini alıyoruz
		opts := badger.DefaultIteratorOptions

		// İteratör oluşturuyoruz ve işlem bittikten sonra kapatmayı unutmuyoruz
		it := txn.NewIterator(opts)
		defer it.Close()

		// UTXO önekine göre arama yapmaya başlıyoruz
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			// İteratörden bir öğe alıyoruz
			item := it.Item()
			k := item.Key()
			v, err := item.ValueCopy(nil)
			Handle(err)                         // Hata durumunu yönetim işlevi ile ele alıyoruz
			k = bytes.TrimPrefix(k, utxoPrefix) // Önekten kurtuluyoruz
			txID := hex.EncodeToString(k)       // Transaction ID'yi hex formatına dönüştürüyoruz
			outs := DeserializeOutputs(v)       // Çıkışları Deserialize ediyoruz

			// Çıkışları döngüye alarak kontrol ediyoruz
			for outIdx, out := range outs.Outputs {
				// Çıkışın bu anahtarla kilidini kontrol ediyoruz ve istenen miktardan azsa ekliyoruz
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOuts[txID] = append(unspentOuts[txID], outIdx)
				}
			}
		}
		return nil
	})
	Handle(err)                     // Hata durumunu yönetim işlevi ile ele alıyoruz
	return accumulated, unspentOuts // Biriktirilen miktarı ve kullanılmamış çıkışları döndürüyoruz
}

// FindUTXO fonksiyonu, bir kripto para biriminin UTXO (Kullanılmamış İşlem Çıkışları) üzerinde, belirli bir anahtara (pubKeyHash) ait olan kullanılmamış çıkışları bulur.
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TxOutput {
	// Kullanılmamış işlem çıkışlarını tutacak bir slice oluşturuyoruz
	var UTXOs []TxOutput

	// Veritabanı bağlantısı için Blockchain'den veritabanı erişimini alıyoruz
	db := u.Blockchain.Database

	// Veritabanında sorgu yapmak için bir View işlevi başlatıyoruz
	err := db.View(func(txn *badger.Txn) error {
		// Badger için varsayılan Iterator seçeneklerini alıyoruz
		opts := badger.DefaultIteratorOptions

		// İteratör oluşturuyoruz ve işlem bittikten sonra kapatmayı unutmuyoruz
		it := txn.NewIterator(opts)
		defer it.Close()

		// UTXO önekine göre arama yapmaya başlıyoruz
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			// İteratörden bir öğe alıyoruz
			item := it.Item()
			v, err := item.ValueCopy(nil)
			Handle(err)                   // Hata durumunu yönetim işlevi ile ele alıyoruz
			outs := DeserializeOutputs(v) // Çıkışları Deserialize ediyoruz

			// Çıkışları döngüye alarak kontrol ediyoruz
			for _, out := range outs.Outputs {
				// Çıkışın bu anahtarla kilidini kontrol ediyoruz
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out) // UTXOs slice'ına uygun çıkışı ekliyoruz
				}
			}
		}

		return nil
	})
	Handle(err) // Hata durumunu yönetim işlevi ile ele alıyoruz

	return UTXOs // Bulunan tüm uygun (locked with key) UTXO'ları döndürüyoruz
}

// CountTransactions fonksiyonu, bir kripto para biriminin UTXO (Kullanılmamış İşlem Çıkışları) içindeki islemlerin sayısını döndürür.
func (u UTXOSet) CountTransactions() int {
	// Veritabanı bağlantısı için Blockchain'den veritabanı erişimini alıyoruz
	db := u.Blockchain.Database

	// İşlem sayısını tutacak bir sayaç başlatıyoruz
	counter := 0

	// Veritabanında sorgu yapmak için bir View işlevi başlatıyoruz
	err := db.View(func(txn *badger.Txn) error {
		// Badger için varsayılan Iterator seçeneklerini alıyoruz
		opts := badger.DefaultIteratorOptions

		// İteratör oluşturuyoruz ve işlem bittikten sonra kapatmayı unutmuyoruz
		it := txn.NewIterator(opts)
		defer it.Close()

		// UTXO önekine göre arama yaparak işlem sayısını artırıyoruz
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			counter++
		}

		return nil
	})

	Handle(err) // Hata durumunu yönetim işlevi ile ele alıyoruz

	return counter // Toplam işlem sayısını döndürüyoruz
}

// Reindex fonksiyonu, UTXO (Kullanılmamış İşlem Çıkışları) haritasını yeniden doldurur. TXO setini yeniden indekslemek için kullanılır.
func (u UTXOSet) Reindex() {
	// Veritabanı bağlantısı için Blockchain'den veritabanı erişimini alıyoruz
	db := u.Blockchain.Database

	// UTXO setini yeniden indekslemek için önce mevcut önekle başlayan tüm verileri sileriz
	u.DeleteByPrefix(utxoPrefix)

	// Blockchain üzerindeki tüm UTXO'ları yeniden alıyoruz
	UTXO := u.Blockchain.FindUTXO()

	// Veritabanında güncelleme işlemi başlatıyoruz
	err := db.Update(func(txn *badger.Txn) error {
		// Her bir transaction ID ve çıkışlar için UTXO haritasını döngüye alıyoruz
		for txID, outs := range UTXO {
			// Transaction ID'yi hex formatına dönüştürüyoruz
			key, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}
			// UTXO öneki ile birleştirerek anahtarı oluşturuyoruz
			key = append(utxoPrefix, key...)

			// Anahtar ve değer (serialize edilmiş çıkışlar) çiftini veritabanına ekliyoruz
			err = txn.Set(key, outs.Serialize())
			Handle(err) // Hata durumunu yönetim işlevi ile ele alıyoruz
		}

		return nil
	})
	Handle(err) // Hata durumunu yönetim işlevi ile ele alıyoruz
}

func (u *UTXOSet) Update(block *Block) {
	// Veritabanı bağlantısı için Blockchain'den veritabanı erişimini alıyoruz
	db := u.Blockchain.Database

	// Veritabanında güncelleme işlemi başlatıyoruz
	err := db.Update(func(txn *badger.Txn) error {
		// Blok içindeki her bir işlemi döngüye alıyoruz
		for _, tx := range block.Transactions {
			// Coinbase işlemi değilse devam ediyoruz
			if !tx.IsCoinbase() {
				// İşlemdeki her bir girdiyi döngüye alıyoruz
				for _, in := range tx.Inputs {
					// Güncellenmiş çıkışları tutacak bir TxOutputs yapısı oluşturuyoruz
					updatedOuts := TxOutputs{}
					// Girdinin ID'sini alıp UTXO önekle birleştirerek anahtar oluşturuyoruz
					inID := append(utxoPrefix, in.ID...)
					// Veritabanından ilgili girdiye ait veriyi alıyoruz
					item, err := txn.Get(inID)
					Handle(err) // Hata durumunu yönetim işlevi ile ele alıyoruz
					v, err := item.ValueCopy(nil)
					Handle(err) // Hata durumunu yönetim işlevi ile ele alıyoruz

					// Deserialize edilmiş çıkışları elde ediyoruz
					outs := DeserializeOutputs(v)

					// Çıkışları döngüye alarak güncellenmiş çıkışları oluşturuyoruz
					for outIdx, out := range outs.Outputs {
						if outIdx != in.Out {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					// Güncellenmiş çıkış listesi boşsa ilgili girdiyi veritabanından siliyoruz
					if len(updatedOuts.Outputs) == 0 {
						if err := txn.Delete(inID); err != nil {
							log.Panic(err)
						}
					} else {
						// Güncellenmiş çıkış listesi boş değilse veritabanına yeni değeri kaydediyoruz
						if err := txn.Set(inID, updatedOuts.Serialize()); err != nil {
							log.Panic(err)
						}
					}
				}
			}

			// Yeni çıkışları tutacak bir TxOutputs yapısı oluşturuyoruz
			newOutputs := TxOutputs{}
			for _, out := range tx.Outputs {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			// İşlem ID'sini UTXO önekle birleştirerek anahtar oluşturuyoruz
			txID := append(utxoPrefix, tx.ID...)
			// Yeni çıkışları veritabanına kaydediyoruz
			if err := txn.Set(txID, newOutputs.Serialize()); err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	Handle(err) // Hata durumunu yönetim işlevi ile ele alıyoruz
}

// DeleteByPrefix, UTXOSet yapısına ait bir metottur ve belirli bir öneki taşıyan tüm anahtarları veritabanından siler.
func (u *UTXOSet) DeleteByPrefix(prefix []byte) {
	// deleteKeys fonksiyonu, belirli anahtarları silmek için kullanılır.
	deleteKeys := func(keysForDelete [][]byte) error {
		// Veritabanı işlemleri güncelleme modunda yapılır.
		if err := u.Blockchain.Database.Update(func(txn *badger.Txn) error {
			// Her bir anahtarı sil.
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	// Anahtarların toplanacağı koleksiyon boyutu.
	collectSize := 100000
	// Veritabanında okuma işlemi için işlev.
	u.Blockchain.Database.View(func(txn *badger.Txn) error {
		// Iterator seçenekleri ayarlanır.
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		// Silinecek anahtarlar için boş bir dilim oluşturulur.
		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0
		// Belirtilen önek ile başlayan anahtarları bulur ve işler.
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			// Anahtarı kopyalar ve silinecek anahtarlar listesine ekler.
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			// Belirli bir koleksiyon boyutuna ulaşıldığında, bu anahtarları silme işlevini çağırır.
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					log.Panic(err)
				}
				// Yeni bir silinecek anahtarlar dilimi oluşturur.
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		// Son toplama işlemi için kalan anahtarları silme işlevini çağırır.
		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}
