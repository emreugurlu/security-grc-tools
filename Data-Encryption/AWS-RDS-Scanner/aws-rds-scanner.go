package main

import (
    "bufio"
    "encoding/csv"
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/rds"
)

func main() {
    accessKeyID := promptUser("Enter AWS Access Key ID:")
    secretAccessKey := promptUser("Enter AWS Secret Access Key:")
    sessionToken := promptUser("Enter AWS Session Token (press Enter if not applicable):")
    region := promptUser("Enter AWS Region:")
    scanType := promptUser("Select Scan Type:\n1. DB Instances\n2. DB Clusters\n3. Both\nEnter your choice (1, 2, or 3):")

    if scanType != "1" && scanType != "2" && scanType != "3" {
        fmt.Println("Invalid choice. Exiting.")
        return
    }

    sess, err := session.NewSession(&aws.Config{
        Region:      aws.String(region),
        Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, sessionToken),
    })
    if err != nil {
        fmt.Println("Error creating AWS session:", err)
        return
    }

    rdsClient := rds.New(sess)

    var instances []*rds.DBInstance
    var clusters []*rds.DBCluster

    if scanType == "1" || scanType == "3" {
        instances, err = scanDBInstances(rdsClient)
        if err != nil {
            fmt.Println("Error scanning DB Instances:", err)
            return
        }
    }

    if scanType == "2" || scanType == "3" {
        clusters, err = scanDBClusters(rdsClient)
        if err != nil {
            fmt.Println("Error scanning DB Clusters:", err)
            return
        }
    }

    if scanType == "1" || scanType == "3" {
        writeToCSV("DBInstances", instances, rdsClient)
    }
    if scanType == "2" || scanType == "3" {
        writeToCSV("DBClusters", clusters, rdsClient)
    }
}

func promptUser(prompt string) string {
    fmt.Print(prompt + " ")
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    return scanner.Text()
}

func scanDBInstances(rdsClient *rds.RDS) ([]*rds.DBInstance, error) {
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

func writeToCSV(filePrefix string, data interface{}, rdsClient *rds.RDS) {
    fileName := fmt.Sprintf("%s_%s.csv", filePrefix, time.Now().Format("20060102_150405"))
    file, err := os.Create(fileName)
    if err != nil {
        fmt.Println("Error creating CSV file:", err)
        return
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    switch d := data.(type) {
    case []*rds.DBInstance:
        writer.Write([]string{"DB Instance", "StorageEncryption", "Availability Zone"})
        for _, instance := range d {
            storageEncrypted := "false"
            if instance.StorageEncrypted != nil {
                storageEncrypted = strconv.FormatBool(*instance.StorageEncrypted)
            }
            availabilityZone := ""
            if instance.AvailabilityZone != nil {
                availabilityZone = *instance.AvailabilityZone
            }
            writer.Write([]string{*instance.DBInstanceIdentifier, storageEncrypted, availabilityZone})
        }
    case []*rds.DBCluster:
        writer.Write([]string{"DB Cluster", "StorageEncrypted", "Availability Zones", "EncryptionInTransit"})
        for _, cluster := range d {
            storageEncrypted := "false"
            if cluster.StorageEncrypted != nil {
                storageEncrypted = strconv.FormatBool(*cluster.StorageEncrypted)
            }

            encryptionInTransit := "Unknown"
            if cluster.DBClusterParameterGroup != nil {
                eit, err := checkEncryptionInTransit(rdsClient, *cluster.DBClusterParameterGroup)
                if err == nil {
                    encryptionInTransit = strconv.FormatBool(eit)
                }
            }

            writer.Write([]string{
                *cluster.DBClusterIdentifier,
                storageEncrypted,
                joinStrings(cluster.AvailabilityZones),
                encryptionInTransit,
            })
        }
    }

    fmt.Printf("Results written to %s\n", fileName)
}

func checkEncryptionInTransit(rdsClient *rds.RDS, parameterGroupName string) (bool, error) {
    var marker *string
    for {
        paramsOutput, err := rdsClient.DescribeDBClusterParameters(&rds.DescribeDBClusterParametersInput{
            DBClusterParameterGroupName: aws.String(parameterGroupName),
            Marker:                      marker,
        })
        if err != nil {
            fmt.Printf("Error retrieving parameters for group %s: %v\n", parameterGroupName, err)
            return false, err
        }

        for _, param := range paramsOutput.Parameters {
            if param.ParameterName != nil && *param.ParameterName == "require_secure_transport" {
                fmt.Printf("Parameter found - Name: %s, Value: %v, Type: %T\n", *param.ParameterName, param.ParameterValue, param.ParameterValue)
		return param.ParameterValue != nil && strings.ToLower(*param.ParameterValue) == "on", nil

            }
        }

        if paramsOutput.Marker == nil {
            break
        }
        marker = paramsOutput.Marker
    }

    fmt.Printf("No require_secure_transport parameter found for %s\n", parameterGroupName)
    return false, nil
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
