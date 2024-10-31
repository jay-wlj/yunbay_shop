package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"testing"
)

func TestRegexp1(t *testing.T) {
	s := "尺寸（S,M,L,XL）"
	res, _ := regexp.Compile("(?<=()S+(?=))")
	a := res.FindString(s)

	fmt.Println(a)
}

func TestRegexp2(t *testing.T) {
	s := `'<div style=\"text-align: center,\"><img alt=\"\" src=\"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2018/08/IMG_1345.JPG\" style=\"width: 380px, height: 588px,\" /></div>'`
	res, _ := regexp.Compile(`src=\\\"([^"]*)\\\".*?width: (\d*)px, height: (\d*)px,`)

	as := res.FindAllStringSubmatch(s, -1)

	vs := []interface{}{}
	for _, v := range as {
		p := descImgs{}
		for i, j := range v {
			switch i {
			case 1:
				p.Path = j
			case 2:
				p.Width, _ = strconv.Atoi(j)
			case 3:
				p.Height, _ = strconv.Atoi(j)
			}
		}
		vs = append(vs, p)
	}
	buf, _ := json.Marshal(vs)
	fmt.Println(string(buf))
}
func TestSlice(t *testing.T) {
	s := []int{23, 54, 123, 435, 65}

	for len(s) > 0 {
		s1 := s[len(s)-1]
		ss := s[:len(s)]

		s = ss
		fmt.Println(s1)
	}
	fmt.Println(s)
}
