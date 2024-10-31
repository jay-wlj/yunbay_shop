package main

import (
	"encoding/json"
	"flag"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"regexp"
	"strconv"
	"strings"
	"time"
	"yunbay/ybgoods/common"
	"yunbay/ybgoods/conf"
	"yunbay/ybgoods/util"

	"github.com/lib/pq"

	"github.com/shopspring/decimal"

	"github.com/jinzhu/gorm"

	"github.com/tealeg/xlsx"

	"github.com/jie123108/glog"
)

func init() {
	conf, err := conf.LoadConfig("../conf/config.yml")
	if err != nil {
		return
	}
	if _, err := db.InitPsqlDb(conf.Server.PSQLUrl, conf.Server.Debug); err != nil {
		panic(err.Error())
	}

}

type descImgs struct {
	Path   string `json:"path"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

const (
	COL_ONE_TYPE = iota
	COL_TWO_TYPE
	COL_TWO_IMG
	COL_GOODS_TITLE
	COL_GOODS_IMG
	COL_GOODS_DESCIMG
	COL_SKU1
	COL_SKU2
	COL_PRICE
	COL_FEE
	COL_SALE_PRICE
	COL_STOKE
	COL_REBAT
	COL_USERID
)

// delete from product_type where id not in(1, 25, 46, 47);
// detete from product where user_id not in(55159, 118994);
// delete from product_model where product_id not in(526,527, 272);

type Orders struct {
	db.Model
	ProductSkuId int64    `json:"product_sku_id"`
	Product      pq.JSONB `json:"product"`
	CurrencyType int      `json:"currency_type"`
	//Product interface{} `json:"product" gorm:"-"`
}

func (Orders) TableName() string {
	return "orders"
}

// 导入商品数据库
func main() {
	var user_id int64
	var publish_area int
	var excelFile string
	var orders bool
	flag.IntVar(&publish_area, "publish_area", 1, "publish_area")
	flag.Int64Var(&user_id, "user_id", 119109, "user_id")
	flag.StringVar(&excelFile, "conf", "./a.xlsx", "conf path")
	flag.BoolVar(&orders, "order", false, "orders product handle")
	flag.Parse()

	// 修改订单商品结构
	if orders {
		vs := []Orders{}
		db := db.GetTxDB(nil)
		if err := db.Select("id,product,product_sku_id,currency_type").Order("id asc").Find(&vs).Error; err != nil {
			glog.Error("get orders fail! err=", err)
			return
		}
		for _, v := range vs {
			p := v.Product
			if vm, ok := p["models"].([]interface{}); ok {
				for _, vmv := range vm {
					if m, ok := vmv.(map[string]interface{}); ok {
						id := int64(m["id"].(float64))
						if id != v.ProductSkuId {
							continue
						}
						pay_price := common.PayPrice{}
						pay_price.PayType = v.CurrencyType
						pay_price.Price = decimal.NewFromFloat(m["sale_price"].(float64))
						pay_price.PredictYbt, _ = m["predict_ybt"].(float64)
						pay_price.Coin = common.GetCurrencyName(pay_price.PayType)

						combine := make(map[string]interface{})
						combine["规格"] = m["title"].(string)

						//p["pay_price"] = pay_price
						p["rebat"] = m["rebat"]
						skus := []map[string]interface{}{}
						s := make(map[string]interface{})
						s["id"] = id
						s["combines"] = []map[string]interface{}{combine}
						s["price"] = decimal.NewFromFloat(m["sale_price"].(float64))

						if imgs, ok := m["images"].([]interface{}); ok {
							if len(imgs) > 0 {
								s["img"], _ = imgs[0].(string)
							}
						}
						s["pay_price"] = []common.PayPrice{pay_price}
						skus = append(skus, s)
						delete(p, "models")
						p["skus"] = skus
						if len(skus) > 0 {
							delete(p, "pay_price")
						}
						v.Product = p

						// 更新product及rebat
						if err := db.Model(&v).Updates(base.Maps{"product": p}).Error; err != nil {
							glog.Error("get orders fail! err=", err)
							return
						}
					}
				}
			}
		}
		db.Commit()
		return
	}
	// 解析execel
	if excelFile == "" {
		fmt.Println("exefile is empty!")
		return
	}
	//excelFile := "./a.xlsx"
	xlsFile, err := xlsx.OpenFile(excelFile)
	if err != nil {
		glog.Error("OpenFile fail! err=", err)
		return
	}
	xlsData, err := xlsFile.ToSlice()
	if err != nil {
		glog.Error("ToSlice fail! err=", err)
		return
	}
	if len(xlsData) == 0 {
		return
	}
	xls := xlsData[0]

	db := db.GetTxDB(nil)
	mOneProductType := make(map[string]int64)
	mSecondType := make(map[string]int64)
	vs := []common.Product{}
	fs := []Uploadfile{}
	for i, v := range xls {
		if i == 0 {
			continue
		}
		contact := make(map[string]interface{})
		json.Unmarshal([]byte(`{"contact_name": "周瀚", "contact_email": "a13338976405@163.com", "contact_phone": "13338976405"}`), &contact)
		// 读取每一行
		p := common.Product{Status: 1, PublishArea: publish_area, UserId: user_id, Contact: contact} // KT商品

		//p := common.Product{Status: 1, PublishArea: 0, UserId: 51871, CheckStatus: 2, Contact: contact} // 折扣商品

		//skus := [][]string{}
		mattrs := make(map[int]string)
		vattrs := [][]string{}
		var one_type_id int64
		var second_type_name string
		for k, d := range v {
			switch k {
			case COL_ONE_TYPE: // 一级分类
				if _, ok := mOneProductType[d]; !ok {
					var r common.ProductCategory
					// 是否已有一级分类
					if err = db.Select("id, parent_id").Find(&r, "title = ? and parent_id=0", d).Error; err != nil && err != gorm.ErrRecordNotFound {
						glog.Error("get ProductType fail! err=", err)
						return
					}
					if err == gorm.ErrRecordNotFound {
						r = common.ProductCategory{Title: d}
						if err = db.Save(&r).Error; err != nil {
							glog.Error("product_type save fail! err=", err)
							return
						}
					}
					mOneProductType[d] = r.Id
				}
				one_type_id = mOneProductType[d]
			case COL_TWO_TYPE: // 二级分类
				second_type_name = d
			case COL_TWO_IMG:
				if _, ok := mSecondType[second_type_name]; !ok {
					var r common.ProductCategory
					// 是否已有一级分类
					if err = db.Select("id, parent_id").Find(&r, "parent_id=? and title = ?", one_type_id, second_type_name).Error; err != nil && err != gorm.ErrRecordNotFound {
						glog.Error("get ProductType fail! err=", err)
						return
					}
					if err == gorm.ErrRecordNotFound {
						r = common.ProductCategory{Title: second_type_name, ParentId: one_type_id, Picture: d}
						if err = db.Save(&r).Error; err != nil {
							glog.Error("product_type save fail! err=", err)
							return
						}
					}
					mSecondType[second_type_name] = r.Id
				}
				p.CategoryId = mSecondType[second_type_name]
			case COL_GOODS_TITLE: // 标题
				p.Title = d
			case COL_GOODS_IMG: // 商品图片列表
				p.Images = strings.Split(d, ",")
			case COL_GOODS_DESCIMG: // 商品详情图
				res, _ := regexp.Compile(`src=\\\"([^"]*)\\\".*?width: (\d*)px, height: (\d*)px,`)

				as := res.FindAllStringSubmatch(d, -1)

				vs := []interface{}{}
				for _, v := range as {
					p := descImgs{}
					for i, j := range v {
						switch i {
						case 1:
							p.Path = j
						case 2:
							p.Width, _ = strconv.Atoi(j)
						case 3:
							p.Height, _ = strconv.Atoi(j)
						}
					}
					vs = append(vs, p)

					f := Uploadfile{Path: p.Path, Width: p.Width, Height: p.Height, Hash: util.Sha1hex([]byte(p.Path)), AppId: "upload"}
					f.Rid = util.HashToRid("upload", f.Hash)
					fs = append(fs, f)
				}

				p.Descimgs = vs

			case COL_SKU1, COL_SKU2: // 规格1

				res, _ := regexp.Compile(`(.*?)\((.*)\)`)
				as := res.FindStringSubmatch(d)
				// if len(as) < 3 {
				// 	glog.Error("can't split sku: ", d)
				// 	continue
				// }
				if len(as) > 2 {
					vattrs = append(vattrs, strings.Split(as[2], ","))
					mattrs[len(vattrs)-1] = as[1]
				}

				if COL_SKU2 == k {
					// 分配sku
					ms := []*common.ProductSku{}
					lsku := len(vattrs)

					mskus := [][]map[string]string{}
					switch lsku {
					case 1, 2:
						for _, l := range vattrs[0] {

							if lsku > 1 { // 含有多个属性
								for _, ll := range vattrs[1] {
									msku := []map[string]string{}
									s1 := make(map[string]string)
									s1[mattrs[0]] = l
									msku = append(msku, s1)

									s2 := make(map[string]string)
									s2[mattrs[1]] = ll
									msku = append(msku, s2)

									mskus = append(mskus, msku)
								}
							} else {
								// 只含一个属性
								msku := []map[string]string{}
								s1 := make(map[string]string)
								s1[mattrs[0]] = l
								msku = append(msku, s1)
								mskus = append(mskus, msku)
							}
						}
						// case 0:
						// 	msku := make(map[string]string)
						// 	msku["规格"] = "默认"
						// 	mskus = append(mskus, msku)
					}

					for _, m := range mskus {
						sku := &common.ProductSku{Img: p.Images[0]}
						//sku.Combines.RawMessage, _ = json.Marshal(m)
						buf, _ := json.Marshal(m)
						if err := json.Unmarshal(buf, &sku.Combines); err != nil {
							glog.Error("args err=", err)
						}

						ms = append(ms, sku)
					}

					p.Skus = ms
				}
				//strings.Trim(a, "(")

			//case 5: // 规格2
			case COL_PRICE:
				price, e := decimal.NewFromString(d)
				if e != nil {
					glog.Error("StringToFloat64 fail! err=", e)
					return
				}
				p.CostPrice = p.CostPrice.Add(price) // 商品成本价
				p.CostPrice = p.CostPrice.Round(3)   // 小数后2位
				for i := range p.Skus {
					//p.Skus[i].CostPrice += price
					p.Skus[i].CostPrice = p.CostPrice // 更新规格成本价
				}

			case COL_FEE:
			case COL_SALE_PRICE:
				price, e := decimal.NewFromString(d)
				if e != nil {
					glog.Error("StringToFloat64 fail! err=", e)
					return
				}
				for i := range p.Skus {
					//p.Skus[i].CostPrice += price
					p.Skus[i].Price = price
				}
				p.Price = price
			case COL_STOKE: // 库存量
				amount, e := base.StringToInt64(d)
				if e != nil {
					glog.Error("StringToFloat64 fail! err=", e)
					return
				}
				for i := range p.Skus {
					p.Skus[i].Stock = amount
				}
				p.Stock = amount
			case COL_REBAT:
				rebat, e := base.StringToFloat64(d)
				if e != nil {
					glog.Error("StringToFloat64 fail! err=", e)
					return
				}
				rebat = float64((int)(rebat*100)) / 100
				p.Rebat = decimal.NewFromFloat(rebat)
			case COL_USERID:
				if d == "" {
					continue
				}
				uid, e := base.StringToInt64(d)
				if e != nil {
					glog.Error("StringToFloat64 fail! err=", e)
					return
				}
				if uid > 0 {
					p.UserId = uid
				}
			default:

			}

		}
		vs = append(vs, p)
	}

	// 录入数据库
	if err = SaveDescImages(db, fs); err != nil {
		glog.Error("SaveToDb fail! err=", err)
		return
	}
	db.Commit()

	// 打印
	buf, err := json.Marshal(vs)
	if err != nil {
		glog.Error("json marshal err=", err)
		return
	}
	fmt.Println(string(buf))

	if err = SaveToDb(vs); err != nil {
		glog.Error("SaveToDb fail! err=", err)
		return
	}

}

func SaveToDb(vs []common.Product) (err error) {
	return util.AddSomeProduct(vs)
	// for _, v := range vs {
	// 	if err = db.Save(&v).Error; err != nil {
	// 		glog.Error("SaveToDb fail! err=", err)
	// 		return
	// 	}

	// 	// 更新默认的def_model_id
	// 	v.DefSkuId = v.Skus[0].Id
	// 	if err = db.Model(&v).Update(base.Maps{"default_model_id": v.DefSkuId}).Error; err != nil {
	// 		glog.Error("SaveToDb fail! err=", err)
	// 		return
	// 	}
	// }
	return nil
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

func SaveDescImages(db *db.PsqlDB, vs []Uploadfile) (err error) {
	db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (rid) DO update set update_time=%v", time.Now().Unix()))
	for _, v := range vs {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("SaveToDb fail! err=", err)
			return
		}
	}
	return nil
}
