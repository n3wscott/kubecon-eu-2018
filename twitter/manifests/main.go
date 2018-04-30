package main

import (
	"fmt"
	"strings"
)

var instance = `apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: %s
  namespace: ledhouse
spec:
  clusterServiceClassExternalName: %s
  clusterServicePlanExternalName: %s
---
`

func instances() {
	rooms := []string{
		"1A", "1B", "1C",
		"2A", "2B", "2C",
		"3A", "3B", "3C",
		"4A",
	}
	colors := []string{"Red", "Green", "Blue"}

	for _, r := range rooms {
		for _, c := range colors {
			name := fmt.Sprintf("%s-%s", strings.ToLower(r), strings.ToLower(c))

			yaml := fmt.Sprintf(instance, name, r, c)

			fmt.Print(yaml)
		}
	}
}

var binding = `apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  name: %s
  namespace: ledhouse
spec:
  instanceRef:
    name: %s
---
`

func bindings() {
	rooms := []string{
		"1A", "1B", "1C",
		"2A", "2B", "2C",
		"3A", "3B", "3C",
		"4A",
	}
	colors := []string{"Red", "Green", "Blue"}

	for _, r := range rooms {
		for _, c := range colors {
			name := fmt.Sprintf("%s-%s", strings.ToLower(r), strings.ToLower(c))

			yaml := fmt.Sprintf(binding, name, name)

			fmt.Print(yaml)
		}
	}
}

var token = `            - name: %s
              valueFrom:
                secretKeyRef:
                  name: %s
                  key: token
`

func tokens() {
	rooms := []string{
		"1A", "1B", "1C",
		"2A", "2B", "2C",
		"3A", "3B", "3C",
		"4A",
	}
	colors := []string{"Red", "Green", "Blue"}

	for _, r := range rooms {
		for _, c := range colors {
			envName := fmt.Sprintf("TOKEN_%s_%s", strings.ToUpper(r), strings.ToUpper(c))
			name := fmt.Sprintf("%s-%s", strings.ToLower(r), strings.ToLower(c))

			yaml := fmt.Sprintf(token, envName, name)

			fmt.Print(yaml)
		}
	}
}

func main() {
	//instances()
	//bindings()
	tokens()
}
