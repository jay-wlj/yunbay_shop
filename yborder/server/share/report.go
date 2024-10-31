package share

import (
	"bytes"
	"encoding/json"
	"strings"
	"yunbay/yborder/common"
	"yunbay/yborder/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
	"github.com/tealeg/xlsx"
)

func get_logistic_ids(ids []int64) (vs []common.Logistics, err error) {
	if err = db.GetDB().Select("id,company,number").Find(&vs, "id in(?)", ids).Error; err != nil {
		glog.Error("Orders_Report fail! err=", err)
		return
	}
	return
}

// 订单表格导出
func OrdersReport(vs *[]common.Orders) (buff bytes.Buffer, err error) {

	// 处理订单商品的成本、售价及规格
	vPids := []int64{}
	vLogIds := []int64{}
	for _, v := range *vs {
		vPids = append(vPids, v.ProductId)
		vLogIds = append(vLogIds, v.LogisticsId)
	}
	vPids = base.UniqueInt64Slice(vPids)
	vLogIds = base.UniqueInt64Slice(vLogIds)

	mPrices, er := util.ListProductPriceByIds(vPids)
	if err = er; err != nil {
		glog.Error("Orders_Report fail! err=", err)
		return
	}

	// 获取物流信息
	vls, err := get_logistic_ids(vLogIds)
	mLogs := make(map[int64]*common.Logistics)
	for i, v := range vls {
		mLogs[v.Id] = &vls[i]
	}

	xlsFile := xlsx.NewFile()
	if err != nil {
		glog.Error("OpenFile fail! err=", err)
		return
	}
	sheeet, _ := xlsFile.AddSheet("帐单结算")
	headers := sheeet.AddRow()
	header := []string{"订单号", "收件人", "商品名称", "规格属性", "数量", "成本价", "零售价", "运费", "贡献值", "下单日期", "订单状态", "买家ID", "运单号", "实际支付币种", "实际支付金额", "交易总额(￥)", "供应商结算总额(￥)", "销售利润总额(￥)", "欧飞扣款", "备注"}
	for _, v := range header {
		headers.AddCell().SetValue(v)
	}

	var fee float64 = 4
	db := db.GetDB()
	for _, v := range *vs {
		row := sheeet.AddRow()
		row.AddCell().SetValue(v.Id) // 订单号
		//if name, _ := v.AddressInfo[""]; name
		var recever string
		recever, _ = v.AddressInfo["receiver"].(string)
		row.AddCell().SetValue(recever) // 收件人

		buf, _ := json.Marshal(v.Product)
		var p common.Product
		if err = json.Unmarshal(buf, &p); err != nil {
			glog.Error("Orders_Report fail! err=", err)
			return
		}
		row.AddCell().SetValue(p.Title)
		sku_cell := row.AddCell()
		if v.ProductSkuId > 0 {
			for _, s := range p.Skus {
				if s.Id == v.ProductSkuId {
					combines := []map[string]string{}
					buf, _ = json.Marshal(s.Combines)
					if err = json.Unmarshal(buf, &combines); err != nil {
						glog.Error("Orders_Report fail! err=", err)
						return
					}
					var s string
					for _, v := range combines {
						for k, val := range v {
							s = k + ":" + val
						}
						s += " "
					}
					s = strings.TrimRight(s, " ")
					sku_cell.SetValue(s) // 规格属性
				}
			}
		}
		row.AddCell().SetValue(v.Quantity) // 数量

		var cost_price decimal.Decimal

		if m, ok := mPrices[v.ProductId]; ok {
			if mm, ok := m[v.ProductSkuId]; ok {
				cost_price = mm.CostPrice
			}
		}
		row.AddCell().SetValue(cost_price.String()) // 成本

		// 规格售价
		price := p.Price
		if v.ProductSkuId > 0 {
			for _, s := range p.Skus {
				if s.Id == v.ProductSkuId {
					price = s.Price
				}
			}
		} else {
			price = p.Price
			// for _, v := range p.PayPrice {
			// 	if v.PayType == common.CURRENCY_KT {
			// 		price = v.SalePrice
			// 	}
			// }
		}

		row.AddCell().SetValue(price)                              // 售价
		row.AddCell().SetValue(fee)                                // 运费
		row.AddCell().SetValue(v.RebatAmount)                      // 贡献值
		row.AddCell().SetValue(v.Date)                             // 下单日期
		row.AddCell().SetValue(common.GetOrderStatusTxt(v.Status)) // 订单状态
		row.AddCell().SetValue(v.UserId)                           // 买家id
		cell := row.AddCell()
		if m, ok := mLogs[v.LogisticsId]; ok {
			cell.SetValue(m.Company + " " + m.Number) // 物流号
		}
		row.AddCell().SetValue(common.GetCurrencyName(v.CurrencyType)) // 实际支付币种
		row.AddCell().SetValue(v.TotalAmount)                          // 实际支付金额
		kt_aount := v.TotalAmount
		if v.CurrencyType != common.CURRENCY_KT {
			kt_aount, _ = price.Mul(decimal.New(int64(v.Quantity), 0)).Float64()
		}
		row.AddCell().SetValue(kt_aount) // 结算kt金额

		//总交易额
		row.AddCell().SetValue(cost_price.Mul(decimal.New(int64(v.Quantity), 0)).Add(decimal.NewFromFloat(fee)).Round(2)) //结算总额
		//row.AddCell().SetValue(p.Rebat)

		cost, _ := cost_price.Float64()
		row.AddCell().SetValue(kt_aount - float64(v.Quantity)*cost - v.RebatAmount - fee) // 利润

		if v.ProductType > common.GOODS_TYPE_PHYSICAL {
			var of_order common.OfOrder
			if err = db.Find(&of_order, "game_state=? and order_id=?", common.STATUS_OK, v.Id).Error; err != nil && err != gorm.ErrRecordNotFound {
				glog.Error("Orders_Report fail! err=", err)
				return
			}
			row.AddCell().SetValue(of_order.Ordercash.String()) // 欧飞实际扣款
		}
	}

	// buf := []byte{}
	// writer := bytes.NewBuffer(buf)

	if err = xlsFile.Write(&buff); err != nil {
		glog.Error("Orders_Report fail! err=", err)
		return
	}

	//data = writer.Bytes()
	return
	//c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", writer.Bytes())
}
