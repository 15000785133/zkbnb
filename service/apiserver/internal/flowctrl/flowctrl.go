package flowctrl

import (
	"github.com/apolloconfig/agollo/v4"
	apollo "github.com/apolloconfig/agollo/v4/env/config"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
)

func FlowControlHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		// Parse the form before reading the parameter
		request.ParseForm()

		// If not forbidden
		// continue to do the next process
		next(writer, request)
	}
}

func InitFlowControlConfiguration(config config.Config) {
	apolloConfig := &apollo.AppConfig{
		AppID:          config.Apollo.AppID,
		Cluster:        config.Apollo.Cluster,
		IP:             config.Apollo.ApolloIp,
		NamespaceName:  config.Apollo.Namespace,
		IsBackupConfig: config.Apollo.IsBackupConfig,
	}

	client, err := agollo.StartWithConfig(func() (*apollo.AppConfig, error) {
		return apolloConfig, nil
	})
	if err != nil {
		logx.Severef("Fail to initiate apollo client,appId:%s,cluster:%s,namespace:%s",
			apolloConfig.AppID, apolloConfig.Cluster, apolloConfig.NamespaceName)
	}

	configUpdater := &FlowControlUpdater{}
	client.AddChangeListener(configUpdater)

	logx.Info("Initiate Apollo Configuration Successfully")
}
