# Mini-wallet
Mini Wallet is a small project for managing simple wallet. This project is for completion of JULO technical assessment test

## Local Setup
Make sure you already install PostgreSQL & Go

create database 
```bash
  createdb -U <your account name> miniwallet
```

go to project directory
```bash
  cd mini-wallet
```

run psql command to create table
```bash
  psql -U <your account name> -f setup.sql -d wallet
```

make sure to fill the username and password for your database in [here](https://github.com/rahimyarza/mini-wallet/blob/938f78e4b6aebf3f7d2d4e0fa1e247c86a5286ee/init.go#L15)

## Run Locally
run go build command on the project directory to compile the code into an executable
```bash
  go build
```

run the executable
```bash
  ./miniwallet
```

## Documentation
For the API documentation you can refer to [here](https://documenter.getpostman.com/view/8411283/SVfMSqA3?version=latest#auth-info-d31009cb-f9bc-4c83-a75f-7a414be1586d
) 
