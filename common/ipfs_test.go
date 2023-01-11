package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	shell "github.com/ipfs/go-ipfs-api"
	file "github.com/ipfs/go-ipfs-files"
	"github.com/mr-tron/base58/base58"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// 交易结构体(未来的通道)
type Transaction struct {
	Person1      string `json:"person1,omitempty" xml:"person1"`
	Person2      string `json:"person2,omitempty" xml:"person2"`
	Person1money string `json:"person1Money,omitempty" xml:"person1Money"`
	Person2money string `json:"person2Money,omitempty" xml:"person2Money"`
}

// 数据上传到ipfs
func UploadIPFS(shell *shell.Shell, str string) string {
	hash, err := shell.Add(bytes.NewBufferString(str))
	if err != nil {
		fmt.Println("上传ipfs时错误：", err)
	}
	return hash
}

// 数据上传到ipfs
func UploadFile(shell *shell.Shell) string {
	tmppath, err := os.MkdirTemp("/Users/user", "files-test2")
	if err != nil {
		return ""
	}
	defer os.RemoveAll(tmppath)
	hash, err := shell.AddDir(tmppath)
	if err != nil {
		fmt.Println("上传ipfs时错误：", err)
	}
	return hash
}

// 数据上传到ipfs
func UploadJson(shell *shell.Shell, str string, index int64) string {
	tmppath, err := os.MkdirTemp("/Users/user", "files-test2")
	if err != nil {
		return ""
	}
	defer os.RemoveAll(tmppath)
	path := filepath.Join(tmppath, strconv.FormatInt(index, 10))
	b, err := json.Marshal(str)
	err = file.WriteTo(file.NewBytesFile(b), path)
	if err != nil {
		return ""
	}
	hash, err := shell.AddDir(tmppath)
	if err != nil {
		fmt.Println("上传ipfs时错误：", err)
	}
	return hash
}

// 从ipfs下载数据
func CatIPFS(shell *shell.Shell, hash string) string {

	read, err := shell.Cat(hash)
	if err != nil {
		fmt.Println(err)
	}
	body, err := ioutil.ReadAll(read)

	return string(body)
}

// 通道序列化
func marshalStruct(transaction Transaction) []byte {

	data, err := json.Marshal(&transaction)
	if err != nil {
		fmt.Println("序列化err=", err)
	}
	return data
}

// 数据反序列化为通道
func unmarshalStruct(str []byte) Transaction {
	var transaction Transaction
	err := json.Unmarshal(str, &transaction)
	if err != nil {
		fmt.Println("unmarshal err:", err)
	}
	return transaction
}

func TestGet1Txs(t *testing.T) {
	sh := shell.NewShell("10.23.34.36:5001")
	var v = "{\"TxType\":2212222,\"NftIndex\":0,\"AccountNameHash\":\"IUGc18NfYlBnnB/OCsdC52AOnWDybw0EPf6CufmcoPM=\",\"AccountIndex\":0,\"CreatorAccountIndex\":0,\"CreatorTreasuryRate\":0,\"CreatorAccountNameHash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"NftContentHash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"CollectionId\":0}"
	hash := UploadJson(sh, v, 1)
	fmt.Println("文件hash是", hash)
}

func TestKeyIpns(t *testing.T) {
	sh := shell.NewShell("localhost:5001")
	key1, _ := sh.KeyGen(context.Background(), "cid+index", shell.KeyGen.Type("ed25519"))
	fmt.Println(key1)
}

func TestCidCode(t *testing.T) {
	v0 := "QmRwPwcyZxH8H6cqNFeEvZgDsRMMs9RsFipR2YX4bhiSfh"
	b0, _ := base58.Decode(v0)
	hex := hexutil.Encode(b0)
	lowerHex := strings.ToLower(hex)
	fmt.Println("v0", strings.Replace(lowerHex, "0x1220", "", 1))
	h, _ := hexutil.Decode(strings.ToLower(hex))
	hs := base58.Encode(h)
	fmt.Println("hs", hs)
}

func TestPublish(t *testing.T) {
	sh := shell.NewShell("localhost:5001")
	resp, err := sh.PublishWithDetails("/ipfs/"+"hash", "cid+index", 0, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp.Value)
}

func TestCid(t *testing.T) {
	var hash = "0x1220" + "90ba66b30af6928ec12db9c0551b5990a192bd9c1ac97ed2419c48c595c04a67"
	b, _ := hexutil.Decode(hash)
	cid := base58.Encode(b)
	fmt.Println(cid)
}