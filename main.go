package main

/*
#cgo CFLAGS: -Icore/snow/include
#cgo LDFLAGS: -L${SRCDIR}/lib  -lsnow
#include <snow.h>
*/
import "C"
import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/grpc"
)

type SnowFlake struct {
}

func (snowFlake SnowFlake) todo() float64 {
	return 1.0
}

func main() {
	var snowflakeIdWorker C.struct_SnowflakeIdWorker
	C.snowflakeIdWorkerInit(&snowflakeIdWorker, 1, 1)
	var mutex C.pthread_mutex_t

	C.pthread_mutex_init(&mutex, nil)

	// create a Dapr service server
	s, err := daprd.NewService(":50001")
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}

	// add some topic subscriptions
	sub := &common.Subscription{
		PubsubName: "messages",
		Topic:      "topic1",
	}
	if err := s.AddTopicEventHandler(sub, eventHandler); err != nil {
		log.Fatalf("error adding topic subscription: %v", err)
	}

	// add a service to service invocation handler
	if err := s.AddServiceInvocationHandler("echo", echoHandler); err != nil {
		log.Fatalf("error adding invocation handler: %v", err)
	}
	snowHandler := func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
		if in == nil {
			err = errors.New("nil invocation parameter")
			return
		}

		log.Printf(
			"snow - ContentType:%s, Verb:%s, QueryString:%s, %s",
			in.ContentType, in.Verb, in.QueryString, in.Data,
		)

		var send = make(map[string]interface{})
		send["id"] = int64(C.nextId(&snowflakeIdWorker, &mutex))
		send["timestamp"] = time.Now().Minute()
		send_str, _ := json.Marshal(send)
		out = &common.Content{
			Data:        []byte(send_str),
			ContentType: in.ContentType,
			DataTypeURL: in.DataTypeURL,
		}
		return
	}

	if err := s.AddServiceInvocationHandler("snow", snowHandler); err != nil {
		log.Fatalf("error adding invocation handler: %v", err)
	}
	// add a binding invocation handler
	if err := s.AddBindingInvocationHandler("run", runHandler); err != nil {
		log.Fatalf("error adding binding handler: %v", err)
	}

	// start the server
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func eventHandler(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	log.Printf("event - PubsubName:%s, Topic:%s, ID:%s, Data: %s", e.PubsubName, e.Topic, e.ID, e.Data)
	return false, nil
}

func echoHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	if in == nil {
		err = errors.New("nil invocation parameter")
		return
	}
	log.Printf(
		"echo - ContentType:%s, Verb:%s, QueryString:%s, %s",
		in.ContentType, in.Verb, in.QueryString, in.Data,
	)
	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return
}

func runHandler(ctx context.Context, in *common.BindingEvent) (out []byte, err error) {
	log.Printf("binding - Data:%s, Meta:%v", in.Data, in.Metadata)
	return nil, nil
}
