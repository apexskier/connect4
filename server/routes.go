package server

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{"UserGet", "GET", "/api/user", UserGet},
	Route{"UserLogout", "DELETE", "/api/user", UserLogout},
	Route{"UsersGet", "GET", "/api/users", UsersGet},
	Route{"UsersNew", "POST", "/api/users", UsersPost},

	Route{"GamesGet", "GET", "/api/games", GamesGet},
	Route{"GamesNew", "POST", "/api/games", GamesNew},

	Route{"GameAccept", "POST", "/api/games/{gameid}", GameAccept},
	Route{"GameStatus", "GET", "/api/games/{gameid}", GameStatus},
	Route{"GameMove", "PUT", "/api/games/{gameid}", GameMove},
}
