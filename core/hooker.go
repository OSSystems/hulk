package core

type Hooker struct {
	subscriber *Subscriber
}

func NewHooker(subscriber *Subscriber) *Hooker {
	return &Hooker{
		subscriber: subscriber,
	}
}

func (h *Hooker) OnPublish(topic string, payload []byte) error {
	return h.subscriber.ExecuteHook(h.subscriber.unit.Hooks.OnPublish, payload)
}

func (h *Hooker) OnSubscribe(topic string) error {
	return h.subscriber.ExecuteHook(h.subscriber.unit.Hooks.OnSubscribe, []byte(topic))
}

func (h *Hooker) OnSubscribeFail(topic string) error {
	return h.subscriber.ExecuteHook(h.subscriber.unit.Hooks.OnSubscribeFail, []byte(topic))
}
