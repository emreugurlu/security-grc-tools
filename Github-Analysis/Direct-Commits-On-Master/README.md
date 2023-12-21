# GitHub Direct Commits Audit Script

## Overview
This Bash script is designed to audit direct commits to the master branch in a specified GitHub repository. It identifies commits that have been made without associated pull requests and gathers relevant details about these commits.

## Prerequisites
Before running this script, ensure you have the following installed:
- **Bash:** The script is written for a Bash environment.
- **curl:** Used for making API requests.
- **jq:** A lightweight and flexible command-line JSON processor. [Install jq](https://stedolan.github.io/jq/download/)
- **GNU Parallel:** A shell tool for executing jobs in parallel. [Install GNU Parallel](https://www.gnu.org/software/parallel/)

Additionally, you will need:
- **GitHub Access Token:** Ensure you have a GitHub access token with appropriate permissions to access repository data.
- **GitHub Enterprise URL:** If using GitHub Enterprise, have your enterprise URL ready.

## Installation
1. Download or copy the script to your local machine.
2. Open a terminal and navigate to the directory where the script is located.
3. To make the script executable, run: chmod +x direct_commits_audit.sh


## Running the Script
1. In the terminal, navigate to the script's directory if you haven't already.
2. Execute the script by typing: ./direct_commits_audit.sh

If you encounter a permission error, ensure the script is executable (refer to the Installation section).
3. Follow the on-screen prompts to enter:
- Your GitHub Enterprise URL (include https://)
- The name of the organization
- The name of the repository
- Your GitHub Access Token
- The start date for the audit (format: yyyy-mm-dd)

## Output
The script outputs a CSV file named `<repository_name>_directcommits.csv`. This file contains details of all direct commits made without associated pull requests, including the PR link, author, date, committer, and their roles in the organization and repository.

## Notes
- The script may take some time to run, depending on the size of the repository and the number of commits.
- Ensure that your GitHub Access Token has sufficient permissions to access the necessary repository data.

## Contributions
Feedback and contributions to the script are welcome. Please feel free to fork, modify, and make pull requests as needed.


