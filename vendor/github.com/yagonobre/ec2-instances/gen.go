// +build ignore
// This program generates instances_data.go. It can be invoked by running
// go generate

package main

import (
	"bytes"
	"encoding/json"
	"go/format"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type InstanceSize struct {
	InstanceType string          `json:"instance_type"`
	VCPURaw      json.RawMessage `json:"vCPU"`
	VCPU         int
	Memory       float64 `json:"memory"`
}

const instancesPath = "hack/instances.json"

func main() {
	instancesFileName, err := os.Open(instancesPath)
	checkError(err)
	defer instancesFileName.Close()

	instancesFile, err := ioutil.ReadAll(instancesFileName)
	checkError(err)

	var instancesData []InstanceSize
	err = json.Unmarshal(instancesFile, &instancesData)
	checkError(err)
	instancesData = getCPUInfo(instancesData)

	instancesDataFile, err := os.Create("instances_data.go")
	var codeBuffer bytes.Buffer

	fileTemplate.Execute(&codeBuffer, struct {
		Timestamp time.Time
		Instances []InstanceSize
	}{
		Timestamp: time.Now(),
		Instances: instancesData,
	})

	code, err := format.Source(codeBuffer.Bytes())
	checkError(err)
	instancesDataFile.Write(code)
	checkError(err)
}

func getCPUInfo(instances []InstanceSize) []InstanceSize {
	var res []InstanceSize
	for _, instance := range instances {
		var VCPU int
		if err := json.Unmarshal(instance.VCPURaw, &VCPU); err == nil {
			instance.VCPU = VCPU
		}
		res = append(res, instance)
	}
	return res
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var fileTemplate = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots at
// {{ .Timestamp }}
// using data from
// https://www.ec2instances.info/instances.json
package ec2instances

// Instances is a map from instance type to InstanceSize
var Instances = map[string]InstanceSize{
	{{- range .Instances }}
		"{{.InstanceType}}" : { {{.VCPU}}, {{ printf " %.2f }," .Memory }}
		
	{{- end }}
	}
`))
