package man

import (
	"regexp"
	"strings"
	"time"
	"yunbay/ybapi/common"
	"yunbay/ybapi/dao"
	"yunbay/ybapi/server/share"
	"yunbay/ybapi/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

// 抽奖活动列表
func Lotterys_List(c *gin.Context) {
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	p_id, _ := base.CheckQueryInt64Field(c, "p_id")

	vs := []common.Lotterys{}
	db := db.GetDB()
	if p_id > 0 {
		db.DB = db.Where("p_id=?", p_id)
	}
	var total int
	var err error
	if err = db.Model(&common.Lotterys{}).Count(&total).Error; err != nil {
		glog.Error("Lotterys fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if err = db.ListPage(page, page_size).Order("status asc, create_time desc").Find(&vs).Error; err != nil {
		glog.Error("Lotterys fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"list": vs, "total": total, "list_ended": base.IsListEnded(page, page_size, len(vs), total)})
}

// 抽奖活动详情
func Lotterys_Info(c *gin.Context) {
	id, _ := base.CheckQueryInt64Field(c, "id")
	if id <= 0 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	var v common.Lotterys
	if err := db.GetDB().Find(&v, "id=?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, yf.ERR_NOT_EXISTS)
			return
		}
		glog.Error("Lotterys_Info fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	v.Now = time.Now().Unix()
	yf.JSON_Ok(c, v)
}

// 抽奖记录列表
func Lotterys_Record(c *gin.Context) {
	id, _ := base.CheckQueryInt64Field(c, "id")
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	if id <= 0 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}

	v := dao.Lotterys{}
	ret, err := v.ListRecord(id, page, page_size)
	if err != nil {
		glog.Error("Lotterys fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, ret)
}

// 抽奖活动修改接口
func Lotterys_Upsert(c *gin.Context) {
	var req common.Lotterys
	if err := c.ShouldBindJSON(&req); err != nil {
		yf.JSON_Fail(c, err.Error())
		return
	}
	// if ok := util.UnmarshalReq(c, &req); !ok {
	// 	return
	// }
	db := db.GetDB()
	if req.Id > 0 {
		// 判断只有开始前才可以修改
		db.DB = db.Where("status=?", common.STATUS_INIT)

	}

	// 每人支付 = 奖品金额*数量 / 参与人数
	req.Amount = req.Price.Mul(decimal.New(int64(req.Num), 0)).Div(decimal.New(int64(req.Stock), 0)).Round(4)

	// 选择相应的商品
	p, err := util.GetProductInfo(req.PId, 0)
	if err != nil {
		glog.Error("Lotterys_Upsert fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if common.PUBLISH_AREA_LOTTERYS != p.PublishArea {
		// 非抽奖专区的商品禁止添加
		yf.JSON_Fail(c, "禁止添加非抽奖专区商品")
		return
	}

	req.Product = base.FilterStruct(p, true, "title", "images").(map[string]interface{}) // 保留商品的标题及图

	req.Pertimes = 10 //默认限制参与次数

	res := db.Save(&req)

	if err := res.Error; err != nil {
		glog.Error("Lotterys_Upsert fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if 0 == res.RowsAffected {
		yf.JSON_Fail(c, common.ERR_FORBIDDENT_MODIFY)
		return
	}
	v := dao.Lotterys{}
	v.Refresh(req.Id) // 刷新缓存
	share.GetLottery().Notify()
	yf.JSON_Ok(c, gin.H{})
}

func Lotterys_Hid(c *gin.Context) {
	var req idstatus
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	db := db.GetTxDB(c)
	res := db.Model(&common.Lotterys{}).Where("id=?", req.Id).Updates(base.Maps{"hid": req.Status})
	if res.Error != nil {
		glog.Error("Lotterys_Del fail! err=", res.Error)
		yf.JSON_Fail(c, res.Error.Error())
		return
	}
	if res.RowsAffected > 0 {
		db.AfterCommit(func() {
			v := &dao.Lotterys{}
			v.Refresh(req.Id)
		})
	}
	yf.JSON_Ok(c, gin.H{})
}

type idstatus struct {
	Id     int64 `json:"id" binding:"gt=0"`
	Status int   `json:"status"`
}

type OrderRebatSt struct {
	OrderId int64  `json:"order_id" binding:"gt=0"`
	TxHash  string `json:"tx_hash" binding:"required"`
}

func Lotterys_RecordHash(c *gin.Context) {
	var req OrderRebatSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	numHash := req.TxHash
	// 将字母去掉
	// 将获得hash值非(1～9)的字符去掉后，取末尾4位数字作为小数点后的数字以生成小数，再通过小数转换为百分数
	r, _ := regexp.Compile("[^0-9]")
	bt := r.ReplaceAll([]byte(req.TxHash), []byte{}) // 将非0-9的字符剔除
	numHash = string(bt)
	numHash = strings.TrimPrefix(numHash, "0") // 移动前面的0
	db := db.GetTxDB(c)
	res := db.Model(&common.LotterysRecord{}).Where("id=?", req.OrderId).Updates(base.Maps{"hash": req.TxHash, "num_hash": numHash})
	if err := res.Error; err != nil {
		glog.Error("Lotterys_RecordHash fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if 0 == res.RowsAffected {
		yf.JSON_Ok(c, gin.H{})
		return
	}
	var v common.LotterysRecord
	if err := db.Find(&v, "id=?", req.OrderId).Error; err != nil {
		glog.Error("Lotterys_RecordHash fail! err=", err)
		return
	}
	sv := share.Lotterys(v.LotterysId)
	if err := sv.HandleLotterys(db); err != nil {
		glog.Error("Lotterys_RecordHash fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	db.AfterCommit(func() {
		// 刷新购买记录缓存
		cl := &dao.Lotterys{}
		cl.RefreshRecord(v.LotterysId, 0) // 刷新抽奖下的购买记录缓存
		cl.RefreshUserRecord(v.UserId)    // 刷新用户购买记录列表

		notify := &util.LotterysNotify{}
		notify.NotifyHash(v.UserId, v.Id, v.NumHash) // 通知前端更新该记录hash
	})
	db.Commit()
	yf.JSON_Ok(c, gin.H{})
}
