package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

func main() {
	chain := InitBlockChain() //zinciri başlattım

	chain.AddBlock("1. Blok bir öncesi genesis") //zincire blok ekledik
	chain.AddBlock("2. Blok bir öncesi 1. blok")
	chain.AddBlock("3. Blok bir öncesi 2. blok")

	//zinciri okuycaz

	for _, block := range chain.blocks {
		fmt.Printf("Önceki Blog Hasi :%x\n", block.PrevHash)
		fmt.Printf("Blog Verisi      :%s\n", block.Data)
		fmt.Printf("Blog Hasi        :%x\n", block.Hash)
		fmt.Println()
	}
}

type BlockChain struct {
	blocks []*Block
}

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
}

func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{}) //son prarametre de bos byte oldugu ıcın  data ve prevHash drekt baglanır aralarında bırsey olmaz
	hash := sha256.Sum256(info)                                //blok verılerını sha256 ya gore sıfreler
	b.Hash = hash[:]                                           //şifrelenmiş veriyi blogun hasıne ekler
}

func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash} //[]byte(data) kısmı strıng ıfadeyi byte dizisine donduruyor
	block.DeriveHash()                                //blogun hasını hesaplıyor
	return block
}

// AddBlock  block zincirine  blok elememızı saglar
func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.blocks[len(chain.blocks)-1] //en son eklenen blogu bulur
	new := CreateBlock(data, prevBlock.Hash)       //blogu olusturuyoruz
	chain.blocks = append(chain.blocks, new)       //block zıncırıne yenı blogu eklıyoruz
}

func Genesis() *Block {
	return CreateBlock("Genesis Block", []byte{})
}

func InitBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}} //block zıncırını baslatıyoruz ve ılk blogu genesis fonksıyonu ıle olusturup ıcerısıne atıyoruz ılk blogumuz olustu artık
}
