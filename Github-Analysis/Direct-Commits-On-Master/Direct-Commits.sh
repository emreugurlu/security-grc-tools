#!/bin/bash

# Prompt user for necessary information
read -p "Enter your Github Enterprise URL (include https://): " GITHUB_ENTERPRISE_URL
read -p "Enter the name of the organization: " orgnamen
read -p "Enter the name of the repo: " actualname
read -p "Enter your Github Access Token: " ACCESS_TOKEN
read -p "Enter the name of the master branch you want to check: " TARGET_BRANCH
read -p "Enter the date you want to check from (yyyy-mm-dd): " SINCE

# Combine organization and repository name for full repo path
REPO="${orgname}/${actualname}"

# Define the output file name
OUTPUT_FILE="${actualname}_directcommits.csv"

# Initialize output file with header row
echo "PR Link,Author,Date,Committer,Org Role,Repo Role" > "$OUTPUT_FILE"

# Function to process each commit
process_commit() {
    # Assign arguments to readable variable names
    sha=$1
    url=$2
    author=$3
    date=$4
    committer=$5
    GITHUB_ENTERPRISE_URL=$6
    REPO=$7
    ACCESS_TOKEN=$8
    OUTPUT_FILE=$9

    # Check if the commit is associated with any pull requests
    pr_associated=$(curl -s -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: token $ACCESS_TOKEN" \
        "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/commits/$sha/pulls" | jq 'length')

    # Process only commits with no associated pull request
    if [ "$pr_associated" -eq "0" ]; then
        # Fetch organization role of the committer
        org_role=$(curl -s -H "Accept: application/vnd.github.v3+json" \
            -H "Authorization: token $ACCESS_TOKEN" \
            "$GITHUB_ENTERPRISE_URL/api/v3/orgs/$orgname/memberships/$committer" | jq -r '.role')

        # Fetch repository role of the committer
        repo_role=$(curl -s -H "Accept: application/vnd.github.v3+json" \
            -H "Authorization: token $ACCESS_TOKEN" \
            "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/collaborators/$committer/permission" | jq -r '.permission')

        # Get the HTML URL of the commit
        pr_link=$(curl -s -H "Accept: application/vnd.github.v3+json" \
            -H "Authorization: token $ACCESS_TOKEN" \
            "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/commits/$sha" | jq -r '.html_url')

        # Append commit details to the output file
        echo "$pr_link ,$author,$date,$committer,$org_role,$repo_role" >> "$OUTPUT_FILE"
    fi
}

# Export the function for parallel processing
export -f process_commit

# Initialize page counter for API pagination
page=1

# Loop through all pages of commits
while true; do
    # Fetch commit data from GitHub API
    commit_data=$(curl -s -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: token $ACCESS_TOKEN" \
        "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/commits?since=$SINCE&sha=$TARGET_BRANCH&page=$page" | jq -r '.[] | select(.parents | length == 1) | [.sha, .html_url, .commit.author.name, .commit.author.date, .author.login] | @csv' | tr -d '"')

    # Break loop if no more commits are found
    if [ -z "$commit_data" ]; then
        echo "No more direct commits found on page $page."
        break
    fi

    # Process each commit in parallel
    echo "$commit_data" | parallel --jobs 0 -C ',' process_commit {1} {2} {3} {4} {5} $GITHUB_ENTERPRISE_URL $REPO $ACCESS_TOKEN "$OUTPUT_FILE"

    # Increment page number for next iteration
    ((page++))
done

# Print completion message with output file details
echo "Script completed. Direct commits to master with details are listed in $OUTPUT_FILE (CSV format)."
