package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"goXChain/blockchain"
	"goXChain/kvstore"
	"goXChain/maker"
	"goXChain/statdb"
	"goXChain/statemachine"
	"goXChain/trie"
	"goXChain/txpool"
	"goXChain/types"
	"goXChain/utils/hash"
	"goXChain/utils/rlp"
	"time"
)

func main() {

	fmt.Println("")
	fmt.Println("初始化...")
	fmt.Println("")
	db := kvstore.NewLevelDB("./testdb")
	state := trie.NewState(db, trie.EmptyHash)
	stateDb := statdb.NewStatDb(*state)
	fmt.Println("原始树根:", stateDb.All().Root())
	exec := statemachine.StateMachine{}
	pool := txpool.NewDefaultPool(stateDb)
	config := maker.ChainConfig{
		Duration:   2 * time.Second,                          // 10秒生成一个区块
		Coinbase:   *types.AddressFromBytes([]byte("0x123")), // Coinbase 地址
		Difficulty: 1,
	}

	fmt.Println("输入数字选择操作：")
	fmt.Println("1.创建账户")
	fmt.Println("2.交易转账")
	fmt.Println("3.查询账户余额")
	fmt.Println("")

	option := 0

	for {
		fmt.Scan(&option)
		switch option {
		case 1:
			privateKey, addr := CreateAccount(stateDb, 200, 10)
			fmt.Println("创建的公钥：", privateKey)
			fmt.Println("账户地址：", addr)

		case 2:
			readyTx()
		}
		if option == -1 {
			return
		}

	}

	fmt.Println("创建账户：")
	privateKey1, add1 := CreateAccount(stateDb, 200, 10)
	privateKey2, add2 := CreateAccount(stateDb, 110, 10)
	fmt.Println("转帐前账户1余额：", stateDb.Load(add1).Amount)
	fmt.Println("转帐前账户2余额：", stateDb.Load(add2).Amount)
	txs, pubk := Tx(privateKey1, privateKey2, &add1, &add2)
	pool.NotifyTxEvent(txs, pubk)

	blockmaker := maker.NewBlockMaker(pool, stateDb, exec, config)
	header := blockchain.Header{
		Root:       stateDb.All().Root(),
		ParentHash: hash.Hash{},
		Height:     0,
		Coinbase:   types.Address{},
		Timestamp:  0,
		Nonce:      0,
	}
	block1 := blockmaker.NewBlock(header)
	block2 := blockmaker.NewBlock(block1.CurrentHeader)
	block3 := blockmaker.NewBlock(block2.CurrentHeader)
	fmt.Println("树根一：", block1.CurrentHeader.Root)
	fmt.Println("时间戳：", block1.CurrentHeader.Timestamp)
	fmt.Println("哈希：", block1.CurrentHeader.Hash())
	fmt.Println("树根二：", block2.CurrentHeader.Root)
	fmt.Println("时间戳：", block2.CurrentHeader.Timestamp)
	fmt.Println("哈希：", block2.CurrentHeader.ParentHash)
	fmt.Println("第三个区块状态树根：", block3.CurrentHeader.Root)
	fmt.Println("时间戳：", block3.CurrentHeader.Timestamp)
	fmt.Println("哈希：", block3.CurrentHeader.Hash())
	fmt.Println("最终树根:", stateDb.All().Root())

	fmt.Println("转帐后账户1余额：", stateDb.Load(add1).Amount)
	fmt.Println("转帐后账户2余额：", stateDb.Load(add2).Amount)
}

// 创建账户
func CreateAccount(stateDb *statdb.StatDb, accBanlance uint64, accNonce uint64) (*ecdsa.PrivateKey, types.Address) {
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err)
	}
	byte1, _ := rlp.EncodeToBytes(privateKey.PublicKey)
	add1 := types.PubKeyToAddress(byte1)
	account1 := &types.Account{
		Amount: accBanlance,
		Nonce:  accNonce,
		Root:   stateDb.All().Root(),
	}
	stateDb.Store(add1, *account1)

	return privateKey, add1
}

// 准备交易
func readyTx() {
	var add1 *types.Address
	fmt.Printf("请输入发送方地址：\n")
	fmt.Printf("请输入发送方地址：\n")
	fmt.Scanf("%v,%v",&add1)
	var add2 *types.Address
	fmt.Printf("请输入接受方地址：")
	fmt.Scanf("%v",&add2)
	var privateKey1 *ecdsa.PrivateKey
	fmt.Printf("请输入发送方pk：")
	var privateKey2 *ecdsa.PrivateKey
	fmt.Scanf("%v",&privateKey1)
	fmt.Printf("请输入接受方pk：")
	fmt.Scanf("%v",&privateKey2)
	Tx(privateKey1,privateKey2,add1,add2)
	

}

func Tx(privateKey *ecdsa.PrivateKey, privateKey2 *ecdsa.PrivateKey, add1 *types.Address, add2 *types.Address) ([]*types.Transaction, []ecdsa.PublicKey) {
	txdata := types.Txdata{
		From:     *add1,
		To:       *add2,
		Nonce:    17,
		Value:    5,
		Gas:      12,
		GasPrice: 10,
		Input:    nil,
	}
	byte1, _ := rlp.EncodeToBytes(txdata)
	
	signature, err := types.SignData(privateKey, byte1)
	if err != nil {
		panic(err)
	}
	tx := types.Transaction{
		Txdata:    txdata,
		Signature: *signature,
	}
	var txs []*types.Transaction
	var pubk []ecdsa.PublicKey
	txs = append(txs, &tx)
	pubk = append(pubk, privateKey.PublicKey, privateKey2.PublicKey, privateKey2.PublicKey)
	return txs, pubk
}

