package settings

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type ThemeConfig struct {
	PrimaryColor  string `json:"primary_color"`
	SidebarBg     string `json:"sidebar_bg"`
	SidebarText   string `json:"sidebar_text"`
	SidebarActive string `json:"sidebar_active"`
	FontSize      string `json:"font_size"`
}

type SystemConfig struct {
	AppName string      `json:"app_name"`
	Theme   ThemeConfig `json:"theme"`
}

const configFile = "data/system.json"

func loadConfig() *SystemConfig {
	cfg := &SystemConfig{
		AppName: "东芳美诊所管理系统",
		Theme: ThemeConfig{
			PrimaryColor:  "#409EFF",
			SidebarBg:     "#304156",
			SidebarText:   "#bfcbd9",
			SidebarActive: "#409EFF",
			FontSize:      "14px",
		},
	}
	data, err := os.ReadFile(configFile)
	if err == nil {
		json.Unmarshal(data, cfg)
	}
	return cfg
}

func saveConfig(cfg *SystemConfig) error {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(configFile, data, 0644)
}

func GetSystemName() string {
	return loadConfig().AppName
}

func GetSystemConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, loadConfig())
	}
}

func UpdateSystemConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SystemConfig
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if req.AppName == "" {
			req.AppName = "东芳美诊所管理系统"
		}
		if err := saveConfig(&req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "保存成功", "config": req})
	}
}
