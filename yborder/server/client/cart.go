package client

import (
	"yunbay/yborder/common"
	"yunbay/yborder/dao"
	"yunbay/yborder/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"

	//"github.com/jinzhu/gorm"
	"time"
	//"yunbay/yborder/conf"

	"fmt"
	"yunbay/yborder/server/share"
)

// 添加修改订单信息
type cartitem struct {
	ProductId    int64 `json:"product_id" binding:"gt=0"`
	ProductSkuId int64 `json:"product_sku_id"`
	PublishArea  *int  `json:"publish_area"`
	Quantity     int   `json:"quantity" binding:"gt=0"`
}

// 添加购物车
func Cart_Add(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var args cartitem
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	country := util.GetCountry(c)

	area := common.CURRENCY_KT // 默认为KT专区商品
	if args.PublishArea != nil {
		area = *args.PublishArea
	}
	db := db.GetTxDB(c)
	now := time.Now().Unix()

	// 获取该商品是否在购物车内
	v := common.Cart{UserId: user_id, PublishArea: area, ProductId: args.ProductId, ProductSkuId: args.ProductSkuId, Quantity: args.Quantity /*SellerUserId:product.UserId,*/, Country: country}

	db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (user_id, product_id, product_sku_id) DO update set quantity=cart.quantity+%v, update_time=%v", args.Quantity, now))
	if err := db.Save(&v).Error; err != nil {
		glog.Errorf("Orders_Upsert Save err=%v", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	db.AfterCommit(func() {
		(&dao.Cart{}).RefreshCache(user_id)
	})

	yf.JSON_Ok(c, v)
}

func CartUpdate(v *common.Cart, quantity int) (err error) {

	v.Quantity = quantity
	return
}

type quantitySt struct {
	Id       int64 `json:"cart_id" binding:"gt=0"`
	Quantity int   `json:"quantity" binding:"gt=0"`
}

// 购物车更新
func Cart_Update(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var args quantitySt
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	db := db.GetTxDB(c)
	var v common.Cart
	// 如果是修改订单 判断订单状态是否为未支付状态

	if err := db.Where("id=? and user_id=?", args.Id, user_id).Find(&v).Error; err != nil {
		glog.Error("ERR_CARTID_NOT_EXIST id=", args.Id, " err=", err)
		yf.JSON_Fail(c, common.ERR_CARTID_NOT_EXIST)
		return
	}

	if err := CartUpdate(&v, args.Quantity); err != nil {
		glog.Error("OrderUpdate err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	now := time.Now().Unix()
	err := db.Model(&v).Updates(map[string]interface{}{"quantity": v.Quantity /*"total_amount": v.TotalAmount, "rebat_amount": v.RebatAmount,*/, "update_time": now}).Error
	if err != nil {
		glog.Errorf("Orders_Upsert Save err=%v", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	db.AfterCommit(func() {
		(&dao.Cart{}).RefreshCache(user_id)
	})
	yf.JSON_Ok(c, v)
}

type cartidsParam struct {
	Ids []int64 `json:"cart_ids" binding:"gt=0"`
}

// 删除购物车
func Cart_Del(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var args cartidsParam
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	args.Ids = base.UniqueInt64Slice(args.Ids) // 去重
	if len(args.Ids) == 0 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	// 购物车删除直接物理删除

	// 删除订单只有在购物车,未付款情况下可删除
	db := db.GetTxDB(c)
	res := db.Delete(common.Cart{}, "user_id=? and id in(?)", user_id, args.Ids)
	if res.Error != nil {
		glog.Error("Order_Del fail! err=", db.Error)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if res.RowsAffected != int64(len(args.Ids)) {
		glog.Error("ERR_ORDER_FORBIDDEN_DEL fail!")
		yf.JSON_Fail(c, common.ERR_ORDER_FORBIDDEN_DEL)
		return
	}

	db.AfterCommit(func() {
		(&dao.Cart{}).RefreshCache(user_id)
	})

	yf.JSON_Ok(c, gin.H{})
}

func ListCartByIds(user_id int64, ids []int64) (vs []common.Cart, err error) {
	vs = []common.Cart{}
	if err = db.GetDB().Find(&vs, "user_id=? and id in(?)", user_id, ids).Error; err != nil {
		glog.Error("ListCartByIds fail! err=", err)
		return
	}
	return
}

// 购物车列表
func Cart_List(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	publish_area, _ := base.CheckQueryIntDefaultField(c, "publish_area", 1)

	v := &dao.Cart{}
	ret, err := v.List(user_id, publish_area, page, page_size)
	if err != nil {
		glog.Error("Cart_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if err := share.GetCartsProductInfos(ret.List); err != nil {
		glog.Error("getUserCartList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, ret)
}
