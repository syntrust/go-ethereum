package tests

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"gotest.tools/assert"
)

func newAccount() *common.Address {
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	addr := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)
	return &addr
}

func createState(db ethdb.Database, address []common.Address) (common.Hash, error) {
	sdb := state.NewDatabase(db)
	statedb, _ := state.New(common.Hash{}, sdb, nil)
	for _, addr := range address {
		statedb.AddBalance(addr, common.Big1)
	}

	root := statedb.IntermediateRoot(false)
	hash, err := statedb.Commit(false)
	if err != nil{
		return common.Hash{}, err
	}
	err = statedb.Database().TrieDB().Commit(root, true, nil)
	return hash, err
}

func TestCreateState(t *testing.T) {
	accountNum := 10000
	defer duration(track(fmt.Sprintf("start create state for %v accounts...", accountNum)))
	r, err := rand.Int(rand.Reader, new(big.Int).SetUint64(uint64(accountNum)))
	assert.NilError(t, err)
	addresses := make([]common.Address, 0)
	var lucky common.Address
	for i := 0; i < accountNum; i++ {
		addr := newAccount()
		if r.Int64() == int64(i) {
			lucky = *addr
			log.Printf("will check the %v th address: %x\n", r, lucky)
		}
		addresses = append(addresses, *addr)
	}
	dbname := "statetest.db"
	assert.NilError(t, err)
	db, err := rawdb.NewLevelDBDatabase(dbname, 0, 0, "")
	assert.NilError(t, err)
	root, err := createState(db, addresses)
	assert.NilError(t, err)
	log.Println("root=",root.Hex())
	sdb := state.NewDatabase(db)
	statedb, err := state.New(root, sdb, nil)
	assert.NilError(t, err)
	assert.Equal(t, statedb.GetBalance(lucky).Uint64(), uint64(1))
	assert.NilError(t, err)

	//iter := db.NewIterator(nil, nil)
	//for iter.Next() {
	//	log.Printf("key:%s, value:%s\n", iter.Key(), iter.Value())
	//}
	//iter.Release()
	err = os.RemoveAll(dbname)
}

func track(msg string) (string, time.Time) {
	return msg, time.Now()
}
func duration(msg string, start time.Time) {
	log.Printf("%v: %v\n", msg, time.Since(start))
}
