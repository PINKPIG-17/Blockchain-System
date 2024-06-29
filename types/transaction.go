package types

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"goXChain/crypto/secp256k1"
	"goXChain/crypto/sha3"
	"goXChain/utils/hexutil"
	"goXChain/utils/rlp"
	"hash"
	"math/big"
)


type Log []struct {
	Address Address
	Topics []hash.Hash
	Data []byte
}

type Receiption struct {
	TxHash hash.Hash
	Status int
	GasUsed uint64
}

type Transaction struct {
	Txdata
	Signature
}


type Txdata struct {
	From     Address
	To       Address
	Nonce    uint64
	Value    uint64
	Gas      uint64
	GasPrice uint64
	Input    []byte
}

type Signature struct {
	R, S *big.Int
	V    uint8
}

func (tx Transaction) From() Address {
	txdata := tx.Txdata
	toSign, err := rlp.EncodeToBytes(txdata)
	fmt.Println(hexutil.Encode(toSign),err)
	msg := sha3.Keccak256(toSign)
	sig := make([]byte,65)
	pubkey, err := secp256k1.RecoverPubkey(msg[:],sig)
	if err != nil {
		panic(err)
	}
	return PubKeyToAddress(pubkey)
}

// SignData 使用私钥对数据进行签名，并返回签名结果
func SignData(privateKey *ecdsa.PrivateKey, data []byte) (*Signature, error) {
	hash := sha256.Sum256(data)

	// 对数据的 SHA-256 哈希进行 ECDSA 签名
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, err
	}

	// 构造签名结构体
	signature := &Signature{
		R: new(big.Int).Set(r),
		S: new(big.Int).Set(s),
		V: 27, // 这里可以根据需要设置 V 字段的值
	}

	return signature, nil
}

// VerifySignature 验证签名的有效性
func VerifySignature(publicKey *ecdsa.PublicKey, data []byte, signature *Signature) bool {
	// 使用 SHA-256 哈希函数计算数据的哈希值
	hash := sha256.Sum256(data)

	// 构造 ECDSA 签名的 r 和 s 值
	r := signature.R
	s := signature.S

	// 使用公钥进行签名验证
	isValid := ecdsa.Verify(publicKey, hash[:], r, s)
	return isValid
}