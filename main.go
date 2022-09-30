package main

import (
	//"fmt"
	"sync"
	"net/http"
	"log"
	"time"
	"os"
	"io/ioutil"
	"strings"
)


type loger interface{
	Print(...any)
}

type UrlCount struct{
	Url string
	CountGo int
}

// для подсчёта количества вхождений слова в строке
func getCountWord(st string, findWord string)int{
	input := strings.Fields(st)
	var count int
    for _, word := range input {
        if word == findWord {
            count++
        }
    }
    return count
}


func generatorUrlCount(urls []string, countGoruting int, chUrlCount chan UrlCount, word string, log loger){
	// для контроля за одновременно выполняющимися гоурутинами, не более countGoruting буферизованный канал
	chCountGoruting := make(chan int, countGoruting) 
	wg := sync.WaitGroup{}
	var httpClient = http.Client{Timeout: time.Second * 5}

	for _, url := range urls{
		wg.Add(1)
		chCountGoruting <- 0  

		go func(i string) {
			resp, err := httpClient.Get(url)
			if err != nil {
				log.Print("goroutin went away with url ", url)
				return  
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Print("goroutin went away with url ", url)
				return
			}

			// читаем тело ответа 
			body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
				log.Print("goroutin went away with url ", url)
				return  
			}

			chUrlCount <- UrlCount{
				Url: url,
				CountGo: getCountWord(string(body), word),
			}

			<- chCountGoruting
			wg.Done()
			 
		}(url)
	}

	go func(){
		wg.Wait()
		close(chUrlCount)
	}()
 }


func getUrlsGo(urls []string, countGoruting int, word string, log loger){
	if len(urls) == 0 { // пограничный случай
		return
	}

	chUrlCount := make(chan UrlCount)

	go generatorUrlCount(urls,countGoruting, chUrlCount, word, log)

	var sum int
	for i := range chUrlCount{
		log.Print(i.Url, " ", i.CountGo)
		sum += i.CountGo
	}
	log.Print(sum)	

}


func main(){
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	infoLog.Print("start")

	// вводные данные
	var data = []string{
		"https://golang.org",
		"https://golang.org",
	}
	countGoruting := 5
	wordExample := "Go"

	//основная функция
	getUrlsGo(data, countGoruting, wordExample, infoLog)  

	infoLog.Print("end")
}