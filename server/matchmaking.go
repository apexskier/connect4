package server

import (
	"encoding/json"
	// "fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"
)

var (
	Rows int32 = 6
	Cols int32 = 7
	Winl int32 = 4

	PendingGamesById = make(map[int32]PendingGame)

	UsersByUsername = make(map[string]Person)
	UsersById       = make(map[int32]Person)
	UsersGames      = make(map[int32]int32)
)

type UserGetRet struct {
	User Person `json:"user"`
	Game int32  `json:"game"`
}

func UserGet(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			json.NewEncoder(w).Encode("")
		}
	}()
	var (
		gameid int32
	)
	user := UserFromRequest(r)
	if gameid = UsersGames[user.Id]; gameid <= 0 {
		gameid = -1
	}
	ret := UserGetRet{*user, gameid}
	json.NewEncoder(w).Encode(ret)
}

func UserLogout(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			json.NewEncoder(w).Encode("")
		}
	}()
	user := UserFromRequest(r)
	delete(UsersById, user.Id)
	delete(UsersByUsername, user.Username)

	cookie := &http.Cookie{Name: "connect4id", Value: "", Expires: time.Unix(0, 0)}
	http.SetCookie(w, cookie)
	r.AddCookie(cookie)
}

func UsersGet(w http.ResponseWriter, r *http.Request) {
	ret := make([]Person, 0, len(UsersById))
	for _, v := range UsersById {
		ret = append(ret, v)
	}
	json.NewEncoder(w).Encode(ret)
}

func UsersPost(w http.ResponseWriter, r *http.Request) {
	var (
		user     Person
		exists   bool
		username string
	)
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 2048))
	if err != nil {
		http.Error(w, "", 500)
		return
	}
	if err := r.Body.Close(); err != nil {
		http.Error(w, "", 500)
		return
	}
	if err := json.Unmarshal(body, &username); err != nil {
		http.Error(w, "Invalid request body", 400)
		return
	}
	if len := utf8.RuneCountInString(username); len < 2 || len > 40 {
		http.Error(w, "Invalid username", 409)
		return
	}
	if user, exists = UsersByUsername[username]; !exists {
		user = Person{username, rand.Int31()}
	}
	UsersByUsername[user.Username] = user
	UsersById[user.Id] = user

	cookie := &http.Cookie{Name: "connect4id", Value: strconv.Itoa(int(user.Id))}
	http.SetCookie(w, cookie)
	r.AddCookie(cookie)
	json.NewEncoder(w).Encode(user)
}

func GamesGet(w http.ResponseWriter, r *http.Request) {
	ret := make([]PendingGame, 0, len(PendingGamesById))
	for _, v := range PendingGamesById {
		ret = append(ret, v)
	}
	if err := json.NewEncoder(w).Encode(ret); err != nil {
        panic(err)
    }
}

func GamesNew(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			json.NewEncoder(w).Encode("")
		}
	}()
	user := UserFromRequest(r)
	if _, exists := UsersGames[user.Id]; exists {
		http.Error(w, "Already has a game", 300)
		return
	}
	game := PendingGame{rand.Int31(), user.Username, Board{make(map[string][]string), Rows, Cols}}
	PendingGamesById[game.Id] = game
	UsersGames[user.Id] = game.Id

	json.NewEncoder(w).Encode(game)
}

func UserFromRequest(r *http.Request) *Person {
	cookie, err := r.Cookie("connect4id")
	if err != nil {
		panic(err)
	}
	userid, err := strconv.Atoi(cookie.Value)
	if err != nil {
		panic(err)
	}
	user, exists := UsersById[int32(userid)]
	if !exists {
		panic(err)
	}
	return &user
}
