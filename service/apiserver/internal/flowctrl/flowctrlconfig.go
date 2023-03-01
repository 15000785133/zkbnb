package flowctrl

const (
	ControlTypeWhiteList = "ControlledByWhitelist"
	ControlTypeBlackList = "ControlledByBlacklist"
)

type FLowControlConfigItem struct {
	FlowControlType  string
	WhiteListAddress []string
	BlackListAddress []string
}

type FlowControlConfig struct {
	DefaultFlowControlConfig FLowControlConfigItem
	PathFlowControlConfigMap map[string]FLowControlConfigItem
}

func UpdateFlowControlConfig(key string, content string) (*FlowControlConfig, error) {

	flowControlConfig := &FlowControlConfig{}

	if err := flowControlConfig.ValidateFlowControlConfig(); err != nil {
		return nil, err
	}
	return flowControlConfig, nil
}

func (c *FlowControlConfig) ValidateFlowControlConfig() error {

	return nil
}
