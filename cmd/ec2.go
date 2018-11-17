package cmd

import (
	"fmt"
	"math/big"
	"os"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	humanize "github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	instances "github.com/yagonobre/ec2-instances"
)

var ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "Generate report about ec2 usage.",
}

var instancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Generate report about ec2 instances usage.",
	Run:   ec2Instances,
}

var ebsCmd = &cobra.Command{
	Use:     "ebs",
	Aliases: []string{"storage"},
	Short:   "Generate report about ebs usage.",
	Run:     ec2EBS,
}

func getEC2Client() *ec2.EC2 {
	return ec2.New(session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})))
}

func ec2Instances(cmd *cobra.Command, args []string) {
	iecFormat, err := cmd.Flags().GetBool("iec-format")
	if err != nil {
		fmt.Println("invalid iec-format flag")
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		fmt.Println("invalid verbose flag")
	}

	ec2Svc := getEC2Client()
	res, err := ec2Svc.DescribeInstances(nil)
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}

	countByInstanceType := make(map[string]int)

	for _, reservation := range res.Reservations {
		for _, instance := range reservation.Instances {
			if (*instance.State.Code) == 16 {
				countByInstanceType[*(instance.InstanceType)]++
			}
		}
	}

	var total, totalCPU int
	var totalMemory float64
	instancesInfo := instances.Instances

	w := new(tabwriter.Writer)
	if verbose {
		w.Init(os.Stdout, 8, 8, 0, '\t', 0)
		fmt.Fprintf(w, "%s\t%s\t\n", "Instance Type", "Instance Count")
	}

	for instanceType, count := range countByInstanceType {
		if verbose {
			fmt.Fprintf(w, "%s\t%d\t\n", instanceType, count)
		}

		total += count
		if instanceInfo, ok := instancesInfo[instanceType]; ok {
			totalMemory += (instanceInfo.Memory * float64(count))
			totalCPU += (instanceInfo.VCPU * count)
		} else {
			fmt.Println("InstanceType not found", instanceType)
		}
	}

	if verbose {
		w.Flush()
		fmt.Println()
	}

	fmt.Printf("Instance Count: %d\n", total)
	fmt.Printf("Total Memory: %s\nTotal CPU: %d cores\n", prettyPrintGiB(totalMemory, iecFormat), totalCPU)
}

func ec2EBS(cmd *cobra.Command, args []string) {
	iecFormat, err := cmd.Flags().GetBool("iec-format")
	if err != nil {
		fmt.Println("invalid iec-format flag")
	}

	ec2Svc := getEC2Client()
	res, err := ec2Svc.DescribeVolumes(nil)
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}

	var total int64
	for _, volume := range res.Volumes {
		total += *(volume.Size)
	}

	fmt.Printf("Volume Count: %d\n", len(res.Volumes))
	fmt.Printf("Volume Size: %s\n", prettyPrintGiB(float64(total), iecFormat))
}

func prettyPrintGiB(size float64, iecFormat bool) string {
	sizeMib := int64(size * 1024.00)                //Size in MiB
	sizeBigInt := big.NewInt(sizeMib * 1024 * 1024) //Size in Byte (big.Int)
	if iecFormat {
		return humanize.BigIBytes(sizeBigInt)
	}
	return humanize.BigBytes(sizeBigInt)
}

func init() {
	rootCmd.AddCommand(ec2Cmd)

	ec2Cmd.AddCommand(instancesCmd)
	ec2Cmd.AddCommand(ebsCmd)

	ec2Cmd.PersistentFlags().Bool("iec-format", false, "use iec size format(TiB instead of Tb)")

	instancesCmd.Flags().Bool("verbose", false, "print instance count in verbose mode")
}
