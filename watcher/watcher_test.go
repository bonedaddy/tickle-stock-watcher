package watcher_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/helloworldpark/tickle-stock-watcher/database"
	"github.com/helloworldpark/tickle-stock-watcher/logger"
	"github.com/helloworldpark/tickle-stock-watcher/structs"
	"github.com/helloworldpark/tickle-stock-watcher/watcher"
)

func TestWatcher(t *testing.T) {
	w := watcher.New(nil)
	w.Register(structs.Stock{Name: "Samsung Electronics", StockID: "005930", MarketType: structs.KOSPI})
	w.Register(structs.Stock{Name: "Korean Air", StockID: "003490", MarketType: structs.KOSPI})
	w.Register(structs.Stock{Name: "Hanwha Chemicals", StockID: "009830", MarketType: structs.KOSPI})

	fmt.Printf("%v Started!!!!\n", time.Now().Unix())
	timer := time.NewTimer(time.Duration(10) * time.Second)
	go func() {
		<-timer.C
		w.StopWatchingStock("003490")
	}()
	handle := w.StartWatching(time.Duration(500) * time.Millisecond)
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

	w := watcher.New(client)
	w.Register(structs.Stock{Name: "Samsung Electronics", StockID: "005930", MarketType: structs.KOSPI})
	w.Register(structs.Stock{Name: "Korean Air", StockID: "003490", MarketType: structs.KOSPI})
	w.Register(structs.Stock{Name: "Hanwha Chemicals", StockID: "009830", MarketType: structs.KOSPI})

	w.Collect(time.Duration(1000)*time.Millisecond, time.Duration(250)*time.Millisecond)
	logger.Info("Finished!!")
}

func TestReflect(t *testing.T) {
	aa := reflect.TypeOf([]structs.Stock{structs.Stock{}})
	fmt.Println(aa)
	fmt.Printf("Name: %v Kind: %v Elem: %v", aa.Name(), aa.Kind(), aa.Elem())
}
