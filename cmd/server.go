package cmd

import (
	"fmt"
	"log"

	"github.com/chain5j/chain5j-pkg/cli"
	"github.com/chain5j/logger"
	"github.com/chain5j/logger/zap"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/models"
	"github.com/xwc1125/apisix-go/internal/serve"
	"github.com/xwc1125/apisix-go/params"
)

var (
	configYml string
)

type ServerCmd struct {
	log     logger.Logger
	rootCli *cli.Cli
	cmd     *cobra.Command
}

// NewServerCmd 初始化cmd
func NewServerCmd(rootCli *cli.Cli) *cobra.Command {
	c := &ServerCmd{
		rootCli: rootCli,
	}
	c.cmd = &cobra.Command{
		Use:          "server",
		Short:        "Start server",
		Example:      params.App() + " server -c conf/config.yaml",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			setup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
	c.addFlags()
	return c.cmd
}

// addFlags 添加flags
func (c *ServerCmd) addFlags() {
	c.cmd.PersistentFlags().StringVarP(&configYml, "config", "c", "conf/config.yaml", "Start server with provided configuration file")
	c.cmd.PersistentFlags().Bool("test_local", false, "Whether to use test-route.json test?")

	// 注册路由 fixme 其他应用的路由，在本目录新建文件放在init方法
	viper.BindPFlags(c.cmd.PersistentFlags())

	// 进行config初始化
	cobra.OnInitialize(func() {
		logViper := viper.Sub("log")
		logConfig := new(logger.LogConfig)
		err := logViper.Unmarshal(&logConfig)
		if err != nil {
			panic(err)
		}
		zap.InitWithConfig(logConfig)
	})
}

func setup() {
	// 1. 读取配置
	// 注册监听函数
	usageStr := fmt.Sprintf(`starting %s server...`, params.App())
	log.Println(usageStr)
}

func run() error {
	var serverConfig models.ServerConfig
	err := viper.UnmarshalKey("server", &serverConfig)
	if err != nil {
		logger.Fatal(err)
	}
	proxyServe, err := serve.NewProxyServe()
	if err != nil {
		log.Fatal(err)
	}

	endpoint := fmt.Sprintf("%s:%d", serverConfig.Host, serverConfig.Port)
	logger.Info("proxy server", "endpoint", endpoint)
	if err := fasthttp.ListenAndServe(endpoint, proxyServe.ProxyHandler); err != nil {
		log.Fatal(err)
	}
	return nil
}
