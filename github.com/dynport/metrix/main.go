package main

import (
	"fmt"
	"os"
)

type Config struct {
	OpenTSDBUrl string `json:"opentsdb_url"`
}

var OUTPUTS = []string{"graphite", "opentsdb"}

func processCollector(mh *MetricHandler, output *OutputHandler, c MetricCollector) (e error) {
	all, e := mh.Collect(c)
	if e != nil {
		return
	}
	e = output.WriteMetrics(all)
	return
}

const (
	KEYS = "keys"
	HELP = "help"

	HOSTNAME = "hostname"

	OPENTSDB = "opentsdb"
	GRAPHITE = "graphite"
)

func init() {
	parser.Add(HELP, "true", "Print this usage page")
	parser.Add(KEYS, "true", "Only list all known keys")
	parser.AddKey(OPENTSDB, "Report metrics to OpenTSDB host.\nEXAMPLE: opentsdb.host:4242")
	parser.AddKey(GRAPHITE, "Report metrics to Graphite host.\nEXAMPLE: graphite.host:2003")
	parser.AddKey(HOSTNAME, "Hostname to be used for tagging. If blank the local hostname is used")
}

func main() {
	if os.Getenv("DEBUG") == "true" {
		logger.LogLevel = DEBUG
	} else {
		logger.LogLevel = WARN
	}

	if e := parser.ProcessAll(os.Args[1:]); e != nil {
		logger.Error(e.Error())
		os.Exit(1)
	}

	output := &OutputHandler{}
	if a := parser.Get(OPENTSDB); a != "" {
		output.OpenTSDBAddress = a
	}
	if a := parser.Get(GRAPHITE); a != "" {
		output.GraphiteAddress = a
	}
	if a := parser.Get(HOSTNAME); a != "" {
		output.Hostname = a
	}

	collectors := map[string]MetricCollector{
		ELASTICSEARCH: &ElasticSearch{Url: parser.Get(ELASTICSEARCH)},
		REDIS:         &Redis{Address: parser.Get(REDIS)},
		CPU:           &Cpu{},
		LOADAVG:       &LoadAvg{},
		MEMORY:        &Memory{},
		DISK:          &Disk{},
		DF:            &Df{},
		NET:           &Net{},
		FREE:          &Free{},
		PROCESSES:     &Processes{},
		RIAK:          &Riak{Address: parser.Get(RIAK)},
		PGBOUNCER:     &PgBouncer{Address: parser.Get(PGBOUNCER)},
		POSTGRES:      &PostgreSQLStats{Uri: parser.Get(POSTGRES)},
		NGINX:         &Nginx{Address: parser.Get(NGINX)},
		FILES:         &Files{},
	}

	mh := &MetricHandler{}

	if parser.Get(HELP) == "true" || len(parser.Values) == 0 {
		parser.PrintDefaults()
		os.Exit(0)
	}

	if parser.Get(KEYS) == "true" {
		for k, _ := range parser.Values {
			if col, ok := collectors[k]; ok {
				for _, m := range mh.Keys(col) {
					fmt.Println(m)
				}
			}
		}
		return
	}

	for key, _ := range parser.Values {
		if collector, ok := collectors[key]; ok {
			if e := processCollector(mh, output, collector); e != nil {
				logger.Error("processong collector", key, e.Error())
			}
		} else {
			isOutput := false
			for _, out := range OUTPUTS {
				if out == key {
					isOutput = true
					break
				}
			}
			if !isOutput {
				logger.Error("ERROR: no collector found for " + key)
			}
		}
	}
}
