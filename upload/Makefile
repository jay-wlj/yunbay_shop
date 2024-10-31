all: linux mac

mac: main.go 
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o upload.mac main.go

linux: main.go
	GOOS=linux GOARCH=amd64  go build -o upload.linux main.go


# 把Go程序变小的办法是：
# go build -ldflags "-s -w" (go install类似)
# -s去掉符号表（然后panic时候的stack trace就没有任何文件名/行号信息了
# 这个等价于普通C/C++程序被strip的效果）
# -w去掉DWARF调试信息，得到的程序就不能用gdb调试了

# 比如，server.go是一个简单的http server，用了net/http包