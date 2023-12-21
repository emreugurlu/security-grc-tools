# Unapproved Merged PRs Audit Script

## Overview
This Bash script audits merged pull requests in a specified GitHub repository to check if they were merged without proper approval. It identifies unapproved PRs and gathers detailed information about them.

## Prerequisites
Before running this script, make sure you have:
- **Bash:** The script is written for a Bash environment.
- **curl:** For making API requests.
- **jq:** A command-line JSON processor. [Install jq](https://stedolan.github.io/jq/download/)
- **GNU Parallel:** For executing jobs in parallel. [Install GNU Parallel](https://www.gnu.org/software/parallel/)

You will also need:
- **GitHub Access Token:** A token with permissions to access repository data.
- **GitHub Enterprise URL:** The URL of your GitHub Enterprise instance, if applicable.

## Installation
1. Download or copy the script to your local machine.
2. Open a terminal and navigate to the directory where the script is located.
3. Make the script executable: chmod +x unapproved_prs_audit.sh


## Running the Script
1. In the terminal, navigate to the script's directory.
2. Start the script: ./unapproved_prs_audit.sh

3. Enter the requested information when prompted:
- GitHub Enterprise URL
- Name of the organization
- Name of the repository
- Your GitHub Access Token
- The branch you want to check

## Output
The script generates a CSV file named `<repository_name>.csv` containing details of unapproved merged PRs. It includes information like user, permission level, organization role, PR number, merge date, committer, and a link to the PR.

## Additional Note
To assist users, the script provides ongoing updates on the dates being processed. This feature allows users to track the progress and decide when to manually terminate the run if they have reached the desired date range.

### Manually Terminating the Script
If you wish to stop the script before it completes processing all data:
- Simply press `Ctrl+C` in the terminal where the script is running. This will stop the script immediately.
- The CSV file will contain all data processed up to that point, allowing you to review the results even if the script did not complete its full run.

## Notes
- Processing time depends on the number of PRs in the repository.
- Ensure your GitHub Access Token has the required permissions.

## Contributions
Your feedback and contributions are appreciated. Feel free to fork, modify, and make pull requests to improve the script.


