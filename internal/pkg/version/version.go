// Package version maker
//
// @author: xwc1125
package version

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	// VersionMeta_Stable 稳定版本
	VersionMeta_Stable = "stable"
)

var (
	FilePath string = "./"
	// AppName 应用名称
	AppName string
	// Version 版本号
	Version string
	// VersionMeta :"stable","beta","alpha"
	// 如果为stable，那么返回主版本：0.0.1
	// 否则，那么返回子版本：0.0.1.1
	VersionMeta string
	// BuildTime 编译时间
	BuildTime string
	// GitHash 当前的Git Hash码
	GitHash string
	// GitDate 当前的Git 提交时间
	GitDate string
	// BuildNumber 编译次数
	BuildNumber string

	buildHistory        = "BuildHistory.json"
	buildNumberFileName = "BuildNumber"
	buildVersion        = "Version"
	defaultVersion      = "0.0.0"
)

// Build 编译
// preMsg 打印的内容
func Build(preMsg string) bool {
	if len(os.Args) < 2 {
		return false
	}
	cmdStr := os.Args[1]
	switch cmdStr {
	case "version", "v":
		if len(preMsg) > 0 {
			fmt.Println(preMsg)
		}
		fmt.Println("Version: " + VersionWithCommit())
		fmt.Println("BuildTime: " + BuildTime)
		return true
	case "out-version":
		fmt.Println(GetVersion())
		return true
	case "out-version-meta":
		fmt.Println(VersionWithMeta())
		return true
	case "build":
		err := makeBuildNumberFile(false)
		if err != nil {
			log.Println(err)
			return false
		}
		return true
	case "rebuild":
		err := makeBuildNumberFile(true)
		if err != nil {
			log.Println(err)
			return false
		}
		return true
	}
	return false
}

func getFilePath(filename string) string {
	return path.Join(FilePath, filename)
}

// isNotExistMkDir 检查文件夹是否存在
// 如果不存在则新建文件夹
func isNotExistMkDir(src string) error {
	_, err := os.Stat(src)
	notExist := os.IsNotExist(err)
	if notExist {
		err := os.MkdirAll(src, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// makeBuildNumberFile 生成编译文件
func makeBuildNumberFile(rebuild bool) error {
	if err := isNotExistMkDir(FilePath); err != nil {
		return err
	}
	if Version != "" {
		defaultVersion = Version
	}
	version, err := readVersion(buildVersion, defaultVersion)
	if err != nil {
		return err
	}
	buildNumberMap := readBuildNumberMap(buildHistory)
	buildNumber := buildNumberMap[version]
	if rebuild && buildNumber > 0 {
		buildNumber--
	}

	// 先保存编译次数文件，再增加编译次数
	// 所以，json文件保存的是下一次编译的次数
	err = saveBuildNumberFile(buildNumber, buildNumberFileName)
	if err != nil {
		return err
	}
	if !rebuild {
		buildNumberMap[version]++
	}

	return saveBuildNumberMap(buildNumberMap, buildHistory)
}

// GetVersion 获取版本号
func GetVersion() string {
	if Version == "" && BuildNumber == "" {
		return defaultVersion
	}
	if BuildNumber == "" {
		return Version
	}
	if len(VersionMeta) == 0 || strings.EqualFold(VersionMeta_Stable, VersionMeta) {
		return Version
	}
	return Version + "." + BuildNumber
}

// VersionWithMeta 带meta的版本号
var VersionWithMeta = func() string {
	v := GetVersion()
	if VersionMeta != "" {
		v += "-" + VersionMeta
	}
	return v
}

// ArchiveVersion 归档版本号:带commit
var ArchiveVersion = func() string {
	vsn := GetVersion()
	if VersionMeta != "" {
		vsn += "-" + VersionMeta
	}
	if len(GitHash) >= 8 {
		vsn += "-" + GitHash[:8]
	}
	return vsn
}

// VersionWithCommit 带commit的版本
var VersionWithCommit = func() string {
	vsn := VersionWithMeta()
	if len(GitHash) >= 8 {
		vsn += "-" + GitHash[:8]
	}
	if GitDate != "" {
		vsn += "-" + GitDate
	}
	return vsn
}

// 获取主版本号的信息
// 如果没有保存主版本号信息的文件，就自动生成一个
func readVersion(filename, defaultVersion string) (string, error) {
	filePath := getFilePath(filename)
	version, err := ioutil.ReadFile(filePath)
	if err != nil {
		version = []byte(defaultVersion)
		fmt.Printf("%s不存在，或者读取%s文件时出错，设置主版本号为“%s”。\n", filePath, filePath, defaultVersion)
		err := os.WriteFile(filePath, version, 0777)
		if err != nil {
			return "", err
		}
	}
	return string(version), nil
}

// 每个主版本号的编译次数，都保存在`BuildHistory.json`当中。
func readBuildNumberMap(filename string) map[string]int {
	buildNumberMap := map[string]int{}
	filePath := getFilePath(filename)
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("%s不存在，或者读取%s失败，正在将其置零。\n", filePath, filePath)
		return buildNumberMap
	}

	if err := json.Unmarshal(bytes, &buildNumberMap); err != nil {
		fmt.Println("转换Json文件失败，正在将其置零。")
	}

	return buildNumberMap
}

// 在相应的主版本号的编译次数++后，需要再把编译记录保存到json文件
func saveBuildNumberMap(bmap map[string]int, filename string) error {
	filePath := getFilePath(filename)
	bytes, err := json.Marshal(bmap)
	if err != nil {
		fmt.Printf("转换Json失败，不保存%s文件\n", filePath)
		return err
	}

	return os.WriteFile(filePath, bytes, 0777)
}

// 把当前编译次数保存到文件中，以便makefile读取。
func saveBuildNumberFile(number int, filename string) error {
	filePath := getFilePath(filename)
	if err := os.WriteFile(filePath, []byte(strconv.Itoa(number)), 0777); err != nil {
		fmt.Println("无法保存BuildNumber.", filePath)
		return err
	}
	return nil
}
