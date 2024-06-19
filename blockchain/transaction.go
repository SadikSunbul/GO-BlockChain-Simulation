package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Transaction struct {
	ID      []byte     //transectıon hası
	Inputs  []TxInput  //bu transectıondakı ınputlar
	Outputs []TxOutput //bu transectıondakı outputlar
}

func CoinbaseTx(to, data string) *Transaction {
	if data == "" { //data boş ise gir
		data = fmt.Sprintf("Coins to %s", to) //paralar to da der
	}

	txin := TxInput{[]byte{}, -1, data} //hıcbır cıktıya referabs vermez ,cıkıs endexi -1 aynı referans yok , sadce data mesajı vardır
	txout := TxOutput{100, to}          //100 tokeni to ya gonderırı

	tx := Transaction{nil, []TxInput{txin}, []TxOutput{txout}} //transectıonı olustururuz
	tx.SetID()                                                 //Transectıon Id sını olustururuz
	return &tx
}

func (tx *Transaction) SetID() { //Id olusturur transectıonun
	var encoded bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx) //transectıonu encode edıyoruz
	Handle(err)

	hash = sha256.Sum256(encoded.Bytes()) //transectıonu byte seklınde sha256 ıle sıfrelıyoruz ve ıd yı urettık
	tx.ID = hash[:]
}

// NewTransaction, belirtilen bir adresten başka bir adrese belirtilen miktar token transferi yapacak yeni bir işlem oluşturur.
func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TxInput   // İşlem girişleri (input'ları) için boş slice
	var outputs []TxOutput // İşlem çıktıları (output'ları) için boş slice

	// Gönderen adresten belirtilen miktar kadar harcanabilir çıktıları bul
	acc, validOutputs := chain.FindSpendableOutputs(from, amount)

	// Eğer hesaplanan toplam miktar istenilen miktardan az ise hata ver ve işlemi sonlandır
	if acc < amount {
		log.Panic("Error: yeterli fon yok")
	}

	// Geçerli (harcanabilir) çıktılar üzerinde döngü
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid) // İşlem ID'sini byte dizisine dönüştür
		Handle(err)                         // Hata varsa işlemi sonlandır

		// Çıktı endeksleri üzerinde döngü
		for _, out := range outs {
			input := TxInput{txID, out, from} // Yeni bir işlem girişi oluştur
			inputs = append(inputs, input)    // Oluşturulan girişi input'lar listesine ekle
		}
	}

	// Yeni bir çıktı oluştur ve belirtilen adrese belirtilen miktarı gönder
	outputs = append(outputs, TxOutput{amount, to})

	// Eğer hesaplanan toplam miktar istenilen miktardan fazlaysa, geri kalan miktarı gönderen adrese geri gönder
	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from})
	}

	// Yeni bir işlem (transaction) oluştur
	tx := Transaction{nil, inputs, outputs}

	// İşlem ID'sini hesapla ve işlem nesnesine ata
	tx.SetID()

	return &tx // Oluşturulan işlem nesnesini işaretçi olarak döndür
}

/*
IsCoinbase fonksiyonu, bir işlemin coinbase işlemi olup olmadığını belirler. Coinbase işlemi, yeni bir bloğun ödülünü
alan ilk işlemdir ve genellikle tek bir girişe sahiptir. Bu fonksiyon, giriş sayısını, girişin kimlik uzunluğunu ve
çıkışını kontrol ederek bir işlemin coinbase olup olmadığını belirler.
*/
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
	// Bu fonksiyon, bir işlemin coinbase işlemi olup olmadığını kontrol eder.
	// Bir coinbase işlemi sadece bir girişe sahiptir.
	// Bu girişin işlem kimliği (ID) uzunluğu 0'a eşit olmalıdır ve çıkış (Out) -1 olmalıdır.
}
