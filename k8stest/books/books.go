package books

import (
	"io/ioutil"
	"net/http"
)

type Book struct {
	Title  string
	Author string
	Pages  int
}

func (b *Book) CategoryByLength() string {

	if b.Pages >= 300 {
		return "NOVEL"
	}

	return "SHORT STORY"
}

func (b *Book) SendHttp() ([]byte, error) {
	client := &http.Client{}
	request, err := http.NewRequest("GET", "127.0.0.1:80", nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck
	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}
