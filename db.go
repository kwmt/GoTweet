package main

import (
	"github.com/ziutek/mymysql/mysql"
	"os"
	//_ "github.com/ziutek/mymysql/native" // Native engine
	_ "github.com/ziutek/mymysql/thrsafe" // Thread safe engine

	"fmt"
	"github.com/mrjones/oauth"
)

func printOK() {
	fmt.Println("OK")
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func checkedResult(rows []mysql.Row, res mysql.Result, err error) ([]mysql.Row, mysql.Result) {
	checkError(err)
	return rows, res
}

var db mysql.Conn

func connectDb() {
	user := "dbuser"
	pass := "C4TuxudR"
	dbname := "tw_kondoiku_go"
	proto := "tcp"
	addr := "127.0.0.1:3306"
	db = mysql.New(proto, "", addr, user, pass, dbname)
	fmt.Printf("dbの型:%T\n",db)
	fmt.Printf("Connect to %s:%s... ", proto, addr)
	checkError(db.Connect())
	printOK()
}

func insertUser(user *UserObject, atoken *oauth.AccessToken) {

	if user != nil {
		fmt.Print("Insert int A... ")
		q:= fmt.Sprintf("insert into users (twtitter_user_id, twitter_screen_name,twitter_profile_image_url, twitter_access_token, twitter_access_toekn_secret ,created, modified) values('%s', '%s', '%s', '%s', '%s', now(), now());",
			user.Id_str,
			user.Screen_name,
			user.Profile_image_url,
			&atoken.Token,
			atoken.Secret)

		fmt.Println(q)
		checkedResult(db.Query(q))
		printOK()
		return 
	}

	return 
}


func existUser(user *UserObject) error {
	if user != nil {
		fmt.Println("user.Screen_name:",user.Screen_name)
		q := fmt.Sprintf("select * from users where twtitter_user_id='%s' limit 1", user.Screen_name)
		fmt.Println(q)
		rows, res := checkedResult(db.Query(q))
		twtitter_user_id := res.Map("twtitter_user_id")
		
		if rows[0].Str(twtitter_user_id) != "" {
			fmt.Println(rows[0].Str(twtitter_user_id))
			return nil //userが存在する
		}
	}
	err := fmt.Errorf("user %q not found", user)
	if err != nil {
		fmt.Println(err)
	}
	return err //userが存在しない
}



