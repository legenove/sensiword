package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
)

func getRandom() int {
	return rand.Int() % 120
}

func main() {
	words := map[string]bool{}
	wordLines := strings.Split(ReadWord("./word.txt"), "\n")
	runes := make([]rune, 50000)
	for _, l := range wordLines {
		runes = append(runes,  []rune(l)...)
	}
	fmt.Println(len(runes))
	fileName := "./word0.txt"
	f, err3 := os.Create(fileName) //创建文件
	if err3 != nil{
		fmt.Println("create file fail")
	}
	w := bufio.NewWriter(f) //创建新的 Writer 对象
	for i:= 0; i < len(runes); i++ {
		for j := 1; j <= 10; j ++ {
			if j == 1 {
				if getRandom() > 50 {
					continue
				}
			} else {
				if getRandom() > (110 - j*j) {
					continue
				}
			}
			if i + j >= len(runes) {
				continue
			}
			if ok := words[string(runes[i:i+j])]; !ok {
				w.WriteString(string(runes[i:i+j]))
				w.WriteString("\n")
				words[string(runes[i:i+j])] = true
			}
		}
	}
	w.Flush()
	f.Close()
}

