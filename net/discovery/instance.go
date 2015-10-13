package discovery

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/satori/go.uuid"
)

type ServiceType string

const (
	DYNAMIC   ServiceType = "DYNAMIC"
	STATIC    ServiceType = "STATIC"
	PERMANENT ServiceType = "PERMANENT"
)

type ServiceInstance struct {
	Name                string      `json:"name"`
	Id                  string      `json:"id"`
	Address             string      `json:"address"`
	Port                *int        `json:"port"`
	SslPort             *int        `json:"sslPort"`
	Payload             *string     `json:"payload"`
	RegistrationTimeUTC int64       `json:"registrationTimeUTC"`
	ServiceType         ServiceType `json:"serviceType"`
	UriSpec             *string     `json:"uriSpec"`
}

func NewSimpleServiceInstance(name, address string, port int) *ServiceInstance {
	return NewServiceInstance(name, address, &port, nil, nil)
}

func NewServiceInstance(name, address string, port, ssl *int, payload *string) *ServiceInstance {
	id := uuid.NewV4().String()
	t := time.Now().UnixNano() / int64(time.Millisecond)
	return &ServiceInstance{name, id, address, port, ssl, payload, t, DYNAMIC, nil}
}

func (i *ServiceInstance) Spec() string {
	if i.Port != nil {
		return fmt.Sprintf("%s:%d", i.Address, *i.Port)
	}
	return i.Address
}

type InstanceSerializer interface {
	Serialize(i *ServiceInstance) ([]byte, error)
	Deserialize(b []byte) (*ServiceInstance, error)
}

type JsonInstanceSerializer struct {
}

func (s *JsonInstanceSerializer) Serialize(i *ServiceInstance) ([]byte, error) {
	return json.Marshal(i)
}

func (s *JsonInstanceSerializer) Deserialize(b []byte) (*ServiceInstance, error) {
	i := new(ServiceInstance)
	err := json.Unmarshal(b, i)
	return i, err
}

var _ InstanceSerializer = (*JsonInstanceSerializer)(nil)
