package main

import (
	"flag"
	"fmt"
	"log"
	"time"
	"os"
	"strings"
	"strconv"
	"net/http"
	"encoding/json"
	"github.com/streadway/amqp"
)

var (
	uri          = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	routingKey   = flag.String("key", "elasticsearch", "AMQP routing key")
	reliable     = flag.Bool("reliable", false, "Wait for the publisher confirmation before exiting")
)

func init() {
	flag.Parse()
}

type Measurement struct {
	UUID               string   `json:"uuid"`
	Measurement_Code   string   `json:"measurement_code"`
	Seq                int64    `json:"seq"`
	Elapsed_Time       int64    `json:"elapsed_time"`
	Server_Time        int64    `json:"server_time"`
	Server             string   `json:"server"`
	Page_Name          string   `json:"page_name"`
	Client_IP          string   `json:"client_ip"`
	User_Agent         string   `json:"user_agent"`
	Browser           *string   `json:"browser"`
	OS                *string   `json:"os"`
	User_Cat1          string   `json:"user_cat1"`
	User_Cat2         *string   `json:"user_cat2"`
}

type ESIndexData struct {
	Index   string  `json:"_index"`
	Type    string  `json:"_type"`
}

type ESIndexCommand struct {
	Index  ESIndexData  `json:"index"`
}

func get_cmd() ESIndexCommand {
	index := fmt.Sprintf("jiffy-%s", time.Now().Format("2006.01.02"))
	entry_type := "measurement"
	cmd := ESIndexCommand{ESIndexData{index, entry_type}}
	return cmd
}

func get_data(req *http.Request) Measurement {
	server, _ := os.Hostname()
	data := Measurement{
		req.FormValue("uid"),                    // uuid
		"",                                      // measurement_code
		0,                                       // seq
		0,                                       // elapsed_time
		time.Now().Unix(),                       // server_time
		server,                                  // server
		req.FormValue("pn"),                     // page_name
		strings.Split(req.RemoteAddr, ":")[0],   // client_addr
		req.UserAgent(), nil, nil,               // user_agent, browser, os
		req.FormValue("sid"), nil,               // user_cat1, user_cat2
	}
	return data
}

func get_elapsed_times(etss string) [][]string {
	var results [][]string
	for _, et := range strings.Split(etss, ",") {
		results = append(results, strings.Split(et, ":"))
	}
	return results
}

func get_connection() *amqp.Connection {
	for err := fmt.Errorf("Not tried to connect"); err != nil ; {
		connection, err := amqp.Dial(*uri)
		if err != nil {
			log.Printf("Error getting connection: %s (Will try to reconnect)", err)
			time.Sleep(time.Second)
		} else {
			return connection
		}
	}
	return nil
}

func main() {
	log.Printf("Opening AMQP connection")
	connection := get_connection()
	//defer connection.Close()

	http.HandleFunc("/add", func(w http.ResponseWriter, req *http.Request) {
		channel, err := connection.Channel()
		if err != nil {log.Fatalf("Error opening channel: %s", err)}

		req.Body.Close()

		cmd := get_cmd()
		data := get_data(req)

		cmd_bytes, err := json.Marshal(cmd)
		if err != nil {log.Fatalf("Error encoding command JSON: %s", err)}

		for _, et := range get_elapsed_times(req.FormValue("ets")) {
			data.Measurement_Code = et[0]
			data.Elapsed_Time, err = strconv.ParseInt(et[1], 10, 0)
			if err != nil {
				log.Printf("Error parsing elapsed time in %s:%s", et[0], et[1])
				continue
			}

			data_bytes, err := json.Marshal(data)
			if err != nil {log.Fatalf("Error encoding data JSON: %s", err)}

			body := fmt.Sprintf("%s\n%s\n", string(cmd_bytes), string(data_bytes))

			err = channel.Publish("", *routingKey, false, false, amqp.Publishing{
				Headers:         amqp.Table{},
				ContentType:     "text/plain",
				ContentEncoding: "",
				Body:            []byte(body),
				DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
				Priority:        0,              // 0-9
				// a bunch of application/implementation-specific fields
			})
			if err != nil {log.Fatalf("Error publishing: %s", err)}
		}

		fmt.Fprintf(w, "ok")
	})

	log.Printf("Launching web server")
	log.Fatal(http.ListenAndServe(":8092", nil))
}
