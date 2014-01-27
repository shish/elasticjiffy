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
	"./ejtypes"
)

var (
	uri          = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	routingKey   = flag.String("key", "elasticsearch", "AMQP routing key")
)

func init() {
	flag.Parse()
}

func main() {
	// since our error handling is "die and restart",
	// try not to restart too quickly
	defer time.Sleep(1 * time.Second)

	log.Printf("Opening AMQP connection")
	connection, err := amqp.Dial(*uri)
	if err != nil {log.Fatalf("Error getting connection: %s", err)}

	server, err := os.Hostname()
	if err != nil {log.Fatalf("Error getting hostname: %s", err)}

	http.HandleFunc("/add", func(w http.ResponseWriter, req *http.Request) {
		channel, err := connection.Channel()
		if err != nil {log.Fatalf("Error opening channel: %s", err)}

		req.Body.Close()

		cmd := ejtypes.ESIndexCommand{ejtypes.ESIndexData{
			fmt.Sprintf("jiffy-%s", time.Now().Format("2006.01.02")),  // ES index to write to
			"measurement",                                             // data _type
		}}
		data := ejtypes.Measurement{
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
			if err != nil {
				log.Printf("Error encoding data JSON: %s", err)
				continue
			}

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

func get_elapsed_times(etss string) [][]string {
	var results [][]string
	for _, et := range strings.Split(etss, ",") {
		results = append(results, strings.Split(et, ":"))
	}
	return results
}

