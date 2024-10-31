package man

import (
	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"

	//base "github.com/jay-wlj/gobaselib"

	"yunbay/yborder/common"
	"github.com/jay-wlj/gobaselib/yf"

	//"time"
	"yunbay/yborder/util"
	"github.com/jay-wlj/gobaselib/db"
	//"fmt"
	//"github.com/jinzhu/gorm"
)

type productPriceSt struct {
	ProductId    int64   `json:"product_id"`
	ProductSkuId int64   `json:"product_sku_id"`
	SalePrice    float64 `json:"sale_price"`
	Rebat        float64 `json:"rebat"`
}

func UpdateProductPrice(c *gin.Context) {
	var args []productPriceSt
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	// 更新购物车的商品价格
	var product_sku_ids []int64
	mProduct := make(map[int64]productPriceSt)
	for _, v := range args {
		product_sku_ids = append(product_sku_ids, v.ProductSkuId)
		mProduct[v.ProductSkuId] = v
	}
	db := db.GetTxDB(c)
	for k, v := range mProduct {
		if err := db.Model(&common.Cart{}).Where("product_sku_id=?", k).Updates(map[string]interface{}{"price": v.SalePrice, "rebat": v.Rebat}).Error; err != nil {
			glog.Error("UpdateProductPrice fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}
	yf.JSON_Ok(c, gin.H{})
}
