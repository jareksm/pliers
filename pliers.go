package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
	//    "flag"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/codegangsta/cli"
)

// Types

// Instance definition type
type Instance struct {
	Provider string            `yaml:"provider"`
	Type     string            `yaml:"type"`
	Subnet   string            `yaml:"subnet"`
	IP       string            `yaml:"ip"`
	Image    string            `yaml:"image"`
	Tags     map[string]string `yaml:"tags,omitempty"`
	Volumes  map[string]int    `yaml:"volumes,omitempty"`
}

type dataMap map[string]Instance

// Global variables

var file string
var w sync.WaitGroup
var data dataMap

func createAWSSession() *ec2.EC2 {
	svc := ec2.New(session.New(), &aws.Config{Region: aws.String("ap-southeast-2")})
	return svc
}

func dieIf(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

// TODO: clean up this mess
func buildInAWS(i Instance) {
	//log.Println("Building ", i)
	svc := createAWSSession()
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		// An Amazon Linux AMI ID for t2.micro instances in the us-west-2 region
		ImageId:      aws.String(i.Image),
		InstanceType: aws.String(i.Type),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
			{
				DeviceIndex: aws.Int64(0),
				SubnetId:    aws.String(i.Subnet),
			},
		},
	})
	if err != nil {
		log.Println("Could not create instance", err)
	} else {
		log.Println("Created instance", *runResult.Instances[0].InstanceId)
	}

	w.Done()
}

func buildThem(vms []string) {
	for _, vmName := range vms {
		instance, ok := data[vmName]
		if ok {
			log.Println("Building", vmName, "... Provider is", instance.Provider)
			switch instance.Provider {
			case "aws":
				w.Add(1)
				go buildInAWS(instance)
			default:
				log.Println("Not sure how to build in", instance.Provider, "... Skipping.")
			}
		} else {
			log.Println(vmName, "is not defined. Skipping.")
		}
	}
	w.Wait()
}

func main() {
	//    file = flag.String("file", "environment.yaml", "file to read the config from")
	//    flag.Parse()

	app := cli.NewApp()
	app.Name = "pliers"
	app.Usage = "manipulate VMs in the cloud"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "file",
			Value:       "environment.yaml",
			Usage:       "file to read the config from",
			Destination: &file,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "build",
			Usage: "build vm [vm [...]]",

			Action: func(c *cli.Context) {
				if c.NArg() > 0 {
					buildThem(c.Args())
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

	data = make(dataMap)
	d, err := ioutil.ReadFile("environment.yaml")
	dieIf("Could not open config file", err)

	err = yaml.Unmarshal([]byte(d), &data)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = app.Run(os.Args)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

}
