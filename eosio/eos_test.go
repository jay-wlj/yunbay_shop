package eosio

import (
	"fmt"
	"testing"
	"github.com/eoscanada/eos-go"
)

var chain *chainClient = &chainClient{}

var (
	API_POINT = "https://proxy.eosnode.tools"
)

func init() {
	//连接cleos
	chain.API = eos.New(API_POINT)

	//获取信息，查看是否连接成功
	_, err := chain.API.GetInfo()
	if err != nil {
		fmt.Println("get info failed,err:", err)
		return
	}
}

//测试产生新地址
func testNewAddress(t *testing.T) {
	fmt.Println("======================testNewAddress==========================\r\n")
	defer fmt.Println("\r\n==============================================================\r\n\r\n")

	addr := chain.NewAddress("testtest4455")
	fmt.Println("new addr:", addr)
}

//测试新建连接
func testNewEOSChain(t *testing.T) {
	fmt.Println("=====================testNewEOSChain=========================\r\n")
	defer fmt.Println("\r\n==============================================================\r\n\r\n")

	NewEOSChain(API_POINT)
	fmt.Println("new eos ok")
}

//测试获取账号余额
func testGetAccountBalance(t *testing.T) {
	fmt.Println("======================testNewEOSChain=========================\r\n")
	defer fmt.Println("\r\n==============================================================\r\n\r\n")

	//传入空字符串
	//_, err := chain.GetAccountBalance("")
	//fmt.Println(err)
	//fmt.Println("\r\n--------------------------------\r\n")

	//传入不存在的账号
	//_, err = chain.GetAccountBalance("****")
	//fmt.Println(err)
	//fmt.Println("\r\n--------------------------------\r\n")

	//正常传入可用账号
	bal, err := chain.GetAccountBalance("testtest4455")
	if err != nil {
		fmt.Println("get account balance failed:", err)
		return
	}
	fmt.Println("get account balance ok,balance:", bal)
}

//测试获取块数量
func testGetBlockCount(t *testing.T) {
	fmt.Println("=====================testGetBlockCount========================\r\n")
	defer fmt.Println("\r\n==============================================================\r\n\r\n")

	count, err := chain.GetBlockCount()
	if err != nil {
		fmt.Println("get block count failed:", err)
		return
	}
	fmt.Println("get block count ok,count =", count)
}

// //测试通过哈希或者num获取区块信息
// func testGetBlockByIdOrNum(t *testing.T) {
// 	fmt.Println("====================testGetBlockByIdOrNum========================\r\n")
// 	defer fmt.Println("\r\n==============================================================\r\n\r\n")

// 	//通过hash获取
// 	block, err := chain.GetBlockByIdOrNum("18474647")
// 	if err != nil {
// 		fmt.Println("block failed:", err)
// 		return
// 	}
// 	fmt.Println("block info:")
// 	fmt.Printf("Num:%d\r\nHash:%s\r\nTimestamp:%v\r\n", block.Num, block.Hash, block.Timestamp)
// 	fmt.Println("------------------------------------")

// 	//通过num获取
// 	block, err = chain.GetBlockByIdOrNum(int64(18474647))
// 	if err != nil {
// 		fmt.Println("block failed:", err)
// 		return
// 	}
// 	fmt.Println("block info:")
// 	fmt.Printf("Num:%d\r\nHash:%s\r\nTimestamp:%v\r\n", block.Num, block.Hash, block.Timestamp)
// }

//测试获取最近不可逆的块
func testGetLastIrreversibleBlock(t *testing.T) {
	fmt.Println("====================testGetLastBlock========================\r\n")
	defer fmt.Println("\r\n==============================================================\r\n\r\n")

	//block, err := chain.GetLastIrreversibleBlock()
	//if err != nil {
	//	fmt.Println("get last block failed:", err)
	//	return
	//}
	//fmt.Println("last block info:")
	//fmt.Printf("Num:%d\r\nHash:%s\r\nTimestamp:%v\r\n", block.Num, block.Hash, block.Timestamp)
}

//测试块检查
func testCheckBlock(t *testing.T) {
	fmt.Println("======================testCheckBlock=========================\r\n")
	defer fmt.Println("\r\n==============================================================\r\n\r\n")

	err := chain.CheckBlock()
	if err != nil {
		fmt.Println("check block failed:", err)
		return
	}
	fmt.Println("check block ok")
}

//测试获取交易
func testGetTransaction(t *testing.T) {
	fmt.Println("====================testGetTransaction========================\r\n")
	defer fmt.Println("\r\n==============================================================\r\n\r\n")

	//传入空
	_, err := chain.GetTransaction("")
	fmt.Println(err)
	fmt.Println("\r\n--------------------------------\r\n")

	//传入不存在的hash
	_, err = chain.GetTransaction("*****")
	fmt.Println(err)
	fmt.Println("\r\n--------------------------------\r\n")


	//正常传入
	transaction, err := chain.GetTransaction("85027b60f012f3e8f69434cc6e075871672febb292f263cf41f1dce67ebc10fe")
	if err != nil {
		fmt.Println("get transaction failed:", err)
		return
	}
	fmt.Println("get transaction ok:")
	fmt.Printf("Hash:%s\r\nCommitAt:%s\r\nblock:%d\r\n", transaction.Hash, transaction.CommitAt, transaction.Block)
	for k, v := range transaction.Data {
		fmt.Printf("transaction sn %d:from(%s),to(%s),amount(%d),Memo(%s)\r\n",
			k, v.From, v.To, v.Amount, v.Memo)
	}

}

//测试获取块上的交易
func testGetBlockTransactions(t *testing.T) {
	fmt.Println("====================testGetBlockTransactions==================\r\n")
	defer fmt.Println("\r\n==============================================================\r\n\r\n")

	//传入不存在的num
	//transactions := chain.GetBlockTransactions(11111111, 11111111)
	//fmt.Println(transactions)
	//fmt.Println("\r\n--------------------------------\r\n")

	//传入相等的值
	//transactions := chain.GetBlockTransactions(26179746, 26179746)
	//transactions := chain.GetBlockTransactions(25936645, 25936645)
	transactions := chain.GetBlockTransactions(190, 191)
	fmt.Printf("transaction: %d\r\n", len(transactions))
	for _, transaction := range transactions {
		fmt.Printf("%s:from(%s),to(%s),amount(%d),symbol(%s),Memo(%s)\r\n",
				transaction.TxId, transaction.From, transaction.To, transaction.Amount, transaction.Symbol, transaction.Memo)

	}
	fmt.Println("\r\n--------------------------------\r\n")

	////传入不相等的值
	//transactions = chain.GetBlockTransactions(1820, 1921)
	//for sn, transaction := range transactions {
	//	fmt.Println("*************************************")
	//	fmt.Printf("block transaction(%d):\r\n", sn)
	//	fmt.Printf("Hash:%s\r\nCommitAt:%s\r\nblock:%d\r\n", transaction.Hash, transaction.CommitAt, transaction.Block)
	//	for k, v := range transaction.Data {
	//		fmt.Printf("transaction(%d):from(%s),to(%s),amount(%d),Memo(%s)\r\n",
	//			k, v.From, v.To, v.Amount, v.Memo)
	//	}
	//}
}

//测试发送交易
func testSendTransaction(t *testing.T) {
	fmt.Println("====================testSendTransaction================\r\n")
	defer fmt.Println("\r\n=============================================================\r\n\r\n")

	var tx EosSendTransaction
	tx.Amount = 105212
	tx.From = "testtest4455"
	tx.To = "testtestin12"
	tx.Memo = "g87mahhfzpk3"
	tx.Key = "5HpWRZcCaxCSkamCt116p3ebXDGfYE1Yf9GHUbNc6Si68NCkFd6"


	//tx.Amount = 280092
	//tx.From = "testtestout3"
	//tx.To = "testtest4455"
	//tx.Key = "5KQALjtq2T9tYE4XVxNmppZhQFbKrgxsumjd77jJ8tro1mHVTys"

	//传空值
	_, err := chain.SendTransaction(nil)
	fmt.Println(err)
	fmt.Println("\r\n-------------------------------\r\n")

	//正常
	txid, err := chain.SendTransaction(&tx)
	if err != nil {
		fmt.Println("send transaction failed:", err)
		return
	}
	fmt.Println("send transaction ok:", txid)
}


func testSendTokenTransaction(t *testing.T) {
	fmt.Println("====================testSendTokenTransaction================\r\n")
	defer fmt.Println("\r\n=============================================================\r\n\r\n")
	//转账一个代币 非eos
	var jungleTx  EosSendTransaction
	jungleTx.Amount = 23000
	jungleTx.From = "testtest4455"
	jungleTx.To = "testtestin12"
	jungleTx.Memo = "g87mahhfzpk3"
	jungleTx.Key = "5HpWRZcCaxCSkamCt116p3ebXDGfYE1Yf9GHUbNc6Si68NCkFd6"
	txid, err := chain.SendTokenTransaction("KT", 8, &jungleTx)

	if err != nil {
		fmt.Println("SendTokenTransaction failed:", err)
		return
	}
	fmt.Println("SendTokenTransaction ok:", txid)
}


func testCheckTransactionConfirm(t *testing.T) {
	check := chain.CheckTransactionConfirm("b837962e46b3962937dae086109a675badd589c351519ebe7616fc64d1b2cfda")
	fmt.Println("check:", check)
}

//测试批量发送交易
func testSendTransactions(t *testing.T) {
	fmt.Println("===================testSendTransactions================\r\n")
	defer fmt.Println("\r\n============================================================\r\n\r\n")

	//传空值
	_, err := chain.SendTransactions(nil)
	fmt.Println(err)
	fmt.Println("\r\n-------------------------------\r\n")

	//传一个值
	var txs []*EosSendTransaction
	var tx EosSendTransaction
	tx.Amount = 5
	tx.From = "yunexeost1"
	tx.To = "yunexeost2"
	tx.Memo = "hello world"
	tx.Key = "5KbmqRYkGhkueAkAuhcihiFt2Hi2W3qbzRsu1YYLqdooveffyBD"
	txs = append(txs, &tx)
	txid, err := chain.SendTransactions(txs)
	if err != nil {
		fmt.Println("send transaction failed:", err)
		return
	}
	fmt.Println("send transaction ok:", txid)
	fmt.Println("\r\n-------------------------------\r\n")

	//传多个值
	for i := 0; i < 5; i++ {
		txm := new(EosSendTransaction)
		txm.Amount = int64(i + 1)
		txm.From = "yunexeost1"
		txm.To = "yunexeost2"
		txm.Memo = "hello world 1->2"
		txm.Key = "5Kg9hf9V55C8wSySY6sFMPUCPm8xvHBKXniLhYA3xGYweb5Nv8C"
		txs = append(txs, txm)
	}

	for i := 0; i < 5; i++ {
		txm := new(EosSendTransaction)
		txm.Amount = int64(i + 1)
		txm.From = "yunexeost2"
		txm.To = "yunexeost1"
		txm.Memo = "hello world 2->1"
		txm.Key = "5KbmqRYkGhkueAkAuhcihiFt2Hi2W3qbzRsu1YYLqdooveffyBD"
		txs = append(txs, txm)
	}
	txid, err = chain.SendTransactions(txs)
	if err != nil {
		fmt.Println("send transaction failed:", err)
		return
	}
	fmt.Println("send transaction ok:", txid)
}

func testGetRam(t *testing.T) {
	fmt.Println("===================testGetRam================\r\n")
	defer fmt.Println("\r\n============================================================\r\n\r\n")

	//获取ram信息
	ram, err := chain.GetRam("yunexeos")
	if err != nil {
		panic(err)
	}
	fmt.Printf("get ram info ok:\r\nquota:%d\r\nused:%d\r\n\r\n",
		ram.Quota, ram.Used)
}

func testGetNet(t *testing.T) {
	fmt.Println("===================testGetNet================\r\n")
	defer fmt.Println("\r\n============================================================\r\n\r\n")

	//获取net信息
	net, err := chain.GetNet("yunexeos")
	if err != nil {
		panic(err)
	}
	fmt.Printf("get net info ok:\r\nweight:%d\r\nused:%d\r\navailable:%d\r\nmax:%d\r\n\r\n",
		net.Weight, net.Used, net.Available, net.Max)
}

func testGetCpu(t *testing.T) {
	fmt.Println("===================testGetNet================\r\n")
	defer fmt.Println("\r\n============================================================\r\n\r\n")

	//获取net信息
	cpu, err := chain.GetCpu("yunexeos")
	if err != nil {
		panic(err)
	}
	fmt.Printf("get cpu info ok:\r\nweight:%d\r\nused:%d\r\navailable:%d\r\nmax:%d\r\n\r\n",
		cpu.Weight, cpu.Used, cpu.Available, cpu.Max)

}

func TestEos(t *testing.T) {
	//测试新产生地址
	testNewAddress(t)

	//测试新建连接
	//testNewEOSChain(t)

	//测试获取账户余额
	//testGetAccountBalance(t)

	//测试获取区块数量
	//testGetBlockCount(t)

	//测试通过哈希或者num获取区块信息
	//testGetBlockByIdOrNum(t)

	//测试获取最后一个区块信息
	//testGetLastIrreversibleBlock(t)

	//测试检查区块
	//testCheckBlock(t)

	//测试获取交易详情
	//testGetTransaction(t)
	//
	//测试获取块上的交易详情
	testGetBlockTransactions(t)

	//测试转账
	//testSendTransaction(t)

	//testSendTokenTransaction(t)


	//测试检查交易确认
	//testCheckTransactionConfirm(t)

	//测试批量转账
	//testSendTransactions(t)

	//测试获取ram信息
	//testGetRam(t)

	//测试获取net信息
	//testGetNet(t)

	//测试获取cpu信息
	//testGetCpu(t)
}
