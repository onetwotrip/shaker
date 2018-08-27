package shaker

import (
	"encoding/json"
	"github.com/bamzi/jobrunner"
	"github.com/bsm/redis-lock"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

func (s *Shaker) getConfig(configFile string) {
	log := MakeLog()
	s.Log().Infof("Reading configuration from %s", configFile)
	config, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Cant't read config file %s", configFile)
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

	s.log.Infof("Reading directory %s", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, file := range files {
		jobFile := dir + "/" + file.Name()
		s.Log().Infof("Reading file for %s jobs %s", jobType, jobFile)
		configByte, err := ioutil.ReadFile(jobFile)
		if err != nil {
			s.Log().Fatalf("Cant't read config file %s", jobFile)
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
	var jobs jobs

	s.log.Infof("Reading directory %s", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		s.log.Fatal(err)
		return
	}

	for _, file := range files {
		jobFile := dir + "/" + file.Name()
		configByte, err := ioutil.ReadFile(jobFile)
		if err != nil {
			s.Log().Fatalf("Cant't read config file %s", jobFile)
		}

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
		return method
	}
	return "get"
}

func findRedisType(method string) string {
	if findType(method) == "redis" {
		switch findMethod(method) {
		case "publish":
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
		s.Log().Infof("Add job %s with lock timeout %d second from file %s", job.Name, lockTimeout, jobFile)

		var username string
		var password string

		if job.User != "" {
			s.Log().Infof("Will use ^%s^ user for job", job.User)
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
		jobrunner.Schedule(job.Cron, RunJob{
			log:     s.Log(),
			lock:    locker,
			request: *request,
			clients: clients,
		})
	}
}
