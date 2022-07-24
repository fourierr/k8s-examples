package ctx_practise

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

// **************************
// context.WithValue

const (
	KEY = "trace_id"
)

func NewRequestID() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)
}

func NewContextWithTraceID() context.Context {
	ctx := context.WithValue(context.Background(), KEY, NewRequestID())
	return ctx
}

func PrintLog(ctx context.Context, message string) {
	fmt.Printf("%s|info|trace_id=%s|%s", time.Now().Format("2006-01-02 15:04:05"), GetContextValue(ctx, KEY), message)
}

func GetContextValue(ctx context.Context, k string) string {
	v, ok := ctx.Value(k).(string)
	if !ok {
		return ""
	}
	return v
}

func ProcessEnter(ctx context.Context) {
	PrintLog(ctx, "Golang梦工厂")
}

// *************************************
// context.WithTimeout

func NewContextWithTimeout() (context.Context,context.CancelFunc) {
	return context.WithTimeout(context.Background(), 13 * time.Second)
}

func HttpHandler()  {
	ctx, cancel := NewContextWithTimeout()
	defer cancel()
	deal(ctx)
}

func deal(ctx context.Context)  {
	for i:=0; i< 10; i++ {
		time.Sleep(1*time.Second)
		select {
		case <- ctx.Done():
			fmt.Println(ctx.Err())
			return
		default:
			fmt.Printf("deal time is %d\n", i)
		}
	}
}

// *************************************
// context.WithCancel

func Speak(ctx context.Context)  {
	for range time.Tick(time.Second){
		select {
		case <- ctx.Done():
			fmt.Println("我要闭嘴了")
			return
		default:
			fmt.Println("balaclava's")
		}
	}
}