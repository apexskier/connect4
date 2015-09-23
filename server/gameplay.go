package server

import (
	"net/http"
	// "net/http/cookiejar"
	"encoding/json"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/gorilla/mux"
)

var (
	ActiveGamesById = make(map[int32]Game)
)

type Game struct {
	Id      int32  `json:"id"`
	Player1 string `json:"player1"`
	Player2 string `json:"player2"`
	Board   Board  `json:"board"`
    LastPlay    string  `json:"lastplay"`
}

type PendingGame struct {
	Id      int32  `json:"id"`
	Player1 string `json:"player1"`
	Board   Board  `json:"board"`
}

type Board struct {
	Slots map[string][]string   `json:"slots"`
	Rows  int32                 `json:"rows"`
	Cols  int32                 `json:"cols"`
}

type Move struct {
	Gameid int32 `json:"gameid"`
	Col    int32 `json:"row"`
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
	game := Game{pendinggame.Id, pendinggame.Player1, user.Username, pendinggame.Board, user.Username}

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

type GameMoveResponse struct {
    Game    Game    `json:"game"`
    Status  string  `json:"status"`
    Details string  `json:"details"`
}
func GameMove(w http.ResponseWriter, r *http.Request) {
	var (
		move Move
        status string
        details string
        col     []string
        exists bool
        user Person
        game Game
        gameid int32
        err error
	)

	user = *UserFromRequest(r)

    defer func() {
        if status == "success" {
            // TODO: check for win
            game.LastPlay = user.Username
            ActiveGamesById[game.Id] = game

            // Check if they've won
            colL := len(col)
            // col is the old col, so this is weird.
            if colL >= int(Winl) - 1 {
                // check column
                check := col[colL - int(Winl - 1):]
                if (check[0] == user.Username &&
                    check[1] == user.Username &&
                    check[2] == user.Username) {
                    status = "win"
                }
            }
            if status != "win" {
                // check row
                start := int32(0)
                if st := move.Col - (Winl - 1); st > start {
                    start = st
                }
                end := move.Col
                if en := Cols - 1; en < end {
                    end = en
                }
                row := len(game.Board.Slots[strconv.Itoa(int(move.Col))]) - 1
                for c := start; c <= end; c++ {
                    num := 0
                    for j := 0; j < int(Winl); j++ {
                        col_ := game.Board.Slots[strconv.Itoa(int(c) + j)]
                        if len(col_) > row && col_[row] == user.Username {
                            num++
                        } else {
                            break
                        }
                    }
                    if num == 4 {
                        status = "win"
                        break
                    }
                }
            }
            if status != "win" {
                // check diagonal
                startCol := int32(0)
                if st := move.Col - (Winl - 1); st > startCol {
                    startCol = st
                }
                endCol := move.Col
                if en := Cols - 1; en < endCol {
                    endCol = en
                }
                for c := startCol; c <= endCol; c++ {
                    startRow := (len(game.Board.Slots[strconv.Itoa(int(move.Col))]) - 1) - int(move.Col - startCol)
                    num := 0
                    for j := 0; j < int(Winl); j++ {
                        col_ := game.Board.Slots[strconv.Itoa(int(c) + j)]
                        row := startRow + j
                        if (row < 0) {
                            break
                        }
                        if len(col_) > row && col_[row] == user.Username {
                            num++
                        } else {
                            break
                        }
                    }
                    if num == 4 {
                        status = "win"
                        break
                    }
                }
            }
        }
        response := GameMoveResponse{game, status, details}
        json.NewEncoder(w).Encode(response)

        if status == "win" {
            delete(ActiveGamesById, game.Id)
            delete(PendingGamesById, game.Id)
            delete(UsersGames, UsersByUsername[game.Player1].Id)
            delete(UsersGames, UsersByUsername[game.Player2].Id)
        }
    }()

	gameid, err = getGameId(w, r)
	if err != nil {
        status = "illegal"
        details = "Game doesn't exist"
        return
	}

	game, exists = ActiveGamesById[gameid]
	if !exists {
        status = "illegal"
        details = "Game not started"
        return
	}

    if user.Username != game.Player1 && user.Username != game.Player2 {
        status = "illegal"
        details = "not your game"
        return
    }


	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 2048))
	if err != nil {
        status = "illegal"
        details = "invalid request"
        return
	}
	if err := r.Body.Close(); err != nil {
        status = "illegal"
        details = "invalid request"
        return
	}
	if err := json.Unmarshal(body, &move); err != nil {
        status = "illegal"
        details = "invalid request"
        return
	}

	defer func() {
		if r := recover(); r != nil {
			json.NewEncoder(w).Encode("")
		}
	}()

    if game.LastPlay == user.Username {
        status = "illegal"
        details = "not your turn"
        return
    }

    if move.Col >= Cols || move.Col < 0 {
        status = "illegal"
        details = "not a valid column"
        return
    }

    colIdx := strconv.Itoa(int(move.Col))
    if col, exists = game.Board.Slots[colIdx]; !exists {
        game.Board.Slots[colIdx] = make([]string, 1)
        game.Board.Slots[colIdx][0] = user.Username
        status = "success"
        return
    } else {
        if len(col) >= int(Rows) {
            status = "illegal"
            details = "row is full"
            return
        }
        game.Board.Slots[colIdx] = append(col, user.Username)
        status = "success"
        return
    }
}

func GameDelete(w http.ResponseWriter, r *http.Request) {
	user := *UserFromRequest(r)

	gameid, err := getGameId(w, r)
	if err != nil {
        return
	}

	game, exists := ActiveGamesById[gameid]
	if exists {
        if user.Username != game.Player1 && user.Username != game.Player2 {
            return
        }
        delete(ActiveGamesById, game.Id)
        delete(UsersGames, UsersByUsername[game.Player1].Id)
        delete(UsersGames, UsersByUsername[game.Player2].Id)
        return
	} else {
        game, exists := PendingGamesById[gameid]
        if exists {
            if user.Username != game.Player1 {
                return
            }
            delete(PendingGamesById, game.Id)
            delete(UsersGames, user.Id)
        }
    }
}
