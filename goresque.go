package goresque

import (
	redis "github.com/hoisie/redis"
	"fmt"
	"os"
	"strconv"
	"encoding/json"
)

type Resque struct {
	Server  string
	Namespace	string
	Port    int
	Db      int
	Queues  []Queue
	Workers []Worker
	client  *redis.Client
}

type Queue struct {
	Id     int
	Namespace string
	Name   string
	client *redis.Client
}

type Worker struct {
	Id     int
	Name   string
	client *redis.Client
}

type Job struct {
	*Queue
	Class string "class"
	Args  []interface{} "args"
	Worker		*Worker
}


func (self *Queue) pop() (job *Job, err os.Error) {
	//decode redis.lpop("queue:#{queue}")
	key := fmt.Sprintf("%squeue:%s", self.Namespace,self.Name)
	data, err := self.client.Lpop(key)
	if err != nil {
		return job, err
	}
	job = new(Job)
	err = json.Unmarshal(data, job)
	job.Queue = self
	return job, err

}

func (self *Queue) size() (int, os.Error) {
	key := fmt.Sprintf("%squeue:%s", self.Namespace,self.Name)
	return self.client.Llen(key)
}

func(self *Resque) watchQueue(queue string)(ok bool, err os.Error){
		ok,err = self.client.Sadd("%squeues", []byte(queue))
		return
}

func (self *Resque) Enqueue(queue string, job *Job)(err os.Error) {
 	self.watchQueue(queue)
	
	outjson, _ := json.Marshal(job)
	key := fmt.Sprintf("%squeue:%s",self.Namespace,queue)
 	err = self.client.Rpush(key,outjson)
	return
}

func (self *Resque) Reserve(queue string) (job *Job, err os.Error) {
	//decode redis.lpop("queue:#{queue}")
	key := fmt.Sprintf("%squeue:%s", self.Namespace, queue)
	data, err := self.client.Lpop(key)
	if err != nil {
		return job, err
	}
	job = new(Job)
	err = json.Unmarshal(data, job)
	return job, err

}

func (self *Resque) getStat(name string) (int, os.Error) {
	key := fmt.Sprintf("%sstat:%s", self.Namespace, name)
	val, err := self.client.Get(key)
	strval := string(val)
	intval, _ := strconv.Atoi(strval)
	return intval, err
}

func (self *Resque) getWorkers() []Worker {
	workers, err := self.client.Smembers(self.Namespace + "workers")
	if err != nil {
		fmt.Println(err)
	}
	var w Worker
	qs := make([]Worker, 1000)
	for i, val := range workers {
		w = Worker{Id: i, Name: string(val)}
		w.client = self.client
		qs[i] = w
	}
	self.Workers = qs
	return self.Workers[0:len(workers)]
}


func (self *Resque) getQueues() []Queue {
	members, err := self.client.Smembers(self.Namespace + "queues")
	if err != nil {
		fmt.Println(err)
	}
	var q Queue
	qs := make([]Queue, 100)
	for i, val := range members {
		q = Queue{Id: i, Namespace: self.Namespace, Name: string(val)}
		q.client = self.client
		qs[i] = q
	}
	self.Queues = qs
	return self.Queues[0:len(members)]
}

func NewResque(server string, port int, db int, namespace string) (resque *Resque) {
	resque = new(Resque)
	resque.Server = server
	resque.Port = port
	resque.Db = db
	if len(namespace) > 0 {
		resque.Namespace = namespace + ":"
	} else {
		resque.Namespace = ""
	}
	resque.client = new(redis.Client)
	resque.client.Db = db
	resque.Queues = make([]Queue, 0)
	resque.Workers = make([]Worker, 0)
	address := fmt.Sprintf("%s:%d", resque.Server, resque.Port)
	resque.client.Addr = address
	keys, err := resque.client.Keys("*")
	if err != nil {
		fmt.Println(err.String())
	}
	fmt.Println("KEYDUMP:")
	for _,y := range keys {
		fmt.Println("Key : ", y)
	}
	fmt.Println(resque)
	return resque
}
