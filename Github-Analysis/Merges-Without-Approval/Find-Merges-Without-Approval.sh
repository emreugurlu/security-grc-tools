#!/bin/bash
SCRIPT_PID=$$
export SCRIPT_PID
read -p "Enter your Github Enterprise URL: " GITHUB_ENTERPRISE_URL
read -p "Enter the name of the organization: " orgname
read -p "Enter the name of the repo: " actualname
read -p "Enter your Github Access Token: " ACCESS_TOKEN
read -p "Enter the branch you want to check: " TARGET_BRANCH
REPO="${orgname}/${actualname}"
OUTPUT_FILE="${actualname}.csv"
PER_PAGE=100

> "$OUTPUT_FILE"

export GITHUB_ENTERPRISE_URL REPO ACCESS_TOKEN OUTPUT_FILE

process_pr() {
    pr_number=$1
    user=$2
    merge_commit_sha=$3
    merged_at=$4
    branch_name=$5

    permissions=$(curl -s -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: token $ACCESS_TOKEN" \
        "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/collaborators/$user/permission" | jq -r '.permission')

    org_roles=$(curl -s -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: token $ACCESS_TOKEN" \
        "$GITHUB_ENTERPRISE_URL/api/v3/orgs/plaid/memberships/$user" | jq -r '.role')

    merged_by=$(curl -s -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: token $ACCESS_TOKEN" \
        "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/commits/$merge_commit_sha" | jq -r '.commit.author.name')

    reviews_page=1
    approval_found=false
    dismissed_approval=false
    revert_approval=false

    while true; do
        reviews=$(curl -s -H "Accept: application/vnd.github.v3+json" \
            -H "Authorization: token $ACCESS_TOKEN" \
            "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/pulls/$pr_number/reviews?page=$reviews_page" | jq '.[] | .state')

        if [[ $reviews =~ "APPROVED" ]]; then
            approval_found=true
            break  # Exit loop if approval found
        elif [[ $reviews =~ "DISMISSED" ]]; then
            dismissed_approval=true
        fi

        if [[ $branch_name == *"revert"* ]]; then
            revert_approval=true
        fi

        if [ -z "$reviews" ] || [ "$reviews" == "[]" ]; then
            break
        fi

        ((reviews_page++))
    done

    pr_link="$GITHUB_ENTERPRISE_URL/$REPO/pull/$pr_number"

    if [ "$approval_found" == false ]; then
        echo "$user,$permissions,$org_roles,$pr_number,$merged_at,$user,$merged_by,$branch_name,$dismissed_approval,$revert_approval,$pr_link" >> "$OUTPUT_FILE"
        echo "Merge Date Of Unapproved PR (end process manually if needed):" $merged_at
    fi
}

export -f process_pr

# Print CSV headers
echo "User,Repository Permission Level,Organization Role,PR Number,Merged Without Approval Date,Committed by,Merged by,Branch the PR came from,Dismissed Approval Status,Revert Approval Status,PR Link" >> "$OUTPUT_FILE"

page=1
while true; do
    apiResponse=$(curl -s -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: token $ACCESS_TOKEN" \
        "$GITHUB_ENTERPRISE_URL/api/v3/repos/$REPO/pulls?state=closed&per_page=$PER_PAGE&page=$page")

    if [ -z "$apiResponse" ] || [ $(echo "$apiResponse" | jq length) -eq 0 ]; then
        echo "No more PRs found on page $page."
        break
    fi

    pr_data=$(echo "$apiResponse" | jq -r --arg TARGET_BRANCH "$TARGET_BRANCH" '.[] | select(.merged_at != null and .base.ref == $TARGET_BRANCH) | [.number, .user.login, .merge_commit_sha, .merged_at, .head.ref] | @csv' | tr -d '"')

    echo "$pr_data" | parallel --jobs 0 -C ',' process_pr {1} {2} {3} {4} {5}

    ((page++))
done

echo "Script completed. Details of unapproved merged PRs are listed in $OUTPUT_FILE (CSV format)."
