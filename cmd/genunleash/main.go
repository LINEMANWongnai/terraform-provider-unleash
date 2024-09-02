package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"

	"git.wndv.co/LINEMANWongnai/terraform-provider-unleash/internal/generator"
	"git.wndv.co/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

type Config struct {
	BaseURL            string `envconfig:"UNLEASH_BASE_URL" required:"true"`
	AuthorizationToken string `envconfig:"UNLEASH_AUTHORIZATION_TOKEN" required:"true"`
}

func main() {
	cfg := Config{}
	envconfig.MustProcess("APP", &cfg)

	err := run(cfg, os.Args)
	if err != nil {
		panic(err)
	}
}

func run(cfg Config, args []string) error {
	startTs := time.Now()

	if len(args) != 2 {
		return fmt.Errorf("usage: %s <project_id>", args[0])
	}

	client, err := unleash.CreateClient(cfg.BaseURL, cfg.AuthorizationToken)
	if err != nil {
		return err
	}

	tfWriter, cleanFn1, err := createWriter("gen.out.tf")
	defer cleanFn1()
	if err != nil {
		return err
	}

	importWriter, cleanFn2, err := createWriter("gen-import.out.tf")
	defer cleanFn2()
	if err != nil {
		return err
	}

	err = generator.Generate(client, args[1], tfWriter, importWriter)
	if err != nil {
		return err
	}

	err = tfWriter.Flush()
	if err != nil {
		return err
	}

	err = importWriter.Flush()
	if err != nil {
		return err
	}

	fmt.Printf("Successfully generate gen.out.tf and gen-import.out.tf files in %v seconds\n", time.Since(startTs).Seconds())

	return nil
}

func createWriter(fileName string) (*bufio.Writer, func(), error) {
	fo, err := os.Create(fileName)
	if err != nil {
		return nil, func() {}, err
	}
	cleanUpFn := func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}
	return bufio.NewWriter(fo), cleanUpFn, nil
}
