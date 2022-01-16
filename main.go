package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

func checkNilErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Data struct {
	Name  string
	Moves []string
}

var DataBase = make(map[string]Data)

func getData() {
	webPage := "https://www.chessgames.com/chessecohelp.html"
	resp, err := http.Get(webPage)
	checkNilErr(err)
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	checkNilErr(err)
	table := doc.Find("tbody")
	table.ChildrenFiltered("tr").Each(func(_ int, row *goquery.Selection) {

		children := row.ChildrenFiltered("td")
		code := children.First().Text()
		second := strings.Split(children.Siblings().First().Text(), "\n")
		name, movesString := second[0], second[1]
		movesArr := strings.Split(movesString, " ")
		moves := []string{}
		for _, move := range movesArr {
			if len(move) <= 1 {
				continue
			}
			move = strings.Replace(move, ",", "", -1)
			moves = append(moves, move)
		}

		DataBase[code] = Data{
			Name:  name,
			Moves: moves,
		}
	})
}

func getAllData(w http.ResponseWriter, r *http.Request) {
	// set 3 minutes cache time
	w.Header().Set("Cache-Control", "max-age=180")
	w.WriteHeader(http.StatusOK)
	response, err := json.Marshal(DataBase)
	checkNilErr(err)

	// Return all data in JSON format
	fmt.Fprintf(w, string(response))
}

func getDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=180")
	w.WriteHeader(http.StatusOK)

	vars := mux.Vars(r)
	code := vars["code"]
	if val, ok := DataBase[code]; ok {
		data, err := json.Marshal(val)
		checkNilErr(err)
		fmt.Fprintf(w, string(data))
	} else {
		fmt.Fprintf(w, "No record found")
	}

}

func main() {
	router := mux.NewRouter()
	getData()
	router.HandleFunc("/", getAllData).Methods("GET")
	router.HandleFunc("/{code}", getDetails).Methods("GET")
	http.Handle("/", router)
	http.ListenAndServe(":", router)
}
