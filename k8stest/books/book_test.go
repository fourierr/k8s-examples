package books

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Book", func() {
	var (
		longBook  Book
		shortBook Book
	)

	BeforeEach(func() {
		longBook = Book{
			Title:  "Les Miserables",
			Author: "Victor Hugo",
			Pages:  2783,
		}

		shortBook = Book{
			Title:  "Fox In Socks",
			Author: "Dr. Seuss",
			Pages:  24,
		}
	})

	Describe("Categorizing book length", func() {
		Context("With more than 300 pages", func() {
			It("should be a novel", func() {
				By("start test `should be a novel`")
				Expect(longBook.CategoryByLength()).To(Equal("NOVEL"))
			})
		})

		Context("With fewer than 300 pages", func() {
			It("should be a short story", func() {
				Expect(shortBook.CategoryByLength()).To(Equal("SHORT STORY"))
			})
		})
	})
})

var _ = Describe("Book", func() {
	var book Book

	BeforeEach(func() {
		book = Book{
			Title:  "Les Miserables",
			Author: "Victor Hugo",
			Pages:  1488,
		}
	})

	It("can be loaded from JSON", func() {
		Expect(book.Title).To(Equal("Les Miserables"))
		Expect(book.Author).To(Equal("Victor Hugo"))
		Expect(book.Pages).To(Equal(1488))
	})

	It("can extract the author's last name", func() {
		Expect(book.Author).To(Equal("Victor Hugo"))
	})
})
