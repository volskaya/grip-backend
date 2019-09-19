package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

const (
	discordAuthURL  = "https://discordapp.com/api/oauth2/authorize"
	discordTokenURL = "https://discordapp.com/api/oauth2/token"
	discordUserURL  = "https://discordapp.com/api/users/@me"
	callbackRoute   = "/auth/callback"
)

type DiscordOauth struct {
	config *oauth2.Config
	state  *State
}

func newOauth(state *State) *DiscordOauth {
	config := &oauth2.Config{
		ClientID:     state.config.DiscordClientID,
		ClientSecret: state.config.DiscordClientSecret,
		RedirectURL:  "http://" + state.config.Address + callbackRoute,
		Scopes:       []string{"identify"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  discordAuthURL,
			TokenURL: discordTokenURL,
		},
	}

	return &DiscordOauth{config, state}
}

// /auth/discord/callback
func (self *DiscordOauth) callbackHandler(w http.ResponseWriter, r *http.Request) {
	var (
		queries  = mux.Vars(r)
		ctx      = r.Context()
		jsonData map[string]interface{}
	)

	if queries["state"] != self.state.config.JwtState {
		fmt.Println("State missmatch prevented in /auth/discord/callback")
		http.Redirect(w, r, "http://127.0.0.1:3000/", http.StatusBadRequest)
		return
	}

	token, err := self.config.Exchange(ctx, queries["code"])

	if err != nil {
		fmt.Println("Oauth2 code exchange failed")
		fmt.Println(err.Error())
		return
	}

	fmt.Println(token.AccessToken)

	// Requests user info
	if res, err := self.config.Client(ctx, token).Get(discordUserURL); err == nil {
		defer res.Body.Close()

		if res.Header.Get("Content-Type") != "application/json" {
			http.Redirect(w, r, "/", http.StatusBadRequest)
			return
		}

		if data, err := ioutil.ReadAll(res.Body); err == nil {
			if err := json.Unmarshal(data, &jsonData); err != nil {
				http.Redirect(w, r, "/", http.StatusBadRequest)
				return
			}

			fmt.Println(safeStr(&jsonData, "username"))

			auth := &DiscordAuth{token, &DiscordUser{
				Username:      safeStr(&jsonData, "username"),
				ID:            safeStr(&jsonData, "id"),
				Avatar:        safeStr(&jsonData, "avatar"),
				Discriminator: safeStr(&jsonData, "discriminator"),
				Email:         safeStrPtr(&jsonData, "email"),
			}}

			jwt, err := self.state.jwt.createToken(auth)

			if err == nil && auth.user.ID != "" {
				// Creates a DB entry for the user
				self.state.db.createUser(auth.user)

				http.SetCookie(w, &http.Cookie{
					Name:     "jwt",
					Value:    *jwt,
					Expires:  time.Now().Add(time.Hour),
					HttpOnly: false,
					Path:     "/",
				})

				http.Redirect(w, r, "http://127.0.0.1:3000/", http.StatusSeeOther)
				return
			} else {
				fmt.Println("Failed to create JWT for user: " + auth.user.Username)
			}
		}
	}

	http.Redirect(w, r, "http://127.0.0.1:3000/", http.StatusBadRequest)
}

// /auth/discord
func (self *DiscordOauth) loginHandler(w http.ResponseWriter, r *http.Request) {
	url := self.config.AuthCodeURL(self.state.config.JwtState)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (self *DiscordOauth) registerRoutes(router *mux.Router) {
	authRoute := router.PathPrefix("/auth").Subrouter()

	authRoute.HandleFunc("/login", self.loginHandler)
	authRoute.HandleFunc("/callback", self.callbackHandler).
		Queries("state", "{state}").
		Queries("code", "{code}")
}

func safeStr(json *map[string]interface{}, key string) string {
	if x, ok := (*json)[key].(string); ok {
		return x
	}

	return ""
}

func safeStrPtr(json *map[string]interface{}, key string) *string {
	if x, ok := (*json)[key].(string); ok {
		return &x
	}

	return nil
}
