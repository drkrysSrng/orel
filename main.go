package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
)

var messages_structured map[string]interface{}

const MaxMsgNum = 100

func Main(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	channelID := r.URL.Query().Get("channelID")
	if channelID == "" {
		log.Println("Channel id is empty")
		return
	}
	log.Println("channelID:", channelID)

	messages_structured = make(map[string]interface{})
	for i := 1; i < MaxMsgNum; i++ {

		url := fmt.Sprintf("https://t.me/s/%s/%d", channelID, i)
		getTGchannel(url, channelID)
	}

	//fmt.Println("messages ", messages_structured)

	jsonResp, err := json.Marshal(messages_structured)
	if err != nil {
		fmt.Printf("Error happened in JSON marshal. Err: %s", err)
	} else {
		w.Write(jsonResp)
	}
	return
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
func getTGchannel(url string, channelID string) {

	res, err := http.Get(url)
	check(err)
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	check(err)

	//messages_structured := make(map[string]interface{})
	doc.Find("div.tgme_widget_message_wrap").Each(func(i int, s *goquery.Selection) {
		message_structured := make(map[string]string)

		converter := md.NewConverter("", true, nil)
		message_structured["markdown"] = converter.Convert(s.Find("div.tgme_widget_message_text"))
		//message_structured["html"], err = s.Find("div.tgme_widget_message_text").Html()
		message_structured["text"] = s.Find("div.tgme_widget_message_text").Text()
		datetime, _ := s.Find(".tgme_widget_message_date time").Attr("datetime")

		// Href to get max message number. Each page has 20 posts and the link associated to this
		// post is in the middle of the page
		//message_structured["href"], _ = s.Find(".tgme_widget_message_date").Attr("href")
		message_structured["owner"] = s.Find(".tgme_widget_message_owner_name").Text()
		message_structured["views"] = s.Find(".tgme_widget_message_views").Text()
		message_structured["href"], _ = s.Find("a.tgme_widget_message_photo_wrap").Attr("href")
		message_structured["style"], _ = s.Find("a.tgme_widget_message_photo_wrap").Attr("style")
		// Unix date to delete UTC
		messages_structured[datetime] = message_structured
	})

}

func main() {
	http.HandleFunc("/", Main)

	log.Println("Server is running...")
	http.ListenAndServe(":8080", nil)
}

//https://api.telegram.org/phone_number=+34610500293&api_id=10296473&api_hash=75379087622dd8fbebeef502932f4ee5
