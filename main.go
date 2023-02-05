package main

import (
	"errors"
	"fmt"
	"log"
	"runtime"

	"github.com/chain5j/chain5j-pkg/cli"
	"github.com/chain5j/chain5j-pkg/color"
	"github.com/chain5j/logger"
	"github.com/chain5j/logger/zap"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/xwc1125/apisix-go/cmd"
	"github.com/xwc1125/apisix-go/internal/pkg/version"
	"github.com/xwc1125/apisix-go/params"
)

func main() {
	if version.Build("App: " + params.App()) {
		return
	}

	cpuNum := runtime.NumCPU() // 获得当前设备的cpu核心数
	fmt.Println("CPU核心数:", cpuNum)
	runtime.GOMAXPROCS(cpuNum) // 设置需要用到的cpu数量

	initCli()
}

// 初始化命令行
func initCli() *cli.Cli {
	rootCli := cli.NewCliWithViper(&cli.AppInfo{
		App:     params.App(),
		Version: params.Version(),
		Welcome: `欢迎使用 ` + color.Green(params.App()) + ` ,可以使用 ` + color.Red(`-h`) + ` 查看命令`,
	}, viper.GetViper())
	rootCli.RootCmd().Args = func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println(`欢迎使用 ` + color.Green(params.App()) + ` 可以使用 ` + color.Red(`-h`) + ` 查看命令`)
			return errors.New(color.Red("requires at least one arg"))
		}
		return nil
	}
	err := rootCli.InitFlags(true, func(rootFlags *pflag.FlagSet) {

	}, func(viper *viper.Viper) {
		// 当config不为空时,才启用
		logViper := viper.Sub("log")
		logConfig := new(logger.LogConfig)
		err := logViper.Unmarshal(&logConfig)
		if err != nil {
			panic(err)
		}
		zap.InitWithConfig(logConfig)
	})
	if err != nil {
		log.Print("initCli", "err", err)
		return nil
	}

	rootCli.AddCommands(cmd.NewServerCmd(rootCli))
	err = rootCli.Execute()
	if err != nil {
		log.Fatal(err)
	}
	return rootCli
}
