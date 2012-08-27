package main

import (
	"flag"
	"fmt"
	"github.com/mrjones/oauth"
	//	"io"
	"encoding/json"
	"io/ioutil"
	"log"
	//	"os"
	"html/template"
	"net/http"
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

// 取得したいパラメータをstructで記述
// 参考 https://dev.twitter.com/docs/api/1/get/statuses/mentions
type TweetObject struct {
	Created_at              string
	Id_str                  string
	Text                    string
	Source                  string
//	In_reply_to_user_id_str string
	User                    UserObject // JSONオブジェクト内のオブジェクトをこのように定義する。
}

// JSONオブジェク内のオブジェクト
type UserObject struct {
	Id_str      string
	Name        string
	Screen_name string
}

func handler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprint(w, "Hello, world")
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
	//	const url = "http://api.twitter.com/1/statuses/user_timeline.json"
	log.Print("GET ", url)
	resp, err := consumer.Get(url, nil, &atoken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalln(err)
	}
	//defer resp.Body.Close()
	//io.Copy(os.Stdout, resp.Body)

	w.Header().Add("Content-type", "text/html charset=utf-8")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalln(err)
	}

	var tweets []TweetObject
//	fmt.Fprintf(w, "%d%T\n", tweets,tweets)
//	fmt.Fprintf(w, "%d%T\n", body,body)
	
	err2 := json.Unmarshal(body,&tweets)
	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		log.Fatalln(err2)
		return
	}

	t, err := template.ParseFiles("template/main.html", "template/tweet.html", "template/sub.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, tweets)

}
func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)

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

/*
1. go run main.go -id="<client id>" -secret="<secret id>"
2012/08/27 14:09:13 Getting Request Token
Visit this URL: https://api.twitter.com/oauth/authorize?oauth_token=<トークン>
Then run this program again with -code=CODE
where CODE is the verification PIN provided by Twitter.

2.  取得した下記のURLにアクセスして、PINを取得
https://api.twitter.com/oauth/authorize?oauth_token=<トークン>

3. 1に -code="<取得したPIN>" を追加して実行
すると、標準出力にツイートが出力される。
*/
