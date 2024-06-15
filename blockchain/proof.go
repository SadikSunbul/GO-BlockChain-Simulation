package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

//Bu kodumuzda proof-work kullanıcaz işkanıtı olarak makınelerın sıfre cozmelerını ıstıycez

//bloktan veri al

//0'dan başlayan bir sayaç (nonce) oluşturun

//veri artı sayacın bir karmasını oluşturun

//bir dizi gereksinimi karşılayıp karşılamadığını görmek için karmayı kontrol edin

// Gereksinimler:
// ilk birkaç bayt 0 içermelidir  (bu zorluk derecesıdır)

const Difficulty = 18

type ProofOfWork struct {
	Block  *Block
	Target *big.Int //blogun hasının bu degerden kucuk olması gerekir
}

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)                  //	buyuk bır ınt degerı olusturduk ve 1 degerı ıel baslattık
	target.Lsh(target, uint(256-Difficulty)) //LSH sol a kaydırma yapar target degerını uint(256-Difficulty) bu kadar sola kaydırır
	pow := &ProofOfWork{b, target}           //iş kanıtı oluşturuldu b blogu ıcın bu target olmalıdır
	return pow
}

// InitData, Proof of Work için gerekli olan verileri hazırlar ve birleştirir.
// Hazırlanan veriler bloğun önceki hash'i, verisi, nonce değeri ve zorluk seviyesini içerir.
func (pow *ProofOfWork) InitData(nonce int) []byte {
	// Verileri birleştirmek için bytes.Join kullanılır.
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,       // Bloğun önceki hash'i
			pow.Block.Data,           // Bloğun verisi
			ToHex(int64(nonce)),      // Nonce değeri (int64 türünde hexadecimal'e dönüştürülür)
			ToHex(int64(Difficulty)), // Zorluk seviyesi (int64 türünde hexadecimal'e dönüştürülür)
		},
		[]byte{}, // Ayraç olarak kullanılacak boş bir byte slice'ı
	)
	return data
}

// ToHex, bir int64 değerini hexadecimal formatına dönüştürür ve byte dizisi olarak döndürür.
func ToHex(num int64) []byte {
	// Yeni bir bytes.Buffer oluşturulur.
	buf := new(bytes.Buffer)

	// binary.Write kullanılarak num değişkeni big-endian formatında yazılır ve buf'a yazılır.
	err := binary.Write(buf, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buf.Bytes() // Byte dizisi olarak buf içeriği döndürülür.
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int //buyuk bır ınt olusturulur
	var hash [32]byte   //hashlenmıs sıfreyı tutucaz burada

	nonce := 0                  //aranıcak degere
	for nonce < math.MaxInt64 { //nerdeyse sonsuz bır dongu olusturuyoruz
		data := pow.InitData(nonce) //nonce degerı ıle datayı olusturuyoruz
		hash = sha256.Sum256(data)  //datayı sıfrelıuoz
		fmt.Printf("\r%x", hash)    //hası ekrana bastırıyoruz
		intHash.SetBytes(hash[:])   //hası buyuk ınte donduruyoruz ıyas yapabılmek ıcın

		if intHash.Cmp(pow.Target) == -1 { //eger targeten kucuk ıse dogru bulundu has cıkabılrısın
			break
		} else { //yanlıs bulundu ıse nonce 1 artır ve devam et
			nonce++
		}
	}
	fmt.Println()
	return nonce, hash[:] //nonce ve hash deerlerını gerıye dondur
}

func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int                   //buyuk bır ınt tanımlanır
	data := pow.InitData(pow.Block.Nonce) //blogun noncenı vererek data olusturulur

	hash := sha256.Sum256(data) //olusturulan datayı haslerız
	intHash.SetBytes(hash[:])   //heshlenmıs datayı int turune donustururu

	return intHash.Cmp(pow.Target) == -1
	/*
		-1: intHash, pow.Target'ten küçüktür.
		0: intHash, pow.Target'e eşittir.
		1: intHash, pow.Target'ten büyüktür.
	*/
}
