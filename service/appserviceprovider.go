package service

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

var services map[interface{}]interface{}

func init() {

	if services == nil {
		services = make(map[interface{}]interface{})
	}

	InjectDefaultServices()
}

func InjectDefaultServices() {
	PutService("IBlockService", data.BlockServiceImpl{})
	PutService("Marshalizer", &marshal.JsonMarshalizer{})
}

func GetService(key interface{}) interface{} {
	return services[key]
}

func PutService(key interface{}, value interface{}) {

	if key == nil || value == nil {
		return
	}

	services[key] = value
}

func GetBlockService() data.IBlockService {
	return GetService("IBlockService").(data.IBlockService)
}

func GetMarshalizerService() marshal.Marshalizer {
	return GetService("Marshalizer").(marshal.Marshalizer)
}