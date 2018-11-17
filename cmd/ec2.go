package cmd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	instances "github.com/yagonobre/ec2-instances"
)

var ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "Generate report about ec2 usage.",
}

var instancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Generate report about ec2 instances.",
	Run:   ec2Instances,
}

func ec2Instances(cmd *cobra.Command, args []string) {
	// Load session from shared config
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create new EC2 client
	ec2Svc := ec2.New(sess)

	res, err := ec2Svc.DescribeInstances(nil)
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}

	countByInstanceType := make(map[string]int)

	for _, reservation := range res.Reservations {
		for _, instance := range reservation.Instances {
			if (*instance.State.Code) == 16 {
				countByInstanceType[*(instance.InstanceType)] += 1
			}
		}
	}

	var total, totalCPU int
	var totalMemory float64
	instancesInfo := instances.Instances

	for instanceType, count := range countByInstanceType {
		fmt.Printf("%s : %d\n", instanceType, count)
		total += count

		if instanceInfo, ok := instancesInfo[instanceType]; ok {
			totalMemory += (instanceInfo.Memory * float64(count))
			totalCPU += (instanceInfo.VCPU * count)
		} else {
			fmt.Println("InstanceType not found", instanceType)
		}
	}

	fmt.Printf("Instance Count: %d\n", total)
	fmt.Printf("Total Memory: %.2f GiB\nTotal CPU: %d cores\n", totalMemory, totalCPU)
}

func init() {
	ec2Cmd.AddCommand(instancesCmd)
	rootCmd.AddCommand(ec2Cmd)
}
