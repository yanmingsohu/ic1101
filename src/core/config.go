package core

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"

	"gopkg.in/yaml.v2"
)

var letterRunes = []rune("abcdef0123456789_-+=$ghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")


type Config struct {
	HttpPort 		int 		`yaml:"httpPort"`
	MongoURL 		string 	`yaml:"mongoURL"`
	MongoDBName string 	`yaml:"mongoDBname"`
	Salt        string  `yaml:"salt"`
}


func RandStringRunes(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[ rand.Intn(len(letterRunes)) ]
    }
    return string(b) 
}


func InitRootUser(u *LoginUser) {
	filename := "root-user.yaml"
	file, _ := os.Open(filename)
	u.IsRoot = true

	if file != nil {
		defer file.Close()
		b, _ := ioutil.ReadAll(file)
		if b != nil {
			yaml.Unmarshal(b, u)
			log.Print("Load root user '", u.Name, "' SUCCESS")
		}
	} else {
		u.Name = "root"
		u.Pass = RandStringRunes(10)
		d, err := yaml.Marshal(u)
		if err != nil {
			log.Print(err)
		}

		err = ioutil.WriteFile(filename, d, 0x700)
		if err != nil {
			log.Print("Init root user fail, ", err)
		} else {
			log.Print("Init root user SUCCESS")
		}
	}
}


func ReadConfig(c *Config, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, c);
}


func DefaultConfig(c *Config) {
	c.HttpPort = 7707
	// mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]
	c.MongoURL = "mongodb://localhost:27017"
	c.MongoDBName = "ic1101"
	c.Salt = RandStringRunes(20)
}
