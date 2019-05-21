package main

import "C"

import (
	"sync"
	"encoding/csv"
	"time"
	"encoding/json"
	"fmt"
	"strconv"
	"math"
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
		Weather []struct {
			Date      string `json:"date"`
			MaxtempC    string `json:"maxtempC"`
			MintempC    string `json:"mintempC"`
			Hourly      []struct {
				Time           string `json:"time"`
				TempC          string `json:"tempC"`
				WindspeedKmph  string `json:"windspeedKmph"`
				WinddirDegree  string `json:"winddirDegree"`
				Winddir16Point string `json:"winddir16Point"`
				WeatherDesc []struct {
					Value string `json:"value"`
				} `json:"weatherDesc"`
				PrecipMM      string `json:"precipMM"`
				Humidity      string `json:"humidity"`
				Visibility    string `json:"visibility"`
				Pressure      string `json:"pressure"`
				Cloudcover    string `json:"cloudcover"`
				HeatIndexC    string `json:"HeatIndexC"`
				DewPointC     string `json:"DewPointC"`
				WindChillC    string `json:"WindChillC"`
				WindGustKmph  string `json:"WindGustKmph"`
				FeelsLikeC    string `json:"FeelsLikeC"`
				UvIndex       string `json:"uvIndex"`
			} `json:"hourly"`
		} `json:"weather"`
	} `json:"data"`
}
 const(
	Index = iota
	City
	DateTime
	MintempC
	MaxtempC
	TempC
	WindspeedKmph
	WeatherDesc
	WinddirDegree
	Winddir16Point
	PrecipMM
	Humidity
	Visibility
	Pressure
	Cloudcover
	HeatIndexC
	DewPointC
	WindChillC
	WindGustKmph
	FeelsLikeC
	UvIndex
)

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

	// call api
	wg := sync.WaitGroup{}
	maxConns := make(chan bool, 100)

	for i := 0; i < len(data); i++ {
		if len(data[i][MintempC]) > 0 && len(data[i][MaxtempC]) > 0 {
			continue
		} else  {
			fmt.Println("searching:", data[i])
		}

		maxConns <- true
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			date := convertDate(data[i][2])
			q := data[i][City]
			res, _ := callAPI(key, q, date)
			if len(res.Data.Weather) > 0 {
				weather := res.Data.Weather[0]
				// find nearest hourly weather
				index := timeIndex(data[i][2])
				hour := weather.Hourly[index]
				// fill in missing values
				data[i][MintempC] = weather.MaxtempC
				data[i][MaxtempC] = weather.MaxtempC
				data[i][TempC] = hour.TempC
				data[i][WindspeedKmph] = hour.WindspeedKmph
				data[i][WeatherDesc] = hour.WeatherDesc[0].Value
				data[i][WinddirDegree] = hour.WinddirDegree
				data[i][Winddir16Point] = hour.Winddir16Point
				data[i][PrecipMM] = hour.PrecipMM
				data[i][Humidity] = hour.Humidity
				data[i][Visibility] = hour.Visibility
				data[i][Pressure] = hour.Pressure
				data[i][Cloudcover] = hour.Cloudcover
				data[i][HeatIndexC] = hour.HeatIndexC
				data[i][DewPointC] = hour.DewPointC
				data[i][WindChillC] = hour.WindChillC
				data[i][WindGustKmph] = hour.WindGustKmph
				data[i][FeelsLikeC] = hour.FeelsLikeC
				data[i][UvIndex] = hour.UvIndex
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


func timeIndex(time string) int {
	times := []int{ 0, 3, 6, 9, 12, 15, 18, 21 }
	// 13/05/2018 12:45 => 1200
	time = strings.Split(time," ")[1]
	hour := strings.Split(time, ":")
	i, _ := strconv.Atoi(hour[0])

	closest := 999999
	closestIndex := 0

	for index, t := range(times) {
		test := int(math.Abs(float64(i - t)))
		if test < closest {
			closestIndex = index
			closest = test
		}
	}

	return closestIndex
}

// go build -o weatherapi.so -buildmode=c-shared weatherapi.go
func main() {}
