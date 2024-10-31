package util

import (

	//"yunbay/ybasset/common"
)


type uidSt struct {
	UserId int64 `json:"user_id"`
}
// 注册用户im帐号
func RegisterIMUser(user_id int64)(err error){
	uri := "/man/user/register"
	v := uidSt{UserId:user_id}
	if err = post_info(uri, "ybim", nil, &v, "", nil, false, EXPIRE_RES_INFO); err != nil {
		return
	}
	return
}

