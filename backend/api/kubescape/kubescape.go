package kubescape

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/komodorio/validkube/backend/api/utils"
	"github.com/komodorio/validkube/backend/internal/routing"
	"sigs.k8s.io/yaml"
)

const Path = "/kubescape"
const Method = routing.POST

func kubescapeWrapper(inputYaml []byte) ([]byte, error) {
	err := utils.CreateDirectory("/tmp/yaml")
	if err != nil {
		return nil, err
	}

	err = utils.WriteFile("/tmp/yaml/target_yaml.yaml", inputYaml)
	if err != nil {
		return nil, err
	}

	outputFile := "/tmp/yaml/output.json"
	exec.Command("kubescape", "scan", "/tmp/yaml/target_yaml.yaml", "-o", outputFile, "-f", "json").Output()

	outputFromKubescapeAsJson, err := ioutil.ReadFile(outputFile)
	if err != nil {
		return nil, err
	}

	outputFromKubescapeAsYaml, err := yaml.JSONToYAML(outputFromKubescapeAsJson)
	if err != nil {
		return nil, err
	}
	return outputFromKubescapeAsYaml, nil
}

func ProcessRequest(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Printf("Erorr has with reading request body: %v", err)
		c.JSON(http.StatusOK, gin.H{"data": "", "err": err.Error()})
		return
	}
	bodyAsMap, err := utils.JsonToMap(body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": "", "err": err.Error()})
		return
	}
	yamlAsInterface := bodyAsMap["yaml"]
	kubescapeOutput, err := kubescapeWrapper(utils.InterfaceToBytes(yamlAsInterface))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": "", "err": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": string(kubescapeOutput), "err": nil})
}
