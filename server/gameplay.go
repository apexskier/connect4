package server

import (
	"net/http"
	// "net/http/cookiejar"
	"encoding/json"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/gorilla/mux"
	// "fmt"
)

var (
	ActiveGamesById = make(map[int32]Game)
)

type Game struct {
	Id      int32  `json:"id"`
	Player1 string `json:"player1"`
	Player2 string `json:"player2"`
	Board   Board  `json:"board"`
}

type PendingGame struct {
	Id      int32  `json:"id"`
	Player1 string `json:"player1"`
	Board   Board  `json:"board"`
}

type Board struct {
	Slots map[int32][]string `json:"slots"`
	Rows  int32              `json:"rows"`
	Cols  int32              `json:"cols"`
}

type Move struct {
	Gameid int32 `json:"gameid"`
	Row    int32 `json:"row"`
}

func getGameId(w http.ResponseWriter, r *http.Request) (int32, error) {
	var (
		gameid int
		err    error
	)
	vars := mux.Vars(r)
	if gameid, err = strconv.Atoi(vars["gameid"]); err != nil {
		return 0, err
	}

	return int32(gameid), nil
}

func GameAccept(w http.ResponseWriter, r *http.Request) {
	gameid, err := getGameId(w, r)
	if err != nil {
		http.Error(w, "", 500)
		return
	}

	pendinggame, exists := PendingGamesById[gameid]
	if !exists {
		http.Error(w, "Game does not exist", 500)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			json.NewEncoder(w).Encode("")
		}
	}()
	user := UserFromRequest(r)
	game := Game{pendinggame.Id, pendinggame.Player1, user.Username, pendinggame.Board}

	delete(PendingGamesById, gameid)
	ActiveGamesById[game.Id] = game
	UsersGames[user.Id] = game.Id

	json.NewEncoder(w).Encode(game)
}

func GameStatus(w http.ResponseWriter, r *http.Request) {
	gameid, err := getGameId(w, r)
	if err != nil {
		http.Error(w, "", 500)
		return
	}

	game, exists := ActiveGamesById[gameid]
	if exists {
		json.NewEncoder(w).Encode(game)
		return
	}
	game2, exists := PendingGamesById[gameid]
	if exists {
		json.NewEncoder(w).Encode(game2)
		return
	}
	http.Error(w, "", 404)
	return
}

func GameMove(w http.ResponseWriter, r *http.Request) {
	var (
		move Move
	)

	gameid, err := getGameId(w, r)
	if err != nil {
		http.Error(w, "", 500)
		return
	}

	game, exists := ActiveGamesById[gameid]
	if !exists {
		http.Error(w, "", 404)
		return
	}

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 2048))
	if err != nil {
		http.Error(w, "", 500)
		return
	}
	if err := r.Body.Close(); err != nil {
		http.Error(w, "", 500)
		return
	}
	if err := json.Unmarshal(body, &move); err != nil {
		http.Error(w, "Invalid request body", 400)
		return
	}

	json.NewEncoder(w).Encode(game)
}
