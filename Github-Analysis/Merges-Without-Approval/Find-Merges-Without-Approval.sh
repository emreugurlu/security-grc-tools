#!/bin/bash

# Store the current script's PID for reference
SCRIPT_PID=$$
export SCRIPT_PID

# Prompt user for necessary information
read -p "Enter your Github Enterprise URL: " GITHUB_ENTERPRISE_URL
read -p "Enter the name of the organization: " orgname
read -p "Enter the name of the repo: " actualname
read -p "Enter your Github Access Token: " ACCESS_TOKEN
read -p "Enter the branch you want to check: " TARGET_BRANCH

# Combine organization and repository name for full repo path
REPO="${orgname}/${actualname}"

# Define the output file name
OUTPUT_FILE="${actualname}.csv"
PER_PAGE=100

# Clear the output file or create it if it doesn't exist
> "$OUTPUT_FILE"

# Export variables for use in the process_pr function
export GITHUB_ENTERPRISE_URL REPO ACCESS_TOKEN OUTPUT_FILE

# Function to process each pull request
process_pr() {
    # Assign arguments to readable variable names
    pr_number=$1
    user=$2
    merge_commit_sha=$3
    merged_at=$4
    branch_name=$5

    # Fetch permissions and roles of the user who created the pull request
    permissions=$(curl -s -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: token $ACCESS_TOKEN" \
        "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/collaborators/$user/permission" | jq -r '.permission')

    org_roles=$(curl -s -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: token $ACCESS_TOKEN" \
        "$GITHUB_ENTERPRISE_URL/api/v3/orgs/$orgname/memberships/$user" | jq -r '.role')

    # Identify the user who merged the PR
    merged_by=$(curl -s -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: token $ACCESS_TOKEN" \
        "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/commits/$merge_commit_sha" | jq -r '.commit.author.name')

    # Initialize variables for review process
    reviews_page=1
    approval_found=false
    dismissed_approval=false
    revert_approval=false

    # Loop to check all reviews for each PR
    while true; do
        reviews=$(curl -s -H "Accept: application/vnd.github.v3+json" \
            -H "Authorization: token $ACCESS_TOKEN" \
            "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/pulls/$pr_number/reviews?page=$reviews_page" | jq '.[] | .state')

        # Check for approved, dismissed reviews and revert approvals
        if [[ $reviews =~ "APPROVED" ]]; then
            approval_found=true
            break  # Exit loop if approval found
        elif [[ $reviews =~ "DISMISSED" ]]; then
            dismissed_approval=true
        fi

        if [[ $branch_name == *"revert"* ]]; then
            revert_approval=true
        fi

        # Break loop if no more reviews found
        if [ -z "$reviews" ] || [ "$reviews" == "[]" ]; then
            break
        fi

        ((reviews_page++))
    done

    # Create the link to the PR
    pr_link="$GITHUB_ENTERPRISE_URL/$REPO/pull/$pr_number"

    # Log PR details if no approval was found
    if [ "$approval_found" == false ]; then
        echo "$user,$permissions,$org_roles,$pr_number,$merged_at,$user,$merged_by,$branch_name,$dismissed_approval,$revert_approval,$pr_link" >> "$OUTPUT_FILE"
        echo "Merge Date Of Unapproved PR (end process manually if needed):" $merged_at
    fi
}

# Export the process_pr function for use with parallel
export -f process_pr

# Print CSV headers to the output file
echo "User,Repository Permission Level,Organization Role,PR Number,Merged Without Approval Date,Committed by,Merged by,Branch the PR came from,Dismissed Approval Status,Revert Approval Status,PR Link" >> "$OUTPUT_FILE"

# Initialize page counter for API pagination
page=1

# Loop through all pages of pull requests
while true; do
    # Fetch PR data from GitHub API
    apiResponse=$(curl -s -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: token $ACCESS_TOKEN" \
        "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/pulls?state=closed&per_page=$PER_PAGE&page=$page")

    # Break loop if no more PRs are found
    if [ -z "$apiResponse" ] || [ $(echo "$apiResponse" | jq length) -eq 0 ]; then
        echo "No more PRs found on page $page."
        break
    fi

    # Process each PR in parallel
    pr_data=$(echo "$apiResponse" | jq -r --arg TARGET_BRANCH "$TARGET_BRANCH" '.[] | select(.merged_at != null and .base.ref == $TARGET_BRANCH) | [.number, .user.login, .merge_commit_sha, .merged_at, .head.ref] | @csv' | tr -d '"')
    echo "$pr_data" | parallel --jobs 0 -C ',' process_pr {1} {2} {3} {4} {5}

    # Increment page number for next iteration
    ((page++))
done

# Print completion message with output file details
echo "Script completed. Details of unapproved merged PRs are listed in $OUTPUT_FILE (CSV format)."
