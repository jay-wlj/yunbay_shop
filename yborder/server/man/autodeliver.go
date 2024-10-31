package man

import (
	"fmt"

	"yunbay/yborder/common"
	"yunbay/yborder/conf"
	"yunbay/yborder/dao"
	"yunbay/yborder/server/share"
	"yunbay/yborder/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/jay-wlj/gobaselib/db"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

type ordersSt struct {
	OrderId int64 `json:"order_id" binding:"gt=0"`
}

// 自动发放代金券
func Orders_AutoDeiver(c *gin.Context) {
	var req ordersSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	var v common.Orders
	var err error
	db := db.GetTxDB(c)

	// 先将状态置为已发货 幂等性
	res := db.Model(&common.Orders{}).Where("id=? and status=? and product_type>?", req.OrderId, common.ORDER_STATUS_PAYED, common.GOODS_TYPE_PHYSICAL).Updates(base.Maps{"status": common.ORDER_STATUS_FINISH})
	//res := db.Model(&common.Orders{}).Where("id=? and status=? and (product->>'type')::int >?", req.OrderId, common.ORDER_STATUS_PAYED, common.GOODS_TYPE_PHYSICAL).Updates(base.Maps{"status": common.ORDER_STATUS_FINISH})
	if err = res.Error; err != nil {
		glog.Error("Orders_AutoDeiver fail err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 无更新直接返回
	if 0 == res.RowsAffected {
		yf.JSON_Ok(c, gin.H{})
		return
	}

	if err = db.Find(&v, "id=?", req.OrderId).Error; err != nil {
		glog.Error("Orders_AutoDeiver fail err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 对虚拟物品进行充值
	switch v.ProductType {
	case common.GOODS_TYPE_TEL_RECHARGE:
		// 话费充值
		// 查询充值的手机号码
		tel, ok := v.ExtInfos["tel"].(string)
		if !ok {
			glog.Error("Orders_AutoDeiver fail! no tel in extinfo=", v.ExtInfos)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}

		// 查询该订单面值
		var man common.ManInfos
		man.ParseJsonb(v.Maninfos)

		// 对此手机号进行充值
		of, err := util.GetOfpay().Tel_Recharge(v.Id, tel, man.OfAmount.IntPart())
		if err != nil {
			glog.Error("Orders_AutoDeiver fail! err=", err)
			yf.JSON_Fail(c, err.Error())
			return
		}
		switch of.Retcode {
		case util.ERR_OFPAY_OK, util.ERR_OFPAY_UNKNOW:
			// 保存虚拟订单信息到maninfo中
			man.TxId = of.OfId
			if err = db.Model(&v).Updates(base.Maps{"maninfo": base.StructToMap(man)}).Error; err != nil {
				glog.Error("Orders_AutoDeiver fail err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}

			// 保存欧飞充值信息
			db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (order_id, of_id) DO update set retcode=%v", of.Retcode))
			game_sate := of.OfOrder.GameState
			of.OfOrder.GameState = 0 // TODO 有待优化 先确保OfRetNotify能够处理正常
			if err = db.Save(&of).Error; err != nil {
				glog.Error("Orders_AutoDeiver fail! err=", err)
				yf.JSON_Fail(c, err.Error())
				return
			}
			of.OfOrder.GameState = game_sate
			db.DB = db.Set("gorm:insert_option", "")

			// 处理欧飞订单充值状态
			if err = share.OfRetNotify(db, &of.OfOrder); err != nil {
				glog.Error("Orders_AutoDeiver fail! Refund err=", err)
				yf.JSON_Fail(c, err.Error())
				return
			}
			o := &dao.Orders{}
			o.RefreshByOrderIds([]int64{req.OrderId})
		default:
			glog.Error("Tel_Recharge fail! retcode=", of.Retcode, " errMsg=", of.ErrMsg)
			yf.JSON_Fail(c, of.ErrMsg)
			return
		}
	case common.GOODS_TYPE_CARD:
		// 卡密商品
		var man common.ManInfos
		man.ParseJsonb(v.Maninfos)

		// 查询该商品的产品编号
		cardid := conf.Config.OfPay.Ofcard[man.OfKey]
		if cardid == "" {
			glog.Error("Orders_AutoDeiver fail! cardid empty, key:", man.OfKey)
			yf.JSON_Fail(c, "cardid empty")
			return
		}
		of, err := util.GetOfpay().CardWidthdraw(v.Id, cardid, v.Quantity)
		if err != nil {
			glog.Error("Orders_AutoDeiver fail! err=", err)
			yf.JSON_Fail(c, err.Error())
			return
		}
		of.OrderId = v.Id
		switch of.Retcode {
		case util.ERR_OFPAY_OK, util.ERR_OFPAY_UNKNOW:
			// 保存虚拟订单信息到maninfo中
			man.TxId = of.OfId
			if err = db.Model(&v).Updates(base.Maps{"maninfo": base.StructToMap(man)}).Error; err != nil {
				glog.Error("Orders_AutoDeiver fail err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}

			// 保存欧飞充值信息
			db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (order_id, of_id) DO update set retcode=%v", of.Retcode))
			if err = db.Save(&of).Error; err != nil {
				glog.Error("Orders_AutoDeiver fail! err=", err)
				yf.JSON_Fail(c, err.Error())
				return
			}
			db.DB = db.Set("gorm:insert_option", "")

			// 处理欧飞订单充值状态
			if err = share.OfRetNotify(db, &of.OfOrder); err != nil {
				glog.Error("Orders_AutoDeiver fail! Refund err=", err)
				yf.JSON_Fail(c, err.Error())
				return
			}
			o := &dao.Orders{}
			o.RefreshByOrderIds([]int64{req.OrderId})
		default:
			glog.Error("Tel_Recharge fail! retcode=", of.Retcode, " errMsg=", of.ErrMsg)
			yf.JSON_Fail(c, of.ErrMsg)
			return
		}
	case common.GOODS_TYPE_YOUBUY:
		// 查询该订单面值
		var man common.ManInfos
		man.ParseJsonb(v.Maninfos)
		if man.Voucher != nil {
			// 调用代金券充值接口
			//order_id, err := util.Voucher_Recharge(util.VoucherSt{Id: v.Id, UserId: v.UserId, Type: man.Voucher.Type, Amount: man.Voucher.Amount})
			order_id, err := util.Voucher_Recharge(util.VoucherSt{Id: v.Id, UserId: man.Voucher.ThirdId, Type: man.Voucher.Type, Amount: man.Voucher.Amount}) // 使用第三方用户id
			if err != nil {
				glog.Error("Orders_AutoDeiver Voucher_Recharge fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
			man.Voucher.TxId = &order_id

			if err = db.Model(&v).Updates(map[string]interface{}{"status": common.ORDER_STATUS_FINISH, "maninfos": base.StructToMap(man)}).Error; err != nil {
				glog.Error("Orders_AutoDeiver fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}

			db.AfterCommit(func() {
				// 刷新订单缓存
				o := &dao.Orders{}
				o.RefreshByOrderIds([]int64{req.OrderId})

				// 完成此笔订单
				v := util.YBAssetStatus{OrderIds: []int64{req.OrderId}, Status: common.ASSET_POOL_FINISH}
				mq := common.MQUrl{Methond: "post", Uri: "/man/asset/payset", AppKey: "ybasset", Data: v, MaxTrys: -1}
				if err := util.PublishMsg(mq); err != nil {
					glog.Error("Orders_AutoDeiver fail! err=", err)
					yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
					return
				}
			})
		}
	default:
	}

	yf.JSON_Ok(c, gin.H{})
	return
}

func OfOrderQuery(c *gin.Context) {
	str_ids, _ := base.CheckQueryStringField(c, "ids")
	ids := base.StringToInt64Slice(str_ids, ",")

	if 0 == len(ids) {
		yf.JSON_Ok(c, gin.H{})
		return
	}
	success_ids := []int64{}

	var err error
	for _, v := range ids {
		of, e := util.GetOfpay().QueryInfo(v)
		if err = e; err != nil {
			glog.Error("OfOrderQuery fail! err=", err)
			continue
		}
		if err = share.OfRetNotify(nil, &of.OfOrder); err != nil {
			glog.Error("OfRetNotify fail! err=", err)
			continue
		}
		success_ids = append(success_ids, v)
	}
	yf.JSON_Ok(c, gin.H{"success": success_ids})
}
