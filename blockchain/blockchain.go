package blockchain

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"time"
)

type BlockChain struct {
	DB *sql.DB
}

type Block struct {
	CurrHash     []byte
	PrevHash     []byte
	Nonce        uint64
	Difficulty   uint8
	Miner        string
	Signature    []byte
	TimeStamp    string
	Transactions []Transaction
	Mapping      map[string]uint64
}

type Transaction struct {
	RandBytes []byte
	PrevBlock []byte
	Sender    string
	Receiver  string
	Value     uint64
	ToStorage uint64
	CurrHash  []byte
	Signature []byte
}

type User struct {
	PrivateKey *rsa.PrivateKey
}

const (
	CREATE_TABLE = `
CREATE TABLE BlockChain (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Hash VARCHAR(44) UNIQUE,
    Block TEXT
);
`
)

const (
	DIFFICULTY     = 20
	RAND_BYTES     = 32
	START_PERCENT  = 10
	STORAGE_REWARD = 1
)

const (
	GENESIS_BLOCK  = "GENESIS-BLOCK"
	STORAGE_VALUE  = 100
	GENESIS_REWARD = 100
	STORAGE_CHAIN  = "STORAGE-CHAIN"
)

func NewChain(filename, receiver string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	file.Close()
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return err
	}
	defer db.Close()
	db.Exec(CREATE_TABLE)
	chain := &BlockChain{
		DB: db,
	}
	genesis := &Block{
		PrevHash:  []byte(GENESIS_BLOCK),
		Mapping:   make(map[string]uint64),
		Miner:     receiver,
		TimeStamp: time.Now().Format(time.RFC3339),
	}
	genesis.Mapping[STORAGE_CHAIN] = STORAGE_VALUE
	genesis.Mapping[receiver] = GENESIS_REWARD
	genesis.CurrHash = genesis.hash()
	chain.AddBlock(genesis)
	return nil
}

func LoadChain(filename string) *BlockChain {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil
	}
	chain := &BlockChain{
		DB: db,
	}
	return chain
}

func NewBlock(miner string, prevHash []byte) *Block {
	return &Block{
		Difficulty: DIFFICULTY,
		PrevHash:   prevHash,
		Miner:      miner,
		Mapping:    make(map[string]uint64),
	}
}

func NewTransaction(user *User, lasthash []byte, to string, value uint64) *Transaction {
	tx := &Transaction{
		RandBytes: GenerateRandomBytes(RAND_BYTES),
		PrevBlock: lasthash,
		Sender:    user.Address(),
		Receiver:  to,
		Value:     value,
	}
	if value > START_PERCENT {
		tx.ToStorage = STORAGE_REWARD
	}
	tx.CurrHash = tx.hash()
	tx.Signature = tx.sign(user.Private())
	return tx
}

func (block *Block) AddTransaction(chain *BlockChain, tx *Transaction) error {
	if tx == nil {
		return errors.New("tx is null")
	}
	if tx.Value == 0 {
		return errors.New("tx value = 0")
	}
	if len(block.Transactions) == TXS_LIMIT && tx.Sender != STORAGE_CHAIN {
		return errors.New("len tx = limit")
	}
	return nil
}

func (user *User) Address() string {
	return StringPublic(user.Public())
}

func (user *User) Private() *rsa.PrivateKey {
	return user.PrivateKey
}

func (tx *Transaction) hash() []byte {
	return HashSum(bytes.Join(
		[][]byte{
			tx.RandBytes,
			tx.PrevBlock,
			[]byte(tx.Sender),
			[]byte(tx.Receiver),
			ToBytes(tx.Value),
			ToBytes(tx.ToStorage),
		},
		[]byte{},
	))
}

func (tx *Transaction) sign(priv *rsa.PrivateKey) []byte {
	return Sign(priv, tx.CurrHash)
}

func StringPublic(pub *rsa.PublicKey) string {
	return Base64Encode(x509.MarshalPKCS1PublicKey(pub))
}

func (user *User) Public() *rsa.PublicKey {
	return &(user.PrivateKey).PublicKey
}

func HashSum(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func ToBytes(num uint64) []byte {
	var data = new(bytes.Buffer)
	err := binary.Write(data, binary.BigEndian, num)
	if err != nil {
		return nil
	}
	return data.Bytes()
}

func Sign(priv *rsa.PrivateKey, data []byte) []byte {
	signdata, err := rsa.SignPSS(rand.Reader, priv, crypto.SHA256, data, nil)
	if err != nil {
		return nil
	}
	return signdata
}

func GenerateRandomBytes(max uint) []byte {
	var slice = make([]byte, max)
	_, err := rand.Read(slice)
	if err != nil {
		return nil
	}
	return slice
}

func (chain *BlockChain) Size() uint64 {
	var size uint64
	row := chain.DB.QueryRow("SELECT Id FROM BlockChain ORDER BY Id DESC")
	row.Scan(&size)
	return size
}

func (chain *BlockChain) AddBlock(block *Block) {
	chain.DB.Exec("INSERT INTO BlockChain (Hash, Block) VALUES ($1, $2)",
		Base64Encode(block.CurrHash),
		SerializeBlock(block),
	)
}

func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func SerializeBlock(block *Block) string {
	jsonData, err := json.MarshalIndent(*block, "", "\t")
	if err != nil {
		return ""
	}
	return string(jsonData)
}
