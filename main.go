package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

type DetailData struct {
	ReceivedBy string    `json:"receiveBy"`
	Histories  []History `json:"histories"`
}

type Status struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Data struct {
	Status Status     `json:"status"`
	Data   DetailData `json:"data"`
}

type Formatted struct {
	CreatedAt string `json:"createdAt"`
}

type History struct {
	Description string    `json:"description"`
	CreatedAt   string    `json:"createdAt"`
	Formatted   Formatted `json:"formatted"`
}

func getReceiver(str string) (result string) {
	if strings.Contains(str, "DELIVERED TO") {
		splitStr := strings.Replace(str, "DELIVERED TO", "", 1)
		splitStr = strings.Replace(splitStr, "[", "", 1)
		splitStr = strings.Replace(splitStr, "]", "", 1)
		newSplit := strings.Split(splitStr, "|")
		result = newSplit[0]
	}
	return
}

func formatDate(str string) (date string) {
	input := str
	layout := "02-01-2006 15:04"
	t, _ := time.Parse(layout, input)
	date = string(t.Format("2006-01-02T15:04:05+0700"))
	return
}

func formatDateWIB(str string) (date string) {
	input := str
	layout := "02-01-2006 15:04"
	t, _ := time.Parse(layout, input)
	date = string(t.Format("02 January 2006, 15:04 WIB"))
	return
}

func getHistory(w http.ResponseWriter, r *http.Request) {
	status := Status{
		Code:    "060101",
		Message: "Delivery tracking detail fetched successfully",
	}

	response, err := http.Get("https://gist.githubusercontent.com/nubors/eecf5b8dc838d4e6cc9de9f7b5db236f/raw/d34e1823906d3ab36ccc2e687fcafedf3eacfac9/jne-awb.html")

	if err != nil {
		fmt.Print(err.Error())
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	tracking := false
	histori := make([]History, 0)
	count := 1
	receiver := ""
	doc.Find("table tr").Children().Each(func(i int, sel *goquery.Selection) {
		if tracking {
			if count == 1 {
				date := string(sel.Text())
				desc := string(sel.Next().Text())
				formatdate := Formatted{
					CreatedAt: formatDateWIB(date),
				}

				historix := new(History)
				historix.Description = desc
				historix.CreatedAt = formatDate(date)
				historix.Formatted = formatdate
				histori = append(histori, *historix)
				count++
				receiver = getReceiver(desc)
			} else {
				count = 1
			}
		}
		if sel.Text() == "History " {
			tracking = true
		}
	})

	datas := DetailData{
		ReceivedBy: receiver,
		Histories:  histori,
	}

	data := Data{Status: status, Data: datas}

	bts, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(bts))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(string(bts))
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/tracking-history", getHistory).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", r))
	// getHistory()
}
