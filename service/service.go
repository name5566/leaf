package service

import (
	"errors"
	"fmt"
	"sync"
)

type Service interface {
	Name() string
	Start()
	Close()
}

var serviceSlice = []Service{}
var serviceMap = map[string]Service{}
var wg sync.WaitGroup

func Start(service Service) error {
	// check
	if _, ok := serviceMap[service.Name()]; ok {
		return errors.New(fmt.Sprintf("%v service already exists",
			service.Name()))
	}

	// start
	wg.Add(1)
	go func() {
		defer wg.Done()
		service.Start()
	}()

	// append
	serviceSlice = append(serviceSlice, service)
	serviceMap[service.Name()] = service
	return nil
}

func Wait() {
	wg.Wait()
}

func Get(name string) Service {
	return serviceMap[name]
}
