package main

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Address             string
	DiscordClientID     string
	DiscordClientSecret string
	JwtSecret           string
	JwtState            string
	PostgresHost        string
	PostgresUser        string
	PostgresPassword    string
	PostgresName        string
	PostgresSSL         string
}

func loadConfig(path string) *Config {
	config := viper.New()

	config.SetConfigName("config")
	config.AddConfigPath(".")

	if err := config.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}

	return &Config{
		Address:             config.Get("address").(string),
		DiscordClientID:     config.Get("discord.client_id").(string),
		DiscordClientSecret: config.Get("discord.client_secret").(string),
		JwtSecret:           config.Get("jwt.secret").(string),
		JwtState:            config.Get("jwt.state").(string),
		PostgresHost:        config.Get("postgres.host").(string),
		PostgresUser:        config.Get("postgres.user").(string),
		PostgresPassword:    config.Get("postgres.password").(string),
		PostgresName:        config.Get("postgres.dbname").(string),
		PostgresSSL:         config.Get("postgres.sslmode").(string),
	}
}
