package proxy

import (
	"bytes"
	"golang.org/x/net/context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ph0m1/porta/config"
	"github.com/ph0m1/porta/logging/gologging"
)

func TestNewDefautFactory(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := gologging.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger: ", err.Error())
		return
	}
	expectedResponse := Response{
		IsComplete: true,
		Data:       map[string]interface{}{"foo": "bar"},
	}
	expectedMethod := "SOME"
	expectedHost := "http://example.com/"
	expectedPath := "/foo"
	expectedURL := expectedHost + strings.TrimLeft(expectedPath, "/")

	URL, err := url.Parse(expectedHost)
	if err != nil {
		t.Error("building the sample url: ", err.Error())
	}

	request := Request{
		Method: expectedMethod,
		Path:   expectedPath,
		URL:    URL,
		Body:   newDummyReadCloser(""), // todo
	}

	assertion := func(ctx context.Context, request *Request) (*Response, error) {
		if request.URL.String() != expectedURL {
			t.Errorf("The middlewares did not update the request URL! want [%s], have [%s]\n", expectedURL, request.URL)
		}
		return &expectedResponse, nil
	}
	factory := NewDefaultFactory(func(_ *config.Backend) Proxy { return assertion }, logger)

	backend := config.Backend{
		URLPattern: expectedPath,
		Method:     expectedMethod,
	}

	endpointSingle := config.EndpointConfig{
		Backend: []*config.Backend{&backend},
	}

	endpointMulti := config.EndpointConfig{
		Backend:         []*config.Backend{&backend, &backend},
		ConcurrentCalls: 3,
	}

	endpointEmpty := config.EndpointConfig{
		Backend:         []*config.Backend{},
		ConcurrentCalls: 3,
	}
	serviceConfig := config.ServiceConfig{
		Version:   1,
		Endpoints: []*config.EndpointConfig{&endpointSingle, &endpointMulti},
		Timeout:   100 * time.Millisecond,
		Host:      []string{expectedHost},
	}
	if err := serviceConfig.Init(); err != nil {
		t.Errorf("Error during the config init: %s\n", err.Error())
	}

	proxyMulti, err := factory.New(&endpointMulti)
	if err != nil {
		t.Errorf("The factory returned an unexpected error: %s\n", err.Error())
	}
	response, err := proxyMulti(context.Background(), &request)
	if err != nil {
		t.Errorf("The proxy middleware propagated an unexpected error: %v\n", response)
	}
	if !response.IsComplete || len(response.Data) != len(expectedResponse.Data) {
		t.Errorf("The proxy middleware propagated an unexpected error: %v\n", response)
	}

	proxySingle, err := factory.New(&endpointSingle)
	if err != nil {
		t.Errorf("The factory returned an unexpected error: %s\n", err.Error())
	}

	response, err = proxySingle(context.Background(), &request)
	if err != nil {
		t.Errorf("The proxy middleware propagated an unexpected error: %v\n", response)
	}
	if !response.IsComplete || len(response.Data) != len(expectedResponse.Data) {
		t.Errorf("The proxy middleware propagated an unexpected error: %v\n", response)
	}

	_, err = factory.New(&endpointEmpty)
	if err != ErrNoBackends {
		t.Errorf("The factory returned an unexpected error: %s\n", err.Error())
	}
}
