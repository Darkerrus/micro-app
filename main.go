package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/nats-io/stan.go"
)

var dat map[string]interface{}

var arr = make(map[string][]byte, 10000)

type info struct {
	Id   string
	Data []byte
}

const (
	host     = "localhost"
	port     = 5432
	user     = "root"
	password = "root"
	dbname   = "my_db"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("SELECT * FROM info")
	if err != nil {
		fmt.Println("The querry has an error or elready exist")
	}

	for rows.Next() {
		inf := info{}
		err := rows.Scan(&inf.Id, &inf.Data)
		if err != nil {
			fmt.Println("The querry has an error or elready exist")
		}
		arr[inf.Id] = inf.Data
	}
	defer rows.Close()
	defer db.Close()

	sc, err := stan.Connect("test-cluster", "sb-1", stan.NatsURL("nats://localhost:14222"))
	if err != nil {
		log.Fatalf("can't connect to Nats: %v", err)
	}
	sc.Subscribe("foo", func(msg *stan.Msg) {
		err := json.Unmarshal(msg.Data, &dat)
		if err != nil {
			fmt.Println("The data has error")
		} else {
			id := fmt.Sprintf("%v", dat["order_uid"])
			_, ok := arr[id]
			if ok {
				fmt.Println("Id already exist")
			} else {
				db, err := sql.Open("postgres", psqlInfo)
				if err != nil {
					log.Fatal(err)
				}
				arr[id] = msg.Data
				json_data := string(msg.Data)
				querry := fmt.Sprintf("INSERT INTO info VALUES ('%s', '%s')", id, json_data)
				rows, err := db.Query(querry)
				if err != nil {
					fmt.Println("The querry has an error or elready exist")
				}
				defer rows.Close()
				defer db.Close()
			}

		}

		// for v, e := range dat {

		// 	switch a := e.(type) {
		// 	case string:
		// 		fmt.Println(v, ":", e,)
		// 	case int:
		// 		fmt.Println(v, ":", e,)
		// 	case float64:
		// 		fmt.Println(v, ":", e,)
		// 	case map[string]interface{}:
		// 		fmt.Println(v, ": {")
		// 		for z, x := range a {
		// 			fmt.Println("	", z, ":", x,)
		// 		}
		// 		fmt.Println("},")

		// 	case []interface{}:
		// 		fmt.Println(v, ": [")
		// 		for _, x := range a {

		// 			switch a := x.(type) {
		// 			case map[string]interface{}:
		// 				fmt.Println("{")
		// 				for z, x := range a {
		// 					fmt.Println("	", z, ":", x,)
		// 				}
		// 				fmt.Println("}")
		// 			}
		// 		}
		// 		fmt.Println("],")
		// 	}

		// }

	})
	http.HandleFunc("/", getRoot)
	log.Fatal(http.ListenAndServe(":3333", nil))

}

func getRoot(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	id := make(map[string]string)
	err := json.Unmarshal(body, &id)
	if err != nil {
		fmt.Println(err)
		return
	}
	res := arr[id["id"]]

	io.WriteString(w, string(res))
}
