GitHub Direct Commits Audit Script
Overview
This Bash script is designed to audit direct commits to the master branch in a specified GitHub repository. It identifies commits that have been made without associated pull requests and gathers relevant details about these commits.

Prerequisites
Before running this script, ensure you have the following installed:

Bash: The script is written for a Bash environment.
curl: Used for making API requests.
jq: A lightweight and flexible command-line JSON processor. Install jq
GNU Parallel: A shell tool for executing jobs in parallel. Install GNU Parallel
Additionally, you will need:

GitHub Access Token: Ensure you have a GitHub access token with appropriate permissions to access repository data.
GitHub Enterprise URL: If using GitHub Enterprise, have your enterprise URL ready.
Installation
No additional installation is needed for the script itself. Simply download or copy the script to your local machine.

Running the Script
Open your terminal.
Navigate to the directory where the script is located.
Run the script using the command:
bash
Copy code
bash direct_commits_audit.sh
Follow the on-screen prompts to enter:
Your GitHub Enterprise URL (include https://)
The name of the organization
The name of the repository
Your GitHub Access Token
The start date for the audit (format: yyyy-mm-dd)
Output
The script outputs a CSV file named <repository_name>_directcommits.csv. This file contains details of all direct commits made without associated pull requests, including the PR link, author, date, committer, and their roles in the organization and repository.

Notes
The script may take some time to run, depending on the size of the repository and the number of commits.
Ensure that your GitHub Access Token has sufficient permissions to access the necessary repository data.
Contributions
Feedback and contributions to the script are welcome. Please feel free to fork, modify, and make pull requests as needed.

