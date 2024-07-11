package consul

import (
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"log"
	"strconv"
	"time"
)

type Service struct {
	consulClient *api.Client
}

type RegisterService struct {
	ServiceName   string
	Address       string
	Port          int
	HeathCheckTTL time.Duration
}

func NewService(address ...string) (*Service, error) {
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
	}, nil
}

func (this *Service) RegisterService(registerParam RegisterService) error {
	healthCheckId := registerParam.ServiceName + "_" + time.Now().String()
	registerInfo := api.AgentServiceRegistration{
		//	Tags:    []string{"aa", "bb"},
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
			err := this.consulClient.Agent().UpdateTTL(
				healthCheckId,
				"online",
				api.HealthPassing,
			)
			if err != nil {
				log.Fatalln(err)
				return
			}
			<-ticker.C
		}
	}()
	return nil
}

func (this *Service) GetServiceAddress(serviceName string) (string, error) {
	service, _, err := this.consulClient.Health().Service(serviceName, "", true, nil)
	if err != nil {
		fmt.Println("get service error: ", err)
		return "", err
	}

	if len(service) > 0 {
		address := service[0].Service.Address + ":" + strconv.Itoa(service[0].Service.Port)
		return address, nil
	}
	return "", errors.New("service not found")
}
