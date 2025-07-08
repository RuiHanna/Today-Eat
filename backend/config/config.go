package config

import (
	"encoding/json"
	"os"
)

// 数据库配置结构
type DBConfig struct {
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBHost     string `json:"db_host"`
	DBPort     int    `json:"db_port"`
	DBName     string `json:"db_name"`
}

// 微信小程序配置结构
type WxConfig struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

// 服务器配置
type ServerConfig struct {
	Domain string `json:"domain"`
}

// AI配置
type AIConfig struct {
	APIKey string `json:"api_key"`
}

// LoadDBConfig 读取数据库配置
func LoadDBConfig(path string) (*DBConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg DBConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadWxConfig 读取微信配置
func LoadWxConfig(path string) (*WxConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg WxConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadServerConfig 读取服务器配置
func LoadServerConfig(path string) (*ServerConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg ServerConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadAIConfig 读取AI配置
func LoadAIConfig(path string) (*AIConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg AIConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
