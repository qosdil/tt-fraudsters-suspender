# tt-fraudsters-suspender

**tt-fraudsters-suspender** is a Go-based CLI application that consists of three commands.

1. fake-users-generator
2. suspend
3. truncate-cognito

Please learn the Commands section below for the explanations of each command.

The main purpose of this app is to demonstrate the fast batch update process to Amazon Cognito and PostgreSQL database simultaneously by leveraging Go's concurrency.

This repo is the rewrite of the original version in which the code was part of the whole User service repository and it is not open source.

> *Note: the original version run as a job worker in a EKS cluster against RDS/PostgreSQL and Cognito in a shared private subnet, it was able to process 1,000 users data in less than 1 minute.*

## Prerequisites

* Go v1.22 or newer
* Amazon Cognito
* PostgreSQL v14 or newer

## Installation

Make sure that your system has met the prerequisites above before running the following installation steps.

### Clone This Repository

Clone this repository in your Go's `/src` directory.

### Create User Table

Use the `/scripts/database/schema.sql` file to create the `users` table in your PostgreSQL instance.

### Set up Environment Variables

Copy the `/.env.example` as `/.env` then fill out the new file with the real values. You can leave the values of `AMAZON_COGNITO_MAX_RPS*` vars as they are if you are not sure about them.

### Install

Go to the app root directory, then do the Go install command. For example:

```
go install .
```

Then the `tt-fraudsters-suspender` app will be available in your Go bin directory.

> Note: If you update the `/.env` file, then you need to run the `go install .` command again.

### Run the App

By assuming that Go bin directory is included in your system's `$PATH` environment variable, you can run the `tt-fraudsters-suspender` from any directories.

Please continue with next section for more details on how to work with app.

## Commands

### Generate Fake Users

This command generates N numbers of fake users (assume they all are fraudsters of the system) and will write to Cognito, PostgreSQL database and a text file.

The following example will generate 1000 fake users:
```
tt-fraudsters-suspender generate-fake-users --num-users=1000 --dest-file=$HOME/Downloads/fraudsters.txt
```

### Suspend Fraudster Users

This command will read the text file that we provide and then update each user's enabled/disabled status on Cognito and in the database.

Command example:
```
tt-fraudsters-suspender suspend --source-file=$HOME/Downloads/fraudsters.txt
```

### Truncate Cognito User Pool

AWS does not seem to provide a tool for us to truncate a Cognito user pool. So, you can use this command to clean up your test Cognito user pool in case you want to start over from fresh.

> WARNING: make sure that you do not put the user pool of real or production systems in the `/.env` file.

Command example:
```
tt-fraudsters-suspender truncate-cognito
```

## An Unpaid Tech Debt

When I developed this app, there was a tech debt that was too risky to pay considering the development time limit. Supposedly Cognito was the only source of truth of a user's enabled/disabled status.

So, the app also updated the `users.is_enabled` field in the PostgreSQL database.

## Help, Feedbacks
If you have any concerns, email me at qosdila@gmail.com.
