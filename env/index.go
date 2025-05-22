package env

import (
	"flag"
	"os"

	"github.com/joho/godotenv"
)

var (
	Port         string
	RootUser     string
	RootPass     string
	ClientId     string
	ClientSecret string
	MongoUri     string
	MinioUri     string
	RedisUri     string
	MailUri      string
	CdnUri       string
)

func init() {
	filepath := flag.String("config", "env/config.env", "config:")
	flag.Parse()
	godotenv.Load(*filepath)

	Port = os.Getenv("PORT")
	RootUser = os.Getenv("ROOT_USER")
	RootPass = os.Getenv("ROOT_PASS")
	ClientId = os.Getenv("CLIENT_ID")
	ClientSecret = os.Getenv("CLIENT_SECRET")
	MongoUri = os.Getenv("MONGO_URI")
	MinioUri = os.Getenv("MINIO_URI")
	RedisUri = os.Getenv("REDIS_URI")
	MailUri = os.Getenv("MAIL_URI")
	CdnUri = os.Getenv("CDN_URI")
}
