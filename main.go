package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/line/line-bot-sdk-go/v8/linebot"
	openai "github.com/sashabaranov/go-openai"
)

// linebot client ptr
var bot *linebot.Client

// OpenAI Api key
var OpenAIApiKey string
var GPTName string

// CompletionModelParam
var MaxTokens int
var Temperature float32
var TopP float32
var FrequencyPenalty float32
var PresencePenalty float32
var ErrEnvVarEmpty = errors.New("getenv: environment variable empty")

func main() {
	var err error
	OpenAIApiKey = os.Getenv("OPENAI_API_KEY")
	GPTName = os.Getenv("GPTName")
	GetModelParamFromEnv()
	bot, err = linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_ACCESS_TOKEN"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	port := "80"
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func GetModelParamFromEnv() {
	var err error
	if MaxTokens, err = getenvInt("max_tokens"); err != nil {
		log.Println("max_tokens", err)
		err = nil
	}
	if Temperature, err = getenvFloat("temperature"); err != nil {
		log.Println("temperature", err)
		err = nil
	}
	if TopP, err = getenvFloat("top_p"); err != nil {
		log.Println("top_p", err)
		err = nil
	}
	if FrequencyPenalty, err = getenvFloat("FrequencyPenalty"); err != nil {
		log.Println("FrequencyPenalty", err)
		err = nil
	}
	if PresencePenalty, err = getenvFloat("PresencePenalty"); err != nil {
		log.Println("PresencePenalty", err)
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

func GetResponse(client *openai.Client, ctx context.Context, question string) string {
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a powerful AI assistant with a rich reserve of knowledge and an answer to all questions.",
			},
			{
				Role:    "user",
				Content: question,
			},
		},
		MaxTokens:        MaxTokens,
		Temperature:      Temperature,
		TopP:             TopP,
		FrequencyPenalty: FrequencyPenalty,
		PresencePenalty:  PresencePenalty,
	})

	if err != nil {
		errString := fmt.Sprintf("Get Open AI Response Error: %s", err)
		log.Println(errString)
		return errString
	}
	answer := resp.Choices[0].Message.Content
	answer = strings.TrimSpace(answer)
	return answer
}

func GetImageResponse(client *openai.Client, ctx context.Context, question string) string {
	resp, err := client.CreateImage(ctx, openai.ImageRequest{
		Prompt:         question,
		Size:           openai.CreateImageSize256x256,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	},
	)

	if err != nil {
		errString := fmt.Sprintf("Image creation error: %s", err)
		log.Println("Image creation error: ", err)
		return errString
	}

	answer := resp.Data[0].URL
	return answer
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
				//var _ID string
				switch {
				case event.Source.GroupID != "":
					//In the group
					//_ID = event.Source.GroupID
					if !strings.HasPrefix(message.Text, GPTName) {
						log.Println("Group", event.Source.GroupID, "message: ", message.Text)
						return
					}

				case event.Source.RoomID != "":
					//In the room
					//_ID = event.Source.RoomID
					if !strings.HasPrefix(message.Text, GPTName) {
						log.Println("Room", event.Source.RoomID, "message: ", message.Text)
						return
					}
				case event.Source.UserID != "":
					//In the personal chat
					//_ID = event.Source.UserID
				}

				// decide the AI object
				var AIObject string
				switch {
				case strings.HasPrefix(message.Text, GPTName):
					AIObject = GPTName
				default:
					AIObject = GPTName
				}
				question = strings.Replace(question, AIObject, "", 1)

				log.Println("Q:", question)

				ctx := context.Background()

				var answer string

				switch AIObject {
				case GPTName:
					client := openai.NewClient(OpenAIApiKey)
					// Handle the special question with image and text
					switch {
					case strings.HasPrefix(question, "Drawing"):
						question = strings.Replace(question, "Drawing", "", 1)
						answer = GetImageResponse(client, ctx, question)
					default:
						answer = GetResponse(client, ctx, question)
					}
				}

				log.Println("A:", answer)

				switch {
				case strings.HasPrefix(answer, "https://") && AIObject == GPTName:
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
