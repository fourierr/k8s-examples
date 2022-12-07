package main

import (
	"context"
	"fmt"
	"runtime"
)

// 责任链模式-错误处理最佳实践
// 实现一个下单逻辑：1.参数校验，2.获取收货地址，地址校验，3.加入购物车，4.检验库存等等后续逻辑

// Handler 处理
type Handler interface {
	// 自身的业务
	Do(c context.Context) error
	// 设置下一个对象
	SetNext(h Handler) Handler
	// 执行
	Run(c context.Context) error
}

// Next 被抽象出来用做合成复用的结构体
// golang没有继承的概念，想达到继承的效果就采用合成复用的机制来实现。
type Next struct {
	// 下一个对象
	nextHandler Handler
}

// SetNext 实现可被复用的SetNext方法
// 返回值是下一个对象 方便写成链式代码优雅
func (n *Next) SetNext(h Handler) Handler {
	n.nextHandler = h
	return h
}

// Run 执行
func (n *Next) Run(c context.Context) (err error) {
	// 由于go无继承的概念 这里无法执行当前handler的Do,不能 n.Do(cache)
	if n.nextHandler != nil {
		// 合成复用下的变种
		// 执行下一个handler的Do
		if err = (n.nextHandler).Do(c); err != nil {
			return
		}
		// 执行下一个handler的Run
		return (n.nextHandler).Run(c)
	}
	return
}

// NullHandler 空Handler
type NullHandler struct {
	// 合成复用Next的`nextHandler`成员属性、`SetNext`成员方法、`Run`成员方法
	Next
}

// Do 空Handler的Do
func (h *NullHandler) Do(c context.Context) (err error) {
	// 空Handler 这里什么也不做 只是载体 do nothing...
	return
}

// ArgumentsHandler 校验参数的handler
type ArgumentsHandler struct {
	// 合成复用Next
	Next
}

// Do 校验参数的逻辑
func (h *ArgumentsHandler) Do(c context.Context) (err error) {
	fmt.Println(runFuncName(), "校验参数成功...")
	return
}

// AddressInfoHandler 地址信息handler
type AddressInfoHandler struct {
	// 合成复用Next
	Next
}

// Do 校验参数的逻辑
func (h *AddressInfoHandler) Do(c context.Context) (err error) {
	fmt.Println(runFuncName(), "获取地址信息...")
	fmt.Println(runFuncName(), "地址信息校验...")
	str = "i am AddressInfoHandler"
	return
}

// CartInfoHandler 获取购物车数据handler
type CartInfoHandler struct {
	// 合成复用Next
	Next
}

// Do 校验参数的逻辑
func (h *CartInfoHandler) Do(c context.Context) (err error) {
	fmt.Println(runFuncName(), "获取购物车数据...")
	fmt.Println("CartInfoHandler: ", str)
	return
}

// StockInfoHandler 商品库存handler
type StockInfoHandler struct {
	// 合成复用Next
	Next
}

// Do 校验参数的逻辑
func (h *StockInfoHandler) Do(c context.Context) (err error) {
	fmt.Println(runFuncName(), "获取商品库存信息...")
	fmt.Println(runFuncName(), "商品库存校验...")
	return
}

// 获取正在运行的函数名
func runFuncName() string {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}

var str string

func main() {
	// 初始化空handler
	nullHandler := &NullHandler{}
	// 链式调用 逻辑关系一览无余
	nullHandler.SetNext(&ArgumentsHandler{}).
		SetNext(&AddressInfoHandler{}).
		SetNext(&CartInfoHandler{}).
		SetNext(&StockInfoHandler{})
	// 开始执行业务
	if err := nullHandler.Run(context.Background()); err != nil {
		// 异常
		fmt.Println("Fail | Error:" + err.Error())
		return
	}
	// 成功
	fmt.Println("Success")
	return
}
