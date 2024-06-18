package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash         []byte
	Transactions []*Transaction //burada data vardı sımdı datalar yerını transectıonlara aldı
	PrevHash     []byte
	Nonce        int
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte // İşlemlerin kimliklerini (hash'lerini) tutmak için bir slice tanımlıyoruz
	var txHash []byte     // Son hesaplanan hash'i tutacak byte slice

	// Bloktaki her bir işlem için döngü
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID) // Her işlemin ID'sini (hash'ini) txHashes slice'ına ekliyoruz
	}

	// txHashes içindeki tüm işlem kimliklerini birleştirip tek bir byte dizisi oluşturuyoruz
	concatenated := bytes.Join(txHashes, []byte{})

	// concatenated byte dizisinin SHA-256 hash'ini hesaplıyoruz
	hash := sha256.Sum256(concatenated)

	// [32]byte türündeki hash'i []byte türüne dönüştürüyoruz
	txHash = hash[:]

	return txHash[:] // Hesaplanan hash'i döndürüyoruz
}

func CreateBlock(tsx []*Transaction, prevHash []byte) *Block {
	block := &Block{[]byte{}, tsx, prevHash, 0} //[]byte(data) kısmı strıng ıfadeyi byte dizisine donduruyor

	pow := NewProof(block)   //yeni bir iş kanıtı olusturuyoruz
	nonce, hash := pow.Run() //bu işkanıtınını çalıştırıyoruz blogunhasını ve nance degerını eklıyoruz
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

//Badger DB sadece byte kabul ettıgı ıcın serılestırme ve deserilize ıslemlerı kolyalastıralım

func (b *Block) Serialize() []byte {
	var res bytes.Buffer // bir bytes.Buffer nesnesi oluşturuluyor

	// encoder adında bir gob.Encoder nesnesi oluşturuluyor
	// Bu encoder, res adındaki bytes.Buffer üzerine kodlanmış (encoded) veri yazacak
	encoder := gob.NewEncoder(&res)

	// encoder.Encode(b) çağrısı, 'b' adındaki bir veriyi (muhtemelen bir struct veya başka bir veri yapısı)
	// encoder kullanarak kodlar (encode) ve bu işlem sırasında bir hata oluşursa 'err' değişkenine atar.
	err := encoder.Encode(b)
	Handle(err) // Handle fonksiyonu, oluşan hatayı işlemek için çağrılır. Hata varsa bu fonksiyon kullanıcı tarafından tanımlanmış olmalıdır.

	// res.Bytes() çağrısı, bytes.Buffer nesnesi 'res' içindeki tüm verileri bir byte dilimine (slice) dönüştürür ve döndürür.
	// Bu dönüştürülmüş byte dilimi, encoder.Encode(b) işlemi sonucunda 'res' nesnesine yazılmış kodlanmış veriyi içerir.
	return res.Bytes()

}

// Deserilize fonksiyonu, verilen byte diliminden (data) bir Block struct'ı oluşturur ve döndürür.
func Deserilize(data []byte) *Block {
	var block Block // Block türünde bir değişken oluşturuluyor

	// bytes.NewReader(data) ile data byte dilimi üzerinde bir okuyucu (reader) oluşturuluyor
	decoder := gob.NewDecoder(bytes.NewReader(data))

	// decoder.Decode(&block) çağrısı, data üzerindeki kodlanmış veriyi Block struct'ına çözümleme (decode) işlemini yapar.
	// Oluşabilecek hatalar err değişkenine atanır ve Handle fonksiyonu ile işlenir.
	err := decoder.Decode(&block)
	Handle(err) // Hata varsa bu fonksiyon kullanıcı tarafından tanımlanmış olmalıdır.

	// Çözümlenen (deserialized) Block struct'ı, bellekte oluşturulan bir yapı olduğu için &
	// ile işaret edilerek ve fonksiyon dışına taşınabilmesi için *Block türünde döndürülür.
	return &block
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
