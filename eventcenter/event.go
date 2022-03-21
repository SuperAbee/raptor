package eventcenter

func NewEvent() *Event {
	return &Event{}
}

type Event struct {
	Header map[string]string
	Body interface{}
}

func (e *Event) WithHeader(key, value string) *Event {
	if e == nil {
		return nil
	}
	if e.Header == nil {
		e.Header = make(map[string]string)
	}
	e.Header[key] = value
	return e
}

func (e *Event) WithBody(body interface{}) *Event {
	if e == nil {
		return nil
	}
	e.Body = body
	return e
}