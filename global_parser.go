package inspeqtor

import (
	"errors"
	"fmt"
	"inspeqtor/conf/global/ast"
	"inspeqtor/conf/global/lexer"
	"inspeqtor/conf/global/parser"
	"inspeqtor/util"
	"io/ioutil"
	"strconv"
)

/*
Parses the global inspeqtor configuration in /etc/inspeqtor/inspeqtor.conf.
*/
type GlobalConfig struct {
	CycleTime uint16
}

var Defaults = GlobalConfig{15}

type AlertRoute struct {
	Name    string
	Channel string
	Config  map[string]string
}

type ConfigFile struct {
	Top         GlobalConfig
	AlertRoutes map[string]AlertRoute
}

func ParseGlobal(rootDir string) (*ConfigFile, error) {
	path := rootDir + "/inspeqtor.conf"
	exists, err := util.FileExists(path)
	if err != nil {
		return nil, err
	}

	if exists {
		util.Debug("Parsing " + path)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		s := lexer.NewLexer([]byte(data))
		p := parser.NewParser()
		obj, err := p.Parse(s)
		if err != nil {
			return nil, err
		}
		ast := obj.(ast.Config)

		var config ConfigFile
		config.Top = Defaults
		if val, has := ast.Variables["cycle_time"]; has {
			time, err := strconv.Atoi(val)
			if err != nil {
				util.Warn("Invalid cycle time: " + val)
				time = 15
			}
			config.Top.CycleTime = uint16(time)
		}
		config.AlertRoutes = map[string]AlertRoute{}
		for _, v := range ast.Routes {
			ar, err := ValidateChannel(v.Name, v.Channel, v.Config)
			if err != nil {
				return nil, err
			}
			if _, ok := config.AlertRoutes[v.Name]; ok {
				return nil, errors.New(fmt.Sprintf("Duplicate alert config for '%s'", v.Name))
			}
			config.AlertRoutes[v.Name] = ar
		}
		return &config, nil
	} else {
		util.Info("No configuration file found at " + rootDir + "/inspector.conf")
		return &ConfigFile{Defaults, nil}, nil
	}
}
