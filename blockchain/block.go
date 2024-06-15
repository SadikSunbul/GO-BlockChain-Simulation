package blockchain

type BlockChain struct {
	Block []*Block
}

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	nonce    int
}

func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash, 0} //[]byte(data) kısmı strıng ıfadeyi byte dizisine donduruyor

	pow := NewProof(block)   //yeni bir iş kanıtı olusturuyoruz
	nonce, hash := pow.Run() //bu işkanıtınını çalıştırıyoruz blogunhasını ve nance degerını eklıyoruz
	block.Hash = hash[:]
	block.nonce = nonce
	return block
}

// AddBlock  block zincirine  blok elememızı saglar
func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.Block[len(chain.Block)-1] //en son eklenen blogu bulur
	new := CreateBlock(data, prevBlock.Hash)     //blogu olusturuyoruz
	chain.Block = append(chain.Block, new)       //block zıncırıne yenı blogu eklıyoruz
}

func Genesis() *Block {
	return CreateBlock("Genesis Block", []byte{})
}

func InitBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}} //block zıncırını baslatıyoruz ve ılk blogu genesis fonksıyonu ıle olusturup ıcerısıne atıyoruz ılk blogumuz olustu artık
}
