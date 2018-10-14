package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/marni/goigc"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var startTime = time.Now()
var urlMap = make(map[int]string)
var mapID int
var uniqueId int




type url struct {
	URL string `json:"url"`
}
var IGC_files []Track
/*

URLs for testing:

	http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc
	http://skypolaris.org/wp-content/uploads/IGS%20Files/Jarez%20to%20Senegal.igc
	http://skypolaris.org/wp-content/uploads/IGS%20Files/Boavista%20Medellin.igc
	http://skypolaris.org/wp-content/uploads/IGS%20Files/Medellin%20Guatemala.igc

 */


type Track struct {
	ID string	`json:"ID"`
	IGC_Track igc.Track `json:"igcTrack"`
}
type MetaInfo struct {
	Uptime string		`json:"uptime"`
	Info string			`json:"info"`
	Version string 		`json:"version"`
}

func findIndex(x map[int]string,y int)bool{
	for k, _ := range x {
		if k == y{
			return false
		}
	}
	return true
}


func searchMap(x map[int]string,y string)int{

	for k, v := range x {
		if v==y{
			return k
		}
	}
	return -1
}

func main(){


	router := mux.NewRouter()

	router.HandleFunc("/igcinfo/", IGCinfo)
	router.HandleFunc("/igcinfo/api",GETapi)
	router.HandleFunc("/igcinfo/api/igc",getApiIGC)
	//me kllapa id edhe field se jon variabile
	router.HandleFunc("/igcinfo/api/igc/{id}", getApiIgcID)
	router.HandleFunc("/igcinfo/api/igc/{id}/{field}", getApiIgcIDField)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	
}


func GetAddr() string {
	var port = os.Getenv("PORT")

	if port == "" {
		port = "4747"
		fmt.Println("No port  variable detected, defaulting to " + port)
	}
	return ":" + port
}



func IGCinfo(w http.ResponseWriter, r *http.Request) {

	http.Error(w, "Error 404: Page not found!", http.StatusNotFound)
	return
}

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

//req eshte gjithmone response qe e marrim nga nje klient
func getApiIGC(w http.ResponseWriter, request *http.Request) {

	//request.method per me dit a e ka bo post a get
	switch request.Method {

	case "GET":
		w.Header().Set("content-type", "application/json")

		URLs := mux.Vars(request)
		if len(URLs) != 0 {
			http.Error(w, "400 - Bad Request!", http.StatusBadRequest)
			return
		}

		trackIDs := make([]string, 0, 0)

		for i := range IGC_files {
			trackIDs = append(trackIDs, IGC_files[i].ID)
		}

		json.NewEncoder(w).Encode(trackIDs)




	case "POST":
		//e bojna qe writerti me na qit json
		w.Header().Set("content-type", "application/json")
		// url qe pe shkrujme ne body si json e shkrun klienti ne post
		URLt := &url{}



		//tash ktu me decode e bon tkunderten filen json e kthen ne strukture
		var error = json.NewDecoder(request.Body).Decode(URLt)
		if error != nil {
			http.Error(w,http.StatusText(400),400)
			return
		}
		rand.Seed(time.Now().UnixNano())



		track, err := igc.ParseLocation(URLt.URL)
		if err != nil {

			http.Error(w,"Bad request!\nMalformed URL!",400)
			return
		}


		mapID = searchMap(urlMap,URLt.URL)
		initialID := rand.Intn(100)

		if mapID == -1{
			if findIndex(urlMap,initialID){
				uniqueId = initialID
				urlMap[uniqueId] = URLt.URL
			} else{
				uniqueId = rand.Intn(100)
				urlMap[uniqueId] = URLt.URL
			}

		} else {
			uniqueId = searchMap(urlMap,URLt.URL)
		}


		igcFile := Track{}
		igcFile.ID = strconv.Itoa(uniqueId)
		igcFile.IGC_Track = track


		if findIndex(urlMap,initialID){

			IGC_files = append(IGC_files, igcFile)
		}


		fmt.Fprint(w,"{\n\t\"id\": \""+igcFile.ID+"\"\n}")


		//not implemented methods-->status:501

	default:
		http.Error(w, "This method is not implemented!", 501)
		return

	}


}

func getApiIgcID(w http.ResponseWriter, request *http.Request) {

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

	for i := range IGC_files {
		//id qe e shkrun nalt  a osht najnjo prej idve qe i kena marr na
		if IGC_files[i].ID == URLt["id"] {
			tDate := IGC_files[i].IGC_Track.Date.String()
			tPilot := IGC_files[i].IGC_Track.Pilot
			tGlider := IGC_files[i].IGC_Track.GliderType
			tGliderId := IGC_files[i].IGC_Track.GliderID
			tTrackLength := fmt.Sprintf("%f",trackLength(IGC_files[i].IGC_Track))
			w.Header().Set("content-type","application/json")
			fmt.Fprint(w,"{\n\"H_date\": \""+tDate+"\",\n\"pilot\": \""+tPilot+"\",\n\"GliderType\": \""+tGlider+"\",\n\"Glider_ID\": \""+tGliderId+"\",\n\"track_length\": \""+tTrackLength+"\"\n}")
		}else{
			http.Error(w,"",404)
		}
	}


}

func getApiIgcIDField(w http.ResponseWriter, request *http.Request) {



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


	for i := range IGC_files {
		if IGC_files[i].ID == URLs["id"] {

			mapping := map[string]string {
				"pilot" : IGC_files[i].IGC_Track.Pilot,
				"glider" : IGC_files[i].IGC_Track.GliderType,
				"glider_id" : IGC_files[i].IGC_Track.GliderID,
				"track_length" : fmt.Sprintf("%f",trackLength(IGC_files[i].IGC_Track)),
				"h_date" : IGC_files[i].IGC_Track.Date.String(),
			}

			field := URLs["field"]
			field = strings.ToLower(field)

			if val, ok := mapping[field]; ok {
				fmt.Fprint(w,val)
			} else {


				http.Error(w, "", 404)

				return
			}

		}

	}
}



func trackLength(track igc.Track) float64 {

	totalDistance := 0.0

	for i := 0; i < len(track.Points)-1; i++ {
		totalDistance += track.Points[i].Distance(track.Points[i+1])
	}

	return totalDistance
}
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