package main

import (
    "bufio"
    "encoding/csv"
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"
	"math"
	"github.com/aws/aws-sdk-go/aws/awserr"


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
    maxAttempts := 2

    for attempt := 0; attempt < maxAttempts; attempt++ {
        instancesInput := &rds.DescribeDBInstancesInput{
            MaxRecords: aws.Int64(100),
        }
        if marker != nil {
            instancesInput.Marker = marker
        }

        instancesOutput, err := rdsClient.DescribeDBInstances(instancesInput)
        if err != nil {
            if aerr, ok := err.(awserr.Error); ok {
                switch aerr.Code() {
                case rds.ErrCodeDBInstanceNotFoundFault:
                    return nil, err
                case "ThrottlingException":
                    fmt.Printf("Throttling error detected, retrying... (Attempt %d/%d)\n", attempt+1, maxAttempts)
                    exponentialBackoff(attempt)
                    continue
                default:
                    return nil, err
                }
            }
        } else {
            totalInstances = append(totalInstances, instancesOutput.DBInstances...)

            if instancesOutput.Marker != nil {
                marker = instancesOutput.Marker
            } else {
                break
            }
        }
    }
    if len(totalInstances) == 0 {
        return nil, fmt.Errorf("max retry attempts reached")
    }
    return totalInstances, nil
}


func scanDBClusters(rdsClient *rds.RDS) ([]*rds.DBCluster, error) {
    var totalClusters []*rds.DBCluster
    var marker *string
    maxAttempts := 2

    for attempt := 0; attempt < maxAttempts; attempt++ {
        clustersInput := &rds.DescribeDBClustersInput{
            MaxRecords: aws.Int64(100),
        }
        if marker != nil {
            clustersInput.Marker = marker
        }

        clustersOutput, err := rdsClient.DescribeDBClusters(clustersInput)
        if err != nil {
            if aerr, ok := err.(awserr.Error); ok {
                switch aerr.Code() {
                case rds.ErrCodeDBClusterNotFoundFault:
                    return nil, err
                case "ThrottlingException":
                    fmt.Printf("Throttling error detected, retrying... (Attempt %d/%d)\n", attempt+1, maxAttempts)
                    exponentialBackoff(attempt)
                    continue
                default:
                    return nil, err
                }
            }
        } else {
            totalClusters = append(totalClusters, clustersOutput.DBClusters...)

            if clustersOutput.Marker != nil {
                marker = clustersOutput.Marker
            } else {
                break
            }
        }
    }

    if len(totalClusters) == 0 {
        return nil, fmt.Errorf("max retry attempts reached")
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
        writer.Write([]string{"DB Instance", "StorageEncryption", "Engine", "Availability Zone"})
        for _, instance := range d {
            storageEncrypted := "false"
            if instance.StorageEncrypted != nil {
                storageEncrypted = strconv.FormatBool(*instance.StorageEncrypted)
            }
            availabilityZone := ""
            if instance.AvailabilityZone != nil {
                availabilityZone = *instance.AvailabilityZone
            }
            engine := ""
            if instance.Engine != nil {
                engine = *instance.Engine
            }

            writer.Write([]string{
                *instance.DBInstanceIdentifier,
                storageEncrypted,
				engine,
                availabilityZone,
            })
        }
    case []*rds.DBCluster:
        writer.Write([]string{"DB Cluster", "StorageEncrypted", "EncryptionInTransit", "Engine", "Availability Zones"})
        for _, cluster := range d {
            storageEncrypted := "false"
            if cluster.StorageEncrypted != nil {
                storageEncrypted = strconv.FormatBool(*cluster.StorageEncrypted)
            }

            var parameterName string
            if strings.Contains(*cluster.Engine, "aurora-postgresql") {
                parameterName = "ssl"
            } else if strings.Contains(*cluster.Engine, "aurora-mysql") {
                parameterName = "require_secure_transport"
            }

            encryptionInTransit, err := checkEncryptionInTransit(rdsClient, *cluster.DBClusterParameterGroup, parameterName)
            if err != nil {
                encryptionInTransit = "Error: " + err.Error()
            }

            availabilityZones := joinStrings(cluster.AvailabilityZones)

            writer.Write([]string{
                *cluster.DBClusterIdentifier,
                storageEncrypted,
                encryptionInTransit,
                *cluster.Engine,
                availabilityZones,
            })
        }
    }

    fmt.Printf("Results written to %s\n", fileName)
}


func checkEncryptionInTransit(rdsClient *rds.RDS, parameterGroupName, parameterName string) (string, error) {
    var marker *string
    maxAttempts := 2

    for attempt := 0; attempt < maxAttempts; attempt++ {
        paramsOutput, err := rdsClient.DescribeDBClusterParameters(&rds.DescribeDBClusterParametersInput{
            DBClusterParameterGroupName: aws.String(parameterGroupName),
            Marker:                      marker,
        })

        if err != nil {
            if aerr, ok := err.(awserr.Error); ok {
                switch aerr.Code() {
                case "ThrottlingException":
                    fmt.Printf("Throttling error detected, retrying... (Attempt %d/%d)\n", attempt+1, maxAttempts)
                    exponentialBackoff(attempt)
                    continue
                default:
                    fmt.Printf("Error retrieving parameters for group %s: %v\n", parameterGroupName, err)
                    return "Error Retrieving Parameters", err
                }
            }
        } else {
            for _, param := range paramsOutput.Parameters {
                if param.ParameterName != nil && *param.ParameterName == parameterName {
                    fmt.Printf("Parameter found - Name: %s, Value: %v, Type: %T\n", *param.ParameterName, param.ParameterValue, param.ParameterValue)
                    if param.ParameterValue != nil {
                        return *param.ParameterValue, nil
                    } else {
                        return "nil", nil
                    }
                }
            }

            if paramsOutput.Marker == nil {
                break
            }
            marker = paramsOutput.Marker
        }
    }

    fmt.Printf("No %s parameter found for %s after %d attempts\n", parameterName, parameterGroupName, maxAttempts)
    return parameterName + " parameter not found", nil
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

func exponentialBackoff(attempt int) {
    time.Sleep(time.Second * time.Duration(math.Pow(2, float64(attempt))))
}
