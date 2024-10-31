package common

import (
	"github.com/jay-wlj/gobaselib/db"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type Notice struct {
	Id         int64  `json:"id" gorm:"primary_key:id"`
	Type       int    `json:"type"`
	UserId     int64  `json:"user_id" gorm:"column:user_id"`
	Title      string `json:"title"`
	Context    string `json:"context"`
	Linkurl    string `json:"linkurl"`
	Status     int    `json:"status"`
	CreateTime int64  `json:"create_time" gorm:"column:create_time"`
	UpdateTime int64  `json:"update_time" gorm:"column:update_time"`
	Country    int    `json:"country"`
}

func (Notice) TableName() string {
	return "notice"
}

type Banner struct {
	Id         int64         `json:"id" gorm:"primary_key:id"`
	Position   int           `json:"position"`
	Content    db.JsonbArray `json"content"`
	CreateTime int64         `json:"create_time" gorm:"column:create_time"`
	UpdateTime int64         `json:"-" gorm:"column:update_time"`
}

func (Banner) TableName() string {
	return "banner"
}

type ProductRecommend struct {
	Id         int64         `json:"id" gorm:"primary_key:id" binding:"gt=0"`
	Type       int           `json:"type"`
	Name       string        `json:"name"`
	Img        string        `json:"img"`
	DescImg    string        `json:"descimg" gorm:"column:descimg"`
	ProductIds pq.Int64Array `json:"product_ids" gorm:"column:product_ids"`
	CreateTime int64         `json:"create_time" gorm:"column:create_time"`
	UpdateTime int64         `json:"-" gorm:"column:update_time"`
	Products   []Product     `json:"product_infos" gorm:"-"`
}

func (ProductRecommend) TableName() string {
	return "product_recommend"
}

type Product struct {
	db.Model
	CategoryId  int64           `json:"category_id,omitempty"`
	UserId      int64           `json:"user_id,omitempty"`
	Title       string          `json:"title"`
	Info        string          `json:"info,omitempty"`
	Images      pq.StringArray  `json:"images"`
	Descimgs    db.JsonbArray   `json:"descimgs,omitempty"`
	Type        int16           `json:"type"`
	Stock       int64           `json:"stock"`
	Sold        int64           `json:"sold"`
	Price       decimal.Decimal `json:"price"`
	Rebat       decimal.Decimal `json:"rebat"`
	Discount    decimal.Decimal `json:"discount"`
	Contact     db.JSONB        `json:"contact,omitempty" gorm:"not null;default:'{}'"`
	Status      int16           `json:"status,omitempty"`
	DefSkuId    int64           `json:"def_sku_id,omitempty"`
	PayType     db.JSONB        `json:"-" gorm:"not null;default:'{}'"`
	PublishArea int             `json:"publish_area,omitempty"`
	Skus        []ProductSku    `json:"skus,omitempty"`
	PredictYbt  float64         `json:"predict_ybt,omitempty" gorm:"-"`
	PayPrice    []PayPrice      `json:"pay_price,omitempty" gorm:"-"`
	Extinfo     db.JSONB        `json:"extinfo,omitempty"`
}

func (Product) TableName() string {
	return "product"
}

type ProductSku struct {
	db.Model
	ProductId int64           `json:"-"`
	Sku       db.Jsonb        `json:"sku"`
	Combines  db.JsonbArray   `json:"combines,omitempty"`
	Stock     int64           `json:"stock"`
	Sold      int64           `json:"sold"`
	Price     decimal.Decimal `json:"price"`
	PayPrice  []PayPrice      `json:"pay_price,omitempty" gorm:"-"`
	Img       string          `json:"img"`
	Extinfo   db.JSONB        `json:"extinfo,omitempty"`
}

func (ProductSku) TableName() string {
	return "product_sku"
}

type PayPrice struct {
	Coin        string          `json:"coin"`
	PayType     int             `json:"pay_type"`
	PredictYbt  decimal.Decimal `json:"predict_ybt,omitempty"`
	OriginPrice decimal.Decimal `json:"origin_price"`
	SalePrice   decimal.Decimal `json:"price"`
}

type Cart struct {
	db.Model
	UserId       int64 `json:"user_id" gorm:"column:user_id"`
	SellerUserId int64 `json:"seller_userid" gorm:"column:seller_userid"`
	ProductId    int64 `json:"product_id" gorm:"column:product_id"`
	ProductSkuId int64 `json:"product_sku_id"`
	Quantity     int   `json:"quantity" gorm:"column:quantity"`
	//OtherAmount  float64  `json:"other_amount" gorm:"column:other_amount"`
	Product     db.JSONB `json:"product" gorm:"-"`
	PublishArea int      `json:"publish_area"`
	Country     int      `json:"country"`
}

func (Cart) TableName() string {
	return "cart"
}

type Orders struct {
	db.Model
	//Id              int64       `json:"id" gorm:"primary_key:id" view:"*"`
	UserId          int64       `json:"user_id"`
	SellerUserId    int64       `json:"seller_userid" gorm:"column:seller_userid"`
	ProductId       int64       `json:"product_id"`
	ProductSkuId    int64       `json:"product_sku_id"`
	ProductType     int16       `json:"product_type"`
	AddressInfo     db.JSONB    `json:"address_info"`
	LogisticsId     int64       `json:"logistics_id"`
	Quantity        int         `json:"quantity"`
	CurrencyType    int         `json:"currency_type"`
	CurrencyPercent float64     `json:"currency_percent"`
	OtherAmount     float64     `json:"other_amount"`
	RebatAmount     float64     `json:"rebat_amount"`
	TotalAmount     float64     `json:"total_amount"`
	Status          int         `json:"status"`
	SaleStatus      int         `json:"sale_status"`
	ExtInfos        db.JSONB    `json:"extinfos" gorm:"column:extinfos;default:'{}'" view:"rebat"`
	AutoCancelTime  int64       `json:"auto_cancel_time"`
	AutoFinishTime  int64       `json:"auto_finish_time"`
	Date            string      `json"date,omitempty"`
	Shield          int16       `json:"-" gorm:"shield"`
	Product         db.JSONB    `json:"product"`
	PublishArea     int         `json:"publish_area"`
	Maninfos        db.JSONB    `json:"-" gorm:"not null;default:'{}'"`
	Country         int         `json:"country"`
	AutoDeliver     bool        `json:"auto_deliver"`
	UserInfo        interface{} `json:"user_info,omitempty" gorm:"-"`
	Now             int64       `json:"now,omitempty" gorm:"-"`
	//Product interface{} `json:"product" gorm:"-"`
}

func (Orders) TableName() string {
	return "orders"
}

type OrderStatus struct {
	db.Model
	OrderId int64  `json:"order_id" gorm:"column:order_id"`
	Status  int    `json:"status"`
	Date    string `json:"date"`
}

func (OrderStatus) TableName() string {
	return "orders_status"
}

type PayPriceSt struct {
	SalePrice  float64 `json:"price"`
	PredictYbt float64 `json:"predict_ybt" view:"order"`
}

type Feedback struct {
	db.Model
	UserId int64          `json:"user_id" gorm:"column:user_id"`
	Email  string         `json:"email"`
	Title  string         `json:"title"`
	Info   string         `json:"info"`
	Affix  pq.StringArray `json:"affix"`
}

func (Feedback) TableName() string {
	return "feedback"
}

type Logistics struct {
	db.Model
	OrderId int64          `json:"order_id" gorm:"column:order_id"`
	UserId  int64          `json:"user_id" gorm:"column:user_id"`
	Company string         `json:"company"`
	Number  string         `json:"number"`
	Infos   pq.StringArray `json:"infos"`
}

func (Logistics) TableName() string {
	return "logistics"
}

type UserAddress struct {
	Id         int64          `json:"id" gorm:"primary_key:id"`
	UserId     int64          `json:"user_id" gorm:"column:user_id"`
	Receiver   string         `json:"receiver"`
	Tel        string         `json:"tel"`
	Address    pq.StringArray `json:"address"`
	Default    bool           `json:"default"`
	CreateTime int64          `json:"create_time" gorm:"column:create_time"`
	UpdateTime int64          `json:"-" gorm:"column:update_time"`
}

func (UserAddress) TableName() string {
	return "user_address"
}

type Invite struct {
	db.Model
	UserId        int64         `gorm:"column:user_id" json:"user_id"`
	Type          int64         `gorm:"column:type" json:"-"`
	InviteUserId  int64         `gorm:"column:invite_userid" json:"invite_userid"`
	InviteTel     string        `gorm:"column:invite_tel" json:"invite_tel"`
	FromInviteIds pq.Int64Array `gorm:"column:recommend_userids" json:"recommend_userids"`
}

func (Invite) TableName() string {
	return "invite"
}

type Business struct {
	db.Model
	UserId         int64          `json:"user_id" gorm:"column:user_id"`
	Type           int16          `json:"type"`
	Company        string         `json:"company"`
	License        pq.StringArray `json:"license"`
	Name           string         `json:"name"`
	Location       string         `json:"location"`
	Certype        int16          `json:"certype"`
	Certid         string         `json:"certid"`
	Certimgs       pq.StringArray `json:"certimgs"`
	ProdutTypes    pq.Int64Array  `json:"product_types" gorm:"column:product_types"`
	Hasbusiness    bool           `json:"hasbusiness"`
	Website        string         `json:"website"`
	Contact        db.JSONB       `json:"contact"`
	Status         int16          `json:"status"`
	Reason         string         `json:"not_pass_cause" gorm:"column:not_pass_cause"`
	TotalTradeFlow float64        `json:"total_tradeflow" gorm:"column:total_tradeflow"`
	TotalRebat     float64        `json:"total_rebat" gorm:"column:total_rebat"`
	TotalYbtFlow   float64        `json:"total_ybtflow" gorm:"column:total_ybtflow"`
	ProTypes       []string       `json:"str_product_types,omitempty" gorm:"-"`
}

func (Business) TableName() string {
	return "business"
}

type Upgrade struct {
	Id         int64          `json:"id" gorm:"primary_key:id"`
	Type       int            `json:"type"`
	Mold       int            `json:"-"`
	Platform   string         `json:"platform"`
	Version    string         `json:"version"`
	VerInt     int64          `json:"-"`
	Url        string         `json:"url"`
	Md5        string         `json:"md5"`
	Upversions pq.StringArray `json:"-"`
	Channels   pq.StringArray `json:"-"`
	Status     int            `json:"-"`
	Title      string         `json:"title"`
	Desc       string         `json:"desc"`
	Country    int            `json:"country"`
	Maner      string         `json:"-"`
	CreateTime int64          `json:"create_time" gorm:"column:create_time"`
	UpdateTime int64          `json:"-" gorm:"column:update_time"`
	Mandatory  db.JSONB       `json:"-" gorm:"not null;default:'{}'"`
}

func (Upgrade) TableName() string {
	return "upgrade"
}

type SectionStruct struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Mandatory struct {
	Section []SectionStruct `json:"section"`
	Specify []string        `json:"specify"`
}

type OfCard struct {
	OrderId int64 `json:"-"`
	//Ordercash  decimal.Decimal `json:"ordercash" xml:"ordercash" gorm:"ordercash"`
	Cardno     string `json:"cardno" xml:"cardno"`
	Cardpws    string `json:"cardpws" xml:"cardpws"`
	Expiretime string `json:"expiretime" xml:"expiretime"`
}

func (OfCard) TableName() string {
	return "of_card"
}

type OfOrder struct {
	db.Model
	OrderId    int64           `json:"sporder_id,string" xml:"sporder_id"`
	OfId       string          `json:"of_id" xml:"orderid"`
	Cardid     string          `json:"cardid" xml:"cardid"`
	Cardname   string          `json:"cardname" xml:"cardname"`
	Cardnum    int             `json:"cardnum,string" xml:"cardnum"`
	Ordercash  decimal.Decimal `json:"ordercash" xml:"ordercash"`
	Cards      []OfCard        `json:"cards" xml:"cards" gorm:"ForeginKey:OrderId;AssociationForeignKey:OrderId"`
	GameUserid string          `json:"game_userid" xml:"game_userid"`
	GameState  int             `json:"game_state" xml:"game_state"`
	ErrMsg     string          `json:"err_msg" xml:"err_msg" gorm:"column:reason"`
	Retcode    int             `json:"retcode,string" xml:"retcode"`
}

func (OfOrder) TableName() string {
	return "of_order"
}

type Lotterys struct {
	db.Model
	PId       int64           `json:"p_id" binding:"required"`
	StartTime int64           `json:"start_time" time_format:"unix" binding:"required,gt=0"`
	EndTime   int64           `json:"end_time" time_format:"unix" binding:"required,gt=0"`
	Coin      int             `json:"coin"`
	Price     decimal.Decimal `json:"price" binding:"required,gt=0"`
	Num       int             `json:"num" binding:"required,gt=0"`
	Stock     int             `json:"stock" binding:"required,gt=0,gtefield=Num"`
	Sold      int             `json:"sold"`
	Amount    decimal.Decimal `json:"amount" binding:"required,gt=0"`
	RewardYbt decimal.Decimal `json:"reward_ybt" binding:"gte=0"`
	Status    int             `json:"status"`
	Hid       int             `json:"hid,omitempty"`
	Product   db.JSONB        `json:"product,omitempty"`
	Pertimes  int             `json:"pertimes"`
	Now       int64           `json:"now,omitempty" gorm:"-"`
	Invalid   bool            `json:"invalid,omitempty" gorm:"-"`
}

func (Lotterys) TableName() string {
	return "lotterys"
}

type LotterysRecord struct {
	db.Model
	LotterysId  int64           `json:"lotterys_id"`
	UserId      int64           `json:"user_id"`
	Amount      decimal.Decimal `json:"amount"`
	Memo        string          `json:"memo"`
	Hash        string          `json:"hash"`
	NumHash     string          `json:"num_hash"`
	Status      int             `json:"status"`
	OrderStatus int             `json:"order_status"`
	Lotterys    interface{}     `json:"lotterys,omitempty" gorm:"-"`
	Url         string          `json:"url" gorm:"-"`
}

func (LotterysRecord) TableName() string {
	return "lotterys_record"
}
