package eosio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
	"github.com/eoscanada/eos-go/token"
	"github.com/jie123108/glog"

	"net/http"
	"net/http/httputil"
)

const (
	EosAddrLen       = 12
	EosAddrSource    = "23456789abcdefghijkmnpqrstuvwxyz"
	EosContract      = "eosio.token"
	EosSymbol        = "EOS"
	CheckBlockTimeMs = 60 * 60
)

func NewEOSChain(host string) Chain {
	client := new(chainClient)
	client.API = eos.New(host)

	//判断是否连接成功
	r, err := client.API.GetInfo()
	if err != nil {
		panic(err)
	}
	bt, _ := json.Marshal(r)
	fmt.Println("api info: ", string(bt))
	return client
}

func (c *chainClient) NewAddress(account string) string {
	source := []byte(EosAddrSource)
	addr := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < EosAddrLen; i++ {
		addr = append(addr, source[r.Intn(len(source))])
	}

	return fmt.Sprintf("%s||%s", account, string(addr))
}

func (c *chainClient) GetAccountBalance(account string) (int64, error) {
	assets, err := c.API.GetCurrencyBalance(eos.AccountName(account), EosSymbol, eos.AccountName(EosContract))
	if err != nil {
		return 0, err
	} else if len(assets) == 0 {
		return 0, errors.New("account not exist")
	}

	return assets[0].Amount, nil
}

func (c *chainClient) GetBlockCount() (int64, error) {
	//获取信息
	info, err := c.API.GetInfo()
	if err != nil {
		return 0, err
	}

	return int64(info.HeadBlockNum), nil
}

//func (c *chainClient) GetBlockByIdOrNum(i interface{}) (*EosBlock, error) {
//	var block *eos.BlockResp
//	var err error
//	switch i.(type) {
//	case int64:
//		num, ok := i.(int64)
//		if !ok {
//			return nil, errors.New("block num type error")
//		}
//		block, err = c.API.GetBlockByNum(uint32(num))
//		if err != nil {
//			return nil, err
//		}
//		break
//	case string:
//		id, ok := i.(string)
//		if !ok {
//			return nil, errors.New("block id type error")
//		}
//		block, err = c.API.GetBlockByID(id)
//		if err != nil {
//			return nil, err
//		}
//		break
//	default:
//		return nil, errors.New("block id or num type error")
//	}
//
//	//信息转换
//	eblock := new(EosBlock)
//	eblock.Num = int64(block.BlockNum)
//	eblock.Hash = hex.EncodeToString(block.ID)
//	eblock.Timestamp = block.Timestamp.Time
//	for _, tx := range block.Transactions {
//
//		txId := hex.EncodeToString(tx.Transaction.ID)
//		transaction, err := c.GetTransaction(txId)
//		glog.Warn("aaaaa:        %+v", txId)
//
//		if err == nil {
//			eblock.Transactions = append(eblock.Transactions, transaction)
//		}
//	}
//
//	return eblock, nil
//}

func (c *chainClient) GetLastIrreversibleBlockNumber() (int64, error) {
	info, err := c.API.GetInfo()
	if err != nil {
		return 0, err
	}
	return int64(info.LastIrreversibleBlockNum), nil
}

func (c *chainClient) CheckBlock() error {
	info, err := c.API.GetInfo()
	if err != nil {
		return err
	}

	if time.Now().Unix()-info.HeadBlockTime.Unix() > CheckBlockTimeMs {
		return errors.New("block is too old")
	}
	return nil
}

func (c *chainClient) SendTransaction(tx *EosSendTransaction) (string, error) {
	if tx == nil {
		return "", errors.New("tx info is empty")
	}

	keyBag := eos.NewKeyBag()
	if err := keyBag.Add(tx.Key); err != nil {
		return "", err
	}
	c.SetSigner(keyBag)

	action := &eos.Action{
		Account: eos.AN(EosContract),
		Name:    eos.ActN("transfer"),
		Authorization: []eos.PermissionLevel{
			{Actor: eos.AccountName(tx.From), Permission: eos.PN("active")},
		},
		ActionData: eos.NewActionData(token.Transfer{
			From:     eos.AccountName(tx.From),
			To:       eos.AccountName(tx.To),
			Quantity: eos.NewEOSAsset(tx.Amount),
			Memo:     tx.Memo,
		}),
	}

	//交易发送
	transaction, err := c.API.SignPushActions(action)
	if err != nil {
		return "", err
	}

	return transaction.TransactionID, nil
}

func (c *chainClient) SendTokenTransaction(symbol string, precisison uint8, tx *EosSendTransaction) (string, error) {
	if tx == nil {
		return "", errors.New("tx info is empty")
	}

	keyBag := eos.NewKeyBag()
	if err := keyBag.Add(tx.Key); err != nil {
		return "", err
	}
	c.SetSigner(keyBag)

	action := &eos.Action{
		Account: eos.AN(EosContract),
		Name:    eos.ActN("transfer"),
		Authorization: []eos.PermissionLevel{
			{Actor: eos.AccountName(tx.From), Permission: eos.PN("active")},
		},
		ActionData: eos.NewActionData(token.Transfer{
			From:     eos.AccountName(tx.From),
			To:       eos.AccountName(tx.To),
			Quantity: eos.Asset{Amount: tx.Amount, Symbol: eos.Symbol{Precision: precisison, Symbol: symbol}},
			Memo:     tx.Memo,
		}),
	}

	//交易发送
	transaction, err := c.API.SignPushActions(action)
	if err != nil {
		return "", err
	}

	return transaction.TransactionID, nil
}

func (c *chainClient) SendTransactions(txs []*EosSendTransaction) (string, error) {
	if txs == nil || len(txs) == 0 {
		return "", errors.New("tx infos is empty")
	}

	keyBag := eos.NewKeyBag()
	if err := keyBag.Add(txs[0].Key); err != nil {
		return "", err
	}
	c.SetSigner(keyBag)

	var actions []*eos.Action
	for _, tx := range txs {
		action := &eos.Action{
			Account: eos.AN(EosContract),
			Name:    eos.ActN("transfer"),
			Authorization: []eos.PermissionLevel{
				{Actor: eos.AccountName(tx.From), Permission: eos.PN("active")},
			},
			ActionData: eos.NewActionData(token.Transfer{
				From:     eos.AccountName(tx.From),
				To:       eos.AccountName(tx.To),
				Quantity: eos.NewEOSAsset(tx.Amount),
				Memo:     tx.Memo,
			}),
		}
		actions = append(actions, action)
	}

	//交易发送
	transaction, err := c.API.SignPushActionsWithOpts(actions, nil)
	if err != nil {
		return "", err
	}

	return transaction.TransactionID, nil
}

func (c *chainClient) GetTransaction(hash string) (*EosTransaction, error) {
	tx, err := c.API.GetTransaction(hash)
	if err != nil {
		return nil, err
	}

	transaction := &EosTransaction{}
	transaction.Hash = hash
	transaction.Block = int64(tx.BlockNum)
	transaction.CommitAt = tx.BlockTime.Time

	if tx.Transaction.Transaction.Transaction == nil {
		return nil, fmt.Errorf("not found transaction.")
	}

	//解析所有动作
	for _, action := range tx.Transaction.Transaction.Actions {
		data := action.Data
		switch data.(type) {
		case map[string]interface{}:
			var tmp map[string]interface{}
			tmp, ok := data.(map[string]interface{})
			if !ok {
				return nil, errors.New("transaction action data type error")
			}
			transactionData := &EosTransactionData{}
			if tmp["from"] == nil || tmp["to"] == nil || tmp["memo"] == nil || tmp["quantity"] == nil {
				return nil, errors.New("transaction data error")
			}
			transactionData.From = tmp["from"].(string)
			transactionData.To = tmp["to"].(string)
			transactionData.Memo = tmp["memo"].(string)
			transactionData.Symbol = strings.ToUpper(strings.Split(tmp["quantity"].(string), " ")[1])

			transactionData.Amount, err = func(quantity string) (int64, error) {
				num := strings.Split(quantity, " ")[0]
				num = strings.Replace(num, ".", "", -1)
				num = strings.TrimLeft(num, "0")
				amount, err := strconv.ParseInt(num, 0, 64)
				return amount, err
			}(tmp["quantity"].(string))
			transaction.Data = append(transaction.Data, transactionData)
		default:
		}
	}
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (c *chainClient) GetTransactionByTxandBlock(hash string, block int64) (*EosTransaction, error) {

	var out *eos.TransactionResp
	err := Call(c, "history", "get_transaction", eos.M{"id": hash, "block_num_hint": uint32(block)}, &out)
	if err != nil {
		return nil, err
	}

	transaction := &EosTransaction{}
	transaction.Hash = hash
	transaction.Block = int64(out.BlockNum)
	transaction.CommitAt = out.BlockTime.Time

	if out.Transaction.Transaction.Transaction == nil {
		return nil, fmt.Errorf("not found transaction.")
	}

	//解析所有动作
	for _, action := range out.Transaction.Transaction.Actions {
		data := action.Data
		switch data.(type) {
		case map[string]interface{}:
			var tmp map[string]interface{}
			tmp, ok := data.(map[string]interface{})
			if !ok {
				return nil, errors.New("transaction action data type error")
			}
			transactionData := &EosTransactionData{}
			if tmp["from"] == nil || tmp["to"] == nil || tmp["memo"] == nil || tmp["quantity"] == nil {
				return nil, errors.New("transaction data error")
			}
			transactionData.ContractAccount = fmt.Sprintf("%s", action.Account)
			transactionData.From = tmp["from"].(string)
			transactionData.To = tmp["to"].(string)
			transactionData.Memo = tmp["memo"].(string)
			transactionData.Symbol = strings.ToUpper(strings.Split(tmp["quantity"].(string), " ")[1])

			transactionData.Amount, err = func(quantity string) (int64, error) {
				num := strings.Split(quantity, " ")[0]
				num = strings.Replace(num, ".", "", -1)
				num = strings.TrimLeft(num, "0")
				amount, err := strconv.ParseInt(num, 0, 64)

				return amount, err
			}(tmp["quantity"].(string))
			transaction.Data = append(transaction.Data, transactionData)
		default:
		}
	}
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (c *chainClient) CheckTransactionConfirm(hash string) bool {
	//获取最后不可逆块num
	lastBlock, err := c.GetLastIrreversibleBlockNumber()
	if err != nil {
		return false
	}

	//获取交易块的num
	transaction, err := c.GetTransaction(hash)
	if err != nil {
		return false
	}

	//交易块num<不可逆块num则交易被确认
	if transaction.Block < lastBlock {
		return true
	}

	return false
}

func (c *chainClient) GetBlockTransactions(from, to int64) []*EosTransactionData {
	if from > to || from < 0 {
		return nil
	}

	var transactions []*EosTransactionData
	for i := from; i <= to; i++ {

		var out interface{}
		err := Call(c, "chain", "get_block", eos.M{"block_num_or_id": fmt.Sprintf("%d", i)}, &out)
		if err != nil {
			glog.Error(err)
			continue
		}

		ss, _ := json.Marshal(out)
		var newBlock Block
		json.Unmarshal(ss, &newBlock)

		for _, detail := range newBlock.Transactions {
			if detail.Status == "executed" {
				transfer := EosTransactionData{}
				transfer.TxId = detail.Trx.Id
				for _, action := range detail.Trx.Transaction.Actions {
					if action.Name == "transfer" {
						transfer.ContractAccount = action.Account
						transfer.From = action.Data.From
						transfer.To = action.Data.To
						transfer.Memo = action.Data.Memo
						transfer.Block = newBlock.BlockNumber
						transfer.BlockTime, _ = time.Parse("2006-01-02T15:04:05", newBlock.Timestamp)
						amount, symbol := func(quantity string) (int64, string) {
							arr := strings.Split(quantity, " ")

							num := strings.Replace(arr[0], ".", "", -1)
							num = strings.TrimLeft(num, "0")
							amount, err := strconv.ParseInt(num, 0, 64)
							if err != nil {
								return 0, ""
							}

							return amount, arr[1]
						}(action.Data.Quantity)
						transfer.Amount = amount
						transfer.Symbol = strings.ToUpper(symbol)

						transactions = append(transactions, &transfer)
					}
				}
			}
		}
	}

	return transactions
}

func (c *chainClient) GetRam(account string) (*EosRam, error) {
	if account == "" {
		return nil, errors.New("account is empty")
	}

	//获取账户信息
	accountInfo, err := c.API.GetAccount(eos.AccountName(account))
	if err != nil {
		return nil, err
	}

	//信息读取
	ram := &EosRam{}
	ram.Quota = accountInfo.RAMQuota
	ram.Used = accountInfo.RAMUsage

	return ram, nil
}

func (c *chainClient) GetNet(account string) (*EosNet, error) {
	if account == "" {
		return nil, errors.New("account is empty")
	}

	//获取账户信息
	accountInfo, err := c.API.GetAccount(eos.AccountName(account))
	if err != nil {
		return nil, err
	}

	//信息读取
	net := &EosNet{}
	net.Weight = int64(accountInfo.NetWeight)
	net.Used = int64(accountInfo.NetLimit.Used)
	net.Available = int64(accountInfo.NetLimit.Available)
	net.Max = int64(accountInfo.NetLimit.Max)

	return net, nil
}

func (c *chainClient) GetCpu(account string) (*EosCpu, error) {
	if account == "" {
		return nil, errors.New("account is empty")
	}

	//获取账户信息
	accountInfo, err := c.API.GetAccount(eos.AccountName(account))
	if err != nil {
		return nil, err
	}

	//信息读取
	cpu := &EosCpu{}
	cpu.Weight = int64(accountInfo.CPUWeight)
	cpu.Used = int64(accountInfo.CPULimit.Used)
	cpu.Available = int64(accountInfo.CPULimit.Available)
	cpu.Max = int64(accountInfo.CPULimit.Max)

	return cpu, nil
}

func (c *chainClient) BuyRam(account, key string, bytes int64) error {
	if account == "" || key == "" || bytes <= 0 {
		return errors.New("buy ram info incomplete")
	}

	//认证
	keyBag := eos.NewKeyBag()
	if err := keyBag.Add(key); err != nil {
		return err
	}
	c.SetSigner(keyBag)

	//购买
	action := system.NewBuyRAMBytes(eos.AccountName(account), eos.AccountName(account), uint32(bytes))
	_, err := c.API.SignPushActions(action)
	if err != nil {
		fmt.Println("buy ram:", bytes)
		return err
	}

	return nil
}

func (c *chainClient) SellRam(account, key string, bytes int64) error {
	if account == "" || key == "" || bytes <= 0 {
		return errors.New("sell ram info incomplete")
	}

	//认证
	keyBag := eos.NewKeyBag()
	if err := keyBag.Add(key); err != nil {
		return err
	}
	c.SetSigner(keyBag)

	//卖出
	action := system.NewSellRAM(eos.AccountName(account), uint64(bytes))
	_, err := c.API.SignPushActions(action)
	if err != nil {
		return err
	}

	return nil
}

func (c *chainClient) DelegateNetCpu(account, key string, netAmount, cpuAmount int64) error {
	if account == "" || key == "" || (netAmount <= 0 && cpuAmount <= 0) {
		return errors.New("delegate account info incomplete")
	}

	//认证
	keyBag := eos.NewKeyBag()
	if err := keyBag.Add(key); err != nil {
		return err
	}

	//抵押
	action := system.NewDelegateBW(eos.AccountName(account), eos.AccountName(account), eos.NewEOSAsset(cpuAmount), eos.NewEOSAsset(netAmount), true)
	_, err := c.API.SignPushActions(action)
	if err != nil {
		return err
	}

	return nil
}

func (c *chainClient) UnDelegateNetCpu(account, key string, netAmount, cpuAmount int64) error {
	if account == "" || key == "" || (netAmount <= 0 && cpuAmount <= 0) {
		return errors.New("delegate account info incomplete")
	}

	//认证
	keyBag := eos.NewKeyBag()
	if err := keyBag.Add(key); err != nil {
		return err
	}

	//解除抵押
	action := system.NewUndelegateBW(eos.AccountName(account), eos.AccountName(account), eos.NewEOSAsset(cpuAmount), eos.NewEOSAsset(netAmount))
	_, err := c.API.SignPushActions(action)
	if err != nil {
		return err
	}

	return nil
}

func Call(api *chainClient, baseAPI string, endpoint string, body interface{}, out interface{}) error {
	jsonBody, err := enc(body)
	if err != nil {
		return err
	}

	targetURL := fmt.Sprintf("%s/v1/%s/%s", api.BaseURL, baseAPI, endpoint)

	req, err := http.NewRequest("POST", targetURL, jsonBody)
	if err != nil {
		return fmt.Errorf("NewRequest: %s", err)
	}

	if api.Debug {
		// Useful when debugging API calls
		requestDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("-------------------------------")
		fmt.Println(string(requestDump))
		fmt.Println("")
	}

	resp, err := api.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %s", req.URL.String(), err)
	}
	defer resp.Body.Close()

	var cnt bytes.Buffer
	_, err = io.Copy(&cnt, resp.Body)
	if err != nil {
		return fmt.Errorf("Copy: %s", err)
	}

	if resp.StatusCode == 404 {
		return errors.New("resource not found")
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("%s: status code=%d, body=%s", req.URL.String(), resp.StatusCode, cnt.String())
	}

	if api.Debug {
		fmt.Println("RESPONSE:")
		fmt.Println(cnt.String())
		fmt.Println("")
	}

	if err := json.Unmarshal(cnt.Bytes(), &out); err != nil {
		return fmt.Errorf("Unmarshal: %s", err)
	}

	return nil
}

func enc(v interface{}) (io.Reader, error) {
	if v == nil {
		return nil, nil
	}

	cnt, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	//fmt.Println("BODY", string(cnt))

	return bytes.NewReader(cnt), nil
}
