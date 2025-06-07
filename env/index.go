package env

import (
	"flag"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	Port             string
	RootUser         string
	RootPass         string
	ClientId         string
	ClientSecret     string
	MongoUri         string
	MinioUri         string
	RedisUri         string
	MailUri          string
	CdnUri           string
	EmailFromName    string
	EmailFromAddress string
	// RsaPrivateKeyPem string // REVERTED - Removed for RS256
	// RsaPublicKeyPem  string // REVERTED - Removed for RS256
)

func init() {
	// Check if we're running in test mode by looking for test-related arguments
	isTest := false
	for _, arg := range os.Args {
		if strings.Contains(arg, "test") || strings.Contains(arg, ".test") || strings.HasSuffix(arg, "_test") {
			isTest = true
			break
		}
	}

	// Also check if we're in a test binary by looking at the executable name
	if !isTest && len(os.Args) > 0 {
		execName := os.Args[0]
		if strings.Contains(execName, "test") || strings.Contains(execName, ".test") {
			isTest = true
		}
	}

	var configPath string
	if isTest {
		// In test mode, use default config path without parsing flags
		configPath = "env/config.env"
		// Also try relative path from test directory
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			configPath = "../env/config.env"
		}
	} else {
		// In normal mode, parse flags
		filepath := flag.String("config", "env/config.env", "config:")
		flag.Parse()
		configPath = *filepath
	}

	err := godotenv.Load(configPath)
	if err != nil {
		// In test mode, try alternative paths
		if isTest {
			altPaths := []string{"env/config.env", "../env/config.env", "../../env/config.env"}
			for _, path := range altPaths {
				if err := godotenv.Load(path); err == nil {
					break
				}
			}
		}
	}

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
	EmailFromName = os.Getenv("EMAIL_FROM_NAME")
	EmailFromAddress = os.Getenv("EMAIL_FROM_ADDRESS")
	// RsaPrivateKeyPem = os.Getenv("RSA_PRIVATE_KEY_PEM") // REVERTED - Removed for RS256
	// RsaPublicKeyPem = os.Getenv("RSA_PUBLIC_KEY_PEM")   // REVERTED - Removed for RS256
}
