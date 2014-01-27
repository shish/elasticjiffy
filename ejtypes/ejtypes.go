package ejtypes

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
	Index string `json:"_index"`
	Type string `json:"_type"`
}

type ESIndexCommand struct {
	Index  ESIndexData  `json:"index"`
}
