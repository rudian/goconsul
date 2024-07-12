package consul

import (
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type Service struct {
	consulClient *api.Client
	EnvTag       string
	registeredId []string
	mutex        *sync.Mutex
}

type RegisterService struct {
	ServiceName   string
	Address       string
	Port          int
	HeathCheckTTL time.Duration
}

func NewService(envTag string, address ...string) (*Service, error) {
	consulConfig := api.DefaultConfig()
	if len(address) >= 1 && address[0] != "" {
		consulConfig.Address = address[0]
	}

	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		return &Service{}, errors.New("consul client error:" + err.Error())
	}
	return &Service{
		consulClient: consulClient,
		EnvTag:       envTag,
		mutex:        &sync.Mutex{},
	}, nil
}

func (this *Service) DeregisterAllService() {
	this.mutex.Lock()
	for _, id := range this.registeredId {
		err := this.consulClient.Agent().ServiceDeregister(id)
		if err != nil {
			continue
		}
	}
	this.registeredId = nil
	this.mutex.Unlock()
}

func (this *Service) RegisterService(registerParam RegisterService) error {
	healthCheckId := this.EnvTag + "_" + registerParam.ServiceName + "_" + time.Now().String()
	this.mutex.Lock()
	this.registeredId = append(this.registeredId, healthCheckId)
	this.mutex.Unlock()

	registerInfo := api.AgentServiceRegistration{
		Tags:    []string{this.EnvTag},
		ID:      healthCheckId,
		Name:    registerParam.ServiceName,
		Address: registerParam.Address,
		Port:    registerParam.Port,
		Check: &api.AgentServiceCheck{
			DeregisterCriticalServiceAfter: registerParam.HeathCheckTTL.String(),
			TLSSkipVerify:                  true,
			TTL:                            registerParam.HeathCheckTTL.String(),
			CheckID:                        healthCheckId,
		},
	}

	errAgent := this.consulClient.Agent().ServiceRegister(&registerInfo)
	if errAgent != nil {
		return errors.New("consul register service error:" + errAgent.Error())
	}

	//update health check
	go func() {
		ticker := time.NewTicker(registerParam.HeathCheckTTL / 2)
		for {
			if this.registeredId == nil {
				return
			}

			err := this.consulClient.Agent().UpdateTTL(
				healthCheckId,
				"online",
				api.HealthPassing,
			)
			if err != nil {
				return
			}
			<-ticker.C
		}
	}()
	return nil
}

func (this *Service) GetServiceAddress(serviceName string) (string, error) {
	services, _, err := this.consulClient.Health().Service(serviceName, this.EnvTag, true, nil)
	if err != nil {
		fmt.Println("get service error: ", err)
		return "", err
	}

	if len(services) > 0 {
		key := 0
		if len(services) > 1 {
			key = rand.Intn(len(services))
		}
		return services[key].Service.Address + ":" + strconv.Itoa(services[key].Service.Port), nil
	}
	return "", errors.New("service not found")
}
