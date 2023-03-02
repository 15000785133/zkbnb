package flowctrl

import "github.com/apolloconfig/agollo/v4/storage"

type FlowControlUpdater struct {
}

func (u *FlowControlUpdater) OnChange(event *storage.ChangeEvent) {

}

func (u *FlowControlUpdater) OnNewestChange(event *storage.FullChangeEvent) {

}
