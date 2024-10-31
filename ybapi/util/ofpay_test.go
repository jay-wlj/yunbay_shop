package util

import (
	"fmt"
	"testing"

	"github.com/jie123108/glog"
)

func TestOfPayTelCheck(t *testing.T) {
	err := GetOfpay().Tel_Check("15418717950", 100)

	if err != nil {
		glog.Error("err=", err)
	}
}

func TestOfTelRecharge(t *testing.T) {
	v, err := GetOfpay().Tel_Recharge(1234, "15818717950", 100)
	if err != nil {
		glog.Error("err=", err)
	}
	fmt.Println("body=", v)
}

func TestCardWidthdraw(t *testing.T) {
	v, err := GetOfpay().CardWidthdraw(1234, "1711287", 3)
	if err != nil {
		glog.Error("err=", err)
	}
	fmt.Println("body=", v)
}
