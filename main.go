package main

import "log"

func main() {
	if err := downloadRawRedditData("reddit_creds.yaml", "raw_reddit_data"); err != nil {
		log.Panic(err)
	}
}
