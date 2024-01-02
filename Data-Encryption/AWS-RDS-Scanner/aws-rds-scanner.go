package main

import (
    // Importing necessary packages for the script
    "bufio"
    "encoding/csv"
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"

    // AWS SDK packages for Go
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/rds"
)

// main function - the entry point of the Go script
func main() {
    // Prompting the user for necessary AWS credentials and configuration
    accessKeyID := promptUser("Enter AWS Access Key ID:")
    secretAccessKey := promptUser("Enter AWS Secret Access Key:")
    sessionToken := promptUser("Enter AWS Session Token (press Enter if not applicable):")
    region := promptUser("Enter AWS Region:")
    scanType := promptUser("Select Scan Type:\n1. DB Instances\n2. DB Clusters\n3. Both\nEnter your choice (1, 2, or 3):")

    // Validate user's scan type choice
    if scanType != "1" && scanType != "2" && scanType != "3" {
        fmt.Println("Invalid choice. Exiting.")
        return
    }

    // Establishing a new AWS session with the provided credentials
    sess, err := session.NewSession(&aws.Config{
        Region:      aws.String(region),
        Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, sessionToken),
    })
    if err != nil {
        fmt.Println("Error creating AWS session:", err)
        return
    }

    // Creating a new RDS client from the AWS session
    rdsClient := rds.New(sess)

    // Initializing slices to store instance and cluster data
    var instances []*rds.DBInstance
    var clusters []*rds.DBCluster

    // Performing scans based on user choice
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

    // Writing the scan results to CSV files
    if scanType == "1" || scanType == "3" {
        writeToCSV("DBInstances", instances, rdsClient)
    }
    if scanType == "2" || scanType == "3" {
        writeToCSV("DBClusters", clusters, rdsClient)
    }
}

// promptUser function to handle user input
func promptUser(prompt string) string {
    fmt.Print(prompt + " ")
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    return scanner.Text()
}

// scanDBInstances scans for RDS DB Instances and returns them
func scanDBInstances(rdsClient *rds.RDS) ([]*rds.DBInstance, error) {
    var totalInstances []*rds.DBInstance
    var marker *string

    // Looping through all instances
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

// scanDBClusters scans for RDS DB Clusters and returns them
func scanDBClusters(rdsClient *rds.RDS) ([]*rds.DBCluster, error) {
    var totalClusters []*rds.DBCluster
    var marker *string

    // Looping through all clusters
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

// writeToCSV writes the scan results to a CSV file
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

    // Handling data based on its type (instance or cluster)
    switch d := data.(type) {
    case []*rds.DBInstance:
        // Writing headers for instance data
        writer.Write([]string{"DB Instance", "StorageEncryption", "EncryptionInTransit", "Engine", "Availability Zone"})
        for _, instance := range d {
            // Extracting relevant data from each instance
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

            // Checking encryption in transit for each instance
            encryptionInTransit := checkInstanceEncryptionInTransit(rdsClient, instance)

            // Writing instance data to CSV
            writer.Write([]string{
                *instance.DBInstanceIdentifier,
                storageEncrypted,
                encryptionInTransit,
                engine,
                availabilityZone,
            })
        }
    case []*rds.DBCluster:
        // Writing headers for cluster data
        writer.Write([]string{"DB Cluster", "StorageEncrypted", "EncryptionInTransit", "Engine", "Availability Zones"})
        for _, cluster := range d {
            // Extracting relevant data from each cluster
            storageEncrypted := "false"
            if cluster.StorageEncrypted != nil {
                storageEncrypted = strconv.FormatBool(*cluster.StorageEncrypted)
            }
            encryptionInTransit, err := checkClusterEncryptionInTransit(rdsClient, *cluster.DBClusterParameterGroup, *cluster.Engine)
            if err != nil {
                encryptionInTransit = "Error: " + err.Error()
            }

            availabilityZones := joinStrings(cluster.AvailabilityZones)

            // Writing cluster data to CSV
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

// CheckClusterEncryptionInTransit checks the encryption in transit setting for a cluster
func checkClusterEncryptionInTransit(rdsClient *rds.RDS, parameterGroupName, engine string) (string, error) {
    var marker *string
    var parameterName string
    // Identifying the parameter based on engine type
    if strings.Contains(engine, "neptune") {
        return "Encryption in transit is automatically enabled for all connections to an Amazon Neptune database. ", nil
    } else if strings.Contains(engine, "aurora-postgresql") {
        parameterName = "ssl"
    } else if strings.Contains(engine, "aurora-mysql") {
        parameterName = "require_secure_transport"
    } else {
        return "Encryption in Transit parameter check not yet applicable for this engine", nil
    }
    for {
        time.Sleep(50 * time.Millisecond) // To avoid hitting AWS rate limits
        paramsOutput, err := rdsClient.DescribeDBClusterParameters(&rds.DescribeDBClusterParametersInput{
            DBClusterParameterGroupName: aws.String(parameterGroupName),
            Marker:                      marker,
        })
        if err != nil {
            fmt.Printf("Error retrieving parameters for group %s: %v\n", parameterGroupName, err)
            return "Error Retrieving Parameters", err
        }

        // Looping through parameters to find the relevant one
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

    fmt.Printf("No %s parameter found for %s\n", parameterName, parameterGroupName)
    return parameterName + " parameter not found", nil
}

// CheckInstanceEncryptionInTransit checks the encryption in transit setting for an instance
func checkInstanceEncryptionInTransit(rdsClient *rds.RDS, instance *rds.DBInstance) string {
    engine := aws.StringValue(instance.Engine)
    var parameterName string
    // Identifying the parameter based on engine type
    if strings.Contains(engine, "aurora") {
        return "Parameter Not Found Because Aurora manages encryption at the cluster level"
    } else if strings.Contains(engine, "neptune") {
        return "Parameter Not Found Because Neptune operates at the cluster level"
    } else if engine == "mysql" {
        parameterName = "require_secure_transport"
    } else if engine == "postgres" {
        parameterName = "rds.force_ssl"
    } else {
        return "Encryption in Transit parameter check not yet applicable for this engine"
    }

    // Looping through parameter groups to find the relevant setting
    for _, dbParameterGroup := range instance.DBParameterGroups {
        groupName := aws.StringValue(dbParameterGroup.DBParameterGroupName)

        var marker *string
        for {
            time.Sleep(50 * time.Millisecond) // To avoid hitting AWS rate limits
            paramsOutput, err := rdsClient.DescribeDBParameters(&rds.DescribeDBParametersInput{
                DBParameterGroupName: aws.String(groupName),
                Marker:               marker,
            })
            if err != nil {
                return "Error: " + err.Error()
            }

            // Looping through parameters to find the relevant one
            for _, param := range paramsOutput.Parameters {
                if param.ParameterName != nil && *param.ParameterName == parameterName {
                    if param.ParameterValue != nil {
                        return *param.ParameterValue
                    } else {
                        return "nil"
                    }
                }
            }

            if paramsOutput.Marker == nil {
                break
            }
            marker = paramsOutput.Marker
        }
    }

    return parameterName + " parameter not found"
}

// joinStrings concatenates a slice of string pointers into a single string
func joinStrings(strPointers []*string) string {
    var strValues []string
    for _, strPointer := range strPointers {
        if strPointer != nil {
            strValues = append(strValues, *strPointer)
        }
    }
    return strings.Join(strValues, ", ")
}
