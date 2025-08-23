package service

import (
	"encoding/json"
	"errors"
	"node/global"
	"node/model"
	"os"
)

func (n *NService) InitChainList() (err error) {
	if err = UpdateChainListFromFile(); err != nil {
		global.NODE_LOG.Info(err.Error())
		return
	}

	return nil
}

func UpdateChainListFromFile() (err error) {
	file, err := os.Open("json/ChainList.json")
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		err = errors.New("can not open chainlist file")
		return
	}
	defer file.Close()

	if err = json.NewDecoder(file).Decode(&model.ChainList); err != nil {
		global.NODE_LOG.Error(err.Error())
		err = errors.New("can not from file to json")
		return
	}

	return nil
}
