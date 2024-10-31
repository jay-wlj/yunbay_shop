package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/jay-wlj/gobaselib/yf"
	"math/rand"
	"strconv"
	"strings"

	"github.com/jie123108/glog"
)

type TokenInfo struct {
	Token      string
	UserType   int16
	UserId     int64
	ExpireTime int64
}

/**
AES 有五种加密模式：
电码本模式（Electronic Codebook Book (ECB)）
密码分组链接模式（Cipher Block Chaining (CBC)）
计算器模式（Counter (CTR)）
密码反馈模式（Cipher FeedBack (CFB)）
输出反馈模式（Output FeedBack (OFB)）
**/

// var encryptor cipher.Stream
// var decryptor cipher.Stream
var token_block cipher.Block

var token_key []byte
var token_iv []byte
var token_magic string
var ascii_table string

var TOKEN_V1_PREFIX string
var err_token_invalid error

func init() {
	var err error
	token_key = []byte("e74b7294770dbef89bdd8437caf1d89f")
	token_iv = []byte("64cd0a53dc63e04e21da472b11fa7278")
	token_magic = "84646d88f26c3c9ed0ce651b45bad235"[:16]
	ascii_table = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	TOKEN_V1_PREFIX = "T1"

	token_block, err = aes.NewCipher(token_key)
	if err != nil {
		glog.Error("NewCipher(", token_key, ") failed! err:", err)
		panic(err.Error())
	}
	token_iv = token_iv[:token_block.BlockSize()]

	err_token_invalid = fmt.Errorf(yf.ERR_TOKEN_INVALID)
}

func RandomSample(letters string, n int) string {
	b := make([]byte, n)
	llen := len(letters)
	for i := range b {
		b[i] = letters[rand.Intn(llen)]
	}
	return string(b)
}

func random_sid() string {
	return RandomSample(ascii_table, 8)
}

func token_cksum(token string) string {
	tmp := fmt.Sprintf("%x", md5.Sum([]byte(token+token_magic)))
	cksum := string(tmp[0:3])
	// fmt.Printf("token: %s, cksum [%s]\n", token, cksum)
	return cksum
}

func add_cksum(token string) string {
	// glog.Error("add_cksum(", token)
	return token + "." + token_cksum(token)
}

func remove_cksum(token_orig string) (token string, err error) {
	token_len := len(token_orig)
	if token_len <= 6 {
		return "", err_token_invalid
	}
	token = token_orig[0 : token_len-4]
	suffix := token_orig[token_len-4:]
	// glog.Info("token: ", token, ", suffix:", suffix)
	if suffix[0:1] != "." {
		return "", err_token_invalid
	}
	cksum_calc := token_cksum(token)
	cksum := suffix[1:]
	if cksum != cksum_calc {
		glog.Error("invalid token [", token_orig, "], the ok cksum is :", cksum_calc)
		return "", err_token_invalid
	}
	return token, nil
}

func TokenEncrypt(user_type int16, user_id int64, expire_time int64) string {
	token_tmp := fmt.Sprintf("%x|%x|%x|%s", user_type, user_id, expire_time, random_sid())
	src := []byte(token_tmp)
	dst := make([]byte, len(src))
	encryptor := cipher.NewCFBEncrypter(token_block, token_iv)
	encryptor.XORKeyStream(dst, src)
	token_tmp = base64.StdEncoding.EncodeToString(dst)
	return add_cksum(TOKEN_V1_PREFIX + token_tmp)
}

func TokenDecrypt(token string) (user_type int16, user_id int64, expire_time int64, err error) {
	ver := token[0:2]
	if ver == TOKEN_V1_PREFIX {
		var ok_token string
		ok_token, err = remove_cksum(token)
		if err != nil {
			return
		}
		ok_token = ok_token[2:]
		var ok_token_bt []byte
		ok_token_bt, err = base64.StdEncoding.DecodeString(ok_token)
		if err != nil {
			glog.Error("base64.DecodeString(", ok_token, ") failed! err:", err)
			err = err_token_invalid
			return
		}
		dst := make([]byte, len(ok_token_bt))
		decryptor := cipher.NewCFBDecrypter(token_block, token_iv)
		decryptor.XORKeyStream(dst, ok_token_bt)
		// glog.Info("ok_token: ", ok_token, ", dst:", dst)
		ok_token = string(dst)
		arr := strings.SplitN(ok_token, "|", 4)
		if len(arr) != 4 {
			glog.Error("invalid token:", ok_token)
			err = err_token_invalid
			return
		}
		var user_type_i64 int64
		user_type_i64, err = strconv.ParseInt(arr[0], 16, 16)
		if err != nil {
			return
		}
		user_type = int16(user_type_i64)
		user_id, err = strconv.ParseInt(arr[1], 16, 64)
		if err != nil {
			return
		}
		expire_time, err = strconv.ParseInt(arr[2], 16, 64)
		if err != nil {
			return
		}
		return
	} else {
		err = err_token_invalid
	}
	fmt.Printf(ver)
	return
}
