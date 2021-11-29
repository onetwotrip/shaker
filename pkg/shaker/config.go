package shaker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/bamzi/jobrunner"
	"github.com/bsm/redis-lock"
	"gopkg.in/yaml.v2"
)

func (s *Shaker) getConfig(configFile string) {
	log := MakeLog()
	s.Log().Infof("reading configuration from %s", configFile)
	config, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("cant't read config file %s", configFile)
	}
	err = yaml.Unmarshal(config, &s.config)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Shaker) isValidConfig() bool {
	if s.validateConfigs("http") && s.validateConfigs("redis") {
		return true
	}
	return false
}

func (s *Shaker) validateConfigs(jobType string) bool {
	var dir string
	var jobs jobs

	switch jobType {
	case "http":
		dir = s.config.Jobs.HTTP.Dir
	case "redis":
		dir = s.config.Jobs.Redis.Dir
	}

	s.log.Infof("reading directory %s", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			s.Log().Warnf("skip file %s due prefix '.'", file.Name())
			continue
		}
		if !strings.HasSuffix(file.Name(), ".json"){
			s.Log().Warnf("DEPRECATED: configs without .json extension will be skipped in future: '%s'", file.Name())
		}
		jobFile := dir + "/" + file.Name()
		s.Log().Infof("reading file for %s jobs %s", jobType, jobFile)
		configByte, err := ioutil.ReadFile(jobFile)
		if err != nil {
			s.Log().Fatalf("cant't read config file %s", jobFile)
			return false
		}

		err = json.Unmarshal(configByte, &jobs)
		if err != nil {
			s.Log().Error(err)
			return false
		}
	}

	return true
}

func (s *Shaker) readConfigDirectory(dir string, jobType string) {
	s.log.Infof("reading directory %s", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		s.log.Fatal(err)
		return
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			s.Log().Warnf("skip file %s due prefix '.'", file.Name())
			continue
		}
		if !strings.HasSuffix(file.Name(), ".json"){
			s.Log().Warnf("DEPRECATED: configs without .json extension will be skipped in future (%s)", file.Name())
		}
		jobFile := dir + "/" + file.Name()
		configByte, err := ioutil.ReadFile(jobFile)
		if err != nil {
			s.Log().Fatalf("cant't read config file %s", jobFile)
		}

		var jobs jobs

		err = json.Unmarshal(configByte, &jobs)
		if err != nil {
			s.Log().Error(err)
		}
		s.loadJobs(jobs, jobFile)
	}

}

func findType(method string) string {
	switch method {
	case "get":
		return "http"
	case "post":
		return "http"
	case "publish":
		return "redis"
	default:
		return "http"
	}
}

func findMethod(method string) string {
	if method != "" {
		return strings.ToUpper(method)
	}
	return "GET"
}

func findRedisType(method string) string {
	if findType(method) == "redis" {
		switch findMethod(method) {
		case "PUBLISH":
			return "pubsub"
		}
	}
	return "default"
}

func (s *Shaker) loadJobs(jobs jobs, jobFile string) {

	for _, job := range jobs.Jobs {
		lockTimeout := 0
		if job.Method != "publish" {
			lockTimeout = 30
			if job.LockTimeout > 0 {
				lockTimeout = job.LockTimeout
			}
		}
		s.Log().Infof("schedule job %s with lock timeout %d second from file %s", job.Name, lockTimeout, jobFile)

		var username string
		var password string

		if job.User != "" {
			s.Log().Infof("will use %s user for job %s", job.User, job.Name)
			username = s.config.Users[job.User].Username
			password = s.config.Users[job.User].Password
		}

		//Creating redis lock
		locker := lock.New(s.connectors.redisStorages["default"], getMD5Hash(urlFormater(jobs.URL, job.URI)), &lock.Options{
			LockTimeout: time.Duration(lockTimeout) * time.Second,
			RetryCount:  0,
			RetryDelay:  time.Microsecond * 100})

		//Creating request
		request := &request{
			name:        job.Name,
			url:         urlFormater(jobs.URL, job.URI),
			body:        job.Body,
			method:      findMethod(job.Method),
			requestType: findType(job.Method),
			username:    username,
			password:    password,
			channel:     job.Channel,
			message:     job.Message,
			timeout:     time.Duration(job.Timeout) * time.Second,
		}

		//Creating Clients
		clients := &clients{
			redisStorage: s.connectors.redisStorages[findRedisType(job.Method)],
			slackClient:  s.connectors.slackConfig,
		}

		//Creating Job with all parameters
		err := jobrunner.Schedule(job.Cron, RunJob{
			log:     s.Log(),
			lock:    locker,
			request: *request,
			clients: clients,
		})
		s.Log().Debugf("jobrunner.Schedule result for '%s' is '%s'", job.Name, err)

		if err == nil {
			s.Log().Infof("success schedule for '%s'", job.Name)
		} else {
			message := fmt.Sprintf("can't schedule job '%s' from '%s', due: '%s'", job.Name, jobFile, err)
			s.Log().Error(message)
			slackSendErrorMessage(s.connectors.slackConfig, "shaker config", message, "", 0)
		}
	}
}
