# AWS-RDS-Scanner

## Overview
Welcome to the `aws-rds-scanner` tool. This Go script allows you to scan Amazon RDS (Relational Database Service) instances and clusters in your AWS account. It provides information about StorageEncryption status, availability zones, and other details.

## Prerequisites
Before running the script, make sure you have the following:

- [Go](https://golang.org/) installed on your machine.
- AWS credentials with appropriate permissions. You can set up credentials using the [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html).

## Getting Started
To get started with the `aws-rds-scanner`:

1. Clone the repository: `git clone https://github.com/yourusername/aws-rds-scanner.git`
2. Navigate to the project directory: `cd aws-rds-scanner`
3. Run the script: `go run rdsdatabasescan.go`

## Script Prompts

When you run the `aws-rds-scanner` script, it will guide you through several prompts to gather necessary information:

1. **AWS Credentials:**
   - You will be asked to input the following:
     - `Enter AWS Access Key ID:`
     - `Enter AWS Secret Access Key:`
     - `Enter AWS Session Token (press Enter if not applicable):`

2. **AWS Region:**
   - Prompt for specifying the AWS Region where the scan will be performed:
     - `Enter AWS Region:`

3. **Type of Scan:**
   - Choose the type of scan you wish to perform, with the options being:
     - `Select Scan Type:`
       1. `DB Instances`
       2. `DB Clusters`
       3. `Both`
     - `Enter your choice (1, 2, or 3):`

Please ensure you have your AWS credentials and the specific region details ready before starting the script. Knowing the type of RDS resources (Instances or Clusters) you intend to audit will also streamline the process.

## Output

Upon successful execution, the `aws-rds-scanner` script generates CSV files that detail the results of the scan. These files are named based on the date and time of the script's execution. Each file contains organized and comprehensive information about the scanned AWS RDS resources. Here’s what you can expect in the output files:

- **For DB Instances:**
  - The CSV file for instances (named like `DBInstances_YYYYMMDD_HHMMSS.csv`) will include columns such as:
    - `DB Instance Identifier`
    - `Storage Encrypted` (Yes/No)
    - `Encryption In Transit` (Status)
    - `Engine Type`
    - `Availability Zone`

- **For DB Clusters:**
  - The CSV file for clusters (named like `DBClusters_YYYYMMDD_HHMMSS.csv`) will include similar details tailored for cluster configurations:
    - `DB Cluster Identifier`
    - `Storage Encrypted` (Yes/No)
    - `Encryption In Transit` (Status)
    - `Engine Type`
    - `Availability Zones` (comma-separated if multiple)

These CSV files serve as a comprehensive audit report, providing a clear view of the encryption status and other vital details of your RDS instances and clusters, aiding in compliance and security analysis.


## How to Contribute
Contributions are welcome! If you have improvements, feature suggestions, or bug fixes, please follow these steps:

1. Fork the repository.
2. Create a new branch for your contribution.
3. Add your changes.
4. Create a pull request with a clear description of your contribution.

## Support and Issues
If you encounter any issues or have questions about the script, please open an issue in the repository.

## Note on Encryption
The script focuses on two main types of encryption:
- **Encryption-At-Storage:** Checks if the `StorageEncrypted` attribute is set to true, ensuring data at rest is encrypted.
- **Encryption-In-Transit:** Audits varying encryption-in-transit configurations, adapting to different AWS RDS engine types and settings.

## Acknowledgements
A special thank you to all contributors who help improve and maintain this tool. Your efforts are appreciated in advancing AWS RDS scanning practices.

---
Thank you for using the `aws-rds-scanner` tool. Together, we can enhance the security and compliance of AWS RDS instances and clusters.
