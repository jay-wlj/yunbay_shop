package share

import (
	"errors"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"math"
	"strings"
	"time"
	"yunbay/ybasset/common"
	"yunbay/ybasset/conf"
	"yunbay/ybasset/dao"
	"yunbay/ybasset/util"

	"github.com/shopspring/decimal"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

var m_drawstatus map[string]int

func init() {
	m_drawstatus = make(map[string]int)
	m_drawstatus["waiting"] = common.TX_STATUS_WAITING
	m_drawstatus["submitted"] = common.TX_STATUS_SUBMIT
	m_drawstatus["confirming"] = common.TX_STATUS_CONFIRM
	m_drawstatus["failed"] = common.TX_STATUS_FAILED
	m_drawstatus["success"] = common.TX_STATUS_SUCCESS
}

// 获取某种币的配置信息
func GetWithDrawConf(txType int) (ret conf.Fee) {
	for _, v := range conf.Config.Drawfees {
		if v.Type == txType {
			ret = v
			break
		}
	}
	return
}

// 保存用户充值地址
func addUserWalletAddress(user_id int64, address string, txType int) (err error) {
	now := time.Now().Unix()
	v := common.UserWallet{UserId: user_id, Type: int16(txType), BindAddress: address, CreateTime: now, UpdateTime: now}

	db := db.GetTxDB(nil)
	if err = db.Save(&v).Error; err != nil {
		db.Rollback()
		glog.Error("AddUserWalletAddress save fail! err=", err)
		return
	}
	db.Commit()
	// 更新到 address_source表中
	AddAddressChannel(address, common.CHANNEL_CHAIN)
	dao.RefleshUserWalletAddress(user_id)
	return
}

// 添加address_source
func AddAddressChannel(address string, channel int) (err error) {

	now := time.Now().Unix()
	v := common.AddressSource{Address: address, Channel: channel, CreateTime: now, UpdateTime: now}
	db := db.GetTxDB(nil)
	db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (address) DO update set channel=%v, update_time=%v", channel, now))
	if err = db.Save(&v).Error; err != nil {
		glog.Error("Wallet_Address save AddressSource fail! err=", err) // 不用返回
		db.Rollback()
	}
	db.Commit()
	return
}

// 获取用户充值地址
func GetAndSaveUserAddress(user_id int64) (address string, err error) {
	address, err = dao.GetUserWalletAddress(user_id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			address, err = util.GetUserWalletAddress(user_id, "kt")
			if err != nil {
				glog.Error("GetUserWalletAddress fail! err=", err)
				return
			}
			if err = addUserWalletAddress(user_id, address, common.CURRENCY_KT); err != nil {
				glog.Error("Wallet_Address save fail! err=", err)
				return
			}
			return
		}
		glog.Error("GetAndSaveUserAddress fail! err=", err, " user_id:", user_id)
		return
	}

	return
}

// 查询提币地址是否为合作渠道地址
func QueryAddressChannel(address string, txType int) (channel int, err error) {
	channel = common.CHANNEL_UNKNOW
	coin := common.GetCurrencyName(txType)
	// 优先查询是否为chain 地址
	var v common.AddressSource
	if err = db.GetDB().Find(&v, "address=?", address).Error; err == nil {
		channel = v.Channel
		return
	}
	var user_id int64
	// 先查询是否为yunex地址
	if user_id, err = util.IsYunexAddress(address, coin); err == nil && user_id > 0 {
		channel = common.CHANNEL_YUNEX
	} else if err != nil {
		glog.Error("QueryAddressChannel IsYunexAddress fail! err=", err)
	}

	// 查询是否yunex地址
	if channel == common.CHANNEL_UNKNOW {
		if user_id, err = util.IsHotCoinAddress(address); err == nil {
			channel = common.CHANNEL_HOTCOIN
		} else {
			glog.Error("QueryAddressChannel IsHotCoinAddress fail! err=", err)
		}
	}

	AddAddressChannel(address, channel) // 保存
	return
}

// 通过提币地址获取手续费用
func GetFeeByAddress(to_uid int64, address string, pChannel *int) (fees []conf.Fee, channel int, err error) {
	// 不能直接赋值给fees 不然修改fees相当于修改了conf.Config.Drawfees了
	fees = make([]conf.Fee, len(conf.Config.Drawfees))
	copy(fees, conf.Config.Drawfees)

	channel = common.CHANNEL_UNKNOW
	if pChannel != nil {
		channel = *pChannel
		if channel >= common.CHANNEL_ALIPAY {
			return
		}
	}

	if to_uid > 0 { // 平台用户
		channel = common.CHANNEL_CHAIN
	}
	if address != "" {
		db := db.GetDB()
		var addr common.AddressSource
		if err = db.Find(&addr, "address=?", address).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				glog.Error("GetFeeByAddress fail! err=", err)
				return
			}
			err = nil
			return
		}
		channel = addr.Channel
	}

	// 平台或第三方将手续费改成0
	if channel >= common.CHANNEL_CHAIN && channel <= common.CHANNEL_YUNEX {
		for i, _ := range fees {
			fees[i].Val = 0
		}
	}

	//glog.Info("GetFeeByAddress address:", address, " fees:", fees)
	return
}

// 获取用户当日可提取数量
func GetUserWithdarwAvalable(user_id int64, txType int, day_max *float64) (max float64, err error) {
	max = -1 // 默认不限额度
	if day_max != nil {
		max = *day_max
	}
	if txType == common.CURRENCY_RMB {
		txType = common.CURRENCY_KT // 提取rmb相当于提取kt
	}
	switch txType {
	case common.CURRENCY_YBT:
		var amount float64 = 0
		if amount, err = dao.GetUserDayWithdraw(user_id, txType); err != nil {
			glog.Error("GetUserWithdarwAvialbe fail! err=", err)
			return
		}
		// 当日剩余可提取金额
		var u common.UserAsset
		if err = db.GetDB().Find(&u, "user_id=?", user_id).Error; err != nil {
			glog.Error("GetUserWithdarwAvialbe fail! err=", err)
			return
		}
		max = u.NormalYbt*GetWithDrawConf(common.CURRENCY_YBT).DayMaxPercent - amount
		if max < 0 {
			max = 0
		}
	default:
		if max > 0 {
			var amount float64 = 0
			if amount, err = dao.GetUserDayWithdraw(user_id, txType); err != nil {
				glog.Error("GetUserWithdarwAvialbe fail! err=", err)
				return
			}
			max -= amount
			if max < 0 {
				max = 0
			}
			return
		}

	}

	return
}

// 根据充值地址获取用户信息
func GetUserInfoByRechargeAddress(address string) (v common.UserWallet, err error) {
	if err = db.GetDB().Find(&v, "bind_address=?", address).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			glog.Error("GetUserInfoByRechargeAddress fail! err=", err, "address=", address)
		}
		return
	}
	return
}

type WithDrawSt struct {
	OrderId    int64   `json:"order_id,string" binding:"required"`
	TxHash     string  `json:"tx_hash"`
	Status     string  `json:"status"`
	Reason     string  `json:"reason"`
	FeeInEther float64 `json:"fee_in_ether,string"`
	Channel    int     `json:"channel"`
	TxType     string  `json:"tx_type"`
}

func WithDrawCallbackHandle(db *db.PsqlDB, req WithDrawSt) (reason string, err error) {
	var v common.WithdrawFlow
	status, ok := m_drawstatus[strings.ToLower(req.Status)]
	if !ok {
		glog.Error("Chain_Withdraw_Callback fail! org status=", req.Status)
		err = fmt.Errorf("INVALID_ARGS")
		return
	}

	now := time.Now().Unix()
	today := time.Now().Format("2006-01-02")

	// 查找是审核通过的订单id
	if req.OrderId > 0 {
		if err = db.Find(&v, "id=? and status>=?", req.OrderId, common.TX_STATUS_CHECKPASS).Error; err != nil {
			glog.Error("Chain_Withdraw_Callback fail! err=", err, " order_id=", req.OrderId)
			reason = "INVALID_ORDER_ID"
			return
		}
	} else if req.TxHash != "" {
		if err = db.Find(&v, "txhash=? and status>=?", req.TxHash, common.TX_STATUS_CHECKPASS).Error; err != nil {
			glog.Error("Chain_Withdraw_Callback fail! err=", err, " order_id=", req.OrderId)
			reason = "INVALID_ORDER_ID"
			return
		}
	}

	// 效验txhash是否一样
	// if req.TxHash != "" && req.TxHash != v.Txhash {
	// 	glog.Error("Chain_Withdraw_Callback fail! err=", err, " txhash=", v.Txhash, " reqHash=", req.TxHash)
	// 	reason = "INVALID_ORDER_ID"
	// 	return
	// }
	// 提币渠道效验
	if req.Channel > 0 && req.Channel != v.Channel {
		s := fmt.Sprintln("Wallet_Withdraw_Callback not same channel! order_id=", req.OrderId)
		glog.Error(s)
		err = fmt.Errorf(s)
		reason = "INVALID_ORDER_ID"
		return
	}
	// 判断当前状态已是交易成功或失败
	if v.Status >= common.TX_STATUS_FAILED {
		if v.Status != status {
			s := fmt.Sprintln("Wallet_Withdraw_Callback fail! old status=", v.Status, " new status=", status)
			glog.Error(s)
			err = fmt.Errorf(s)
			reason = "INVALID_ORDER_ID"
		}

		return
	}

	// 先条件修改提币状态等信息
	res := db.Model(&common.WithdrawFlow{}).Where("id=? and status<?", v.Id, common.TX_STATUS_FAILED).Updates(map[string]interface{}{"txhash": req.TxHash, "status": status, "reason": req.Reason, "feeinether": req.FeeInEther, "update_time": now})
	if err = res.Error; err != nil {
		glog.Error("Chain_Withdraw_Callback update fail! err=", err)
		return
	}
	// 无记录更新
	if 0 == res.RowsAffected {
		glog.Error("Chain_Withdraw_Callback update fail! id=", req.OrderId)
		return
	}

	// 转帐成功或失败
	if status >= common.TX_STATUS_FAILED {
		var al common.AssetLock
		if err = db.Find(&al, "id=?", v.LockAssetId).Error; err != nil {
			glog.Error("Chain_Withdraw_Callback fail! err=", err)
			return
		}
		amount := v.Amount + v.Fee // 提币金额=实际提币金额+手续费
		// 转帐成功
		if status == common.TX_STATUS_SUCCESS {
			if !base.IsEqual(al.LockAmount, amount) {
				s := fmt.Sprintln("Chain_Withdraw_Callback fail! args amount:", v.Amount, " lock_amount:", al.LockAmount)
				glog.Error(s)
				err = fmt.Errorf(s)
				return
			}
			if _, err = WalletTransfer(db, v); err != nil {
				glog.Error("Chain_Withdraw_Callback  WalletTransfer fail! err=", err)
				return
			}

			dao.RefleshTotalWidthDraw() // 刷新总成功提币数
		} else {
			// 转帐失败 直接解冻原帐户资产即可
			ak := common.AssetLock{UserId: v.UserId, Type: v.TxType, LockType: common.ASSET_LOCK_WITHDRAW, LockAmount: -math.Abs(amount), Date: today, CreateTime: now, UpdateTime: now}
			if err = db.Save(&ak).Error; err != nil {
				glog.Error("Chain_Withdraw_Callback AssetLock save fail! err=", err)
				return
			}
		}
		// 更新用户已提取币种数量
		dao.RefleshUserDayWidthDraw(v.UserId, v.TxType)
	}
	return
}

// 内盘转帐
func WalletTransfer(db *db.PsqlDB, v common.WithdrawFlow) (id int64, err error) {

	today := time.Now().Format("2006-01-02")


	amount := v.Amount + v.Fee // 提币金额=实际提币金额+手续费
	if v.LockAssetId > 0 {
		// 解冻相应资产
		ak := common.AssetLock{UserId: v.UserId, Type: v.TxType, LockType: common.ASSET_LOCK_WITHDRAW, LockAmount: -math.Abs(amount), Date: today}
		if err = db.Save(&ak).Error; err != nil {
			glog.Error("Chain_Withdraw_Callback AssetLock save fail! err=", err, " user_id:", v.UserId)
			return
		}
	}

	// 先添加交易流水
	us := []common.UserAssetDetail{}
	us = append(us, common.UserAssetDetail{UserId: v.UserId, Amount: -math.Abs(amount), Type: v.TxType, TransactionType: common.KT_TRANSACTION_PICKUP, Date: today})

	// 将手续用转给平台
	if v.Fee > 0 {
		txFeeType := common.KT_TRANSACTION_FEE // kt和ybt的手续费用类型不一样
		if v.TxType == common.CURRENCY_YBT {
			txFeeType = common.YBT_TRANSACTION_FEE
		}
		us = append(us, common.UserAssetDetail{UserId: 0, Amount: math.Abs(v.Fee), Type: v.TxType, TransactionType: txFeeType, Date: today})
	}

	// 注:如果是提币到第三方平台,则需要将所提币打回到第三方指定帐户中 内盘操作
	if v.FlowType == common.FLOW_TYPE_YUNBAY { // 走的内盘方式调用第三方接口
		if v.Channel == common.CHANNEL_CHAIN {
			//var uw common.UserWallet
			var user_id int64
			if v.ToUserId > 0 {
				user_id = v.ToUserId
				// 国内版没有获取充值地址的必要
				// if err = db.Find(&uw, "user_id=?", v.ToUserId).Error; err != nil {
				// 	glog.Error("WalletTransfer fail! id=", v.Id, " not found user_id by bind_address:", v.Address)
				// 	return
				// }
			} else {
				var uw common.UserWallet
				if err = db.Find(&uw, "bind_address=?", v.Address).Error; err != nil {
					glog.Error("WalletTransfer fail! id=", v.Id, " not found user_id by bind_address:", v.Address)
					return
				}
				user_id = uw.UserId
			}

			//us = append(us, common.UserAssetDetail{UserId:uw.UserId, Amount:math.Abs(v.Amount), Type:v.TxType, TransactionType:common.KT_TRANSACTION_RECHARGE, Date:today, CreateTime:now, UpdateTime:now})
			u := common.UserAssetDetail{UserId: user_id, Amount: math.Abs(v.Amount), Type: v.TxType, TransactionType: common.KT_TRANSACTION_RECHARGE, Date: today}
			if err = db.Save(&u).Error; err != nil {
				glog.Error("WalletTransfer UserAssetDetail save fail! err=", err)
				return
			}
			id = u.Id
			// 查找转帐方地址
			var from_address string
			if addr, err1 := dao.GetUserWalletAddress(v.UserId); err1 == nil {
				from_address = addr
			} else {
				glog.Debug("WalletTransfer UserWallet fail! user_id=", v.UserId)
			}
			// 添加对方用户的充值记录流水
			r := common.RechargeFlow{FlowType: common.FLOW_TYPE_YUNBAY, Channel: common.CHANNEL_CHAIN, UserId: user_id, FromAddress: from_address, Address: v.Address, TxType: v.TxType, AssetId: u.Id, Amount: v.Amount, Date: today,TxHash: util.GetSnowflake().Generate().String()}
			if err = db.Save(&r).Error; err != nil {
				glog.Error("WalletTransfer RechargeFlow save fail! err=", err)
				return
			}

		} else if v.Channel > common.CHANNEL_CHAIN { // 注:如果是提币到第三方平台,则需要将所提币打回到第三方指定帐户中 内盘操作
			third_id := GetThirdWithdarwIdByChannel(v.Channel)
			if third_id <= 0 {
				s := fmt.Sprintln("third channel:", v.Channel, " user_id not define!")
				glog.Error(s)
				err = fmt.Errorf(s)
				return
			}

			us = append(us, common.UserAssetDetail{UserId: third_id, Type: v.TxType, TransactionType: common.KT_TRANSACTION_RECHARGE, Amount: math.Abs(v.Amount), Date: today})
		}
	} else { // 走的合约转帐

	}

	for _, v := range us {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("Chain_Withdraw_Callback UserAssetDetail save fail! err=", err)
			return
		}
	}

	// 更新用户已提取币种数量
	dao.RefleshUserDayWidthDraw(v.UserId, v.TxType)
	return
}

// 内盘转帐
func WalletTransferTo(db *db.PsqlDB, from, to int64, tx_type int, amount float64, country int, extinfos map[string]interface{}) (id int64, err error) {
	now := time.Now().Unix()
	v := common.WithdrawFlow{FlowType: common.FLOW_TYPE_YUNBAY, Channel: common.CHANNEL_CHAIN, Amount: amount, UserId: from, ToUserId: to, TxType: tx_type, Country: country, UpdateTime: now, CreateTime: now, CheckTime: now}
	v.Date = time.Now().Format("2006-01-02")
	v.Status = common.TX_STATUS_SUCCESS
	v.Maner = "system"
	v.Extinfos = extinfos
	id, err = WalletTransfer(db, v)
	return
}

type DrawSt struct {
	Man        bool    `json:"-"`
	Id         *int64  `json:"-"`
	UserId     int64   `json:"-"`
	UserType   int     `json:"-"`
	Token      string  `json:"-"`
	ToUserId   *int64  `json:"-"`
	Channel    *int    `json:"channel"`
	TxType     *int    `json:"tx_type" binding:"required,min=0"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Fee        float64 `json:"fee"`
	AddressId  int64   `json:"address_id"`
	Address    string  `json:"address"`
	ZJPassword string  `json:"zjpassword" binding:"required"`
	Code       string  `json:"code"`
	Country    int
	Extinfos   map[string]interface{} `json:"-"`
}

// 创建转帐订单逻辑
func WithDraw(db *db.PsqlDB, req DrawSt) (reason string, err error) {
	err = fmt.Errorf(yf.ERR_SERVER_ERROR)
	if (req.Channel != nil && *req.Channel > common.CHANNEL_BANK) || req.Amount < 0 || req.Amount < req.Fee {
		reason = yf.ERR_ARGS_INVALID
		err = fmt.Errorf(reason)
		return
	}

	txType := *req.TxType
	if req.Channel != nil && *req.Channel >= common.CHANNEL_ALIPAY && *req.Channel <= common.CHANNEL_BANK {
		txType = common.CURRENCY_RMB // 提现rmb换算成kt
		//req.Amount = req.Amount * GetRatio(common.GetCurrencyName(req.TxType), "kt")	// 兑换成相应的kt
	}
	// 判断提币开关是否打开
	if !CanWidthDraw(txType) {
		reason = common.ERR_FORBIDDEN_WITHDRAW
		err = fmt.Errorf(reason)
		return
	}
	user_id := req.UserId

	// 验证用户帐户是否冻结状态,否则禁止提现和交易
	var lock bool = false
	if lock, err = dao.UserAssetIsLocked(user_id); err != nil || lock {
		if lock {
			reason = common.ERR_USERASSET_LOCK
			err = fmt.Errorf(reason)
		}
		return
	}
	address := req.Address
	if req.AddressId > 0 {
		// 获取钱包地址
		var w common.WalletAddress
		if err = db.Find(&w, "id=?", req.AddressId).Error; err != nil {
			glog.Error("Wallet_Draw fail! err=", err)
			reason = common.ERR_WALLET_ADDRESS_NOT_FOUND
			return
		}
		address = w.Adddress
	}
	if address == "" {
		glog.Error("Wallet_Draw fail! address is empty!")
		reason = common.ERR_WALLET_ADDRESS_NOT_FOUND
		err = fmt.Errorf(reason)
		return
	}

	// 非提现通道需判断是否为自己的地址
	if req.Channel != nil && *req.Channel <= common.CHANNEL_YUNEX {
		if owner_address, err1 := dao.GetUserWalletAddress(user_id); err1 == nil {
			if owner_address != "" && owner_address == address {
				glog.Error("Wallet_Draw fail! ow.BindAddress == address!")
				reason = common.ERR_WALLET_ADDRESS_WITHDRAW_NOTOWNER
				err = fmt.Errorf(reason)
				return
			}
		}
	}

	// 先检测用户帐号可用余额是否足额
	v, err := dao.GetUserAsset(user_id, db.DB)
	if err != nil {
		glog.Error("Wallet_Draw fail! err=", err)
		return
	}
	// 提取kt
	fee, min, channel, err1 := CalcDrawFee(req.Channel, txType, address, req.Amount)
	if err1 != nil {
		err = err1
		glog.Error("Wallet_Draw fail! err=", err)
		return
	}

	// 后台接口调用忽略手续费及最小提币数
	if req.Man {
		fee = 0
		min = 0
	}

	if !base.IsEqual(fee, req.Fee) || req.Amount < req.Fee || (min > 0 && req.Amount < min) {
		// 手续费用与传进来的不符或提取金额小于手续费用 提示用户
		glog.Error("Wallet_Draw fail! req.fee=", req.Fee, " fee:", fee)
		reason = common.ERR_AMOUNT_INVALID
		err = fmt.Errorf(common.ERR_AMOUNT_INVALID)
		return
	}
	amount := req.Amount
	normal_amount := v.NormalKt
	switch txType {
	case common.CURRENCY_YBT:
		normal_amount = v.NormalYbt
	case common.CURRENCY_RMB:
		frate := GetRatio(common.GetCurrencyName(txType), "kt") // 兑换成相应的kt
		amount = amount * frate                                 // 兑换成相应的kt
		fee = fee * frate
		txType = common.CURRENCY_KT
	case common.CURRENCY_SNET:
		normal_amount = v.NormalSnet
	}

	if normal_amount < amount { // 可用余额不足
		glog.Error("Wallet_Draw fail! ERR_MONEY_NOT_MORE amount:", amount, " txType:", txType, " normal:", normal_amount)
		reason = common.ERR_MONEY_NOT_MORE
		err = fmt.Errorf(common.ERR_MONEY_NOT_MORE)
		return
	}

	// 判断是否超过可提金额
	if !req.Man {
		switch txType {
		case common.CURRENCY_YBT:
			// 仅限制商家用户的提币
			if req.UserType == common.USER_TYPE_BUSINESS {
				var max float64 = 0
				if max, err = GetUserWithdarwAvalable(user_id, txType, nil); err != nil {
					glog.Error("Wallet_Draw fail! err=", err)
					return
				}
				if max >= 0 && max < (req.Amount-fee) { // 限额并超过可提金额
					err = fmt.Errorf(common.ERR_AMOUNT_EXCEED)
					return
				}
			}
		default:
			tx_type := txType
			if req.Channel != nil && *req.Channel >= common.CHANNEL_ALIPAY && *req.Channel <= common.CHANNEL_BANK {
				tx_type = common.CURRENCY_RMB
			}
			// 限制用户提现额度
			draw := GetWithDrawConf(tx_type)
			var max float64 = 0
			if max, err = GetUserWithdarwAvalable(user_id, txType, draw.Max); err != nil {
				glog.Error("Wallet_Draw fail! err=", err)
				return
			}
			if max >= 0 && max < (amount-fee) { // 限额并超过可提金额
				reason = common.ERR_AMOUNT_EXCEED
				err = fmt.Errorf(reason)
				return
			}
		}
	}

	// 验证短信码及支付密码
	//token, _ := util.GetHeaderString(c, "X-Yf-Token")
	if req.ZJPassword != "" {
		if err = util.AuthSmsPasswrod(req.Token, req.Code, req.ZJPassword); err != nil {
			glog.Error("AuthSmsCode fail! err=", err)
			reason = err.Error()
			return
		}
	}

	// 先将提取的币置冻结状态
	now := time.Now().Unix()
	today := time.Now().Format("2006-01-02")
	u := common.AssetLock{UserId: user_id, Type: txType, LockAmount: amount, LockType: common.ASSET_LOCK_WITHDRAW, Date: today, CreateTime: now, UpdateTime: now}
	if err = db.Save(&u).Error; err != nil {
		glog.Error("Wallet_Draw fail! err=", err)
		return
	}

	// 注: 地址是yunbay或第三方合作的地址 优先走内盘交易
	wdType := common.FLOW_TYPE_CHAIN
	to_user_id := int64(-1)
	if channel >= 0 {
		wdType = common.FLOW_TYPE_YUNBAY
	}
	if channel == common.CHANNEL_CHAIN {
		// 查询入帐用户的id
		if req.ToUserId != nil {
			to_user_id = *req.ToUserId
		} else {
			var toUser common.UserWallet
			if e := db.First(&toUser, "bind_address=?", req.Address); e == nil {
				to_user_id = toUser.UserId
			}
		}
	}
	// 添加提币申请记录 提币金额需减去手续费用
	// 获取免审额度
	var ybt, kt float64 = 0, 0
	ybt, kt, err = util.GetDrawNoCheckAmountLimit(channel)
	status := common.TX_STATUS_INIT
	if channel >= common.CHANNEL_ALIPAY {
		status = common.TX_STATUS_CHECKPASS // 国内版自动审核通过 由财务手动打款
	}
	if err == nil {
		switch u.Type {
		case common.CURRENCY_KT:
			if kt > 0 && kt >= amount {
				status = common.TX_STATUS_CHECKPASS
				glog.Info("withdraw no check amount:", amount, " conf limit kt:", kt)
			}
		case common.CURRENCY_YBT:
			if ybt > 0 && ybt >= amount {
				status = common.TX_STATUS_CHECKPASS
				glog.Info("withdraw no check amount:", amount, " conf limit ybt:", ybt)
			}

		}
	}
	f := common.WithdrawFlow{Channel: channel, FlowType: wdType, UserId: user_id, ToUserId: to_user_id, LockAssetId: u.Id, TxType: u.Type, Address: address, Amount: base.Round2(amount-fee, base.FLOAT_MIN_PRECISION), Fee: fee, Status: status, Date: today, Country: req.Country, CreateTime: now, UpdateTime: now}
	if req.Extinfos != nil {
		f.Extinfos = req.Extinfos
	}
	if status == common.TX_STATUS_CHECKPASS {
		f.Maner = "system"
		f.CheckTime = now
	}

	if err = db.Save(&f).Error; err != nil {
		glog.Error("Wallet_Draw fail! err=", err)
		return
	}
	dao.RefleshTotalWidthDraw()                   // 刷新平台总提数量
	dao.RefleshUserDayWidthDraw(v.UserId, u.Type) // 刷新用户已提数量

	if req.Id != nil {
		*req.Id = f.Id
	}
	err = nil
	return
}

type TransferPaySt struct {
	common.TransferPool
	ZJPassword string `json:"zjpassword"`
	Token      string `json:"token"`
}

// 创建转帐订单逻辑
func (req *TransferPaySt) Transfer(db *db.PsqlDB) (err error) {
	err = fmt.Errorf(yf.ERR_SERVER_ERROR)

	txType := req.CoinType
	user_id := req.From

	// 验证用户帐户是否冻结状态,否则禁止提现和交易
	var lock bool = false
	if lock, err = dao.UserAssetIsLocked(user_id); err != nil || lock {
		if lock {
			err = fmt.Errorf(common.ERR_USERASSET_LOCK)
		}
		return
	}

	// 先检测用户帐号可用余额是否足额
	v, err := dao.GetUserAsset(user_id, db.DB)
	if err != nil {
		glog.Error("LotterysPay fail! err=", err)
		return
	}

	amount := req.Amount
	normal_amount := v.NormalKt
	switch txType {
	case common.CURRENCY_YBT:
		normal_amount = v.NormalYbt
	case common.CURRENCY_RMB:
		frate := GetRatio(common.GetCurrencyName(txType), "kt") // 兑换成相应的kt
		amount = amount.Mul(decimal.NewFromFloat(frate))        // 兑换成相应的kt
	case common.CURRENCY_SNET:
		normal_amount = v.NormalSnet
	}

	if amount.GreaterThan(decimal.NewFromFloat(normal_amount)) { // 可用余额不足
		glog.Error("LotterysPay fail! ERR_MONEY_NOT_MORE amount:", amount, " txType:", txType, " normal:", normal_amount)
		err = fmt.Errorf(common.ERR_MONEY_NOT_MORE)
		return
	}

	// 验证短信码及支付密码
	//token, _ := util.GetHeaderString(c, "X-Yf-Token")
	if req.ZJPassword != "" {
		if err = util.AuthUserZJPassword(req.Token, req.ZJPassword); err != nil {
			glog.Error("LotterysPay AuthSmsCode fail! err=", err)
			return
		}
	}

	tType := common.TRANSACTION_CONSUME
	switch req.CoinType {
	case common.CURRENCY_KT:
		tType = common.KT_TRANSACTION_CONSUME
	case common.CURRENCY_YBT:
		tType = common.YBT_TRANSACTION_CONSUME
	}
	us := []common.UserAssetDetail{}

	fAmount, _ := req.Amount.Float64()
	// 将用户的资产转移给抽奖帐号
	us = append(us, common.UserAssetDetail{UserId: req.From, Type: req.CoinType, TransactionType: tType, Amount: -fAmount, Date: base.GetCurDay()})
	us = append(us, common.UserAssetDetail{UserId: req.To, Type: req.CoinType, TransactionType: tType, Amount: fAmount, Date: base.GetCurDay()})

	for _, u := range us {
		if err = db.Save(&u).Error; err != nil {
			glog.Error("LotterysPay fail! err=", err)
			return
		}
	}

	req.Id = 0
	req.Status = common.STATUS_OK
	if err = db.Save(&req.TransferPool).Error; err != nil {
		glog.Error("LotterysPay fail! err=", err)
		return
	}
	err = nil
	return
}

type TransferRefundSt struct {
	Id  int64  `json:"id"`
	Key string `json:"key"`
}

func (req *TransferRefundSt) Refund(db *db.PsqlDB) (err error) {
	if 0 == req.Id && "" == req.Key {
		err = errors.New(yf.ERR_ARGS_INVALID)
		return
	}
	db.DB = db.Where("status=?", common.STATUS_OK)
	if req.Id > 0 {
		db.DB = db.Where("id=?", req.Id)
	}
	if req.Key != "" {
		db.DB = db.Where("key=?", req.Key)
	}
	vs := []common.TransferPool{}
	if err = db.Find(&vs).Error; err != nil {
		glog.Error("Refund fail! err=", err)
		return
	}
	if len(vs) == 0 {
		return
	}
	// 原路退还

	us := []common.UserAssetDetail{}
	for _, v := range vs {
		fAmount, _ := v.Amount.Float64()

		tType := common.TRANSACTION_CONSUME
		switch v.CoinType {
		case common.CURRENCY_KT:
			tType = common.KT_TRANSACTION_CONSUME
		case common.CURRENCY_YBT:
			tType = common.YBT_TRANSACTION_CONSUME
		}

		us = append(us, common.UserAssetDetail{UserId: v.To, Type: v.CoinType, TransactionType: tType, Amount: -fAmount, Date: base.GetCurDay()})
		us = append(us, common.UserAssetDetail{UserId: v.From, Type: v.CoinType, TransactionType: tType, Amount: fAmount, Date: base.GetCurDay()})
	}

	// 转帐
	for _, v := range us {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("Refund fail! err=", err)
			return
		}

	}

	// 更新转帐状态为退还状态
	if err = db.Model(&vs).Updates(base.Maps{"status": common.STATUS_FAIL}).Error; err != nil {
		glog.Error("Refund fail! err=", err)
		return
	}
	return
}
