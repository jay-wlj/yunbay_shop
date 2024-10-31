package dao

import (
	"fmt"
	"time"
	"encoding/json"
)


var expires time.Duration
func init() {
	var err error		
	expires, err = time.ParseDuration("720h")
	if err != nil {
		return
	}
}

type imToken struct {
	Accid string `json:"accid"`
	Token string `json:"im_token"`
}

func SaveImToken(user_id int64, accid, token string) (err error) {
	keyid := "imt:" + fmt.Sprintf("%v", user_id)
	v := imToken{Accid:accid, Token:token}
	body, _ := json.Marshal(v)

	cache, err1 := GetIMCache()
	if err1 != nil {
		err = err1
		return
	}

	err = cache.Set(keyid, string(body), expires)
	return
}

func GetImToken(user_id int64) (accid, token string, err error) {
	keyid := "imt:" + fmt.Sprintf("%v", user_id)
	var val string
	cache, err1 := GetIMCache()
	if err1 != nil {
		err = err1
		return
	}

	val, err = cache.Get(keyid)
	if err == nil {
		var v imToken
		if err = json.Unmarshal([]byte(val), &v); err == nil {
			accid = v.Accid
			token = v.Token
		}
	}
	return
}

func DelIm(user_id int64){
	keyid := "imt:" + fmt.Sprintf("%v", user_id)
	cache, err := GetIMCache()
	if err != nil {
		return
	}

	cache.Del(keyid)
	return
}
