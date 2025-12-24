package config

import (
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type BotConfig struct{
	TelegramToken string
	BaseAdminChatID int64
 	DatabaseURL     string
}

var instance *BotConfig
var once sync.Once

func GetBotConfig() *BotConfig{
	once.Do(func() {
		instance = &BotConfig{}

		if err := godotenv.Load(); err != nil {
			logrus.Fatalf("error loading env variables: %s", err.Error())
		}

		instance.TelegramToken = getEnv("TELEGRAM_BOT_TOKEN", "")
		if instance.TelegramToken == ""{
			logrus.Fatal("could not get bot token")
		}

		instance.BaseAdminChatID = getEnvAsInt("BASE_ADMIN_CHAT_ID", -2)
		if instance.BaseAdminChatID == -2{
			logrus.Fatal("could not get admin chat id")
		}
		
		instance.DatabaseURL = getEnv("DATABASE_URL", "")
		if instance.TelegramToken == ""{
			logrus.Fatal("could not get db url")
		}
	})

	return instance
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

func getEnvAsInt(name string, defaultVal int64) int64{
	valStr := getEnv(name, "")
	if val, err := strconv.Atoi(valStr); err == nil{
		return int64(val)
	}
	
	return defaultVal
}