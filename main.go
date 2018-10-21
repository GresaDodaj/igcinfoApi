package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/marni/goigc"
)

//time since the server starts
var startTime = time.Now()
var urlMap = make(map[int]string)
var mapID int
var initialID int
var uniqueID int

type url struct {
	URL string `json:"url"`
}

//IGCfiles saves the igc files tracks
var IGCfiles []Track

//Track is a struct that saves the ID and igcTrack data
type Track struct {
	ID       string    `json:"ID"`
	IGCtrack igc.Track `json:"igcTrack"`
}

//MetaInfo is a struct that saves meta information about the server
type MetaInfo struct {
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
	Version string `json:"version"`
}

// this function returns true if the index is not found and false otherwise
func findIndex(x map[int]string, y int) bool {
	for k := range x {
		if k == y {
			return false
		}
	}
	return true
}

//this function the key of the string if the map contains it, or -1 if the map does not contain the string
func searchMap(x map[int]string, y string) int {

	for k, v := range x {
		if v == y {
			return k
		}
	}
	return -1
}

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/igcinfo/", IGCinfo)
	router.HandleFunc("/igcinfo/api", GETapi)
	router.HandleFunc("/igcinfo/api/igc", getAPIIGC)
	router.HandleFunc("/igcinfo/api/igc/{id}", getAPIIgcID)
	router.HandleFunc("/igcinfo/api/igc/{id}/{field}", getAPIIgcIDField)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), router)
	//	if err := http.ListenAndServe(":8080", router); err != nil {
	if err != nil {
		//log.Fatal(err)
		log.Fatal("ListenAndServe: ", err)
	}

}

//GetAddr is a  function that  gets the port assigned by heroku
func GetAddr() string {
	var port = os.Getenv("PORT")

	if port == "" {
		port = "8080"
		fmt.Println("No port  variable detected, defaulting to " + port)
	}
	return ":" + port
}
//IGCinfo is a function that responds to requests made to the root
func IGCinfo(w http.ResponseWriter, r *http.Request) {

	http.Error(w, "Error 404: Page not found!", http.StatusNotFound)
	return
}
//GETapi returns the meta information of an igc track
func GETapi(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("content-type", "application/json")

	URLs := mux.Vars(request)
	if len(URLs) != 0 {
		http.Error(w, "400 - Bad Request!", http.StatusBadRequest)
		return
	}

	metaInfo := &MetaInfo{}
	metaInfo.Uptime = FormatSince(startTime)
	metaInfo.Info = "Service for IGC tracks"
	metaInfo.Version = "version 1.0"

	json.NewEncoder(w).Encode(metaInfo)
}

//request is what we get from the client
func getAPIIGC(w http.ResponseWriter, request *http.Request) {

	//request.method gives us the method selected by the client, in this api there are two methods
	//that are implemented GET and POST, requests made for other methods will result to an error 501
	//501 is an HTTP  error for not implemented
	switch request.Method {

	case "GET":
		w.Header().Set("content-type", "application/json")

		URLs := mux.Vars(request)
		if len(URLs) != 0 {
			http.Error(w, "400 - Bad Request!", 400)
			return
		}

		trackIDs := make([]string, 0, 0)

		for i := range IGCfiles {
			trackIDs = append(trackIDs, IGCfiles[i].ID)
		}

		json.NewEncoder(w).Encode(trackIDs)

	case "POST":
		// Set response content-type to JSON
		w.Header().Set("content-type", "application/json")

		URLt := &url{}

		//Url is given to the server as JSON and now we decode it to a go structure
		var error = json.NewDecoder(request.Body).Decode(URLt)
		if error != nil {
			http.Error(w, http.StatusText(400), 400)
			return
		}

		//making a random unique ID for the track files
		rand.Seed(time.Now().UnixNano())

		track, err := igc.ParseLocation(URLt.URL)
		if err != nil {

			http.Error(w, "Bad request!\nMalformed URL!", 400)
			return
		}

		mapID = searchMap(urlMap, URLt.URL)
		initialID = rand.Intn(100)

		if mapID == -1 {
			if findIndex(urlMap, initialID) {
				uniqueID = initialID
				urlMap[uniqueID] = URLt.URL

				igcFile := Track{}
				igcFile.ID = strconv.Itoa(uniqueID)
				igcFile.IGCtrack = track
				IGCfiles = append(IGCfiles, igcFile)
				fmt.Fprint(w, "{\n\t\"id\": \""+igcFile.ID+"\"\n}")
				return
			}
				rand.Seed(time.Now().UnixNano())
				uniqueID = rand.Intn(100)
				urlMap[uniqueID] = URLt.URL
				igcFile := Track{}
				igcFile.ID = strconv.Itoa(uniqueID)
				igcFile.IGCtrack = track
				IGCfiles = append(IGCfiles, igcFile)
				fmt.Fprint(w, "{\n\t\"id\": \""+igcFile.ID+"\"\n}")
				return

		}
			uniqueID = searchMap(urlMap, URLt.URL)
			fmt.Fprint(w, "{\n\t\"id\": \""+fmt.Sprintf("%d", uniqueID)+"\"\n}")
			return


	default:
		http.Error(w, "This method is not implemented!", 501)
		return

	}

}

func getAPIIgcID(w http.ResponseWriter, request *http.Request) {

	w.Header().Set("content-type", "application/json")

	URLt := mux.Vars(request)
	if len(URLt) != 1 {
		http.Error(w, "400 - Bad Request!", http.StatusBadRequest)
		return
	}

	if URLt["id"] == "" {
		http.Error(w, "400 - Bad Request!", http.StatusBadRequest)
		return
	}

	for i := range IGCfiles {
		//The requested meta information about a particular track based on the ID given in the url
		//checking if the meta information about it is in memory if so the meta information will be returned
		//otherwise it will return error 404, not found
		if IGCfiles[i].ID == URLt["id"] {
			tDate := IGCfiles[i].IGCtrack.Date.String()
			tPilot := IGCfiles[i].IGCtrack.Pilot
			tGlider := IGCfiles[i].IGCtrack.GliderType
			tGliderID := IGCfiles[i].IGCtrack.GliderID
			tTrackLength := fmt.Sprintf("%f", trackLength(IGCfiles[i].IGCtrack))
			w.Header().Set("content-type", "application/json")
			fmt.Fprint(w, "{\n\"H_date\": \""+tDate+"\",\n\"pilot\": \""+tPilot+"\",\n\"GliderType\": \""+tGlider+"\",\n\"Glider_ID\": \""+tGliderID+"\",\n\"track_length\": \""+tTrackLength+"\"\n}")
		} else {
			http.Error(w, "", 404)
		}
	}

}

func getAPIIgcIDField(w http.ResponseWriter, request *http.Request) {

	URLs := mux.Vars(request)
	if len(URLs) != 2 {
		w.Header().Set("content-type", "application/json")
		http.Error(w, "Error 400 : Bad Request!", http.StatusBadRequest)
		return
	}

	if URLs["id"] == "" {
		w.Header().Set("content-type", "application/json")
		http.Error(w, "Error 400 : Bad Request!\n You did not enter an ID.", http.StatusBadRequest)
		return
	}

	if URLs["field"] == "" {
		w.Header().Set("content-type", "application/json")
		http.Error(w, "Error 400 : Bad Request!\n You did not  enter a field.", http.StatusBadRequest)
		return
	}

	for i := range IGCfiles {
		if IGCfiles[i].ID == URLs["id"] {

			mapping := map[string]string{
				"pilot":        IGCfiles[i].IGCtrack.Pilot,
				"glider":       IGCfiles[i].IGCtrack.GliderType,
				"glider_id":    IGCfiles[i].IGCtrack.GliderID,
				"track_length": fmt.Sprintf("%f", trackLength(IGCfiles[i].IGCtrack)),
				"h_date":       IGCfiles[i].IGCtrack.Date.String(),
			}

			field := URLs["field"]
			field = strings.ToLower(field)

			if val, ok := mapping[field]; ok {
				fmt.Fprint(w, val)
			} else {

				http.Error(w, "", 404)

				return
			}

		}

	}
}

//function calculating the total  distance of the flight, from the start point until end point(geographical coordinates)
func trackLength(track igc.Track) float64 {

	totalDistance := 0.0

	for i := 0; i < len(track.Points)-1; i++ {
		totalDistance += track.Points[i].Distance(track.Points[i+1])
	}

	return totalDistance
}

// FormatSince is a function that returns the current uptime of the service, format as specified by ISO 8601.
func FormatSince(t time.Time) string {
	const (
		Decisecond = 100 * time.Millisecond
		Day        = 24 * time.Hour
	)
	ts := time.Since(t)
	sign := time.Duration(1)
	if ts < 0 {
		sign = -1
		ts = -ts
	}
	ts += +Decisecond / 2
	d := sign * (ts / Day)
	ts = ts % Day
	h := ts / time.Hour
	ts = ts % time.Hour
	m := ts / time.Minute
	ts = ts % time.Minute
	s := ts / time.Second
	ts = ts % time.Second
	f := ts / Decisecond
	y := d / 365
	return fmt.Sprintf("P%dY%dD%dH%dM%d.%dS", y, d, h, m, s, f)
}
