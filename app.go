package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// App is used for
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// Initialize is use for connecting DB and HTTP handlers
func (a *App) Initialize(username, password, dbname string) {
	connectionString := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", username, password, dbname)
	fmt.Println("SQL :: connection Strings --", connectionString)
	var err error
	a.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	var layOuts layOutConfigManagerImpl
	layOutPlan, err := layOuts.get(a.DB)
	if err != nil {
		fmt.Println("err :", err)
	}
	fmt.Println("layOutPlan :", layOutPlan)

	err = layOuts.insert(a.DB, layOutPlan)
	if err != nil {
		fmt.Println("err :", err)
	}

	var seatings seatingConfigManagerImpl
	seatingPlan, err := seatings.get(a.DB)
	if err != nil {
		fmt.Println("err :", err)
	}
	fmt.Println("seatingPlan :", seatingPlan)

	err = seatings.insert(a.DB, seatingPlan)
	if err != nil {
		fmt.Println("err :", err)
	}

	var tots totConfigManagerImpl
	totPlan, err := tots.get(a.DB)
	if err != nil {
		fmt.Println("err :", err)
	}
	fmt.Println("totPlan :", totPlan)

	err = tots.insert(a.DB, totPlan)
	if err != nil {
		fmt.Println("err :", err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()

}

// Run is use to start the application
func (a *App) Run(addr string) {
	fmt.Println("HTTP server is running on port :8080")
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/booking/{date}/{time}", a.getReservation).Methods("Post")
	a.Router.HandleFunc("/checkout/{tableid}", a.processCheckOut).Methods("Post")
}

func (a *App) getReservation(w http.ResponseWriter, r *http.Request) {
	var resRes reservationResult
	vars := mux.Vars(r)

	err := resRes.parseTime(vars)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = resRes.parseDate(vars)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	var resDetail reservationDetail
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&resDetail); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}
	defer r.Body.Close()

	resRes.Name = resDetail.Name
	resRes.Phone = resDetail.Phone
	resRes.NoOfSeats = resDetail.PartySize
	err = resRes.getAvgTOT(a.DB)
	if err != nil {
		fmt.Println("error while getting avgtot")
	}

	err = resRes.checkTableAvailability(a.DB)
	if err != nil {
		respondWithError(w, http.StatusOK, err.Error())
		return
	}

	err = resRes.confirmBooking(a.DB)
	if err != nil {
		respondWithError(w, http.StatusOK, err.Error())
		return
	}

	fmt.Println("reservationDetails :", resDetail)
	fmt.Println("Res Result :", resRes)
	fmt.Println("TableStatus :", tableStatus)

	respondWithJSON(w, http.StatusOK, resRes)

	//Schedule a default checkout using Average_Table_Occupency_Time
	err = resRes.scheduleCheckOut(a.DB)
	if err != nil {
		fmt.Println("Error while schduling tot checkout")
	}
}

func (a *App) processCheckOut(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Temp Not available")
	/*var resRes reservationResult
	var err error
	vars := mux.Vars(r)
	resRes.TableID, err = strconv.Atoi(vars["tableid"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid TableID")
		return
	}

	err = resRes.processCheckOut(a.DB)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"checkout ": "succes"})*/
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error ": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
