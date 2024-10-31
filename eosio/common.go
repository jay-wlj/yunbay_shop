package eosio

import (
	"time"

	eos "github.com/eoscanada/eos-go"
)

type Chain interface {
	// 账号下创建新地址
	NewAddress(account string) string

	// 获取系统的余额
	GetAccountBalance(account string) (int64, error)

	// 获取区块数量
	GetBlockCount() (int64, error)

	//通过区块num或hash获取区块信息
	//GetBlockByIdOrNum(i interface{}) (*EosBlock, error)

	// 获取最后一个区块
	//GetLastIrreversibleBlock() (*EosBlock, error)
	//获取最新不可逆区块号
	GetLastIrreversibleBlockNumber() (int64, error)

	//区块检查
	CheckBlock() error

	//发送转账
	SendTransaction(tx *EosSendTransaction) (string, error)

	//发送eosio.token标准合约代币转账
	SendTokenTransaction(symbol string, precisison uint8, tx *EosSendTransaction) (string, error)
	
	//一次发送多个交易
	SendTransactions(txs []*EosSendTransaction) (string, error)

	//根据交易哈希获取交易详情
	GetTransaction(hash string) (*EosTransaction, error)

	GetTransactionByTxandBlock(tx string, block int64) (*EosTransaction, error)

	//检查交易是否被确认
	CheckTransactionConfirm(hash string) bool

	//获取块的交易详情
	GetBlockTransactions(from, to int64) []*EosTransactionData

	//获取ram信息
	GetRam(account string) (*EosRam, error)

	//获取带宽信息
	GetNet(account string) (*EosNet, error)

	//获取CPU信息
	GetCpu(account string) (*EosCpu, error)

	//eos购买ram
	BuyRam(account, key string, bytes int64) error

	//eos卖ram
	SellRam(account, key string, bytes int64) error

	//eos抵押带宽或CPU
	DelegateNetCpu(account, key string, netAmount, cpuAmount int64) error

	//eos赎回带宽或CPU
	UnDelegateNetCpu(account, key string, netAmount, cpuAmount int64) error
}

type chainClient struct {
	*eos.API
}

type EosTransactionData struct {
	ContractAccount string
	TxId  string
	From string
	To  string
	Amount int64
	Memo   string
	Symbol  string
	Block   int64
	BlockTime time.Time
}

type EosTransaction struct {
	Hash     string
	CommitAt time.Time
	Block    int64

	Data []*EosTransactionData
}

type EosSendTransaction struct {
	Key  string //交易账号的active私钥
	From string
	To   string
	Memo string

	Amount int64
}

type EosBlock struct {
	Num          int64
	Hash         string
	Timestamp    time.Time
	Transactions []*EosTransaction
}


type Block struct {
	BlockNumber int64  `json:"block_num"`
	Timestamp  string  `json:"timestamp"`
	Transactions []struct{
		Status string `json:"status"`
		Trx   struct{
			Id  string `json:"id"`
			Transaction struct{
				Actions []struct{
					Account string `json:"account"`
					Name    string `json:"name"`
					Data    struct{
						From  string `json:"from"`
						To    string `json:"to"`
						Quantity string `json:"quantity"`
						Memo   string `json:"memo"`
					} `json:"data"`
				} `json:"actions"`
			}  `json:"transaction"`
		} `json:"trx"`

	} `json:"transactions"`
}


type EosRam struct {
	Quota int64
	Used  int64
}

type EosNet struct {
	Weight    int64
	Used      int64
	Available int64
	Max       int64
}

type EosCpu struct {
	Weight    int64
	Used      int64
	Available int64
	Max       int64
}
