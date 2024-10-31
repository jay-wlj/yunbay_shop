package util

import (
	"github.com/spaolacci/murmur3"
)

func MMHash(buf []byte) uint64 {
	return murmur3.Sum64WithSeed(buf, 0)
}

// import "fmt"
// func main() {
// 	for i := 0; i < 10; i++ {
// 		str := fmt.Sprintf("test-%d", i)
// 		fmt.Printf("%x\n", MMHash([]byte(str)))
// 	}
// 	fmt.Println(MMHash([]byte("test")))
// 	fmt.Println(MMHash([]byte("test2")))
// }
