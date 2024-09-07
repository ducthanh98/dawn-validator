package main

import (
	"crypto/tls"
	"dawn-validator/constant"
	"dawn-validator/request"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/mattn/go-colorable"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"math/rand"
	"sync"
	"time"
)

var lock struct {
	sync.Mutex // <-- this mutex protects
}

var logger *zap.Logger

func main() {
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(config),
		zapcore.AddSync(colorable.NewColorableStdout()),
		zapcore.DebugLevel,
	))

	viper.SetConfigFile("./conf.toml")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	proxies := viper.GetStringSlice("proxies.data")

	var accounts []request.Authentication
	err = viper.UnmarshalKey("data.auth", &accounts)
	if err != nil {
		logger.Error("Error unmarshalling config: %v\n", zap.Error(err))
		return
	}

	for i, acc := range accounts {
		go ping(proxies[i%len(proxies)], acc)
	}

	select {}

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
		SetHeader("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")

	//loginRequest := request.LoginRequest{
	//	Username: authInfo.Email,
	//	Password: authInfo.Password,
	//	Logindata: struct {
	//		V        string `json:"_v"`
	//		Datetime string `json:"datetime"`
	//	}(struct {
	//		V        string
	//		Datetime string
	//	}{V: "1.0.7", Datetime: time.Now().Format("2006-01-02 15:04:05")}),
	//}
	//var loginResponse request.LoginResponse
	//res, err := client.R().
	//	SetBody(loginRequest).
	//	SetResult(&loginResponse).
	//	Post(constant.LoginURL)
	//if err != nil {
	//	logger.Error("Login error", zap.String("acc", authInfo.Email), zap.Error(err))
	//	time.Sleep(1 * time.Minute)
	//	go ping(proxyURL, authInfo)
	//	return
	//}
	//lastLogin := time.Now()
	//
	//logger.Info("Login successfully", zap.String("email", authInfo.Email), zap.Any("res", res))
	//if loginResponse.Data.Token == "" {
	//	time.Sleep(5 * time.Second)
	//	go ping(proxyURL, authInfo)
	//	return
	//}

	keepAliveRequest := map[string]interface{}{
		"username":     authInfo.Email,
		"extensionid":  "fpdkjdnhkakefebpekbdhillbhonfjjp",
		"numberoftabs": 0,
		"_v":           "1.0.7",
	}
	for {
		//if time.Now().Sub(lastLogin) > 2*time.Hour {
		//	loginRequest.Logindata.Datetime = time.Now().Format("2006-01-02 15:04:05")
		//	_, err := client.R().
		//		SetBody(loginRequest).
		//		SetResult(&loginResponse).
		//		Post(constant.LoginURL)
		//	if err != nil {
		//		logger.Error("Login error", zap.String("acc", authInfo.Email), zap.Error(err))
		//		time.Sleep(1 * time.Minute)
		//		go ping(proxyURL, authInfo)
		//		return
		//	}
		//}

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
		}
		logger.Info("Get point success", zap.String("acc", authInfo.Email), zap.String("res", res.String()))

		time.Sleep(2 * time.Minute)
	}

}
