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
	"net"
	"net/http"
	"net/http/fcgi"
)

var (
	clientid     = flag.String("id", "jdit7Chpc6RyZVNe3r2akA", "OAuth Client ID")
	clientsecret = flag.String("secret", "W4XuOzeu1dbrWKelmbVzA81q0B9IvOq0tMxM9ZxAvY", "OAuth Client Secret")
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
	Created_at string
	Id_str     string
	Text       string
	Source     string
	//	In_reply_to_user_id_str string
	User UserObject // JSONオブジェクト内のオブジェクトをこのように定義する。
}

// JSONオブジェク内のオブジェクト
type UserObject struct {
	Id_str      string
	Profile_image_url string
	Screen_name string
	Access_token string
	Access_token_secret string
}

var atoken oauth.AccessToken
var rtoken oauth.RequestToken
var me = new(UserObject)
//access_token は毎回取る必要がある。
func login(w http.ResponseWriter, r *http.Request) {

	flag.Parse()
	if *clientid == "" || *clientsecret == "" {
		flag.Usage()
		return
	}
	consumer := oauth.NewConsumer(*clientid, *clientsecret, provider)

	connectDb()
	err := existUser(me)
	if err != nil {
		//DBに格納
		fmt.Println("DBに格納")
		//var rtoken oauth.RequestToken
		rtoken, url, err := consumer.GetRequestTokenAndUrl("http://localhost/app/callback")
		fmt.Println("rtoken:",rtoken)
		if err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, url, http.StatusFound)
		return
	}	

}


func tl(w http.ResponseWriter, r *http.Request){
	consumer := oauth.NewConsumer(*clientid, *clientsecret, provider)


	//	const url = "http://api.twitter.com/1/statuses/mentions.json"
	const url = "http://api.twitter.com/1/statuses/user_timeline.json"
	//const url = "http://api.twitter.com/1/account/verify_credentials.json"
	log.Print("GET ", url)
	resp, err := consumer.Get(url, nil, &atoken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	w.Header().Add("Content-type", "text/html charset=utf-8")
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("body")
	fmt.Println(string(body))
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalln(err)
	}

	var tweets []TweetObject

	err2 := json.Unmarshal(body, &tweets)
	fmt.Println("tweets")
	//fmt.Println(tweets[0].Id_str)
	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		log.Fatalln(err2)
		return
	}

	t, err := template.ParseFiles("template/index.html", "template/tweet.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, tweets)
}

func post(w http.ResponseWriter, r *http.Request) {
	consumer := oauth.NewConsumer(*clientid, *clientsecret, provider)
	content := r.FormValue("content")
	response, err3 := consumer.Post(
		"http://api.twitter.com/1/statuses/update.json",
		"",
		map[string]string{
			"key":    "YgV7Rq8CyfvvfANEbFxZA",
			"status": content,
		},
		&atoken)
	if err3 != nil {
		log.Fatal(err3)
	}
	defer response.Body.Close()
	http.Redirect(w, r, "/app", http.StatusMovedPermanently)
	return
}

func logout(w http.ResponseWriter, r *http.Request) {
	consumer := oauth.NewConsumer(*clientid, *clientsecret, provider)

	const url = "https://api.twitter.com/1/account/end_session.json"
	fmt.Print(w, "logoutしました")

	log.Print("Delete ", url)
	resp, err := consumer.Delete(url, nil, &atoken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalln(err)
	}
	defer resp.Body.Close()

}

func callback(w http.ResponseWriter, r *http.Request) {
	//consumer := oauth.NewConsumer(*clientid, *clientsecret, provider)
	fmt.Println("callback")
	fmt.Println(r)
	*code = r.FormValue("oauth_verifier")

	http.Redirect(w, r, "/app", http.StatusMovedPermanently)
	return
}


func index(w http.ResponseWriter, r *http.Request){
	consumer := oauth.NewConsumer(*clientid, *clientsecret, provider)
	const url = "http://api.twitter.com/1/account/verify_credentials.json"
	log.Print("GET ", url)
	resp, err := consumer.Get(url, nil, &atoken)
	if err != nil {
		
		t, err := template.ParseFiles("template/index.html","template/main.html", "template/tweet.html", "template/sub.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.Execute(w, me)
		return
		
		
	}
	
	tok, err := consumer.AuthorizeToken(&rtoken, *code)
	if err != nil {
		log.Fatal(err)
	}

	atoken = *tok
	defer resp.Body.Close()
	w.Header().Add("Content-type", "text/html charset=utf-8")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalln(err)
	}
	err2 := json.Unmarshal(body, &me)
	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		log.Fatalln(err2)
		return
	}

	insertUser(me, &atoken)

	t, err := template.ParseFiles("template/index.html","template/main.html", "template/tweet.html", "template/sub.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, me)


//	tl(w,r)


}


func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/app/", index)
	mux.HandleFunc("/app/login", login)
	mux.HandleFunc("/app/callback", callback)
	mux.Handle("/app/js/", http.StripPrefix("/app/js/", http.FileServer(http.Dir("js"))))
	mux.Handle("/app/img/", http.StripPrefix("/app/img/", http.FileServer(http.Dir("img"))))
	mux.Handle("/app/css/", http.StripPrefix("/app/css/", http.FileServer(http.Dir("css"))))
	mux.HandleFunc("/app/logout", logout)
	mux.HandleFunc("/app/post", post)
	l, _ := net.Listen("tcp", ":9000")
	if l == nil {
		fmt.Println("listener is nil")
		return
	}
	fcgi.Serve(l, mux)
	fmt.Println("end")
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
