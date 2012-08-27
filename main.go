package main

import (
	"flag"
	"fmt"
	"github.com/mrjones/oauth"
	"io"
	"io/ioutil"
	"encoding/json"
	"log"
	"os"
)

var (
	clientid     = flag.String("id", "", "OAuth Client ID")
	clientsecret = flag.String("secret", "", "OAuth Client Secret")
	rtokenfile   = flag.String("request", "request.json", "Request token file name")
	atokenfile   = flag.String("access", "access.json", "Access token file name")
	code         = flag.String("code", "", "Verification code")
)

var provider = oauth.ServiceProvider{
	RequestTokenUrl:   "http://api.twitter.com/oauth/request_token",
	AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
	AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
}

func main() {
	flag.Parse()
	if *clientid == "" || *clientsecret == "" {
		flag.Usage()
		return
	}
	consumer := oauth.NewConsumer(*clientid, *clientsecret, provider)

	var atoken oauth.AccessToken
	err := readToken(&atoken, *atokenfile)
	if err != nil {
		log.Print("Couldn't read token:", err)

		var rtoken oauth.RequestToken
		err := readToken(&rtoken, *rtokenfile)
		if err != nil {
			log.Print("Couldn't read token:", err)
			log.Print("Getting Request Token")
			rtoken, url, err := consumer.GetRequestTokenAndUrl("oob")
			if err != nil {
				log.Fatal(err)
			}
			err = writeToken(rtoken, *rtokenfile)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Visit this URL:", url)
			fmt.Println("Then run this program again with -code=CODE")
			fmt.Println("where CODE is the verification PIN provided by Twitter.")
			return
		}

		log.Print("Getting Access Token")
		if *code == "" {
			fmt.Println("You must supply a -code parameter to get an Access Token.")
			return
		}
		tok, err := consumer.AuthorizeToken(&rtoken, *code)
		if err != nil {
			log.Fatal(err)
		}
		err = writeToken(tok, *atokenfile)
		if err != nil {
			log.Fatal(err)
		}
		atoken = *tok
	}

	const url = "http://api.twitter.com/1/statuses/mentions.json"
	log.Print("GET ", url)
	resp, err := consumer.Get(url, nil, &atoken)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}

func readToken(token interface{}, filename string) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, token)
}

func writeToken(token interface{}, filename string) error {
	b, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, b, 0666)
}
