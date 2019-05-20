package main

import "C"

import (
	"sync"
	"encoding/csv"
	"time"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"net/http"
	"net/url"
	"strings"
	"io/ioutil"
)

// worldWeather API response
type worldWeather struct {
	Data struct {
		Error []struct{
			Msg string `json:"msg"`
		} `json:"error"`
		Request []struct {
			Type  string `json:"type"`
			Query string `json:"query"`
		} `json:"request"`
		Weather []struct {
			Date     string `json:"date"`
			MaxtempC string `json:"maxtempC"`
			MaxtempF string `json:"maxtempF"`
			MintempC string `json:"mintempC"`
			MintempF string `json:"mintempF"`
		} `json:"weather"`
	} `json:"data"`
}

func callAPI(key, q, date string) (*worldWeather, error) {
	q = url.QueryEscape(q)
	url := fmt.Sprintf("http://api.worldWeatheronline.com/premium/v1/past-weather.ashx?key=%s&q=%s&format=json&date=%s", key, q, date)

	res, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	data := worldWeather{}
	json.Unmarshal(body, &data)

	if len(data.Data.Error) == 0 || (res.StatusCode != 200) {
		return &data, fmt.Errorf("API request failed")
	}

	return &data, nil
}

//export GetCityTemps
func GetCityTemps(ckey, cfilename *C.char) {
	key := C.GoString(ckey)
	filename := C.GoString(cfilename)
	csvFile, err := os.OpenFile(filename, os.O_RDWR, 0755)
	if err != nil {
		log.Fatalln("Didn't open csv file", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	reader := csv.NewReader(csvFile)

	data, err := reader.ReadAll()
	if err != nil {
		log.Fatalln("Failed reading csv file", err)
	}

	// Find first position where value is missing
	i := 0
	for ; i < len(data); i++ {
		missingMinMaxTemp := len(data[i][3]) == 0 && len(data[i][4]) == 0
		if missingMinMaxTemp {
			break
		}
	}

	if i == len(data) {
		fmt.Println("All cities have temps")
		return 
	}

	// call api
	wg := sync.WaitGroup{}
	maxConns := make(chan bool, 100)

	for ; i < len(data); i++ {
		if len(data[i][3]) > 0 && len(data[i][4]) > 0 {
			continue
		}

		maxConns <- true
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			date := convertDate(data[i][2])
			q := data[i][1]
			res, _ := callAPI(key, q, date)

			if len(res.Data.Weather) > 0 {
				// fill in missing values
				data[i][3] = res.Data.Weather[0].MaxtempC
				data[i][4] = res.Data.Weather[0].MaxtempC
			}

			<-maxConns
		}(i)
		time.Sleep(100)
	}

	wg.Wait()

	// Write all to file
	csvFile.Truncate(0)
	csvFile.Seek(0,0)
	writer.WriteAll(data)
	if err := writer.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}

	return // still cities without temps
}

//export convertDate
func convertDate(date string) string {
	// "13/05/2018 15:00" => "2014-08-16"
	d := strings.Split(date, " ")
	d = strings.Split(d[0], "/")

	return fmt.Sprintf("%s-%s-%s", d[2], d[1], d[0])
}

// go build -o weatherapi.so -buildmode=c-shared weatherapi.go
func main() {}
