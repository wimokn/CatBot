package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	photoprism "github.com/kris-nova/photoprism-client-go"
	"github.com/kris-nova/photoprism-client-go/api/v1"
	"github.com/line/line-bot-sdk-go/v8/linebot"
)

// linebot client ptr
var bot *linebot.Client

// OpenAI Api key
var OpenAIApiKey string
var GPTName string
var PUser string
var PPass string

// CompletionModelParam
var MaxTokens int
var Temperature float32
var TopP float32
var FrequencyPenalty float32
var PresencePenalty float32
var ErrEnvVarEmpty = errors.New("getenv: environment variable empty")

func main() {
	var err error
	PUser = os.Getenv("PHOTOPRISM_USER")
	PPass = os.Getenv("PHOTOPRISM_PASS")
	OpenAIApiKey = os.Getenv("OPENAI_API_KEY")
	GPTName = os.Getenv("GPT_NAME")
	GetModelParamFromEnv()
	bot, err = linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_ACCESS_TOKEN"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	port := "80"
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func GetImageResponse(question string) string {
	client := photoprism.New("https://photoprism.dreamyard.dev")
	err := client.Auth(photoprism.NewClientAuthLogin(PUser, PPass))
	if err != nil {
		return "The Cat Is Missing"
	}

	keywords := "cat & " + question

	options := api.PhotoOptions{
		Count: 10,
		//AlbumUID:
		Filter: keywords,
	}

	result, err := client.V1().GetPhotos(&options)
	if err != nil {
		return "The Cat Is Missing"
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	rnd := r1.Intn(len(result))

	uuid := result[rnd].PhotoUID

	// ---
	// GetPhoto()
	//
	photo, err := client.V1().GetPhoto(uuid)
	if err != nil {
		return "The Cat Is Missing"
	}

	answer := "https://photoprism.dreamyard.dev/api/v1/t/" + photo.Files[0].FileHash + "/2qoiyfaq/fit_4096"
	return answer
}

func GetModelParamFromEnv() {
	var err error
	if MaxTokens, err = getenvInt("MAX_TOKENS"); err != nil {
		log.Println("MAX_TOKENS", err)
		err = nil
	}
	if Temperature, err = getenvFloat("TEMPERATURE"); err != nil {
		log.Println("TEMPERATURE", err)
		err = nil
	}
	if TopP, err = getenvFloat("TOP_P"); err != nil {
		log.Println("TOP_P", err)
		err = nil
	}
	if FrequencyPenalty, err = getenvFloat("PRESENCE_PENALTY"); err != nil {
		log.Println("PRESENCE_PENALTY", err)
		err = nil
	}
	if PresencePenalty, err = getenvFloat("FREQUENCY_PENALTY"); err != nil {
		log.Println("FREQUENCY_PENALTY", err)
		err = nil
	}
}

func getenvStr(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return v, ErrEnvVarEmpty
	}
	log.Println(key, v)
	return v, nil
}

func getenvInt(key string) (int, error) {
	s, err := getenvStr(key)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func getenvFloat(key string) (float32, error) {
	s, err := getenvStr(key)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return 0, err
	}
	return float32(v), nil
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			// Handle only on text message
			case *linebot.TextMessage:

				question := message.Text
				log.Println("Q:", question)
				var answer string
				answer = GetImageResponse(question)
				log.Println("A:", answer)

				switch {
				case strings.HasPrefix(answer, "https://"):
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewImageMessage(answer, answer)).Do(); err != nil {
						log.Print(err)
					}
				default:
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(answer)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	}
}
