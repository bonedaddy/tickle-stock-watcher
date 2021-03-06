package watcher_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/helloworldpark/tickle-stock-watcher/commons"
	"github.com/helloworldpark/tickle-stock-watcher/database"
	"github.com/helloworldpark/tickle-stock-watcher/logger"
	"github.com/helloworldpark/tickle-stock-watcher/structs"
	"github.com/helloworldpark/tickle-stock-watcher/watcher"
)

func TestWatcher(t *testing.T) {
	w := watcher.New(nil, time.Millisecond*500)
	w.Register(structs.Stock{Name: "Samsung Electronics", StockID: "005930", MarketType: structs.KOSPI})
	w.Register(structs.Stock{Name: "Korean Air", StockID: "003490", MarketType: structs.KOSPI})
	w.Register(structs.Stock{Name: "Hanwha Chemicals", StockID: "009830", MarketType: structs.KOSPI})

	fmt.Printf("%v Started!!!!\n", time.Now().Unix())
	timer := time.NewTimer(time.Duration(10) * time.Second)
	go func() {
		<-timer.C
		w.StopWatchingStock("003490")
	}()
	handle := w.StartWatchingStock("003490")
	for v := range handle {
		fmt.Println(v)
	}
	fmt.Printf("%v Finished!!!!\n", time.Now().Unix())

	// timer = time.NewTimer(time.Duration(10) * time.Second)
	// go func() {
	// 	<-timer.C
	// 	w.StopWatching()
	// }()
	// handle = w.StartWatching(time.Duration(500) * time.Millisecond)
	// for v := range handle {
	// 	fmt.Println(v)
	// }
	// fmt.Printf("%v Finished Again!!!!\n", time.Now().Unix())
}

func TestCrawlPast(t *testing.T) {
	credential := database.LoadCredential("/Users/shp/Documents/projects/tickle-stock-watcher/credee.json")
	client := database.CreateClient()
	client.Init(credential)
	client.Open()

	defer client.Close()

	client.RegisterStructFromRegisterables([]database.DBRegisterable{
		structs.Stock{},
		structs.StockPrice{},
		structs.WatchingStock{},
	})

	w := watcher.New(client, time.Millisecond*500)
	w.Register(structs.Stock{StockID: "271980", MarketType: structs.KOSPI})
	w.Register(structs.Stock{StockID: "272450", MarketType: structs.KOSPI})
	w.Register(structs.Stock{StockID: "272550", MarketType: structs.KOSPI})
	w.Register(structs.Stock{StockID: "280360", MarketType: structs.KOSPI})
	w.Register(structs.Stock{StockID: "281820", MarketType: structs.KOSPI})
	w.Register(structs.Stock{StockID: "282330", MarketType: structs.KOSPI})
	w.Register(structs.Stock{StockID: "282690", MarketType: structs.KOSPI})
	w.Register(structs.Stock{StockID: "284740", MarketType: structs.KOSPI})

	now := time.Now().Unix()
	w.Collect()
	elapsed := time.Now().Unix() - now
	logger.Info("Finished!! %d mins", elapsed/60)
}

func TestReflect(t *testing.T) {
	aa := reflect.TypeOf([]structs.Stock{structs.Stock{}})
	fmt.Println(aa)
	fmt.Printf("Name: %v Kind: %v Elem: %v", aa.Name(), aa.Kind(), aa.Elem())
}

func TestWatchAndOff(t *testing.T) {
	type testCrawler struct {
		sentinel chan struct{}
	}

	crawlers := make(map[string]*testCrawler)
	const key = "aaa"
	crawlers[key] = &testCrawler{
		sentinel: make(chan struct{}),
	}

	dummyjob := func(id string, zero int) int {
		fmt.Printf("%v: %s\n", time.Now(), id)
		return zero
	}
	out := make(chan int)
	commons.InvokeGoroutine("Watcher_StartWatchingStock_"+key, func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			select {
			case out <- dummyjob(key, 9):
				continue
			case <-crawlers[key].sentinel:
				fmt.Println("Got Sentinel")
				return
			}
		}
		fmt.Printf("[Watcher] Finish TestWatchAndOff: %s", key)
		close(out)
	})

	go func() {
		for {
			select {
			case v, ok := <-out:
				if ok {
					fmt.Println(v)
				} else {
					fmt.Println("Ruined")
				}
				fmt.Println("I'm running?")
			}
			fmt.Println("Machine Running")
		}
	}()

	timer := time.NewTimer(10 * time.Second)
	<-timer.C

	b := "aaa"
	crawlers[b].sentinel <- struct{}{}
	// close(crawlers[b].sentinel)

	timer = time.NewTimer(10 * time.Second)
	<-timer.C

	fmt.Println("Finished")
}
