package main

import (
	"yunbay/ybapi/common"
	"yunbay/ybapi/conf"
	"yunbay/ybapi/dao"
	"yunbay/ybapi/util"
	"encoding/json"
	"flag"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"regexp"
	"strconv"
	"strings"
	"time"

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
)

// begin;
// delete from product_type where id not in(1, 25, 46, 47);
// detete from product where user_id not in(55159, 118994);
// delete from product_model where product_id not in(526,527, 272);
// commit

// 导入商品数据库
func main() {
	var user_id int64
	var publish_area int64
	flag.Int64Var(&publish_area, "publish_area", 1, "publish_area")
	flag.Int64Var(&user_id, "user_id", 118994, "user_id")

	flag.Parse()

	_, err := dao.GetApiCache()
	if err != nil {
		glog.Error("GetApiCache fail! err=", err)
		return
	}

	// 解析execel
	excelFile := "./a.xlsx"
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
		p := common.Product{PayType: []string{"KT"}, Status: 1, PublishArea: 1, UserId: user_id, CheckStatus: 2, Contact: contact} // KT商品

		// 折扣商品
		if publish_area == 0 {
			p = common.Product{PayType: []string{"YBT", "SNET"}, Status: 1, PublishArea: 0, UserId: user_id, CheckStatus: 2, Contact: contact} // 折扣商品
		}
		//p := common.Product{PayType: []string{"KT", "YBT", "SNET"}, Status: 1, PublishArea: 0, UserId: 51871, CheckStatus: 2, Contact: contact} // 折扣商品

		var one_type_id int64
		var second_type_name string
		for k, d := range v {
			switch k {
			case COL_ONE_TYPE: // 一级分类
				if _, ok := mOneProductType[d]; !ok {
					var r common.ProductType
					// 是否已有一级分类
					if err = db.Select("id, parent_id").Find(&r, "title = ? and parent_id=0", d).Error; err != nil && err != gorm.ErrRecordNotFound {
						glog.Error("get ProductType fail! err=", err)
						return
					}
					if err == gorm.ErrRecordNotFound {
						r = common.ProductType{Title: d}
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
					var r common.ProductType
					// 是否已有一级分类
					if err = db.Select("id, parent_id").Find(&r, "parent_id=? and title = ?", one_type_id, second_type_name).Error; err != nil && err != gorm.ErrRecordNotFound {
						glog.Error("get ProductType fail! err=", err)
						return
					}
					if err == gorm.ErrRecordNotFound {
						r = common.ProductType{Title: second_type_name, ParentId: one_type_id, Picture: d}
						if err = db.Save(&r).Error; err != nil {
							glog.Error("product_type save fail! err=", err)
							return
						}
					}
					mSecondType[second_type_name] = r.Id
				}
				p.ProductTypeId = mSecondType[second_type_name]
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
			case COL_SKU1: // 规格1
				ms := []common.ProductModel{}
				res, _ := regexp.Compile(`(.*?)\((.*)\)`)
				ss := res.FindStringSubmatch(d)
				if len(ss) > 2 {
					names := strings.Split(ss[2], ",")
					for _, v := range names {
						ms = append(ms, common.ProductModel{Title: v, Images: []string{p.Images[0]}})
					}
				} else {
					ms = append(ms, common.ProductModel{Title: "默认", Images: []string{p.Images[0]}})
				}
				//strings.Trim(a, "(")
				p.Models = ms
			case COL_SKU2: // 规格2

			//case 5: // 规格2
			case COL_PRICE, COL_FEE: // 价格
			case COL_SALE_PRICE:
				price, e := base.StringToFloat64(d)
				if e != nil {
					glog.Error("StringToFloat64 fail! err=", e)
					return
				}
				for i := range p.Models {
					p.Models[i].CostPrice += price
					p.Models[i].SalePrice += price
				}

			case COL_STOKE: // 库存量
				amount, e := strconv.Atoi(d)
				if e != nil {
					glog.Error("StringToFloat64 fail! err=", e)
					return
				}
				for i := range p.Models {
					p.Models[i].Quantity = amount

				}
			case COL_REBAT:
				rebat, e := base.StringToFloat64(d)
				if e != nil {
					glog.Error("StringToFloat64 fail! err=", e)
					return
				}
				rebat = float64((int)(rebat*100)) / 100
				for i := range p.Models {
					p.Models[i].Rebat = rebat

				}
			default:

			}

		}
		vs = append(vs, p)
	}

	// 打印
	buf, err := json.Marshal(fs)
	if err != nil {
		glog.Error("json marshal err=", err)
		return
	}
	fmt.Println(string(buf))

	// 录入数据库
	// if err = SaveToDb(db, vs); err != nil {
	// 	glog.Error("SaveToDb fail! err=", err)
	// 	return
	// }

	if err = SaveDescImages(db, fs); err != nil {
		glog.Error("SaveToDb fail! err=", err)
		return
	}

	// TODO 缓存处理
	db.Commit()
}

func SaveToDb(db *db.PsqlDB, vs []common.Product) (err error) {

	for _, v := range vs {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("SaveToDb fail! err=", err)
			return
		}

		// 更新默认的def_model_id
		v.DefModelId = v.Models[0].Id
		if err = db.Model(&v).Update(base.Maps{"default_model_id": v.DefModelId}).Error; err != nil {
			glog.Error("SaveToDb fail! err=", err)
			return
		}
	}
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
