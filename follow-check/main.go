package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/joho/godotenv"
)

const DB_NAME = "follow_users.sqlite3"

type user struct {
	id         int64
	screenName string
	name       string
	protected  bool
	verified   bool
	createdAt  time.Time
}

type app struct {
	api *anaconda.TwitterApi

	dm *databaseManager
}

func init() {
	if err := godotenv.Load("twitter.env"); err != nil {
		fmt.Errorf("%v", err)
	}
}

func main() {
	accessToken := os.Getenv("ACCESS_TOKEN")
	accessSecret := os.Getenv("ACCESS_SECRET")
	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")

	if accessToken == "" ||
		accessSecret == "" ||
		consumerKey == "" ||
		consumerSecret == "" {
		log.Fatalf("access token or consumer token and so on are empty.")
	}

	a := &app{
		api: anaconda.NewTwitterApiWithCredentials(
			accessToken,
			accessSecret,
			consumerKey,
			consumerSecret,
		),
		dm: &databaseManager{
			name: DB_NAME,
		},
	}

	if err := a.openDB(); err != nil {
		log.Fatalf("can not open database: %s", err)
	}

	if err := a.createDB(); err != nil {
		log.Fatalf("can not create database: %s", err)
	}

	a.closeDB()

	a.startApp()
}

func (a app) startApp() {
	if err := a.openDB(); err != nil {
		log.Printf("can not open database: %s", err)
		return
	}

	now := time.Now()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		pageChan := a.api.GetFriendsListAll(url.Values{})
		for page := range pageChan {
			if page.Error != nil {
				return
			}

			friends := make([]user, len(page.Friends))
			for i, friend := range page.Friends {
				friends[i] = user{
					id:         friend.Id,
					screenName: friend.ScreenName,
					name:       friend.Name,
					protected:  friend.Protected,
					verified:   friend.Verified,
					createdAt:  now,
				}
			}

			a.saveFriends(friends)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		pageChan := a.api.GetFollowersListAll(url.Values{})
		for page := range pageChan {
			if page.Error != nil {
				return
			}

			followers := make([]user, len(page.Followers))
			for i, follower := range page.Followers {
				followers[i] = user{
					id:         follower.Id,
					screenName: follower.ScreenName,
					name:       follower.Name,
					protected:  follower.Protected,
					verified:   follower.Verified,
					createdAt:  now,
				}
			}

			a.saveFollowers(followers)
		}
	}()

	wg.Wait()
}
