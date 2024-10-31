package common

import (
	"github.com/jay-wlj/gobaselib/db"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// type UserLock struct {
// 	Id int64 `json:"-" gorm:"primary_key:id"`
// 	UserId int64 `json:"user_id" gorm:"column:user_id"`
// 	Status int `json:"status"`
// 	CreateTime int64 `json:"create_time" gorm:"column:create_time"`
// 	UpdateTime int64 `json:"update_time" gorm:"column:update_time"`
// }

// func (UserLock) TableName() string {
// 	return "user_lock"
// }

type AssetLock struct {
	Id         int64   `json:"id" gorm:"primary_key:id"`
	UserId     int64   `json:"user_id" gorm:"column:user_id"`
	Type       int     `json:"type" gorm:"column:type"`
	LockType   int     `json:"lock_type" gorm:"column:lock_type"`
	LockAmount float64 `json:"lock_amount" gorm:"column:lock_amount"`
	UnlockTime int64   `json:"unlock_time" gorm:"column:unlock_time"`
	Date       string  `json:"date"`
	CreateTime int64   `json:"create_time" gorm:"column:create_time"`
	UpdateTime int64   `json:"-" gorm:"column:update_time"`
}

func (AssetLock) TableName() string {
	return "asset_lock"
}

type UserAsset struct {
	Id     int64 `json:"id" gorm:"primary_key:id"`
	UserId int64 `json:"user_id" gorm:"column:user_id"`
	// Date string `json:"date"`
	TotalYbt  float64 `json:"total_ybt" gorm:"column:total_ybt"`
	NormalYbt float64 `json:"normal_ybt" gorm:"column:normal_ybt"`
	LockYbt   float64 `json:"lock_ybt" gorm:"column:lock_ybt"`
	//LockYbtBonus float64 `json:"lock_ybt_bonus" gorm:"column:lock_ybt_bonus"`
	FreezeYbt  float64         `json:"freeze_ybt" gorm:"column:freeze_ybt"`
	TotalKt    float64         `json:"total_kt" gorm:"column:total_kt"`
	NormalKt   float64         `json:"normal_kt" gorm:"column:normal_kt"`
	LockKt     float64         `json:"lock_kt" gorm:"column:lock_kt"`
	TotalSnet  float64         `json:"total_snet"`
	NormalSnet float64         `json:"normal_snet"`
	LockSnet   float64         `json:"lock_snet"`
	Status     int16           `json:"status" gorm:"status"`
	CreateTime int64           `json:"create_time" gorm:"column:create_time"`
	UpdateTime int64           `json:"update_time" gorm:"column:update_time"`
	RmbRatio   interface{}     `json:"rmbratio,omitempty" gorm:"-"`
	AssetList  []UserAssetType `json:"asset_list,omitempty" gorm:"-"`
}

func (UserAsset) TableName() string {
	return "user_asset"
}

type UserAssetDetail struct {
	db.Model
	UserId          int64   `json:"user_id" gorm:"column:user_id"`
	Type            int     `json:"type" gorm:"column:type"`
	TransactionType int     `json:"transaction_type" gorm:"column:transaction_type"`
	Amount          float64 `json:"amount" gorm:"column:amount"`
	// LockAmount float64 `json:"lock_amount" gorm:"column:lock_amount"`
	// LockType int `json:"lock_type" gorm:"column:lock_type"`
	// UnlockTime int `json:"unlock_time" gorm:"column:unlock_time"`
	Date string `json:"date"`
}

func (UserAssetDetail) TableName() string {
	return "user_asset_detail"
}

type YBAsset struct {
	Id                int64   `json:"id" gorm:"primary_key:id"`
	TotalAmount       float64 `json:"total_kt" gorm:"column:total_kt"`
	TotalProfit       float64 `json:"total_kt_profit" gorm:"column:total_kt_profit"`
	TotalIssuedYbt    float64 `json:"total_issue_ybt" gorm:"column:total_issue_ybt"`
	TotalDestroyedYbt float64 `json:"total_destroyed_ybt" gorm:"column:total_destroyed_ybt"`
	TotalMining       float64 `json:"total_mining"`
	TotalAirDrop      float64 `json:"total_air_drop"`
	TotalAirUnlock    float64 `json:"total_air_unlock"`
	TotalAirRecover   float64 `json:"total_air_recover"`
	TotalActivity     float64 `json:"total_activity"`
	TotalProject      float64 `json:"total_project"`
	TotalPerynbay     float64 `json:"total_perynbay"`
	//TotalLockdYbt float64 `json:"total_lock_ybt" gorm:"total_lock_ybt"`
	//TotalFanliAmount float64 `json:"total_fanli_ybt" gorm:"column:total_fanli_ybt"`
	//TotalLockFanliYbt float64 `json:"total_lock_fanli_ybt" gorm:"column:total_lock_fanli_ybt"`
	Date       string `json:"date"`
	CreateTime int64  `json:"create_time" gorm:"column:create_time"`
	UpdateTime int64  `json:"update_time" gorm:"column:update_time"`
}

func (YBAsset) TableName() string {
	return "yunbay_asset"
}

type YBAssetProfitType struct {
	Id           int64   `json:"id" gorm:"primary_key:id"`
	CurrencyType int     `json:"currency_type"`
	Amount       float64 `json:"amount"`
	Profit       float64 `json:"profit"`
	Date         string  `json:"date"`
}

func (YBAssetProfitType) TableName() string {
	return "yunbay_asset_type"
}

type YBAssetDetail struct {
	Id           int64   `json:"id" gorm:"primary_key:id"`
	Amount       float64 `json:"amount"`
	Profit       float64 `json:"profit"`
	Mining       float64 `json:"mining"`
	Project      float64 `json:"project"`
	AirUnlock    float64 `json:"air_unlock"`
	Activity     float64 `json:"activity"`
	AirDrop      float64 `json:"air_drop"`
	LockYbt      float64 `json:"lock_ybt"`
	AirRecover   float64 `json:"air_recover"`
	IssueYbt     float64 `json:"issue_ybt" gorm:"column:issue_ybt"`
	DestoryedYbt float64 `json:"destroyed_ybt" gorm:"column:destoryed_ybt"`
	FreezeYbt    float64 `json:"freeze_ybt"`
	BonusYbt     float64 `json:"bonus_ybt" gorm:"column:bonus_ybt"`
	Perynbay     float64 `json:"perynbay" gorm:"column:perynbay"`
	Date         string  `json:"date"`
	Period       int64   `json:"period"`
	Difficult    float64 `json:"difficult"`
	Miners       int     `json:"miners"`
	Bonusers     int     `json:"bonusers"`
	KtStatus     int     `json:"kt_status"`
	YbtStatus    int     `json:"ybt_status"`
	CreateTime   int64   `json:"create_time" gorm:"column:create_time"`
	UpdateTime   int64   `json:"update_time" gorm:"column:update_time"`
	ProfitRate   float64 `json:"profit_rate" gorm:"-"`
}

func (YBAssetDetail) TableName() string {
	return "yunbay_asset_detail"
}

type YBAssetPool struct {
	db.Model
	OrderId      int64    `json:"order_id" gorm:"column:order_id"`
	CurrencyType int      `json:"currency_type" gorm:"column:currency_type"`
	PayerUserId  int64    `json:"payer_userid" gorm:"column:payer_userid"`
	PayAmount    float64  `json:"pay_amount" gorm:"column:pay_amount"`
	SellerUserId int64    `json:"seller_userid" gorm:"column:seller_userid"`
	SellerAmount float64  `json:"seller_amount" gorm:"column:seller_amount"`
	RebatAmount  float64  `json:"rebat_amount" gorm:"column:rebat_amount"`
	SellerKt     float64  `json:"seller_kt"`
	Status       int      `json:"status"`
	Country      int      `json:"country"`
	PublishArea  int      `json:"publish_area"`
	Extinfos     db.JSONB `json:"extinfos" gorm:"not null;default:'{}'"`
	Date         string   `json:"date"`
}

func (YBAssetPool) TableName() string {
	return "yunbay_asset_pool"
}

type Ordereward struct {
	Id           int64   `json:"id" gorm:"primary_key:id"`
	OrderId      int64   `json:"order_id" gorm:"column:order_id"`
	Ybt          float64 `json:"ybt" gorm:"column:ybt"`
	BuyerUserId  int64   `json:"buyer_userid" gorm:"column:buyer_userid"`
	BuyerYbt     float64 `json:"buyer_ybt" gorm:"column:buyer_ybt"`
	SellerUserId int64   `json:"seller_userid" gorm:"column:seller_userid"`
	SellerYbt    float64 `json:"seller_ybt" gorm:"column:seller_ybt"`
	SellerStatus int     `json:"seller_status"`
	ReUserId     int64   `json:"recommender_userid" gorm:"column:recommender_userid"`
	ReYbt        float64 `json:"recommender_ybt" gorm:"column:recommender_ybt"`
	Re2UserId    int64   `json:"recommender2_userid" gorm:"column:recommender2_userid"`
	Re2Ybt       float64 `json:"recommender2_ybt" gorm:"column:recommender2_ybt"`
	YunbayUserId int64   `json:"yunbay_userid" gorm:"column:yunbay_userid"`
	YunbayYbt    float64 `json:"yunbay_ybt" gorm:"column:yunbay_ybt"`
	Date         string  `json:"date" gorm:"column:date"`
	Valid        int     `json:"valid" gorm:"valid"`
	CreateTime   int64   `json:"create_time" gorm:"column:create_time"`
	UpdateTime   int64   `json:"update_time" gorm:"column:update_time"`
}

func (Ordereward) TableName() string {
	return "ordereward"
}

// type YbtIssue struct {
// 	Id int64 `json:"id" gorm:"primary_key:id"`
// 	Amount float64 `json:"amount"`
// 	TotalAmount float64 `json:"total_amount"`
// 	Date string `json:"date"`
// 	CreateTime int64 `json:"create_time" gorm:"column:create_time"`
// 	UpdateTime int64 `json:"-" gorm:"column:update_time"`
// }

// func (YbtIssue) TableName() string {
// 	return "ybt_issue"
// }

type BonusKtDetail struct {
	Id         int64   `json:"-" gorm:"primary_key:id"`
	UserId     int64   `json:"-" gorm:"column:user_id"`
	Ybt        float64 `json:"ybt"`
	Kt         float64 `json:"kt"`
	Date       string  `json:"date"`
	CreateTime int64   `json:"-" gorm:"column:create_time"`
	UpdateTime int64   `json:"-" gorm:"column:update_time"`
}

func (BonusKtDetail) TableName() string {
	return "bonus_kt"
}

type BonusYbtDetail struct {
	Id         int64    `json:"-" gorm:"primary_key:id"`
	UserId     int64    `json:"-" gorm:"column:user_id"`
	Infos      db.JSONB `json:"infos"`
	TotalYbt   float64  `json:"total_ybt"`
	Date       string   `json:"date"`
	CreateTime int64    `json:"-" gorm:"column:create_time"`
	UpdateTime int64    `json:"-" gorm:"column:update_time"`
}

func (BonusYbtDetail) TableName() string {
	return "bonus_ybt"
}

type YbtBonusTypeAmount struct {
	Consume  float64 `json:"consume"`
	Activity float64 `json:"activity"`
	Invite   float64 `json:"invite"`
	Seller   float64 `json:"seller"`
	AirDrop  float64 `json:"airdrop"`
}

type TradeFlow struct {
	Id          int64   `json:"-" gorm:"primary_key:id"`
	TotalOrders int64   `json:"total_orders" gorm:"column:total_orders"`
	PayedOrders int64   `json:"payed_orders" gorm:"column:payed_orders"`
	TotalPayers int64   `json:"total_payers" gorm:"column:total_payers"`
	TotalAmount float64 `json:"total_amount" gorm:"column:total_amount"`
	TotalProfit float64 `json:"total_profit" gorm:"column:total_profit"`
	Perynbay    float64 `json:"perynbay"`
	Country     int     `json:"country"`
	Date        string  `json:"date"`
	CreateTime  int64   `json:"-" gorm:"column:create_time"`
	UpdateTime  int64   `json:"update_time" gorm:"column:update_time"`
}

func (TradeFlow) TableName() string {
	return "tradeflow"
}

type TransferPool struct {
	db.Model
	Key      string          `json:"key"`
	CoinType int             `json:"coin_type"`
	From     int64           `json:"from"`
	To       int64           `json:"to"`
	Amount   decimal.Decimal `json:"amount"`
	Status   int             `json:"status"`
}

func (TransferPool) TableName() string {
	return "transfer_pool"
}

type UserWallet struct {
	Id          int64  `json:"id" gorm:"primary_key:id"`
	Type        int16  `json:"type"`
	UserId      int64  `json:"user_id" gorm:"column:user_id"`
	BindAddress string `json:"bind_address" gorm:"column:bind_address"`
	CreateTime  int64  `json:"create_time" gorm:"column:create_time"`
	UpdateTime  int64  `json:"-" gorm:"column:update_time"`
}

func (UserWallet) TableName() string {
	return "user_wallet"
}

type WalletAddress struct {
	Id         int64  `json:"id" gorm:"primary_key:id"`
	Type       uint16 `json:"type"`
	UserId     int64  `json:"user_id" gorm:"column:user_id"`
	Name       string `json:"name"`
	Adddress   string `json:"adddress"`
	Default    bool   `json:"default"`
	CreateTime int64  `json:"create_time" gorm:"column:create_time"`
	UpdateTime int64  `json:"-" gorm:"column:update_time"`
}

func (WalletAddress) TableName() string {
	return "wallet_address"
}

type WithdrawFlow struct {
	Id          int64    `json:"id" gorm:"primary_key:id"`
	Channel     int      `json:"channel"`
	FlowType    int      `json:"flow_type" gorm:"column:flow_type"`
	UserId      int64    `json:"user_id" gorm:"column:user_id"`
	LockAssetId int64    `json:"-" gorm:"column:lock_asset_id"`
	TxType      int      `json:"tx_type" gorm:"column:tx_type"`
	ToUserId    int64    `json:"to_user_id"`
	Address     string   `json:"address"`
	Amount      float64  `json:"amount"`
	Fee         float64  `json:"fee"`
	FeeInEther  float64  `json:"feeinether" gorm:"column:feeinether"`
	Txhash      string   `json:"txhash"`
	Status      int      `json:"status"`
	Reason      string   `json:"reason"`
	Date        string   `json:"date"`
	Country     int      `json:"country"`
	Maner       string   `json:"maner" gorm:"column:maner"`
	CheckTime   int64    `json:"check_time" gorm:"column:check_time"`
	CreateTime  int64    `json:"create_time" gorm:"column:create_time"`
	UpdateTime  int64    `json:"update_time" gorm:"column:update_time"`
	Extinfos    db.JSONB `json:"extinfos" gorm:"not null;default:'{}'"`
}

func (WithdrawFlow) TableName() string {
	return "withdraw_flow"
}

type RechargeFlow struct {
	db.Model
	FlowType    int     `json:"flow_type" gorm:"column:flow_type"`
	Channel     int     `json:"channel"`
	UserId      int64   `json:"user_id,string" gorm:"column:user_id"`
	AssetId     int64   `json:"asset_id"`
	TxHash      string  `json:"tx_hash" gorm:"column:txhash"`
	FromAddress string  `json:"from_address" gorm:"column:from_address"`
	TxType      int     `json:"tx_type" gorm:"column:tx_type"`
	Address     string  `json:"address" gorm:"column:address"`
	Amount      float64 `json:"amount"`
	BlockTime   string  `json:"-"`
	Date        string  `json:"date"`
	Country     int     `json:"country"`
}

func (RechargeFlow) TableName() string {
	return "recharge_flow"
}

type Ybt struct {
	Id             int64   `json:"id" gorm:"primary_key:id"`
	Reward         float64 `json:"reward"`
	Project        float64 `json:"project"`
	Minepool       float64 `json:"minepool"`
	NormalReward   float64 `json:"normal_reward"`
	LockReward     float64 `json:"lock_reward"`
	LockProject    float64 `json:"lock_project"`
	LockMinepool   float64 `json:"lock_minepool"`
	UnlockReward   float64 `json:"unlock_reward"`
	UnlockProject  float64 `json:"unlock_project"`
	UnlockMinepool float64 `json:"unlock_minepool"`
	CreateTime     int64   `json:"create_time"`
	UpdateTime     int64   `json:"-"`
}

func (Ybt) TableName() string {
	return "ybt"
}

type YbtDayFlow struct {
	Id             int64   `json:"id" gorm:"primary_key:id"`
	Reward         float64 `json:"reward"`
	UnlockReward   float64 `json:"unlock_reward"`
	UnlockProject  float64 `json:"unlock_project"`
	UnlockMinepool float64 `json:"unlock_minepool"`
	Date           string  `json:"date"`
	CreateTime     int64   `json:"create_time"`
	UpdateTime     int64   `json:"-"`
}

func (YbtDayFlow) TableName() string {
	return "ybt_day_flow"
}

type YbtFlow struct {
	Id          int64   `json:"id" gorm:"primary_key:id"`
	Type        int     `json:"type"`
	UserId      int64   `json:"user_id"`
	Amount      float64 `json:"amount"`
	UserAssetId int64   `json:"user_asset_id"`
	Date        string  `json:"date"`
	Maner       string  `json:"maner"`
	CreateTime  int64   `json:"create_time"`
	UpdateTime  int64   `json:"-"`
}

func (YbtFlow) TableName() string {
	return "ybt_flow"
}

type RewardRecord struct {
	Id          int64   `json:"id" gorm:"primary_key:id"`
	Type        int     `json:"type"`
	ReleaseType int     `json:"release_type"`
	Fixdays     int     `json:"fixdays"`
	UserId      int64   `json:"user_id"`
	InviteId    int64   `json:"invite_id"`
	Amount      float64 `json:"amount"`
	Reason      string  `json:"reason"`
	Maner       string  `json:"maner"`
	Date        string  `json:"date"`
	Status      int     `json:"status"`
	Lock        bool    `json:"lock"`
	CreateTime  int64   `json:"create_time" gorm:"column:create_time"`
	UpdateTime  int64   `json:"update_time" gorm:"column:update_time"`
}

func (RewardRecord) TableName() string {
	return "reward_record"
}

type KtBonusDetail struct {
	UserAsset
	Mining       float64 `json:"mining"`
	AirUnlock    float64 `json:"air_unlock"`
	Project      float64 `json:"project"`
	BonusYbt     float64 `json:"bonus_ybt"`
	BonusPercent float64 `json:"bonus_percent"`
	KtBonus      float64 `json:"kt_bonus"`
	CheckStatus  int     `json:"check_status"`
	Date         string  `json:"date"`
	ThirdBonus   int     `json:"third_bonus"`
}

func (KtBonusDetail) TableName() string {
	return "kt_bonus_detail"
}

type YbtUnlockDetail struct {
	Id           int64   `json:"id" gorm:"primary_key:id"`
	UserId       int64   `json:"user_id"`
	Mining       float64 `json:"mining"`
	Consume      float64 `json:"consume"`
	Sale         float64 `json:"sale"`
	Invite       float64 `json:"invite"`
	Activity     float64 `json:"activity"`
	AirDrop      float64 `json:"air_drop"`
	AirUnlock    float64 `json:"air_unlock"`
	Project      float64 `json:"project"`
	TotalUnlock  float64 `json:"total_unlock"`
	YbtPercent   float32 `json:"ybt_percent"`
	Rebat        float64 `json:"rebat"`
	RebatPercent float32 `json:"rebat_percent"`
	SaleRebat    float64 `json:"sale_rebat"`
	SalePercent  float32 `json:"sale_percent"`
	CheckStatus  int     `json:"check_status"`
	Date         string  `json:"date"`
	CreateTime   int64   `json:"create_time" gorm:"column:create_time"`
	UpdateTime   int64   `json:"update_time" gorm:"column:update_time"`
}

func (YbtUnlockDetail) TableName() string {
	return "ybt_unlock_detail"
}

type ThirdBonus struct {
	Id         int64   `json:"id" gorm:"primary_key:id"`
	Uid        int64   `json:"uid" gorm:"column:uid"`
	Tid        int64   `json:"tid" gorm:"column:tid"`
	Ybt        float64 `json:"ybt"`
	Kt         float64 `json:"kt"`
	Status     int     `json:"status" gorm:"column:status"`
	Date       string  `json:"date" gorm:"column:date"`
	CreateTime int64   `json:"-" gorm:"column:create_time"`
	UpdateTime int64   `json:"-" gorm:"column:update_time"`
}

func (ThirdBonus) TableName() string {
	return "third_bonus"
}

type AddressSource struct {
	Id         int64  `json:"id" gorm:"primary_key:id"`
	Address    string `json:"address"`
	Channel    int    `json:"channel"`
	CreateTime int64  `json:"create_time" gorm:"column:create_time"`
	UpdateTime int64  `json:"-" gorm:"column:update_time"`
}

func (AddressSource) TableName() string {
	return "address_source"
}

type RmbRecharge struct {
	Id          int64         `json:"id" gorm:"primary_key:id"`
	Channel     int           `json:"channel"`
	UserId      int64         `json:"user_id,string" gorm:"column:user_id"`
	OrderIds    pq.Int64Array `json:"order_ids"`
	Subject     string        `json:"subject"`
	AssetId     int64         `json:"asset_id"`
	TxHash      string        `json:"tx_hash" gorm:"column:txhash"`
	TxType      int           `json:"tx_type" gorm:"column:tx_type"`
	Account     string        `json:"address"`
	Amount      float64       `json:"amount"`
	Date        string        `json:"date"`
	Status      int           `json:"status"`
	OrderStatus int           `json:"order_status"`
	OverTime    int64         `json:"over_time"`
	CreateTime  int64         `json:"create_time" gorm:"column:create_time"`
	UpdateTime  int64         `json:"-" gorm:"column:update_time"`
}

func (RmbRecharge) TableName() string {
	return "rmb_recharge"
}

type YBAssetAll struct {
	YBAssetDetail
	TotalAmount       float64 `json:"total_kt" gorm:"column:total_kt"`
	TotalProfit       float64 `json:"total_kt_profit" gorm:"column:total_kt_profit"`
	TotalIssuedYbt    float64 `json:"total_issue_ybt" gorm:"column:total_issue_ybt"`
	TotalDestroyedYbt float64 `json:"total_destroyed_ybt" gorm:"column:total_destroyed_ybt"`
	TotalMining       float64 `json:"total_mining"`
	TotalAirDrop      float64 `json:"total_air_drop"`
	TotalAirUnlock    float64 `json:"total_air_unlock"`
	TotalAirRecover   float64 `json:"total_air_recover"`
	TotalActivity     float64 `json:"total_activity"`
	TotalProject      float64 `json:"total_project"`
	TotalPerynbay     float64 `json:"total_perynbay"`
}

func (YBAssetAll) TableName() string {
	return "ybasset_all"
}

type Voucher struct {
	Id         int64       `json:"id" gorm:"primary_key:id"`
	UserId     int64       `json:"-"`
	Type       int         `json:"type" view:"record"`
	Amount     float64     `json:"amount"`
	UnlockTime int64       `json:"unlock_time,omitempty"`
	CreateTime int64       `json:"-"`
	UpdateTime int64       `json:"update_time"`
	Info       interface{} `json:"info" gorm:"-"`
}

func (Voucher) TableName() string {
	return "voucher"
}

type VoucherInfo struct {
	Id         int64       `json:"id" gorm:"primary_key:id" view:"man"`
	Type       int         `json:"type" view:"record"`
	Title      string      `json:"title" view:"record"`
	Context    string      `json:"context"`
	ProductId  int64       `json:"product_id"`
	CreateTime int64       `json:"create_time" view:"man"`
	UpdateTime int64       `json:"update_time" view:"man"`
	Voucher    interface{} `json:"voucher,omitempty" gorm:"-"`
}

func (VoucherInfo) TableName() string {
	return "voucher_info"
}

type VoucherRecord struct {
	Id          int64       `json:"-" gorm:"primary_key:id"`
	VoucherId   int64       `json:"-"`
	ToUid       int64       `json:"-"`
	Summary     string      `json:"summary"`
	Amount      float64     `json:"amount"`
	CreateTime  int64       `json:"-"`
	UpdateTime  int64       `json:"update_time"`
	VoucherInfo interface{} `json:"voucher" gorm:"-"`
}

func (VoucherRecord) TableName() string {
	return "voucher_record"
}

type CurrencyRate struct {
	db.Model
	Key string `json:"key"`
	//UserId  int64           `json:"user_id"`
	From    string          `json:"from"`
	To      string          `json:"to"`
	Ratio   decimal.Decimal `json:"ratio"`
	Source  string          `json:"-"`
	Digital bool            `json:"-"`
	Auto    bool            `json:"-"`
}

func (CurrencyRate) TableName() string {
	return "currency_rate"
}

type UserAssetType struct {
	Id           int64   `json:"-" gorm:"primary_key:id"`
	UserId       int64   `json:"-"`
	Type         int     `json:"type"`
	TotalAmount  float64 `json:"total_amount"`
	NormalAmount float64 `json:"normal_amount,omitempty"`
	LockAmount   float64 `json:"lock_amount,omitempty"`
	FreezeAmount float64 `json:"freeze_amount,omitempty"`
	CreateTime   int64   `json:"-"`
	UpdateTime   int64   `json:"update_time"`
}

func (UserAssetType) TableName() string {
	return "user_asset_type"
}
