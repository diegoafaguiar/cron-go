/** create a cron job to run every 5 seconds and call an post request to the server and save it to the database (postgresql) and send a notification to the user if the data is not normal */
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Data struct {
	Id          int     `json:"id"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Normal      bool    `json:"normal"`
}

// create a function to schedule the cron job every 5 seconds and another cron job to run every 50 minutes

func main() {
	fmt.Println("Starting cron job...")
	for {
		time.Sleep(time.Second * 5)
		go cronJob()
	}
}

func cronJob() {
	fmt.Println("Running cron job...")
	// get data from the server
	data := getData()
	// save data to the database
	saveData(data)
	// send notification to the user
	sendNotification(data)
}

func getData() Data {
	fmt.Println("Getting data from the server...")
	resp, err := http.Get("http://localhost:3000/data")
	if err != nil {
		fmt.Println("Error getting data from the server")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading data from the server")
	}
	var data Data
	json.Unmarshal(body, &data)
	return data
}

func saveData(data Data) {
	fmt.Println("Saving data to the database...")
	db, err := sql.Open("postgres", "user="+os.Getenv("POSTGRES_USER")+" dbname="+os.Getenv("POSTGRES_DB")+" password="+os.Getenv("POSTGRES_PASSWORD")+" sslmode=disable")
	if err != nil {
		fmt.Println("Error connecting to the database")
	}
	defer db.Close()
	_, err = db.Exec("insert into data (temperature, humidity, normal) values ($1, $2, $3)", data.Temperature, data.Humidity, data.Normal)
	if err != nil {
		fmt.Println("Error inserting data to the database")
	}
}

func sendNotification(data Data) {
	if !data.Normal {
		fmt.Println("Sending notification to the user...")
		url := "https://fcm.googleapis.com/fcm/send"
		var jsonStr = []byte(`{"to": "/topics/news","notification": {"title": "Alerta de temperatura e umidade", "body": "Temperatura: ` + fmt.Sprintf("%f", data.Temperature) + ` - Umidade: ` + fmt.Sprintf("%f", data.Humidity) + `"}}`)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		if err != nil {
			fmt.Println("Error sending notification to the user")
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "key=AAAAZ7V4kqo:APA91bG0s2G3l3g9Qx8k7fP6v4Z0vDj3I3g2RcU1wZf3u0YQbWc7jY5W0JZ9ZqYQZ1YmYm9Kt7A4aZqY9qJQ5z4sUQ1Yq9cL4l7X6y5XfX7yG6p5g8Zl4n7b1v0Z9D1e6L4pYX7xq0Dj")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending notification to the user")
		}
		defer resp.Body.Close()
	}
}
