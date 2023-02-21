package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {

	bot, err := tg.NewBotAPI("6293373314:AAFVkfHFUowX1FpcRML5frzcAylXEeEMB9I")

	if err != nil {
		log.Println(err)
		return
	} else {
		log.Println("Bot is running")
	}
	updateConfig := tg.NewUpdate(0)
	updateConfig.Timeout = 1
	updates, err := bot.GetUpdatesChan(updateConfig)
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	go r.Run()
	if err != nil {
		log.Println(err)
		return
	} else {
		log.Println("Getting updates")
	}

	for update := range updates {
		if update.Message != nil {

			// typing 10 second
			if update.Message.PinnedMessage != nil {
				continue
			}

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
					msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
					bot.Send(msg)
				case "help":
					msg := tg.NewMessage(update.Message.Chat.ID, "Type /start to get started, send me word and i will give you definition, /help to get help")
					bot.Send(msg)
				default:
					msg := tg.NewMessage(update.Message.Chat.ID, "I don't know that command")
					bot.Send(msg)
				}
			} else {
				t := uzb(update.Message.Text)
				if t != "error" {
					msg := tg.NewMessage(update.Message.Chat.ID, t)
					msg.ReplyToMessageID = update.Message.MessageID
					msg.ParseMode = "markdown"
					res, err := bot.Send(msg)
					if err != nil {
						log.Println(err)
					}
					bot.Send(tg.PinChatMessageConfig{
						ChatID:              res.Chat.ID,
						MessageID:           res.MessageID,
						DisableNotification: false,
					})
				} else {

					msg := tg.NewMessage(update.Message.Chat.ID, "Sorry, i can't tranlate this")
					msg.ParseMode = "markdown"
					bot.Send(msg)
				}

				rest := req(bot, update.Message.Text)
				if rest == nil {
					msg := tg.NewMessage(update.Message.Chat.ID, "Defination: Sorry, i can't find this ")
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

func uzb(text string) string {

	// Generated by curl-to-Go: https://mholt.github.io/curl-to-go

	// curl 'https://translate.yandex.net/api/v1/tr.json/translate?id=fdc8be1d.63f2d4dc.c4357712.74722d74657874-3-0&srv=tr-text&source_lang=en&target_lang=uz&reason=paste&format=text&ajax=1&yu=5678177601665486337&yum=1665688354940623045' \
	//   -H 'authority: translate.yandex.net' \
	//   -H 'accept: */*' \
	//   -H 'accept-language: en-GB,en-US;q=0.9,en;q=0.8,uz;q=0.7' \
	//   -H 'content-type: application/x-www-form-urlencoded' \
	//   -H 'origin: https://translate.yandex.com' \
	//   -H 'referer: https://translate.yandex.com/?source_lang=en&target_lang=uz' \
	//   -H 'sec-ch-ua: "Chromium";v="110", "Not A(Brand";v="24", "Google Chrome";v="110"' \
	//   -H 'sec-ch-ua-mobile: ?0' \
	//   -H 'sec-ch-ua-platform: "Linux"' \
	//   -H 'sec-fetch-dest: empty' \
	//   -H 'sec-fetch-mode: cors' \
	//   -H 'sec-fetch-site: cross-site' \
	//   -H 'user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36' \
	//   -H 'x-retpath-y: https://translate.yandex.com' \
	//   --data-raw 'text=School&options=4' \
	//   --compressed
	params := url.Values{}
	params.Add("text", text)
	params.Add("options", `4`)
	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", "https://translate.yandex.net/api/v1/tr.json/translate?id=fdc8be1d.63f2d4dc.c4357712.74722d74657874-3-0&srv=tr-text&source_lang=en&target_lang=uz&reason=paste&format=text&ajax=1&yu=5678177601665486337&yum=1665688354940623045", body)
	if err != nil {
		log.Println(err)
		return "error"
	}
	req.Header.Set("Authority", "translate.yandex.net")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8,uz;q=0.7")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://translate.yandex.com")
	req.Header.Set("Referer", "https://translate.yandex.com/?source_lang=en&target_lang=uz")
	req.Header.Set("Sec-Ch-Ua", "\"Chromium\";v=\"110\", \"Not A(Brand\";v=\"24\", \"Google Chrome\";v=\"110\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"Linux\"")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36")
	req.Header.Set("X-Retpath-Y", "https://translate.yandex.com")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return "error"
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {

		var data Uzb
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			log.Fatal(err)
		}
		return data.Text[0]

	} else {
		return "error"
	}
}

type Uzb struct {
	Align []string `json:"align"`
	Code  int      `json:"code"`
	Lang  string   `json:"lang"`
	Text  []string `json:"text"`
}
