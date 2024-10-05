package main

import (
	"context"
	"crypto/tls"
	"dawn-validator/constant"
	"dawn-validator/request"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mattn/go-colorable"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

type PointResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ReferralPoint struct {
			Commission float64 `json:"commission"`
		} `json:"referralPoint"`
		RewardPoint struct {
			Points           float64 `json:"points"`
			RegisterPoints   float64 `json:"registerpoints"`
			SignInPoints     float64 `json:"signinpoints"`
			TwitterXIDPoints float64 `json:"twitter_x_id_points"`
			DiscordIDPoints  float64 `json:"discordid_points"`
			TelegramIDPoints float64 `json:"telegramid_points"`
			BonusPoints      float64 `json:"bonus_points"`
		} `json:"rewardPoint"`
	} `json:"data"`
}

func calculateTotalPoints(jsonResponse string) (float64, error) {
	var response PointResponse
	err := json.Unmarshal([]byte(jsonResponse), &response)
	if err != nil {
		return 0, fmt.Errorf("error parsing JSON: %v", err)
	}

	if !response.Status {
		return 0, fmt.Errorf("error fetching points: %s", response.Message)
	}

	totalPoints := response.Data.RewardPoint.Points +
		response.Data.RewardPoint.RegisterPoints +
		response.Data.RewardPoint.SignInPoints +
		response.Data.RewardPoint.TwitterXIDPoints +
		response.Data.RewardPoint.DiscordIDPoints +
		response.Data.RewardPoint.TelegramIDPoints +
		response.Data.RewardPoint.BonusPoints +
		response.Data.ReferralPoint.Commission

	return totalPoints, nil
}

func formatPoints(points float64) string {
	roundedPoints := math.Round(points*1000) / 1000
	return fmt.Sprintf("%s points", humanizeFloat(roundedPoints))
}

func humanizeFloat(f float64) string {
	intPart, fracPart := math.Modf(f)
	intStr := fmt.Sprintf("%d", int(intPart))

	for i := len(intStr) - 3; i > 0; i -= 3 {
		intStr = intStr[:i] + "," + intStr[i:]
	}

	if fracPart == 0 {
		return intStr
	}

	return fmt.Sprintf("%s.%03d", intStr, int(fracPart*1000))
}

func ping(proxyURL string, authInfo request.Authentication) {
	rand.Seed(time.Now().UnixNano())
	client := resty.New().SetProxy(proxyURL).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		SetHeader("content-type", "application/json").
		SetHeader("origin", "chrome-extension://fpdkjdnhkakefebpekbdhillbhonfjjp").
		SetHeader("accept", "*/*").
		SetHeader("accept-language", "en-US,en;q=0.9").
		SetHeader("priority", "u=1, i").
		SetHeader("sec-fetch-dest", "empty").
		SetHeader("sec-fetch-mode", "cors").
		SetHeader("sec-fetch-site", "cross-site").
		SetHeader("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36")

	keepAliveRequest := map[string]interface{}{
		"username":     authInfo.Email,
		"extensionid":  "fpdkjdnhkakefebpekbdhillbhonfjjp",
		"numberoftabs": 0,
		"_v":           "1.0.7",
	}
	for {
		res, err := client.R().
			SetHeader("authorization", fmt.Sprintf("Bearer %v", authInfo.Password)).
			SetBody(keepAliveRequest).
			Post(constant.KeepAliveURL)
		if err != nil {
			logger.Error("Keep alive error", zap.String("acc", authInfo.Email), zap.Error(err))
		}
		logger.Info("Keep alive success", zap.String("acc", authInfo.Email), zap.String("res", res.String()))

		res, err = client.R().
			SetHeader("authorization", fmt.Sprintf("Bearer %v", authInfo.Password)).
			Get(constant.GetPointURL)
		if err != nil {
			logger.Error("Get point error", zap.String("acc", authInfo.Email), zap.Error(err))
		} else {
			logger.Info("Get point success", zap.String("acc", authInfo.Email), zap.String("res", res.String()))

			totalPoints, err := calculateTotalPoints(res.String())
			if err != nil {
				logger.Error("Error calculating total points", zap.String("acc", authInfo.Email), zap.Error(err))
			} else {
				formattedPoints := formatPoints(totalPoints)
				logger.Info("Total points calculated", zap.String("acc", authInfo.Email), zap.String("total_points", formattedPoints))
			}
		}

		time.Sleep(1 * time.Minute)
	}
}

// telegram logic
func logUserInteraction(update *models.Update, action string) {
	username := update.Message.From.Username
	if username == "" {
		username = fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName)
	}
	log.Printf("User %s %s", username, action)
}

func handleStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	logUserInteraction(update, "started the bot")
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Welcome, please send the /point command to check your points",
	}); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func handlePoint(ctx context.Context, b *bot.Bot, update *models.Update, accounts []request.Authentication, proxies []string) {
	logUserInteraction(update, "requested point information")
	sendTelegramNotification(ctx, b, update.Message.Chat.ID, accounts, proxies[0])
}

func sendTelegramNotification(ctx context.Context, b *bot.Bot, chatID int64, accounts []request.Authentication, proxyURL string) {
	client := resty.New().SetProxy(proxyURL).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		SetHeader("content-type", "application/json").
		SetHeader("origin", "chrome-extension://fpdkjdnhkakefebpekbdhillbhonfjjp").
		SetHeader("accept", "*/*").
		SetHeader("accept-language", "en-US,en;q=0.9").
		SetHeader("priority", "u=1, i").
		SetHeader("sec-fetch-dest", "empty").
		SetHeader("sec-fetch-mode", "cors").
		SetHeader("sec-fetch-site", "cross-site").
		SetHeader("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36")

	var totalPoints []string
	for _, acc := range accounts {
		res, err := client.R().
			SetHeader("authorization", fmt.Sprintf("Bearer %v", acc.Password)).
			Get(constant.GetPointURL)
		if err != nil {
			logger.Error("Get point error", zap.String("acc", acc.Email), zap.Error(err))
			totalPoints = append(totalPoints, fmt.Sprintf("%s: Error fetching points", acc.Email))
		} else {
			points, err := calculateTotalPoints(res.String())
			if err != nil {
				logger.Error("Error calculating total points", zap.String("acc", acc.Email), zap.Error(err))
				totalPoints = append(totalPoints, fmt.Sprintf("%s: Error calculating points", acc.Email))
			} else {
				formattedPoints := formatPoints(points)
				totalPoints = append(totalPoints, fmt.Sprintf("%s: %s", acc.Email, formattedPoints))
			}
		}
	}

	message := fmt.Sprintf("Total accounts: %d\n\n%s", len(accounts), strings.Join(totalPoints, "\n"))
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   message,
	}); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func main() {
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(config),
		zapcore.AddSync(colorable.NewColorableStdout()),
		zapcore.DebugLevel,
	))

	viper.SetConfigFile("./conf.toml")
	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatal("Fatal error config file", zap.Error(err))
	}

	proxies := viper.GetStringSlice("proxies.data")

	var accounts []request.Authentication
	err = viper.UnmarshalKey("data.auth", &accounts)
	if err != nil {
		logger.Error("Error unmarshalling config", zap.Error(err))
		return
	}

	botToken := viper.GetString("telegram.bot_token")

	b, err := bot.New(botToken)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, handleStart)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/point", bot.MatchTypeExact, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		handlePoint(ctx, b, update, accounts, proxies)
	})

	go func() {
		for i, acc := range accounts {
			go ping(proxies[i%len(proxies)], acc)
		}
	}()

	b.Start(context.Background())
}
