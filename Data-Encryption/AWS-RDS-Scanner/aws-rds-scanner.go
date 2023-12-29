package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
)

func main() {
	// Prompt user for AWS credentials
	accessKeyID := promptUser("Enter AWS Access Key ID:")
	secretAccessKey := promptUser("Enter AWS Secret Access Key:")
	sessionToken := promptUser("Enter AWS Session Token (press Enter if not applicable):")

	// Prompt user for AWS region
	region := promptUser("Enter AWS Region:")

	// Prompt user for scan type
	scanType := promptUser("Select Scan Type:\n1. DB Instances\n2. DB Clusters\n3. Both\nEnter your choice (1, 2, or 3):")

	// Validate user input for scan type
	if scanType != "1" && scanType != "2" && scanType != "3" {
		fmt.Println("Invalid choice. Exiting.")
		return
	}

	// Create an AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, sessionToken),
	})
	if err != nil {
		fmt.Println("Error creating AWS session:", err)
		return
	}

	// Create an RDS service client
	rdsClient := rds.New(sess)

	// List all RDS instances and clusters based on user-selected scan type
	switch scanType {
	case "1":
		instances, err := scanDBInstances(rdsClient)
		if err != nil {
			fmt.Println("Error scanning DB Instances:", err)
			return
		}
		writeToCSV("DBInstances", instances)

	case "2":
		clusters, err := scanDBClusters(rdsClient)
		if err != nil {
			fmt.Println("Error scanning DB Clusters:", err)
			return
		}
		writeToCSV("DBClusters", clusters)

	case "3":
		instances, err := scanDBInstances(rdsClient)
		if err != nil {
			fmt.Println("Error scanning DB Instances:", err)
			return
		}
		writeToCSV("DBInstances", instances)

		clusters, err := scanDBClusters(rdsClient)
		if err != nil {
			fmt.Println("Error scanning DB Clusters:", err)
			return
		}
		writeToCSV("DBClusters", clusters)
	}
}

func scanDBInstances(rdsClient *rds.RDS) ([]*rds.DBInstance, error) {
	// Fetch information about DB instances
	var totalInstances []*rds.DBInstance
	var marker *string

	for {
		instancesInput := &rds.DescribeDBInstancesInput{
			MaxRecords: aws.Int64(100),
		}
		if marker != nil {
			instancesInput.Marker = marker
		}

		instancesOutput, err := rdsClient.DescribeDBInstances(instancesInput)
		if err != nil {
			return nil, err
		}

		totalInstances = append(totalInstances, instancesOutput.DBInstances...)

		if instancesOutput.Marker != nil {
			marker = instancesOutput.Marker
		} else {
			break
		}
	}

	return totalInstances, nil
}

func scanDBClusters(rdsClient *rds.RDS) ([]*rds.DBCluster, error) {
	// Fetch information about DB clusters
	var totalClusters []*rds.DBCluster
	var marker *string

	for {
		clustersInput := &rds.DescribeDBClustersInput{
			MaxRecords: aws.Int64(100),
		}
		if marker != nil {
			clustersInput.Marker = marker
		}

		clustersOutput, err := rdsClient.DescribeDBClusters(clustersInput)
		if err != nil {
			return nil, err
		}

		totalClusters = append(totalClusters, clustersOutput.DBClusters...)

		if clustersOutput.Marker != nil {
			marker = clustersOutput.Marker
		} else {
			break
		}
	}

	return totalClusters, nil
}

func writeToCSV(filePrefix string, data interface{}) {
	// Create a file with today's date and current time in the title
	fileName := fmt.Sprintf("%s_%s.csv", filePrefix, time.Now().Format("20060102_150405"))
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating CSV file:", err)
		return
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header based on the type of data
	switch data.(type) {
	case []*rds.DBInstance:
		writer.Write([]string{"DB Instance", "Encryption", "Availability Zone"})
	case []*rds.DBCluster:
		writer.Write([]string{"DB Cluster", "Encryption", "Availability Zones"})
	}

	// Write data to CSV
	switch d := data.(type) {
	case []*rds.DBInstance:
		for _, instance := range d {
			writer.Write([]string{*instance.DBInstanceIdentifier, fmt.Sprintf("%t", *instance.StorageEncrypted), *instance.AvailabilityZone})
		}
	case []*rds.DBCluster:
		for _, cluster := range d {
			writer.Write([]string{*cluster.DBClusterIdentifier, fmt.Sprintf("%t", *cluster.StorageEncrypted), joinStrings(cluster.AvailabilityZones)})
		}
	}

	fmt.Printf("Results written to %s\n", fileName)
}

func joinStrings(strPointers []*string) string {
	var strValues []string
	for _, strPointer := range strPointers {
		if strPointer != nil {
			strValues = append(strValues, *strPointer)
		}
	}
	return strings.Join(strValues, ", ")
}

func promptUser(prompt string) string {
	fmt.Print(prompt + " ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}
