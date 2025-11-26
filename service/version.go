package service

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

const APP_VERSION = "v1.1"

type githubRelease struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
	Assets  []struct {
		BrowserDownloadUrl string `json:"browser_download_url"`
	} `json:"assets"`
}

func CheckOutNewVersion() (bool, string) {
	// return true, "https://github.com/vst93/lazyrdm/releases/latest"
	releaseUrl := "https://api.github.com/repos/vst93/lazyrdm/releases/latest"
	resp, err := http.Get(releaseUrl)
	if err != nil {
		return false, "Failed to get latest version from Github"
	}
	defer resp.Body.Close()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil || release.TagName == "" {
		return false, "Failed to parse Github response"
	}
	// 移除版本号前的 'v' 如果存在
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(APP_VERSION, "v")
	// 字符串转小数
	latestVersionFloat, _ := strconv.ParseFloat(latestVersion, 64)
	currentVersionFloat, _ := strconv.ParseFloat(currentVersion, 64)
	// 比较版本号
	if latestVersionFloat > currentVersionFloat {
		return true, "New version available, please download from " + release.HtmlUrl
	}
	return false, "No new version available"
}
