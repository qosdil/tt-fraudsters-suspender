# tt-fraudsters-suspender

**tt-fraudsters-suspender** is a Go-based CLI application. The main purpose of this app is to demonstrate the fast batch update process to Amazon Cognito and PostgreSQL database simultaneously by leveraging Go's concurrency. Expect it to process 1,000 user data rows in less than 1 minute.

This repo is the rewrite of the original version in which the code was a part of the whole closed-source User service repository.

On the business aspect, this app was used to suspend hundreds of fraudster users in the platform to prevent them to sign in and get benefits from the available marketing programs (e.g. promo codes).

> *Note: the original version run as a job worker in a EKS cluster against Cognito and RDS/PostgreSQL in a shared private subnet.*

## Prerequisites

* Go v1.22 or newer
* Amazon Cognito
* For the IAM user, grant these permissions againts the targeted user pool: `AdminCreateUser`, `AdminDisableUser`, `AdminEnableUser`
* PostgreSQL v14 or newer

## Installation

Make sure that your system has met the prerequisites above before running the following installation steps.

### Clone This Repository

Clone this repository anywhere in your system. For example:
```
git clone <this_repo>
```

### Create User Table

Use the `/scripts/database/schema.sql` file to create the `users` table in your PostgreSQL instance.

### Set up Environment Variables

Copy the `/.env.example` as `/.env` then fill out the new file with the real values. You can leave the values of `AMAZON_COGNITO_MAX_RPS*` vars as they are if you are not sure about them.

### Build

Go to the app root directory, then build the app. For example:

```
go build -o .
```

That will output the binary file in the same directory.

## Commands

### Generate Fake Users

To simulate the presense of fraudsters in your system, you can use this command to generates N numbers of fake users. It will populate the data to Cognito, PostgreSQL database and a text file.

The following example will generate 1000 fake users:
```
./tt-fraudsters-suspender generate-fake-users --num-users=1000 --dest-file=$HOME/Downloads/fraudsters.txt
```

Example output:
```
2024/09/27 15:56:44 start generating 1000 fake users...
2024/09/27 15:57:25 successfully generated 1000 fake users to Cognito, database and batch text file
2024/09/27 15:57:25 done in 41.40s
```

### Suspend Fraudster Users, Fast

This command will read the text file that we provide and then update each user's `Account status` to `Disabled` on Cognito and set `is_enabled` to  `FALSE` in the database.

The `suspend` command updates multiple rows with concurrency. For example:
```
./tt-fraudsters-suspender suspend --source-file=$HOME/Downloads/fraudsters.txt
```

Example output:
```
2024/09/27 16:00:09 start suspending 1000 users...
2024/09/27 16:00:52 batch suspension done, # of rows: 1000, # of successful: 1000, # of failed: 0
2024/09/27 16:00:52 done in 43.54s
```

### Suspend Fraudster Users, Slow

To see the difference on the execution time with the sequential updates, use the `seq-suspend` command. For example:
```
./tt-fraudsters-suspender seq-suspend --source-file=$HOME/Downloads/fraudsters.txt
```

Example output:
```
2024/09/27 16:13:13 start suspending 1000 users...
2024/09/27 16:16:05 batch suspension done, # of rows: 1000, # of successful: 1000, # of failed: 0
2024/09/27 16:16:05 done in 2m51.72s
```

### Truncate Cognito User Pool

AWS does not seem to provide a tool for us to truncate a Cognito user pool. So, you can use this command to clean up your test Cognito user pool in case you want to start over from fresh.

But, unlike the other two commands, the `truncate` command performs slow on the large dataset because the logic has not been implemented with concurrency.

> WARNING: make sure that you do not put the user pool of a real or production system in the `/.env` file.

Before running the command, you need to add the IAM permissions of `ListUsers` and `AdminDeleteUser` againts the targeted user pool.

Command example:
```
./tt-fraudsters-suspender truncate-cognito
```

## An Unpaid Tech Debt

Supposedly, Cognito was the only source of truth of a user's enabled/disabled status, but that was a tech debt that was too risky to pay considering the short development time limit.

So, the app also updated the `users.is_enabled` field in the PostgreSQL database.

## Help, Feedbacks
If you have any concerns, email me at qosdila@gmail.com.
