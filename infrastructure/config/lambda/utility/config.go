package utility

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/amplifon-x/ax-go-application-layer/v5/db/redisclient"
	"github.com/joho/godotenv"
)

var ErrNoEnv = errors.New("env variable not found")
var ErrEnvWrongType = errors.New("parsing env failed")

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		slog.Info("`.env` file not found. Using environment")
	}
}

func LoadENV(envName string) string {
	env, found := os.LookupEnv(envName)
	if !found {
		slog.Warn(envName + " not found")
		return ""
	}

	return os.ExpandEnv(env)
}

func LoadAuthDBConfig() DBConfig {
	user, found := os.LookupEnv("AUTH_DB_USER")
	if !found {
		slog.Error("`DB_USER` not found")
		panic(ErrNoEnv)
	}

	password, found := os.LookupEnv("AUTH_DB_PASSWORD")
	if !found {
		slog.Error("`DB_PASSWORD` not found")
		panic(ErrNoEnv)
	}

	host, found := os.LookupEnv("AUTH_DB_HOST")
	if !found {
		slog.Error("`DB_HOST` not found")
		panic(ErrNoEnv)
	}

	host = os.ExpandEnv(host)

	database, found := os.LookupEnv("AUTH_DB_DATABASE")
	if !found {
		slog.Error("`DB_DATABASE` not found")
		panic(ErrNoEnv)
	}

	portString, found := os.LookupEnv("AUTH_DB_PORT")
	if !found {
		slog.Error("`DB_PORT` not found")
		panic(ErrNoEnv)
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		slog.Error("Error parsing `DB_PORT` to int")
		panic(ErrEnvWrongType)
	}

	return DBConfig{
		User:     user,
		Password: password,
		Host:     host,
		Port:     port,
		Database: database,
	}
}

const PG_DSN_TEMPLATE = "postgresql://%s:%s@%s:%d/%s?sslmode=disable"
const CAS_DSN_TEMPLATE = "%s:%s@tcp(%s:%d)/%s?tls=skip-verify&autocommit=true&parseTime=true"

var ErrMissingEnv = errors.New("missing required env")

type DWHConfig struct {
	User     string
	Password string
	Host     string
}

func LoadRegionalOtkDBConfig() string {
	user, found := os.LookupEnv("ROTK_DB_USER")
	if !found {
		slog.Error("`ROTK_DB_USER` not found")
		panic(ErrNoEnv)
	}

	password, found := os.LookupEnv("ROTK_DB_PASSWORD")
	if !found {
		slog.Error("`ROTK_DB_PASSWORD` not found")
		panic(ErrNoEnv)
	}

	host, found := os.LookupEnv("ROTK_DB_HOST")
	if !found {
		slog.Error("`ROTK_DB_HOST` not found")
		panic(ErrNoEnv)
	}

	host = os.ExpandEnv(host)

	database, found := os.LookupEnv("ROTK_DB_DATABASE")
	if !found {
		slog.Error("`ROTK_DB_DATABASE` not found")
		panic(ErrNoEnv)
	}

	portString, found := os.LookupEnv("ROTK_DB_PORT")
	if !found {
		slog.Error("`ROTK_DB_PORT` not found")
		panic(ErrNoEnv)
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		slog.Error("Error parsing `ROTK_DB_PORT` to int")
		panic(ErrEnvWrongType)
	}

	return fmt.Sprintf(
		PG_DSN_TEMPLATE,
		user,
		password,
		host,
		port,
		database,
	)
}

func LoadGlobalOtkDBConfig() string {
	user, found := os.LookupEnv("OTK_DB_USER")
	if !found {
		slog.Error("`OTK_DB_USER` not found")
		panic(ErrNoEnv)
	}

	password, found := os.LookupEnv("OTK_DB_PASSWORD")
	if !found {
		slog.Error("`OTK_DB_PASSWORD` not found")
		panic(ErrNoEnv)
	}

	host, found := os.LookupEnv("OTK_DB_HOST")
	if !found {
		slog.Error("`OTK_DB_HOST` not found")
		panic(ErrNoEnv)
	}

	host = os.ExpandEnv(host)

	database, found := os.LookupEnv("OTK_DB_DATABASE")
	if !found {
		slog.Error("`OTK_DB_DATABASE` not found")
		panic(ErrNoEnv)
	}

	portString, found := os.LookupEnv("OTK_DB_PORT")
	if !found {
		slog.Error("`OTK_DB_PORT` not found")
		panic(ErrNoEnv)
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		slog.Error("Error parsing `OTK_DB_PORT` to int")
		panic(ErrEnvWrongType)
	}

	return fmt.Sprintf(
		PG_DSN_TEMPLATE,
		user,
		password,
		host,
		port,
		database,
	)
}

func LoadCasDBConfig() string {
	user, found := os.LookupEnv("CAS_DB_USER")
	if !found {
		slog.Error("`CAS_DB_USER` not found")
		panic(ErrNoEnv)
	}

	password, found := os.LookupEnv("CAS_DB_PASSWORD")
	if !found {
		slog.Error("`CAS_DB_PASSWORD` not found")
		panic(ErrNoEnv)
	}

	host, found := os.LookupEnv("CAS_DB_HOST")
	if !found {
		slog.Error("`CAS_DB_HOST` not found")
		panic(ErrNoEnv)
	}

	host = os.ExpandEnv(host)

	database, found := os.LookupEnv("CAS_DB_DATABASE")
	if !found {
		slog.Error("`CAS_DB_DATABASE` not found")
		panic(ErrNoEnv)
	}

	portString, found := os.LookupEnv("CAS_DB_PORT")
	if !found {
		slog.Error("`CAS_DB_PORT` not found")
		panic(ErrNoEnv)
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		slog.Error("Error parsing `CAS_DB_PORT` to int")
		panic(ErrEnvWrongType)
	}

	return fmt.Sprintf(
		CAS_DSN_TEMPLATE,
		user,
		password,
		host,
		port,
		database,
	)
}

func LoadA1PDBConfig() string {
	user, found := os.LookupEnv("A1P_DB_USER")
	if !found {
		slog.Error("`A1P_DB_USER` not found")
		panic(ErrNoEnv)
	}

	password, found := os.LookupEnv("A1P_DB_PASSWORD")
	if !found {
		slog.Error("`A1P_DB_PASSWORD` not found")
		panic(ErrNoEnv)
	}

	host, found := os.LookupEnv("A1P_DB_HOST")
	if !found {
		slog.Error("`A1P_DB_HOST` not found")
		panic(ErrNoEnv)
	}

	host = os.ExpandEnv(host)

	database, found := os.LookupEnv("A1P_DB_DATABASE")
	if !found {
		slog.Error("`A1P_DB_DATABASE` not found")
		panic(ErrNoEnv)
	}

	portString, found := os.LookupEnv("A1P_DB_PORT")
	if !found {
		slog.Error("`A1P_DB_PORT` not found")
		panic(ErrNoEnv)
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		slog.Error("Error parsing `A1P_DB_PORT` to int")
		panic(ErrEnvWrongType)
	}

	return fmt.Sprintf(
		PG_DSN_TEMPLATE,
		user,
		password,
		host,
		port,
		database,
	)
}

func LoadJwtConfig() JwtConfig {
	secret, found := os.LookupEnv("JWT_SECRET")
	if !found {
		slog.Error("`JWT_SECRET` not found")
		panic(ErrNoEnv)
	}

	jwtExpHStr, found := os.LookupEnv("JWT_H_EXPIRATION")
	if !found {
		slog.Error("`JWT_H_EXPIRATION` not found")
		panic(ErrNoEnv)
	}

	jwtExpH, err := strconv.Atoi(jwtExpHStr)
	if err != nil {
		slog.Error("Error parsing `JWT_H_EXPIRATION` to int")
		panic(ErrEnvWrongType)
	}

	jwtRExpMStr, found := os.LookupEnv("R_JWT_M_EXPIRATION")
	if !found {
		slog.Error("`R_JWT_M_EXPIRATION` not found")
		panic(ErrNoEnv)
	}

	rJwtExpM, err := strconv.Atoi(jwtRExpMStr)
	if err != nil {
		slog.Error("Error parsing `R_JWT_M_EXPIRATION` to int")
		panic(ErrEnvWrongType)
	}

	return JwtConfig{
		Secret:                       secret,
		TokenExpirationHours:         jwtExpH,
		RefreshTokenExpirationMonths: rJwtExpM,
	}
}

func LoadLogLevel() slog.Level {
	level, _ := os.LookupEnv("LOG_LEVEL")
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func LoadRedisConfig() redisclient.RedisConfig {
	url, found := os.LookupEnv("REDIS_URL")
	if !found {
		slog.Error("`REDIS_URL` not found")
		panic(ErrNoEnv)
	}

	pass, found := os.LookupEnv("REDIS_PASSWORD")
	if !found {
		slog.Error("`REDIS_PASSWORD` not found")
		pass = ""
	}

	return redisclient.RedisConfig{
		Host:     url,
		Password: pass,
	}
}

func LoadDWHConfig() DWHConfig {
	user, found := os.LookupEnv("DWH_USERNAME")
	if !found {
		slog.Error("`DWH_USERNAME` not found")
		panic(ErrNoEnv)
	}

	password, found := os.LookupEnv("DWH_PASSWORD")
	if !found {
		slog.Error("`DWH_PASSWORD` not found")
		panic(ErrNoEnv)
	}

	host, found := os.LookupEnv("DWH_BASE_URL")
	if !found {
		slog.Error("`DWH_BASE_URL` not found")
		panic(ErrNoEnv)
	}

	return DWHConfig{
		User:     user,
		Password: password,
		Host:     host,
	}
}

func LoadEnableIntegration(envName string) bool {
	env, found := os.LookupEnv(envName)
	if !found {
		slog.Warn(envName + " not found")
		return false
	}

	if env == "true" || env == "TRUE" || env == "1" {
		return true
	}

	return false
}
