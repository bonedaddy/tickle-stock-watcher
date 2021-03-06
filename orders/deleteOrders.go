package orders

import (
	"fmt"

	"github.com/helloworldpark/tickle-stock-watcher/analyser"
	"github.com/helloworldpark/tickle-stock-watcher/commons"
	"github.com/helloworldpark/tickle-stock-watcher/structs"
	"github.com/helloworldpark/tickle-stock-watcher/watcher"
)

type deleteOrder struct {
	action Action
}

func (o *deleteOrder) Name() string {
	return "delete"
}

func (o *deleteOrder) IsValid(args []string) error {
	if len(args) == 0 {
		return newError(fmt.Sprintf("Invalid number of arguments: need more than 1, got %d", len(args)))
	}
	if len(args) == 2 {
		if args[1] != "buy" && args[1] != "sell" {
			return newError("Invalid optional arguments: last argument must be either 'buy' or 'sell'")
		}
	}
	if len(args) == 3 {
		arg1 := args[1]
		arg2 := args[2]
		if arg1 > arg2 {
			arg1, arg2 = arg2, arg1
		}
		if arg1 != "buy" || arg2 != "sell" {
			return newError("Invalid optional arguments: last 2 arguments must be both 'buy' and 'sell'")
		}
	}
	if len(args) > 3 {
		return newError(fmt.Sprintf("Invalid number of arguments: too much, got %d", len(args)))
	}
	return nil
}

func (o *deleteOrder) SetAction(a Action) {
	o.action = a
}

func (o *deleteOrder) OnAction(user structs.User, args []string) error {
	err := o.IsValid(args)
	if err != nil {
		return err
	}
	return o.action(user, args)
}

func (o *deleteOrder) IsAsync() bool {
	return true
}

func (o *deleteOrder) IsPublic() bool {
	return false
}

// NewDeleteOrder order 'delete'
func NewDeleteOrder() Order {
	return &deleteOrder{}
}

// DeleteOrder order for 'delete'
func DeleteOrder(
	broker analyser.BrokerAccess,
	stockinfo watcher.StockAccess,
	price watcher.WatcherAccess,
	onSuccess func(user structs.User, stockname, stockid string)) Action {
	f := func(user structs.User, args []string) error {
		stockvar := args[0]
		stock, ok := stockinfo.AccessStockItem(stockvar)
		if !ok {
			stock, ok = stockinfo.AccessStockItemByName(stockvar)
			if !ok {
				firstCharDiff := stockvar[0] - "0"[0]
				if 0 <= firstCharDiff && firstCharDiff <= 9 {
					return newError(fmt.Sprintf("Invalid stock ID: %s", stockvar))
				}
				return newError(fmt.Sprintf("Invalid stock name: %s", stockvar))
			}
		}
		deleteStrategies := func(orderside int) error {
			err := broker.AccessBroker().DeleteStrategy(user, stock.StockID, orderside)
			if err != nil {
				return err
			}
			ok := price.AccessWatcher().Withdraw(stock)
			if !ok {
				return newError(fmt.Sprintf("Failed to stop watching stock %s(%s)", stock.Name, stock.StockID))
			}
			return nil
		}

		var err error
		switch len(args) {
		case 1, 3:
			strategies := broker.AccessBroker().GetStrategy(user)
			if len(strategies) == 0 {
				return newError("No strategies to delete")
			}
			deleted := 0
			for i := range strategies {
				if strategies[i].StockID != stock.StockID {
					continue
				}
				err = deleteStrategies(strategies[i].OrderSide)
				if err != nil {
					return err
				}
				deleted++
			}
			if deleted == 0 {
				return newError("No strategies to delete")
			}
		case 2:
			if args[1] == "buy" {
				err = deleteStrategies(commons.BUY)
			} else {
				err = deleteStrategies(commons.SELL)
			}
		default:
			break
		}
		if err != nil {
			return err
		}
		onSuccess(user, stock.Name, stock.StockID)
		return nil
	}
	return f
}
