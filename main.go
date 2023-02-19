package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {

	app := gin.Default()
	app.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	go app.Run()
	bot, err := tg.NewBotAPI("1669588541:AAGRZEulyKI_QVRSf14ada1X2jt3xFA7mbU")

	if err != nil {
		log.Println(err)
		return
	} else {
		log.Println("Bot is running")
	}
	updateConfig := tg.NewUpdate(0)
	updateConfig.Timeout = 1
	updates, err := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Println(err)
		return
	} else {
		log.Println("Getting updates")
	}

	for update := range updates {
		if update.Message != nil {

			// typing 10 second

			bot.Send(tg.ChatActionConfig{
				BaseChat: tg.BaseChat{
					ChatID:           update.Message.Chat.ID,
					ChannelUsername:  "osonenglishbot",
					ReplyToMessageID: update.Message.MessageID,
				},
				Action: "searching",
			})

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					msg := tg.NewMessage(update.Message.Chat.ID, "Welcome to the bot, send me word and i will give you definition")
					bot.Send(msg)
				case "help":
					msg := tg.NewMessage(update.Message.Chat.ID, "Type /start to get started, send me word and i will give you definition, /help to get help")
					bot.Send(msg)
				default:
					msg := tg.NewMessage(update.Message.Chat.ID, "I don't know that command")
					bot.Send(msg)
				}
			} else {

				rest := req(bot, update.Message.Text)
				if rest == nil {
					msg := tg.NewMessage(update.Message.Chat.ID, "Word not found, send me English word, I will give you definition")
					msg.ParseMode = "markdown"
					bot.Send(msg)
				} else {
					var result string
					for _, data := range rest {
						result += fmt.Sprintf("**Word:** %s\n", data.Word)

						if len(data.Phonetics) > 0 {
							p := ""
							p += "\t**Phonetic:**"
							audio := make([]string, 0)
							for _, v := range data.Phonetics {
								if v.Text != "" {
									p += v.Text + ", "
									if v.Audio != "" {
										audio = append(audio, fmt.Sprintf("[link](%s)\n", v.Audio))
									} else {
										p += "\n"
									}
								}
							}
							if len(audio) > 0 {
								p += "\n\t**Audio:**" + audio[len(audio)-1]
							}
							result += p
						}

						result += "\n**Meanings:**\n"
						if len(data.Meanings) > 0 {
							for _, v := range data.Meanings {
								result += fmt.Sprintf("\t\tPart of speech: %s\n", v.PartOfSpeech)

								if len(v.Definitions) > 0 {
									result += "\tDefinitions:\n"
									for _, v2 := range v.Definitions {
										result += fmt.Sprintf("\t\t\t\t- %s\n", v2.Definition)
										if v2.Example != "" {
											result += fmt.Sprintf("\t\t\t\t\t\texm: %s\n", v2.Example)
										}
									}
								}

								if len(v.Synonyms) > 0 {
									s := ""
									for _, v := range v.Synonyms {
										s += v + ", "
									}
									result += fmt.Sprintf("\n\t\t\t\tSynonyms: %s\n", s)
								}

								if len(v.Antonyms) > 0 {
									a := ""
									for _, v := range v.Antonyms {
										a += v.(string) + ", "
									}
									result += fmt.Sprintf("\n\t\t\t\tAntonyms: %s\n", a)

								}

								result += "\n"
							}

						}
						if len(result) > 4096 {
							msg := tg.NewMessage(update.Message.Chat.ID, result[:4096])
							msg.ParseMode = "markdown"
							_, err := bot.Send(msg)
							if err != nil {
								log.Println(err)
							}

							result = result[4096:]
							msg = tg.NewMessage(update.Message.Chat.ID, result)
							msg.ParseMode = "markdown"
							_, err = bot.Send(msg)
							if err != nil {
								log.Println(err)
							}
						} else {
							msg := tg.NewMessage(update.Message.Chat.ID, result)
							msg.ParseMode = "markdown"
							_, err := bot.Send(msg)
							if err != nil {
								log.Println(err)
							}
						}

						result = ""
					}

				}
				forward := tg.NewForward(265943548, update.Message.Chat.ID, update.Message.MessageID)
				bot.Send(forward)

				msg := tg.NewMessage(265943548, "From: @"+update.Message.From.UserName+"\nFullname:"+update.Message.From.FirstName+" "+update.Message.From.LastName+fmt.Sprintf("\nid=%d", update.Message.Chat.ID))

				bot.Send(msg)

			}

		}
	}
}

func req(bot *tg.BotAPI, name string) []*Data {
	url := "https://api.dictionaryapi.dev/api/v2/entries/en/" + name
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode == 200 {
		data := make([]*Data, 0)
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Fatal(err)
		}
		return data
	} else {
		return nil
	}
}

type License struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Data struct {
	Word      string `json:"word"`
	Phonetic  string `json:"phonetic,omitempty"`
	Phonetics []struct {
		Text      string  `json:"text"`
		Audio     string  `json:"audio"`
		SourceURL string  `json:"sourceUrl"`
		License   License `json:"license"`
	} `json:"phonetics"`
	Meanings []struct {
		PartOfSpeech string `json:"partOfSpeech"`
		Definitions  []struct {
			Definition string        `json:"definition"`
			Synonyms   []interface{} `json:"synonyms"`
			Antonyms   []interface{} `json:"antonyms"`
			Example    string        `json:"example"`
		} `json:"definitions"`
		Synonyms []string      `json:"synonyms"`
		Antonyms []interface{} `json:"antonyms"`
	} `json:"meanings"`
	License    License  `json:"license"`
	SourceUrls []string `json:"sourceUrls"`
}
