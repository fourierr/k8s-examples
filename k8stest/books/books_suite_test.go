package books

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBooks(t *testing.T) {
	// ginkgo通过调用Fail(description string)函数来发出fail信号，RegisterFailHandler()将Fail函数传递给Gomega。
	// RegisterFailHandler()连接了ginkgo和gomega
	RegisterFailHandler(Fail)
	// RunSpecs(t *testing.T, suiteDescription string) 告诉 Ginkgo 开始执行测试，如果任何一个specs失败则自动返回失败
	RunSpecs(t, "Books Suite")
}
