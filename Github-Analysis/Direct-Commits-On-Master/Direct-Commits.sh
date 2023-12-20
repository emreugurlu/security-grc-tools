#!/bin/bash
read -p "Enter your Github Enterprise URL (include https://): " GITHUB_ENTERPRISE_URL
read -p "Enter the name of the organization: " orgname
read -p "Enter the name of the repo: " actualname
read -p "Enter your Github Access Token: " ACCESS_TOKEN
read -p "Enter the date you want to check from (yyyy-mm-dd): " SINCE
REPO="${orgname}/${actualname}"
OUTPUT_FILE="${actualname}_directcommits.csv"
# Clear the output file or create it if it doesn't exist
echo "PR Link,Author,Date,Committer,Org Role,Repo Role" > "$OUTPUT_FILE"

process_commit() {
    sha=$1
    url=$2
    author=$3
    date=$4
    committer=$5
    GITHUB_ENTERPRISE_URL=$6
    REPO=$7
    ACCESS_TOKEN=$8
    OUTPUT_FILE=$9
    
    pr_associated=$(curl -s -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: token $ACCESS_TOKEN" \
        "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/commits/$sha/pulls" | jq 'length')
    
    if [ "$pr_associated" -eq "0" ]; then
        org_role=$(curl -s -H "Accept: application/vnd.github.v3+json" \
            -H "Authorization: token $ACCESS_TOKEN" \
            "$GITHUB_ENTERPRISE_URL/api/v3/orgs/plaid/memberships/$committer" | jq -r '.role')
            
        repo_role=$(curl -s -H "Accept: application/vnd.github.v3+json" \
            -H "Authorization: token $ACCESS_TOKEN" \
            "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/collaborators/$committer/permission" | jq -r '.permission')
        
        pr_link=$(curl -s -H "Accept: application/vnd.github.v3+json" \
            -H "Authorization: token $ACCESS_TOKEN" \
            "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/commits/$sha" | jq -r '.html_url')
            
        echo "$pr_link ,$author,$date,$committer,$org_role,$repo_role" >> "$OUTPUT_FILE"
    fi
}

export -f process_commit

page=1
while true; do
    commit_data=$(curl -s -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: token $ACCESS_TOKEN" \
        "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/commits?since=$SINCE&sha=master&page=$page" | jq -r '.[] | select(.parents | length == 1) | [.sha, .html_url, .commit.author.name, .commit.author.date, .author.login] | @csv' | tr -d '"')
    
    if [ -z "$commit_data" ]; then
        echo "No more direct commits found on page $page."
        break
    fi
    
    echo "$commit_data" | parallel --jobs 0 -C ',' process_commit {1} {2} {3} {4} {5} $GITHUB_ENTERPRISE_URL $REPO $ACCESS_TOKEN "$OUTPUT_FILE"
    
    ((page++))
done

echo "Script completed. Direct commits to master with details are listed in $OUTPUT_FILE (CSV format)."
