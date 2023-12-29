# AWS-RDS-Scanner

## Overview
Welcome to the `aws-rds-scanner` tool. This Go script allows you to scan Amazon RDS (Relational Database Service) instances and clusters in your AWS account. It provides information about encryption status, availability zones, and other details.

## Prerequisites
Before running the script, make sure you have the following:

- [Go](https://golang.org/) installed on your machine.
- AWS credentials with appropriate permissions. You can set up credentials using the [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html).

## Getting Started
To get started with the `aws-rds-scanner`:

1. Clone the repository: `git clone https://github.com/yourusername/aws-rds-scanner.git`
2. Navigate to the project directory: `cd aws-rds-scanner`
3. Run the script: `go run rdsdatabasescan.go`

Follow the prompts to enter your AWS credentials, region, and select the type of scan (DB Instances, DB Clusters, or Both).

## Output
The script generates CSV files with results, named with the date and time of execution. The CSV files contain organized information about DB Instances and/or DB Clusters, including encryption status and availability zones.

## How to Contribute
Contributions are welcome! If you have improvements, feature suggestions, or bug fixes, please follow these steps:

1. Fork the repository.
2. Create a new branch for your contribution.
3. Add your changes.
4. Create a pull request with a clear description of your contribution.

## Support and Issues
If you encounter any issues or have questions about the script, please open an issue in the repository.

## Acknowledgements
A special thank you to all contributors who help improve and maintain this tool. Your efforts are appreciated in advancing AWS RDS scanning practices.

---
Thank you for using the `aws-rds-scanner` tool. Together, we can enhance the security and compliance of AWS RDS instances and clusters.
