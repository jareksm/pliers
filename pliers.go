package main

import (
        "fmt"
        "log"
//        "io/ioutil"
//        "gopkg.in/yaml.v2"
//        "flag"
        "github.com/codegangsta/cli"
        "os"
        "sync"
)

// Instance definition type
type Instance struct {
    Provider string `yaml:"provider"`
    Type string `yaml:"type"`
    Subnet string `yaml:"subnet"`
    IP string `yaml:"ip"`
    Image string `yaml:"image"`
    Tags map[string]string `yaml:"tags,omitempty"`
    Volumes map[string]int `yaml:"volumes,omitempty"`
}

var file string
var w sync.WaitGroup

func dieIf(message string, err error)  {
    if err != nil {
        log.Fatal(message, err)
    }
}

func build(vm string) {
    log.Println("Building ", vm)
    w.Done()
}

func main() {
//    file = flag.String("file", "environment.yaml", "file to read the config from")
//    flag.Parse()

    app := cli.NewApp()
    app.Name = "pliers"
    app.Usage = "manipulate VMs in the cloud"
    app.Flags = []cli.Flag {
        cli.StringFlag{
            Name: "file",
            Value: "environment.yaml",
            Usage: "file to read the config from",
            Destination: &file,
        },
    }

    app.Commands = []cli.Command{
		{
			Name:  "build",
			Usage: "build vm [vm [...]]",

			Action: func(c *cli.Context) {
                if c.NArg() > 0 {
                    vms := c.Args()
                    w.Add(c.NArg())
                    for _, vm := range(vms) {
                        go build(vm)
                    }
                    w.Wait()
                } else {
                    fmt.Println("Not enough arguments")
                    os.Exit(1)
                }
			},
/*			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "format",
					Usage:       "json or table. table is the default.",
					Destination: &format,
				},
			},
*/
		},
    }
    app.Run(os.Args)
/*    m := make(map[interface{}]Instance)
    data, err := ioutil.ReadFile("environment.yaml")
    dieIf("Could not open config file", err)

    err = yaml.Unmarshal([]byte(data), &m)
    if err != nil {
        log.Fatalf("error: %v", err)
    }
    fmt.Printf("--- m:\n%v\n\n", m)
    d, err := yaml.Marshal(&m)
    if err != nil {
        log.Fatalf("error: %v", err)
    }
    fmt.Printf("--- m dump:\n%s\n\n", string(d))
*/
}
