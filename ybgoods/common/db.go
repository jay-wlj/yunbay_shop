package common

import (
	"github.com/jay-wlj/gobaselib/db"

	"github.com/shopspring/decimal"

	"github.com/lib/pq"
)

type Business struct {
	db.Model
	UserId  int64  `json:"user_id"`
	Type    int    `json:"type"`
	Company string `json:"company"`
	Name    string `json:"name"`
	Status  int16  `json:"status,omitempty" view:"man"`
}

func (Business) TableName() string {
	return "business"
}

type ProductCategory struct {
	db.Model
	ParentId int64             `json:"parent_id"`
	Title    string            `json:"title"`
	Picture  string            `json:"picture,omitempty"`
	Sort     int               `json:"sort"`
	IsShow   int               `json:"is_show"`
	Children []ProductCategory `json:"children,omitempty" gorm:"-"`
}

func (ProductCategory) TableName() string {
	return "product_category"
}

type Attr struct {
	Id     int64              `json:"id"`
	Name   string             `json:"name"`
	Values []ProductAttrValue `json:"values"`
}

type Product struct {
	db.Model
	CategoryId  int64           `json:"category_id,omitempty"`
	UserId      int64           `json:"user_id,omitempty"`
	Title       string          `json:"title" view:"order"`
	Info        string          `json:"info,omitempty"`
	Images      pq.StringArray  `json:"images" view:"order"`
	Descimgs    db.JsonbArray   `json:"descimgs,omitempty"`
	Type        int16           `json:"type"`
	Stock       int64           `json:"stock"`
	Sold        int64           `json:"sold"`
	CostPrice   decimal.Decimal `json:"cost_price,omitempty" gorm:"-"`
	Price       decimal.Decimal `json:"price"`
	Rebat       decimal.Decimal `json:"rebat"`
	Canreturn   bool            `json:"canreturn,omitempty" gorm:"-"`
	Contact     db.JSONB        `json:"contact,omitempty" gorm:"not null;default:'{}'"`
	Status      int16           `json:"status,omitempty" view:"man"`
	IsHid       int16           `json:"is_hid,omitempty"`
	DefSkuId    int64           `json:"def_sku_id,omitempty"`
	PublishArea int             `json:"publish_area,omitempty"`
	Skus        []*ProductSku   `json:"skus,omitempty"`
	Attrs       []Attr          `json:"attrs,omitempty" gorm:"-"`
	PredictYbt  float64         `json:"predict_ybt,omitempty" gorm:"-"`
	PayPrice    []PayPrice      `json:"pay_price,omitempty" gorm:"-"`
	Extinfo     db.JSONB        `json:"extinfo,omitempty"`
	HidCause    string          `json:"hid_cause,omitempty"`
	Discount    decimal.Decimal `json:"discount"`
	Country     int             `json:"-" gorm:"-"`
}

func (Product) TableName() string {
	return "product"
}

type ManProduct struct {
	Product
	Categories []*ProductCategory `json:"categories"`
	DefSku     string             `json:"def_sku_name,omitempty"`
	Business   *Business          `json:"business,omitempty"`
}

type PayPrice struct {
	Coin                 string          `json:"coin"`
	PayType              int             `json:"pay_type"`
	PredictYbt           float64         `json:"predict_ybt,omitempty"`
	OriginPrice          decimal.Decimal `json:"origin_price"`
	Price                decimal.Decimal `json:"price"`
	UnitPrice            decimal.Decimal `json:"-"`
	LowestDiscountPrice  float64         `json:"lowest_discount_price,omitempty"`
	HighestDiscountPrice float64         `json:"highest_discount_price,omitempty"`
}

type ProductSku struct {
	db.Model
	ProductId int64         `json:"-"`
	Sku       db.Jsonb      `json:"sku" gorm:"not null;default:'{}'" view:"order"`
	Combines  db.JsonbArray `json:"combines"`
	//SkuName   string          `json:"sku_name"`
	Stock     int64           `json:"stock"`
	Sold      int64           `json:"sold"`
	CostPrice decimal.Decimal `json:"cost_price,omitempty" gorm:"-"`
	Price     decimal.Decimal `json:"price"`
	PayPrice  []PayPrice      `json:"pay_price,omitempty" gorm:"-"`
	Img       string          `json:"img,omitempty"`
	Extinfo   db.JSONB        `json:"extinfo,omitempty"`
}

func (ProductSku) TableName() string {
	return "product_sku"
}

type ProductAttrKey struct {
	db.Model
	CategoryId int64              `json:"-"`
	Name       string             `json:"name"`
	Values     []ProductAttrValue `json:"values"`
}

func (ProductAttrKey) TableName() string {
	return "product_attr_key"
}

type ProductAttrValue struct {
	db.Model
	ProductAttrKeyId int64  `json:"-"`
	Value            string `json:"value"`
}

func (ProductAttrValue) TableName() string {
	return "product_attr_value"
}

type ProductPrice struct {
	db.Model
	PId       int64           `json:"p_id"`
	PSkuId    int64           `json:"p_sku_id"`
	CostPrice decimal.Decimal `json:"cost_price"`
	Price     decimal.Decimal `json:"price"`
}

func (ProductPrice) TableName() string {
	return "product_price"
}

type Setting struct {
	db.Model
	SetttingKey  string `json:"setting_key"`
	SettingValue string `json:"setting_value"`
}

func (Setting) TableName() string {
	return "setting"
}

type RecommendIndex struct {
	db.Model
	Type       int           `json:"type"`
	Name       string        `json:"name"`
	Img        string        `json:"img"`
	Descimg    string        `json:"descimg"`
	ProductIds pq.Int64Array `json:"product_ids,omitempty"`
	Country    int           `json:"country,omitempty"`
	Rowset     interface{}   `json:"list" gorm:"-"`
	ListEnded  bool          `json:"list_ended" gorm:"-"`
}

func (RecommendIndex) TableName() string {
	return "product_recommend"
}

type Uploadfile struct {
	db.Model
	Rid      string
	AppId    string `gorm:"column:appid"`
	Hash     string `gorm:"column:hash`
	Size     int
	Path     string
	Width    int
	Height   int
	Duration int
}

func (t *Uploadfile) TableName() string {
	return "uploadfile"
}
