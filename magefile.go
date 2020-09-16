// +build mage

package main

import (
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/sumup-oss/go-pkgs/testutils"

	"log"
	"net/http"
	"time"
)

const (
	dockerNeo4jTestContainerName = "tf_neo4j_test_container"
)

var (
	neo4jImageVersions = []string{
		"4.1.1",
	}
)

func Test() {
	mg.Deps(testAgainstStableNeo4j)
}

func testAgainstStableNeo4j() error {
	for _, imageVersion := range neo4jImageVersions {
		err := runTestAgainstNeo4j(imageVersion)
		if err != nil {
			return err
		}
	}

	return nil
}

func runTestAgainstNeo4j(imageVersion string) error {
	// NOTE: Ignore error since we clean optimistically
	sh.Run("docker", "rm", "-fv", dockerNeo4jTestContainerName)

	// NOTE: Change password every time to make sure we're not using a hardcoded one
	password := testutils.RandString(15)

	err := sh.Run(
		"docker",
		"run",
		"-p",
		"7474:7474",
		"-p",
		"7687:7687",
		fmt.Sprintf("--name=%s", dockerNeo4jTestContainerName),
		"-d",
		"--rm",
		"-ti",
		"-e",
		fmt.Sprintf("NEO4J_AUTH=neo4j/%s", password),
		fmt.Sprintf("neo4j:%s", imageVersion),
	)
	if err != nil {
		return err
	}

	// NOTE: Check 15 times with interval 1 second,
	// for healthiness of previously started Vault.
	isHealthy := false

	log.Printf("Waiting for neo4j %s to be healthy\n", imageVersion)
	for i := 0; i < 15; i++ {
		if isNeo4jHealthy() {
			isHealthy = true
			break
		}
		time.Sleep(1 * time.Second)
	}

	if !isHealthy {
		return fmt.Errorf("neo4j %s still not healthy after 15 attempts\n", imageVersion)
	}

	log.Printf("Neo4j %s is healthy. Proceeding with test plan\n", imageVersion)

	args := []string{"test", "./...", "-cover"}
	if mg.Verbose() {
		args = append(args, "-v")
	}

	err = sh.RunWith(
		map[string]string{
			"NEO4J_CONNECTION_URI": "neo4j://localhost:7687",
			"NEO4J_USERNAME":       "neo4j",
			"NEO4J_PASSWORD":       password,
		},
		"go",
		args...,
	)

	if err != nil {
		return err
	}

	sh.Run("docker", "rm", "-fv", dockerNeo4jTestContainerName)
	return nil
}

func Lint() error {
	return sh.Run("golangci-lint", "run", "--timeout=10m")
}

func UnitTests() error {
	args := []string{"test", "./...", "-cover", "-short"}
	if mg.Verbose() {
		args = append(args, "-v")
	}

	return sh.Run("go", args...)
}

func isNeo4jHealthy() bool {
	// HACK: Afaik there's no official health check endpoint, so test if the web UI is up.
	resp, err := http.Get("http://localhost:7474")
	if err != nil {
		return false
	}

	return resp.StatusCode == 200
}
