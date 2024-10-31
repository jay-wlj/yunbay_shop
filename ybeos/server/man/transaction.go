package man

import (
	"fmt"
	"github.com/jay-wlj/gobaselib/yf"
	"yunbay/ybeos/common"
	"yunbay/ybeos/dao"
	"yunbay/ybeos/server/share"
	"yunbay/ybeos/util"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

const (
	EOSTRANS_KEY string = "eostrans"
)

// type transactionSt struct {
// 	OrderId   int64  `json:"order_id"`
// 	ToAccount string `json:"to"`
// 	Amount    int64  `json:"amount"`
// 	Memo      string `json:"memo"`
// }

// func Transaction_Push(c *gin.Context) {
// 	var req transactionSt
// 	if ok := util.UnmarshalReq(c, &req); !ok {
// 		glog.Error("Transaction_Push fail! args invalid!", req)
// 		return
// 	}

// 	field := fmt.Sprintf("%v", req.OrderId)
// 	cache, err := dao.GetApiCache()
// 	if err == nil && cache != nil {
// 		if hash, err := cache.HGet(EOSTRANS_KEY, field); err == nil && hash != "" {
// 			yf.JSON_Ok(c, gin.H{"order_id": req.OrderId, "tx_hash": hash})
// 			return
// 		}
// 	}

// 	txHash, err := share.SendTransaction(req.ToAccount, req.Memo, req.Amount)
// 	if err != nil {
// 		glog.Error("Transaction_Push fail! err=", err)
// 		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
// 		return
// 	}
// 	if cache != nil {
// 		cache.HSet(EOSTRANS_KEY, field, txHash, 0)
// 	}
// 	if err := util.UpdateOrderRebat(req.OrderId, txHash); err != nil {
// 		glog.Error("Transaction_Push fail! err=", err)
// 		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
// 		return
// 	}

// 	yf.JSON_Ok(c, gin.H{"order_id": req.OrderId, "tx_hash": txHash})
// }

type tranSt struct {
	MQUrl *common.MQUrl `json:"mqurl"`
	Memo  string        `json:"memo"`
	Id    int64         `json:"id"`
}

func Transaction_Push(c *gin.Context) {
	var req tranSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		glog.Error("Transaction_Push fail! args invalid!", req)
		return
	}
	field := fmt.Sprintf("%v-%v", req.Id, req.Memo)
	cache, err := dao.GetApiCache()
	if err == nil && cache != nil {
		if hash, err := cache.HGet(EOSTRANS_KEY, field); err == nil && hash != "" {
			if req.MQUrl != nil {
				if err := util.UpdateOrderNotify(req.Id, hash, req.MQUrl); err != nil {
					glog.Error("Transaction_Push fail! err=", err)
					yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
					return
				}
			}
			yf.JSON_Ok(c, gin.H{"order_id": req.Id, "tx_hash": hash})
			return
		}
	}

	txHash, err := share.SendTransaction("", req.Memo, 0)
	if err != nil {
		glog.Error("Transaction_Push fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if cache != nil {
		cache.HSet(EOSTRANS_KEY, field, txHash, 0)
	}

	if req.MQUrl != nil {
		if err := util.UpdateOrderNotify(req.Id, txHash, req.MQUrl); err != nil {
			glog.Error("Transaction_Push fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}

	yf.JSON_Ok(c, gin.H{"id": req.Id, "tx_hash": txHash})
}
