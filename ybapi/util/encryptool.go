package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io/ioutil"
	. "yunbay/ybapi/conf"
	rnd "math/rand"
	"github.com/jie123108/glog"
)

var (
	block           cipher.Block
	ascii_table     string
	aes_iv          []byte
	aes_key         []byte
	rsa_public_key  []byte
	rsa_private_key []byte
)

func init() {
	var err error
	ascii_table = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	aes_key = []byte("0a4e07a222ad782e559235d1a7b30e4c")
	aes_iv = []byte("2b00042f7481c7b056c4b410d28f33cf")
	block, err = aes.NewCipher(aes_key)
	if err != nil {
		glog.Error("NewCipher(", aes_key, ") failed! err:", err)
		panic(err.Error())
	}
	aes_iv = aes_iv[:block.BlockSize()]
}

func RandomSample(letters string, n int) string {
	b := make([]byte, n)
	llen := len(letters)
	for i := range b {
		b[i] = letters[rnd.Intn(llen)]
	}
	return string(b)
}

func ReloadRsaKey() (err error) {
	// 读取公私钥
	rsa_public_key, err = ioutil.ReadFile(Config.RsaPublicKeyFile)
	if err != nil {
		glog.Error("public_key read fail!", err, " path:", Config.RsaPublicKeyFile)
		panic(err.Error())
	}
	//glog.Error("public_Key:", string(public_Key))
	//rsa_public_key, _ = base64.StdEncoding.DecodeString(string(public_Key))

	rsa_private_key, err = ioutil.ReadFile(Config.RsaPrivateKeyFile)
	if err != nil {
		glog.Error("private_key read fail!", err, " path:", Config.RsaPrivateKeyFile)
		panic(err.Error())
	}
	return
}

func TelEncrypt(tel string) string {
	src := []byte(tel)
	dst := make([]byte, len(src))
	encryptor := cipher.NewCFBEncrypter(block, aes_iv)
	encryptor.XORKeyStream(dst, src)
	return base64.StdEncoding.EncodeToString(dst)
}

func TelDecrypt(tel string) (s string, err error) {
	src, err := base64.StdEncoding.DecodeString(tel)
	if err != nil {
		return
	}
	dst := make([]byte, len(src))
	decryptor := cipher.NewCFBDecrypter(block, aes_iv)
	decryptor.XORKeyStream(dst, src)
	s = string(dst)
	return
}

func GetContractPublickey() string {
	return string(rsa_public_key)
}

func DecryptContractKey(data []byte) ([]byte, error) {
	return RsaDecrypt([]byte(data), rsa_private_key)
}

func DecryptContractInfo(data []byte, key []byte) []byte {
	klen := len(key)

	dst := make([]byte, len(data))
	for i, v := range data {
		dst[i] = v ^ key[i%klen]
	}
	return dst
}

func GenRsaKey(bits int) (err error, public_key, private_key string) {
	// 生成私钥文件
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	private_buf := new(bytes.Buffer)
	//file, err := os.Create("private.pem")
	// if err != nil {
	//     return err
	// }
	err = pem.Encode(private_buf, block)
	if err != nil {
		return
	}
	private_key = base64.StdEncoding.EncodeToString(private_buf.Bytes())
	//private_key = string(private_buf.Bytes())

	// 生成公钥文件
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	//file, err = os.Create("public.pem")
	// if err != nil {
	//     return err
	// }
	public_buf := new(bytes.Buffer)
	err = pem.Encode(public_buf, block)
	if err != nil {
		return
	}
	//public_key = string(public_buf.Bytes())
	public_key = base64.StdEncoding.EncodeToString(public_buf.Bytes())
	return
}

// rsa加密
func RsaEncrypt(origData []byte, publicKey []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

// rsa解密
func RsaDecrypt(ciphertext []byte, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}

type XRsa struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

func NewXRsa(publicKey []byte, privateKey []byte) (*XRsa, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	block, _ = pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)

	return &XRsa{
		publicKey:  pub,
		privateKey: priv,
	}, nil
}

// rsa加密
func (xr *XRsa) RsaEncrypt(origData []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, xr.publicKey, origData)
}

// rsa解密
func (xr *XRsa) RsaDecrypt(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, xr.privateKey, ciphertext)
}

func test() {
	err, pub_key, pri_key := GenRsaKey(1024)
	if err != nil {
		glog.Error("GenRsaKey fail! err=", err)
		return
	}
	s := "我们都 love you!"
	glog.Error(pub_key, "\n", pri_key)
	// rs, err := RsaEncrypt([]byte(s), []byte(pub_key))
	// glog.Error(string(rs))

	// ps, err := RsaDecrypt([]byte(rs), []byte(pri_key))
	// glog.Error(string(ps))
	rsa, err := NewXRsa([]byte(pub_key), []byte(pri_key))
	ps, err := rsa.RsaEncrypt([]byte(s))
	glog.Errorf("%v PublicEncrypt: %v", s, string(ps))

	prs, err := rsa.RsaDecrypt([]byte(ps))
	glog.Errorf("%v PrivateDecrypt: %v", ps, string(prs))
}
