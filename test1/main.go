/*
@Time : 2019/10/18 10:21
@Author : Tux
@File : main
@Description :
*/

package main

import (
	"sync"
)

type smap struct {
	sync.Mutex
	data map[string]string
}

func newSMap()  {
	var a error
}

func main() {

}
