package sensiword

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var versionTestData = 1

var sensiWordTestDataS = []string{
	"gcd",
	"naive",
	"敏感词",
}

func ReadWord(path string) string {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("read file fail", err)
		panic("read file fail")
		return ""
	}
	defer f.Close()

	fd, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println("read to fd fail", err)
		panic("read to fd fail")
		return ""
	}

	return string(fd)
}

func QPS40000() (int, int) {
	return 10, 500
}

func QPS200000() (int, int) {
	return 1000, 9300
}

func TestAC_BuildFailurePaths(t *testing.T) {
	var sensiWordTestData2 = []string{
		"PASPD",
		"PHDAB",
		"HDC",
		"DBASK",
		"ASC",
		"ADPAS",
	}
	sensi := NewSensitive(true)
	sensi.Init(strconv.Itoa(versionTestData), sensiWordTestData2...)
	fmt.Println("-------")
	for _, vs := range sensi.ac.FailurePath {
		for k, v := range vs {
			fmt.Println(k, v.ID, string(v.Character), v.Parent.ID, string(v.Parent.Character))
		}
	}
}

func TestNewSensitive(t *testing.T) {
	sensi := NewSensitive(true)
	sensi.Init(strconv.Itoa(versionTestData), sensiWordTestDataS...)
	res := sensi.FindAll("hello, gcd, GDB,nanaive哈哈哈，敏感词, caicai我是谁")
	assert.Equal(t, []string{"gcd", "naive", "敏感词"}, res)
}

func TestLoadSensitive(t *testing.T) {
	words := strings.Split(ReadWord("./test_data/word0.txt"), "\n")
	start := time.Now()
	for i := 0; i < 10; i++ {
		sensi := NewSensitive(true)
		sensi.Init(strconv.Itoa(versionTestData), words...)
	}
	cost := time.Now().Sub(start)
	fmt.Printf("load %d words 10 times, total cost time : %s, avg time : %s",
		len(words), cost.String(), (cost/ 10).String())
	fmt.Println()
}

func TestPartRebuildSensitive(t *testing.T) {
	words := strings.Split(ReadWord("./test_data/word0.txt"), "\n")
	start := time.Now()
	sensi := NewSensitive(true)
	sensi.Init(strconv.Itoa(versionTestData), words...)
	fmt.Println("endLoad ", time.Now().Sub(start).String())
	start = time.Now()
	for i := 0; i < 10; i++ {
		versionTestData++
		sensi.PartRebuild(strconv.Itoa(versionTestData), words[:1000],  words[:1000])
	}
	cost := time.Now().Sub(start)
	fmt.Printf("part rebuild %d words 10 times, total cost time : %s, avg time : %s",
		len(words), cost.String(), (cost/ 10).String())
	fmt.Println()
}

func TestRebuildSensitive(t *testing.T) {
	words := strings.Split(ReadWord("./test_data/word0.txt"), "\n")
	start := time.Now()
	sensi := NewSensitive(true)
	sensi.Init(strconv.Itoa(versionTestData), words...)
	fmt.Println("endLoad ", time.Now().Sub(start).String())
	start = time.Now()
	for i := 0; i < 10; i++ {
		versionTestData++
		sensi.Rebuild(strconv.Itoa(versionTestData), words...)
	}
	cost := time.Now().Sub(start)
	fmt.Printf("rebuild %d words 10 times, total cost time : %s, avg time : %s",
		len(words), cost.String(), (cost/ 10).String())
	fmt.Println()
}

func sensitiveBetchFindAll(sensi *Sensitive, wg *sync.WaitGroup, stop chan struct{}) {
	var total, currentTotal int32
	gcount, mod := QPS200000()
	for i := 0; i <= gcount; i++ {
		go func() {
			wg.Add(1)
			for {
				select {
				case <-stop:
					wg.Done()
					return
				default:
					time.Sleep(time.Duration(rand.Int()%mod) * time.Microsecond)
					atomic.AddInt32(&total, 1)
					sensi.FindAll("hello, gcd, GDB,nanaive哈哈哈，敏感词, caicai我是谁")
				}
			}
		}()
	}
	go func() {
		wg.Add(1)
		for {
			select {
			case <-time.After(1 * time.Second):
				fmt.Println("total qps :", total-currentTotal)
				atomic.StoreInt32(&currentTotal, total)
				fmt.Println("---------------------------")
			case <-stop:
				wg.Done()
				return
			}
		}
	}()
}

func TestSensitiveFindAllWhenRebuild(t *testing.T) {
	words := strings.Split(ReadWord("./test_data/word0.txt"), "\n")
	words = append(words, sensiWordTestDataS...)
	start := time.Now()
	sensi := NewSensitive(true)
	sensi.Init(strconv.Itoa(versionTestData), words...)
	fmt.Println("endLoad ", time.Now().Sub(start).String())
	wg := &sync.WaitGroup{}
	stop := make(chan struct{})
	sensitiveBetchFindAll(sensi, wg, stop)
	start = time.Now()
	for i := 0; i < 10; i++ {
		versionTestData++
		sensi.Rebuild(strconv.Itoa(versionTestData), words...)
	}
	cost := time.Now().Sub(start)
	fmt.Printf("rebuild %d words 10 times, total cost time : %s, avg time : %s",
		len(words), cost.String(), (cost/ 10).String())
	fmt.Println()
	close(stop)
	wg.Wait()
}

func TestSensitiveFindAllWhenPartRebuild(t *testing.T) {
	words := strings.Split(ReadWord("./test_data/word0.txt"), "\n")
	words = append(words, sensiWordTestDataS...)
	start := time.Now()
	sensi := NewSensitive(true)
	sensi.Init(strconv.Itoa(versionTestData), words...)
	fmt.Println("endLoad ", time.Now().Sub(start).String())
	wg := &sync.WaitGroup{}
	stop := make(chan struct{})
	sensitiveBetchFindAll(sensi, wg, stop)
	start = time.Now()
	for i := 0; i < 10; i++ {
		versionTestData++
		sensi.PartRebuild(strconv.Itoa(versionTestData), words[:10000],  words[:10000])
	}
	cost := time.Now().Sub(start)
	fmt.Printf("part rebuild %d words 10 times, total cost time : %s, avg time : %s",
		len(words), cost.String(), (cost/ 10).String())
	fmt.Println()
	close(stop)
	wg.Wait()
}

func TestQps(t *testing.T) {
	words := strings.Split(ReadWord("./test_data/word0.txt"), "\n")
	words = append(words, sensiWordTestDataS...)
	start := time.Now()
	sensi := NewSensitive(true)
	sensi.Init(strconv.Itoa(versionTestData), words...)
	fmt.Println("endLoad ", time.Now().Sub(start).String())
	wg := &sync.WaitGroup{}
	stop := make(chan struct{})
	sensitiveBetchFindAll(sensi, wg, stop)
	time.Sleep(5 * time.Second)
	close(stop)
	wg.Wait()
}

func BenchmarkSensitive_FindAll(b *testing.B) {
	words := strings.Split(ReadWord("./test_data/word0.txt"), "\n")
	words = append(words, sensiWordTestDataS...)
	sensi := NewSensitive(true)
	sensi.Init(strconv.Itoa(versionTestData), words...)
	for i :=0; i < b.N ; i++{
		sensi.FindAll("hello, gcd, GDB,nanaive哈哈哈，敏感词, caicai我是谁")
	}
}

func BenchmarkSensitive_ReplaceAll(b *testing.B) {
	words := strings.Split(ReadWord("./test_data/word0.txt"), "\n")
	words = append(words, sensiWordTestDataS...)
	sensi := NewSensitive(true)
	sensi.Init(strconv.Itoa(versionTestData), words...)
	for i :=0; i < b.N ; i++{
		sensi.Replace("hello, gcd, GDB,nanaive哈哈哈，敏感词, caicai我是谁", []rune("*")[0])
	}
}

func BenchmarkSensitive_Rebuild(b *testing.B) {
	words := strings.Split(ReadWord("./test_data/word0.txt"), "\n")
	sensi := NewSensitive(true)
	sensi.Init(strconv.Itoa(versionTestData), words...)
	for i := 0; i < b.N; i++ {
		versionTestData++
		sensi.Rebuild(strconv.Itoa(versionTestData), words...)
	}
}

func BenchmarkSensitive_PartRebuild(b *testing.B) {
	words := strings.Split(ReadWord("./test_data/word0.txt"), "\n")
	sensi := NewSensitive(true)
	sensi.Init(strconv.Itoa(versionTestData), words...)
	for i := 0; i < b.N; i++ {
		versionTestData++
		sensi.PartRebuild(strconv.Itoa(versionTestData), words[:10000],  words[:10000])
	}
}