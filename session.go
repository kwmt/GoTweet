package main
import(
	"net/http"
	"code.google.com/p/gorilla/sessions"
	"log"
	"fmt"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

func MyHandler(w http.ResponseWriter, r *http.Request){
	fmt.Println("myhandler")
	session,err:=store.Get(r,"user")
	if err != nil {
		fmt.Println("err")
	}
	fmt.Println(*session)
//	session.Values["foo"]="bar"
//	session.Values[42]=43
//	fmt.Println(session.Values["foo"])
//	fmt.Println(session.Values[42])
//	session.Save(r,w)

}

func main(){
	http.HandleFunc("/", MyHandler)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}