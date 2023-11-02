package main

import "log"

func main() {
	if err := downloadRawRedditData("raw_reddit_data"); err != nil {
		log.Panic(err)
	}
}
