package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	interval      time.Duration
	telegramToken string
	chatID        int64
	keys          []string
	keysMap       map[string]int
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the server",
	Long:  `Starts the server`,
	Run: func(cmd *cobra.Command, args []string) {
		interval = viper.GetDuration("interval")
		telegramToken = viper.GetString("telegram-token")
		chatID = viper.GetInt64("chat-id")
		keys = viper.GetStringSlice("keys")

		bot, err := tgbotapi.NewBotAPI(telegramToken)
		if err != nil {
			log.Panic(err)
		}

		log.Println("server started")
		_, err = bot.Send(tgbotapi.NewMessage(chatID, "Server started"))
		if err != nil {
			log.Panic(err)
		}

		keysMap = map[string]int{}
		for i, key := range keys {
			keysMap[key] = i
		}

		if err := check(bot); err != nil { //first check
			log.Panic(err)
		}
		for range time.NewTicker(interval).C {
			if err := check(bot); err != nil {
				log.Panic(err)
				break
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().DurationVarP(&interval, "interval", "i", time.Hour, "Interval duration")
	viper.BindPFlag("interval", serveCmd.PersistentFlags().Lookup("interval"))

	serveCmd.PersistentFlags().StringVarP(&telegramToken, "telegram-token", "t", "", "Telegram token")
	viper.BindPFlag("telegram-token", serveCmd.PersistentFlags().Lookup("telegram-token"))

	serveCmd.PersistentFlags().Int64VarP(&chatID, "chat-id", "c", 0, "Telegram chat ID")
	viper.BindPFlag("chat-id", serveCmd.PersistentFlags().Lookup("chat-id"))

	serveCmd.PersistentFlags().StringSliceVarP(&keys, "keys", "k", nil, "Node keys")
	viper.BindPFlag("keys", serveCmd.PersistentFlags().Lookup("keys"))
	serveCmd.MarkFlagRequired("keys")
}

type Node struct {
	Key        string
	Uptime     int
	Downtime   int
	Percentage float64
	Online     bool
}

func check(bot *tgbotapi.BotAPI) error {
	resp, err := http.Get("https://uptime-tracker.skywire.skycoin.com/uptimes")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var nodes []Node
		err = json.Unmarshal(bodyBytes, &nodes)
		if err != nil {
			return err
		}

		found := 0
		for _, node := range nodes {
			if name, ok := keysMap[node.Key]; ok {
				if node.Online {
					found = found + 1
					fmt.Printf("Node %d has %f%% uptime.\n", name+1, node.Percentage)
				} else {
					fmt.Printf("Node %d is offline with %f%% uptime.\n", name+1, node.Percentage)
				}
			}
		}

		if found != len(keys) {
			log.Println("some nodes are down")
			_, err := bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Some nodes are down %d/%d", found, len(keys))))
			if err != nil {
				return err
			}
			return nil
		}
	} else {
		_, err := bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Service is down. Status %v", resp.StatusCode)))
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
