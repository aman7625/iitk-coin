# iitk-coin
This repository contains the code contributed by me as part of summer project offered by programming club IIT Kanpur(Currently under Development)

## Getting Started
To run the project locally, open the folder inside a terminal and use the following command
```
 go run main.go
```
## API Requests
| Endpoint        |  HTTP Method          | Description  |
| -------------   | --------------------- | :------------: |
| `/signup`       | `POST`                |  `Create new Account`|
| `/login`        | `POST`                |  `Logging into Account` |
| `/secretpage`   | `GET`                 |`Restricted access`|
| `/view`         | `GET`                 |`Find Current Balance`|
| `/reward`       | `POST`                |`Reward certain amount to a user` |
| `/transfer`     | `POST`                |`Transfer certain amount between users` |

## Testing Endpoints
* ### Creating new account

`Request`
```
curl -i --request POST 'http://localhost:8000/signup' \
--header 'Content-Type: application/json' \
--data-raw '{
    "rollno":1000,
    "password":"dummy",
    "name":"John Doe"
}'
```

`Response`
```
{
  "rollno":1000,
  "password":"dummy",
  "name":"John Doe"
  }
``` 

* ### Logging into account

`Request`
```
curl -i --request POST 'http://localhost:8000/login' \
--header 'Content-Type: application/json' \
--data-raw '{
    "rollno":1000,
    "password":"dummy"
}'
```

`Response`

A token is returned which can be used to access the secret page
```
{"token":"${Token}"}
```

* ### Find Current Balance

`Request`
```
curl -i --request GET 'http://localhost:8000/view' \
--header 'Content-Type: application/json' \
--data-raw '{
    "rollno":1000
}'
```

`Response`
```
{"message":"Current Balance: 0"}
```

* ### Reward certain amount to user

`Request`
```
curl -i --request POST 'http://localhost:8000/reward' \
--header 'Content-Type: application/json' \
--data-raw '{
    "rollno":1000,
    "amount_to_send": 50
}'
```

`Response`
```
{"message":"Transaction Successful"}
```

* ### Transfer certain amount between users

`Request`
```
curl -i --request POST 'http://localhost:8000/transfer' \
--header 'Content-Type: application/json' \
--data-raw '{
    "sender_rollno": 1000,
    "reciever_rollno":1001,
    "amount_to_send": 40
}'
```

`Response`
```
{"message":"Transaction Successful"}
```

